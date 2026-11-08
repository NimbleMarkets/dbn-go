// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"log/slog"
	"sync"
	"time"

	"github.com/NimbleMarkets/dbn-go/internal/mcp_meta"
	"golang.org/x/sync/singleflight"
	"nimbleterminal.dev/ntduckbox"
)

type ServerConfig struct {
	ApiKey   string  // Databento API key
	MaxCost  float64 // Max cost for a query
	CacheDir string  // Directory for cached data files
	ReadOnly bool    // When true, fetch_range is disabled (no billing possible)
}

// Server holds state for MCP data tool handlers.
// It embeds *mcp_meta.Server for access to MaxCost and Logger.
type Server struct {
	*mcp_meta.Server
	maxCost  float64
	readOnly bool // When true, fetch_range is not registered

	cacheDir string        // Directory for cached data files
	box      *ntduckbox.DB // Locked-down DuckDB in-memory instance via ntduckbox
	mu       sync.Mutex
	fetches  singleflight.Group

	// queryGate admits one query_cache call at a time. ntduckbox serializes
	// queries on a single connection and starts its QueryTimeout before the
	// connection is acquired, so without admission control a cheap query
	// queued behind a slow one fails with a confusing deadline error.
	queryGate chan struct{}
	queryWait time.Duration // How long to wait for the gate before reporting busy (0 = defaultQueryWait)
}

// NewServer creates a new Server with the given API key, and logger.
func NewServer(config ServerConfig, logger *slog.Logger) *Server {
	return &Server{
		Server:   mcp_meta.NewServer(config.ApiKey, logger),
		maxCost:  config.MaxCost,
		readOnly: config.ReadOnly,
		cacheDir: config.CacheDir,
	}
}

func (s *Server) GetMaxCost() float64 {
	return s.maxCost
}

func (s *Server) GetCacheDir() string {
	return s.cacheDir
}
