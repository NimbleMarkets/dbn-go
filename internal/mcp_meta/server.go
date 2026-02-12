// Copyright (c) 2025 Neomantra Corp

package mcp_meta

import "log/slog"

// Server holds shared state for MCP metadata tool handlers.
type Server struct {
	apiKey string
	Logger *slog.Logger
}

// NewServer creates a new Server with the given API key, and logger.
func NewServer(apiKey string, logger *slog.Logger) *Server {
	return &Server{apiKey: apiKey, Logger: logger}
}

// GetApiKey returns the Databento API key.
func (s *Server) GetApiKey() string {
	return s.apiKey
}
