// Copyright (c) 2025 Neomantra Corp

package mcp_meta

import "log/slog"

// Server holds shared state for MCP metadata tool handlers.
type Server struct {
	ApiKey  string
	MaxCost float64
	Logger  *slog.Logger
}
