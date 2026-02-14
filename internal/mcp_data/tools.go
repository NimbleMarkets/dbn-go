// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"github.com/NimbleMarkets/dbn-go/internal/mcp_meta"
	"github.com/mark3labs/mcp-go/mcp"
	mcp_server "github.com/mark3labs/mcp-go/server"
)

///////////////////////////////////////////////////////////////////////////////

// RegisterDataTools registers data-specific MCP tools (fetch_range + cache tools).
func (s *Server) RegisterDataTools(mcpServer *mcp_server.MCPServer) {
	// fetch_range
	mcpServer.AddTool(
		mcp.NewTool("fetch_range",
			mcp.WithDescription("Fetches market data from Databento and caches it locally as Parquet. Returns metadata about the cached file (path, record count, size, view name). Use query_cache to query the data with SQL. CAUTION: This incurs Databento billing. Call get_cost first to check the cost. Supported schemas: ohlcv-1s/1m/1h/1d, trades, mbp-1, tbbo, imbalance, statistics."),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("dataset",
				mcp.Required(),
				mcp.Description("Dataset to query (e.g. XNAS.ITCH, GLBX.MDP3)"),
				mcp.Enum(mcp_meta.ValidDatasets...),
			),
			mcp.WithString("schema",
				mcp.Required(),
				mcp.Description("Schema to query (e.g. trades, ohlcv-1d, mbp-1)"),
				mcp.Enum(mcp_meta.ValidSchemas...),
			),
			mcp.WithString("symbols",
				mcp.Required(),
				mcp.Description("Comma-separated symbols to query (e.g. 'AAPL' or 'AAPL,TSLA,MSFT'). Up to 2,000 symbols per request."),
			),
			mcp.WithString("stype_in",
				mcp.Description("Input symbology type (default: raw_symbol). Use 'continuous' for futures like ES.c.0."),
				mcp.Enum(mcp_meta.ValidStypes...),
			),
			mcp.WithString("stype_out",
				mcp.Description("Output symbology type for symbol mapping in results (default: instrument_id)."),
				mcp.Enum(mcp_meta.ValidStypes...),
			),
			mcp.WithString("start",
				mcp.Required(),
				mcp.Description("Start of range, as ISO 8601 datetime (e.g. 2024-01-15 or 2024-01-15T09:30:00Z)"),
			),
			mcp.WithString("end",
				mcp.Required(),
				mcp.Description("End of range (exclusive), as ISO 8601 datetime (e.g. 2024-01-16)"),
			),
			mcp.WithBoolean("force",
				mcp.Description("Force re-fetch even if data is already cached (default: false)."),
			),
		),
		s.fetchRangeHandler,
	)
	// query_cache - SQL tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("query_cache",
			mcp.WithDescription("Query cached market data using DuckDB SQL. Returns results as CSV. Data previously fetched via fetch_range is stored locally and can be queried without additional billing. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("sql",
				mcp.Required(),
				mcp.Description("SQL query to execute against cached data"),
			),
		),
		s.queryCacheHandler,
	)
	// list_cache - cache tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("list_cache",
			mcp.WithDescription("Lists all cached datasets with their schema, symbols, date ranges, and size. Use this to see what data is available locally without additional billing. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
		),
		s.listCacheHandler,
	)
	// clear_cache - cache tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("clear_cache",
			mcp.WithDescription("Removes cached data. Optionally filter by dataset and/or schema. With no parameters, clears all cached data."),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("dataset",
				mcp.Description("Optional dataset to clear (e.g. XNAS.ITCH). If omitted, clears all datasets."),
				mcp.Enum(mcp_meta.ValidDatasets...),
			),
			mcp.WithString("schema",
				mcp.Description("Optional schema to clear (e.g. trades, ohlcv-1d). If omitted, clears all schemas."),
				mcp.Enum(mcp_meta.ValidSchemas...),
			),
		),
		s.clearCacheHandler,
	)
}
