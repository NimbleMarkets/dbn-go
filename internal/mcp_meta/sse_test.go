// Copyright (c) 2025 Neomantra Corp

package mcp_meta

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mcp_server "github.com/mark3labs/mcp-go/server"
)

const initializeBody = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`

func newSSETestServer(t *testing.T, config SSEConfig) *httptest.Server {
	t.Helper()
	mcpServer := mcp_server.NewMCPServer("test", "0.0.0")
	ts := httptest.NewServer(NewSSEHandler(mcpServer, config))
	t.Cleanup(ts.Close)
	return ts
}

// openSSE issues GET /sse and returns the status code and, when the stream
// opens, the message endpoint path announced in its first event.
func openSSE(t *testing.T, ts *httptest.Server, headers map[string]string) (status int, endpoint string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/sse", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /sse: %v", err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, ""
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if data, ok := strings.CutPrefix(scanner.Text(), "data: "); ok {
			return resp.StatusCode, strings.TrimSpace(data)
		}
	}
	t.Fatalf("SSE stream ended without an endpoint event: %v", scanner.Err())
	return 0, ""
}

func postMessage(t *testing.T, ts *httptest.Server, endpoint string, headers map[string]string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+endpoint, strings.NewReader(initializeBody))
	req.Header.Set("Content-Type", "text/plain") // what a cross-origin page can send without preflight
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", endpoint, err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func TestSSE_ClientWithoutOriginConnects(t *testing.T) {
	ts := newSSETestServer(t, SSEConfig{})
	status, endpoint := openSSE(t, ts, nil)
	if status != http.StatusOK || endpoint == "" {
		t.Fatalf("GET /sse: status %d, endpoint %q; want 200 and an endpoint", status, endpoint)
	}
	if status := postMessage(t, ts, endpoint, nil); status != http.StatusAccepted {
		t.Errorf("POST message: status %d, want 202", status)
	}
}

func TestSSE_RejectsForeignOrigin(t *testing.T) {
	ts := newSSETestServer(t, SSEConfig{AllowedOrigins: []string{"http://localhost:6274"}})
	for _, origin := range []string{"https://evil.example", "null", "http://localhost:9999", "http://localhost:6274.evil.example"} {
		if status, _ := openSSE(t, ts, map[string]string{"Origin": origin}); status != http.StatusForbidden {
			t.Errorf("GET /sse with Origin %q: status %d, want 403", origin, status)
		}
	}
	// A page must not be able to post to a session it does not own either.
	_, endpoint := openSSE(t, ts, nil)
	if status := postMessage(t, ts, endpoint, map[string]string{"Origin": "https://evil.example"}); status != http.StatusForbidden {
		t.Errorf("POST message with foreign Origin: status %d, want 403", status)
	}
}

func TestSSE_RejectsEveryOriginByDefault(t *testing.T) {
	ts := newSSETestServer(t, SSEConfig{})
	for _, origin := range []string{"https://evil.example", "http://localhost:6274", ts.URL} {
		if status, _ := openSSE(t, ts, map[string]string{"Origin": origin}); status != http.StatusForbidden {
			t.Errorf("GET /sse with Origin %q: status %d, want 403", origin, status)
		}
	}
}

func TestSSE_RejectsCrossSiteRequestWithoutOrigin(t *testing.T) {
	ts := newSSETestServer(t, SSEConfig{})
	for _, site := range []string{"cross-site", "same-site"} {
		if status, _ := openSSE(t, ts, map[string]string{"Sec-Fetch-Site": site}); status != http.StatusForbidden {
			t.Errorf("GET /sse with Sec-Fetch-Site %q: status %d, want 403", site, status)
		}
	}
}

func TestSSE_AllowsConfiguredOrigin(t *testing.T) {
	const origin = "http://localhost:6274"
	ts := newSSETestServer(t, SSEConfig{AllowedOrigins: []string{origin}})
	headers := map[string]string{"Origin": origin, "Sec-Fetch-Site": "cross-site"}
	status, endpoint := openSSE(t, ts, headers)
	if status != http.StatusOK {
		t.Fatalf("GET /sse with allowed Origin: status %d, want 200", status)
	}
	if status := postMessage(t, ts, endpoint, headers); status != http.StatusAccepted {
		t.Errorf("POST message with allowed Origin: status %d, want 202", status)
	}
}

func TestSSE_NeverSendsWildcardCORS(t *testing.T) {
	const origin = "http://localhost:6274"
	for _, config := range []SSEConfig{{}, {AllowedOrigins: []string{origin}}} {
		ts := newSSETestServer(t, config)
		for _, headers := range []map[string]string{nil, {"Origin": origin}} {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/sse", nil)
			for k, v := range headers {
				req.Header.Set(k, v)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if got := resp.Header.Get("Access-Control-Allow-Origin"); got == "*" {
				t.Errorf("config %+v, headers %v: Access-Control-Allow-Origin is *", config, headers)
			}
			resp.Body.Close()
			cancel()
		}
	}
}

func TestSSE_TokenRequiredWhenConfigured(t *testing.T) {
	ts := newSSETestServer(t, SSEConfig{Token: "s3cret-token"})
	for _, auth := range []string{"", "Bearer wrong", "Bearer s3cret-token-and-more", "Basic s3cret-token", "s3cret-token"} {
		headers := map[string]string{}
		if auth != "" {
			headers["Authorization"] = auth
		}
		if status, _ := openSSE(t, ts, headers); status != http.StatusUnauthorized {
			t.Errorf("GET /sse with Authorization %q: status %d, want 401", auth, status)
		}
	}
	good := map[string]string{"Authorization": "Bearer s3cret-token"}
	status, endpoint := openSSE(t, ts, good)
	if status != http.StatusOK {
		t.Fatalf("GET /sse with token: status %d, want 200", status)
	}
	if status := postMessage(t, ts, endpoint, nil); status != http.StatusUnauthorized {
		t.Errorf("POST message without token: status %d, want 401", status)
	}
	if status := postMessage(t, ts, endpoint, good); status != http.StatusAccepted {
		t.Errorf("POST message with token: status %d, want 202", status)
	}
}

func TestIsLoopbackHostPort(t *testing.T) {
	for hostPort, want := range map[string]bool{
		"127.0.0.1:8889": true, "localhost:8889": true, "[::1]:8889": true,
		"0.0.0.0:8889": false, ":8889": false, "192.168.1.5:8889": false, "example.com:80": false, "bogus": false,
	} {
		if got := IsLoopbackHostPort(hostPort); got != want {
			t.Errorf("IsLoopbackHostPort(%q) = %v, want %v", hostPort, got, want)
		}
	}
}
