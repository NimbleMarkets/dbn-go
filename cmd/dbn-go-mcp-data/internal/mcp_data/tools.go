// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"github.com/NimbleMarkets/dbn-go/internal/mcp_meta"
	"github.com/mark3labs/mcp-go/mcp"
	mcp_server "github.com/mark3labs/mcp-go/server"
)

///////////////////////////////////////////////////////////////////////////////

// RegisterDataTools registers data-specific MCP tools (fetch_range + cache tools).
// When ReadOnly is true, fetch_range is not registered, preventing any billing.
func (s *Server) RegisterDataTools(mcpServer *mcp_server.MCPServer) {
	if s.readOnly {
		s.Logger.Info("read-only mode: fetch_range tool disabled")
	} else {
		s.registerFetchRange(mcpServer)
	}

	// query_cache - SQL tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("query_cache",
			mcp.WithDescription("Query cached market data using DuckDB SQL. Data previously fetched via fetch_range is stored locally and can be queried without additional billing. This does not incur any billing.\n\n"+
				"Each cached dataset/schema has two views. Prefer the alias view (plain identifier, e.g. xnas_itch_trades): its timestamp columns are TIMESTAMPTZ, so date_trunc, extract, and time_bucket work directly. "+
				"The raw view (\"XNAS.ITCH/trades\", must be double-quoted) keeps nanosecond TIMESTAMPTZ_NS timestamps, which most date functions reject without a ::TIMESTAMPTZ cast. "+
				"Call describe_cache first to see view names and exact column names/types (e.g. bid_px_00, not bid_px). "+
				"SUMMARIZE works on alias views; on raw views it fails because nanosecond timestamps cannot be averaged. "+
				"Timestamps are UTC. A US equities session (04:00-20:00 New York) spans 08:00-00:00 UTC, so filter or group by (ts_event AT TIME ZONE 'America/New_York')::DATE when the question is about a trading day. "+
				"Prices are doubles; undefined prices and timestamps are NULL. "+
				"Allowed: a single read-only SELECT-style statement (SELECT, WITH, FROM-first, VALUES, DESCRIBE, SUMMARIZE, SHOW TABLES). Not allowed: DDL/DML, PRAGMA/SET, EXPLAIN, PIVOT, multiple statements, and reading files."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("sql",
				mcp.Required(),
				mcp.Description("SQL query to execute against cached data"),
			),
			mcp.WithString("format",
				mcp.Description("Output format: 'csv' (default) returns comma-separated values with header row. 'markdown' returns a Markdown table. 'json_rows' returns an array of row objects. 'json_columnar' returns an object mapping column names to value arrays — ideal for Chart.js datasets."),
				mcp.Enum("csv", "markdown", "json_rows", "json_columnar"),
			),
			mcp.WithNumber("max_rows",
				mcp.Description("Maximum number of rows to return. If set, limits results regardless of any LIMIT in the SQL. Default: 10000."),
			),
		),
		s.queryCacheHandler,
	)
	// describe_cache - schema introspection (no billing)
	mcpServer.AddTool(
		mcp.NewTool("describe_cache",
			mcp.WithDescription("Describes the columns (name and DuckDB type) of cached views. With no arguments, lists every cached view with its alias and the alias view's columns. Pass 'view' (either the raw name like \"XNAS.ITCH/trades\" or the alias like xnas_itch_trades) to describe exactly that view. Call this before query_cache to get column names right. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("view",
				mcp.Description("Optional view to describe: raw name (e.g. XNAS.ITCH/trades) or alias (e.g. xnas_itch_trades). Omit to describe all cached views."),
			),
		),
		s.describeCacheHandler,
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

// registerFetchRange adds the fetch_range tool to the MCP server.
func (s *Server) registerFetchRange(mcpServer *mcp_server.MCPServer) {
	mcpServer.AddTool(
		mcp.NewTool("fetch_range",
			mcp.WithDescription("Fetches market data from Databento and caches it locally as Parquet. Returns metadata about the cached file (path, record count, size, view name). Use query_cache to query the data with SQL. CAUTION: This incurs Databento billing. Call get_cost first to check the cost. Supported schemas: ohlcv-1s/1m/1h/1d, trades, mbp-1, mbp-10, tbbo, cmbp-1, bbo-1s/1m, imbalance, statistics, status."),
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
}
