// Copyright (c) 2025 Neomantra Corp

package mcp_meta

import (
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	mcp_server "github.com/mark3labs/mcp-go/server"
)

// SSEConfig controls who may use the SSE transport.
type SSEConfig struct {
	// AllowedOrigins lists the browser origins (e.g. "http://localhost:6274")
	// allowed to connect. Requests from any other web page are rejected.
	// Clients that are not browsers send no Origin header and are unaffected.
	AllowedOrigins []string

	// Token, when set, must be presented by every client as
	// "Authorization: Bearer <token>".
	Token string
}

// NewSSEHandler returns the HTTP handler for mcpServer's SSE transport,
// restricted according to config.
//
// A local MCP server is reachable from every web page the user visits, and a
// page can open the event stream and post tool calls without any preflight.
// So a request that a browser sent on behalf of a page (it carries an Origin
// header, or Sec-Fetch-Site says it is not same-origin) is refused with 403
// unless its origin is in config.AllowedOrigins. When config.Token is set,
// requests without the matching bearer token are refused with 401.
func NewSSEHandler(mcpServer *mcp_server.MCPServer, config SSEConfig) http.Handler {
	var opts []mcp_server.SSEOption
	if len(config.AllowedOrigins) > 0 {
		opts = append(opts, mcp_server.WithSSECORS(
			mcp_server.WithCORSAllowedOrigins(config.AllowedOrigins...),
			mcp_server.WithCORSAllowedMethods(http.MethodGet, http.MethodPost, http.MethodOptions),
			mcp_server.WithCORSAllowedHeaders("Authorization", "Content-Type"),
		))
	}
	sseServer := mcp_server.NewSSEServer(mcpServer, opts...)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			if !slices.Contains(config.AllowedOrigins, origin) {
				http.Error(w, "Forbidden: origin not allowed", http.StatusForbidden)
				return
			}
		} else if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
			// A browser request made for another site that carries no Origin
			// header, such as a plain cross-site GET.
			http.Error(w, "Forbidden: cross-site request", http.StatusForbidden)
			return
		}

		// CORS preflight requests cannot carry credentials; the request that
		// follows the preflight is checked.
		if config.Token != "" && r.Method != http.MethodOptions {
			token, isBearer := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !isBearer || subtle.ConstantTimeCompare([]byte(token), []byte(config.Token)) != 1 {
				w.Header().Set("WWW-Authenticate", "Bearer")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		sseServer.ServeHTTP(&noWildcardCORSWriter{ResponseWriter: w}, r)
	})
}

// SSETokenEnvVar is the environment variable holding the SSE bearer token.
const SSETokenEnvVar = "DBN_GO_MCP_SSE_TOKEN"

// LoadSSEToken returns the SSE bearer token from tokenFile if it is set, and
// otherwise from the SSETokenEnvVar environment variable (empty if unset).
func LoadSSEToken(tokenFile string) (string, error) {
	if tokenFile == "" {
		return strings.TrimSpace(os.Getenv(SSETokenEnvVar)), nil
	}
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to read SSE token file: %w", err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("SSE token file '%s' is empty", tokenFile)
	}
	return token, nil
}

// IsLoopbackHostPort reports whether hostPort names a loopback listen address.
func IsLoopbackHostPort(hostPort string) bool {
	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// ServeSSE serves mcpServer's SSE transport on hostPort until it fails.
func ServeSSE(mcpServer *mcp_server.MCPServer, hostPort string, config SSEConfig) error {
	httpServer := &http.Server{
		Addr:              hostPort,
		Handler:           NewSSEHandler(mcpServer, config),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return httpServer.ListenAndServe()
}

// noWildcardCORSWriter drops the "Access-Control-Allow-Origin: *" header that
// mcp-go adds to the event stream when no CORS origins are configured, so no
// response ever grants read access to arbitrary web pages.
type noWildcardCORSWriter struct {
	http.ResponseWriter
	checked bool
}

func (w *noWildcardCORSWriter) dropWildcard() {
	if !w.checked {
		w.checked = true
		if w.Header().Get("Access-Control-Allow-Origin") == "*" {
			w.Header().Del("Access-Control-Allow-Origin")
		}
	}
}

func (w *noWildcardCORSWriter) WriteHeader(statusCode int) {
	w.dropWildcard()
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *noWildcardCORSWriter) Write(b []byte) (int, error) {
	w.dropWildcard()
	return w.ResponseWriter.Write(b)
}

func (w *noWildcardCORSWriter) Flush() {
	w.dropWildcard()
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *noWildcardCORSWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
