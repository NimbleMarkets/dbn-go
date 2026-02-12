// Copyright (c) 2025 Neomantra Corp

package mcp_data

import "github.com/NimbleMarkets/dbn-go/internal/mcp_meta"

// Server holds state for MCP data tool handlers.
// It embeds *mcp_meta.Server for access to ApiKey, MaxCost, and Logger.
type Server struct {
	*mcp_meta.Server

	CacheDir string // Directory for cached data files
	CacheDB  string // Path to DuckDB database file
}
