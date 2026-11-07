//go:build !integration

// Copyright (c) 2024-2025 Neomantra Corp

package dbn_hist

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestPostFormRequestSetsAcceptHeader(t *testing.T) {
	var accept, acceptEncoding string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept = r.Header.Get("Accept")
		acceptEncoding = r.Header.Get("Accept-Encoding")
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	body, err := databentoPostFormRequest(ts.URL, "key", url.Values{}, "application/octet-stream")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != "ok" {
		t.Fatalf("unexpected body: %q", body)
	}
	if accept != "application/octet-stream" {
		t.Errorf("Accept = %q, want %q", accept, "application/octet-stream")
	}
	if acceptEncoding == "application/octet-stream" {
		t.Errorf("Accept-Encoding = %q, want a content coding", acceptEncoding)
	}
}

func TestRequestTimesOutWaitingForHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // never send headers
	}))
	defer ts.Close()

	oldClient := httpClient
	httpClient = newHTTPClient(100 * time.Millisecond)
	defer func() { httpClient = oldClient }()

	start := time.Now()
	if _, err := databentoGetRequest(ts.URL, "key"); err == nil {
		t.Fatal("GET: expected timeout error")
	}
	if _, err := databentoPostFormRequest(ts.URL, "key", url.Values{}, ""); err == nil {
		t.Fatal("POST: expected timeout error")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("requests took %v, want them to time out quickly", elapsed)
	}
}

func TestDefaultHTTPClientHasResponseHeaderTimeout(t *testing.T) {
	transport, ok := httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type %T", httpClient.Transport)
	}
	if transport.ResponseHeaderTimeout <= 0 {
		t.Errorf("ResponseHeaderTimeout = %v, want > 0", transport.ResponseHeaderTimeout)
	}
	if httpClient.Timeout != 0 {
		t.Errorf("Timeout = %v, want 0 so large downloads are not cut off", httpClient.Timeout)
	}
}

type customRoundTripper struct{}

func (customRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, nil
}

func TestNewHTTPClientHandlesCustomDefaultTransport(t *testing.T) {
	oldTransport := http.DefaultTransport
	http.DefaultTransport = customRoundTripper{}
	defer func() { http.DefaultTransport = oldTransport }()

	client := newHTTPClient(5 * time.Second)
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type %T", client.Transport)
	}
	if transport.ResponseHeaderTimeout != 5*time.Second {
		t.Errorf("ResponseHeaderTimeout = %v, want 5s", transport.ResponseHeaderTimeout)
	}
}
