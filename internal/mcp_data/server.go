// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"database/sql"
	"log/slog"
	"sync"

	"github.com/NimbleMarkets/dbn-go/internal/mcp_meta"
)

type ServerConfig struct {
	ApiKey   string  // Databento API key
	MaxCost  float64 // Max cost for a query
	CacheDir string  // Directory for cached data files
	CacheDB  string  // Path to DuckDB database file (reserved for future use)
}

// Server holds state for MCP data tool handlers.
// It embeds *mcp_meta.Server for access to MaxCost and Logger.
type Server struct {
	*mcp_meta.Server
	maxCost float64

	cacheDir string  // Directory for cached data files
	cacheDB  string  // Path to DuckDB database file (reserved for future use)
	db       *sql.DB // DuckDB in-memory connection
	mu       sync.Mutex
}

// NewServer creates a new Server with the given API key, and logger.
func NewServer(config ServerConfig, logger *slog.Logger) *Server {
	return &Server{
		Server:   mcp_meta.NewServer(config.ApiKey, logger),
		maxCost:  config.MaxCost,
		cacheDir: config.CacheDir,
		cacheDB:  config.CacheDB,
	}
}

func (s *Server) GetMaxCost() float64 {
	return s.maxCost
}

func (s *Server) GetCacheDir() string {
	return s.cacheDir
}

func (s *Server) GetCacheDB() string {
	return s.cacheDB
}
