// Copyright (c) 2025 Neomantra Corp

package main

import (
	"slices"
	"strings"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/mark3labs/mcp-go/mcp"
	mcp_server "github.com/mark3labs/mcp-go/server"
	"github.com/relvacode/iso8601"
)

///////////////////////////////////////////////////////////////////////////////

// Valid datasets to query. This will serve as an enum in our MCP API.
// via: dbn-go-hist datasets
var validDatasets = []string{
	"ARCX.PILLAR",
	"DBEQ.BASIC",
	"EPRL.DOM",
	"EQUS.MINI",
	"EQUS.SUMMARY",
	"GLBX.MDP3",
	"IEXG.TOPS",
	"IFEU.IMPACT",
	"MEMX.MEMOIR",
	"NDEX.IMPACT",
	"OPRA.PILLAR",
	"XASE.PILLAR",
	"XBOS.ITCH",
	"XCHI.PILLAR",
	"XCIS.TRADESBBO",
	"XNAS.BASIC",
	"XNAS.ITCH",
	"XNYS.PILLAR",
	"XPSX.ITCH",
}

// Valid schemas to query. This will serve as an enum in our MCP API.
// via: dbn-go-hist schemas -d XNAS.ITCH
var validSchemas = []string{
	"mbo",
	"mbp-1",
	"mbp-10",
	"bbo-1s",
	"bbo-1m",
	"tbbo",
	"trades",
	"ohlcv-1s",
	"ohlcv-1m",
	"ohlcv-1h",
	"ohlcv-1d",
	"definition",
	"statistics",
	"status",
	"imbalance",
}

// Valid symbology types for resolve_symbols.
var validStypes = []string{
	"raw_symbol",
	"instrument_id",
	"smart",
	"continuous",
	"parent",
	"nasdaq",
	"cms",
	"isin",
	"us_code",
	"bbg_comp_id",
	"bbg_comp_ticker",
	"figi",
	"figi_ticker",
}

///////////////////////////////////////////////////////////////////////////////

// registerTools registers all MCP tools with the server.
func registerTools(mcpServer *mcp_server.MCPServer) {
	// list_datasets - discovery tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("list_datasets",
			mcp.WithDescription("Lists all available Databento dataset codes. Use this to discover valid dataset values before querying. Optionally filter by a date range to see which datasets have data in that period. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("start",
				mcp.Description("Optional start of date range filter, as ISO 8601 datetime (e.g. 2024-01-01)"),
			),
			mcp.WithString("end",
				mcp.Description("Optional end of date range filter, as ISO 8601 datetime (e.g. 2024-12-31)"),
			),
		),
		listDatasetsHandler,
	)
	// list_schemas - discovery tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("list_schemas",
			mcp.WithDescription("Lists all available data record schemas for a given dataset. Use this to discover which schemas (e.g. trades, ohlcv-1d, mbp-1) are available before querying. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("dataset",
				mcp.Required(),
				mcp.Description("Dataset code (e.g. XNAS.ITCH, GLBX.MDP3). Use list_datasets to discover valid values."),
				mcp.Enum(validDatasets...),
			),
		),
		listSchemasHandler,
	)
	// list_fields - discovery tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("list_fields",
			mcp.WithDescription("Lists all field names and types for a given schema. Use this to understand the structure of records before querying with get_range. Returns fields for JSON encoding. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("schema",
				mcp.Required(),
				mcp.Description("Schema to inspect (e.g. trades, ohlcv-1d, mbp-1). Use list_schemas to discover valid values."),
				mcp.Enum(validSchemas...),
			),
		),
		listFieldsHandler,
	)
	// list_publishers - discovery tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("list_publishers",
			mcp.WithDescription("Lists all Databento publishers with their publisher_id, dataset code, venue, and description. Use this to discover which venues and data sources are available, and to map publisher IDs seen in records back to their source. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
		),
		listPublishersHandler,
	)
	// get_dataset_range - discovery tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("get_dataset_range",
			mcp.WithDescription("Returns the available date range (start and end) for a dataset. Use this to determine what time period of data is available before querying. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("dataset",
				mcp.Required(),
				mcp.Description("Dataset code (e.g. XNAS.ITCH, GLBX.MDP3). Use list_datasets to discover valid values."),
				mcp.Enum(validDatasets...),
			),
		),
		getDatasetRangeHandler,
	)
	// get_dataset_condition - discovery tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("get_dataset_condition",
			mcp.WithDescription("Returns the data quality and availability condition for each day in a dataset's date range. Conditions are: 'available', 'degraded', 'pending', 'missing', or 'intraday'. Use this to check if data exists and is reliable before querying. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("dataset",
				mcp.Required(),
				mcp.Description("Dataset code (e.g. XNAS.ITCH, GLBX.MDP3). Use list_datasets to discover valid values."),
				mcp.Enum(validDatasets...),
			),
			mcp.WithString("start",
				mcp.Description("Optional start of date range filter, as ISO 8601 datetime (e.g. 2024-01-01). Defaults to beginning of dataset."),
			),
			mcp.WithString("end",
				mcp.Description("Optional end of date range filter, as ISO 8601 datetime (e.g. 2024-01-31). Defaults to end of dataset."),
			),
		),
		getDatasetConditionHandler,
	)
	// list_unit_prices - discovery tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("list_unit_prices",
			mcp.WithDescription("Lists the unit prices in US dollars per gigabyte for each schema and feed mode in a dataset. Use this to understand relative costs of different schemas before querying. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("dataset",
				mcp.Required(),
				mcp.Description("Dataset code (e.g. XNAS.ITCH, GLBX.MDP3). Use list_datasets to discover valid values."),
				mcp.Enum(validDatasets...),
			),
		),
		listUnitPricesHandler,
	)
	// resolve_symbols - discovery tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("resolve_symbols",
			mcp.WithDescription("Resolves symbols from one symbology type to another for a given dataset and date range. For example, convert a raw symbol like 'AAPL' to its instrument_id, or resolve a continuous futures contract. Returns mappings with date-range validity, plus lists of partial and not-found symbols. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("dataset",
				mcp.Required(),
				mcp.Description("Dataset code (e.g. XNAS.ITCH, GLBX.MDP3). Use list_datasets to discover valid values."),
				mcp.Enum(validDatasets...),
			),
			mcp.WithString("symbols",
				mcp.Required(),
				mcp.Description("Comma-separated list of symbols to resolve (e.g. 'AAPL' or 'AAPL,TSLA,QQQ')"),
			),
			mcp.WithString("stype_in",
				mcp.Description("Input symbology type (default: raw_symbol). One of: raw_symbol, instrument_id, smart, continuous, parent, nasdaq, cms, isin, us_code, bbg_comp_id, bbg_comp_ticker, figi, figi_ticker"),
				mcp.Enum(validStypes...),
			),
			mcp.WithString("stype_out",
				mcp.Description("Output symbology type (default: instrument_id). One of: raw_symbol, instrument_id, smart, continuous, parent, nasdaq, cms, isin, us_code, bbg_comp_id, bbg_comp_ticker, figi, figi_ticker"),
				mcp.Enum(validStypes...),
			),
			mcp.WithString("start",
				mcp.Required(),
				mcp.Description("Start of date range, as ISO 8601 datetime (e.g. 2024-01-01)"),
			),
			mcp.WithString("end",
				mcp.Description("Optional end of date range (exclusive), as ISO 8601 datetime (e.g. 2024-12-31). Defaults to current date."),
			),
		),
		resolveSymbolsHandler,
	)
	// get_cost
	mcpServer.AddTool(
		mcp.NewTool("get_cost",
			mcp.WithDescription("Returns the estimated cost in USD, billable data size in bytes, and record count for a query. Always call this before get_range to understand cost implications. This does not incur any billing."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("dataset",
				mcp.Required(),
				mcp.Description("Dataset to query (e.g. XNAS.ITCH, GLBX.MDP3)"),
				mcp.Enum(validDatasets...),
			),
			mcp.WithString("schema",
				mcp.Required(),
				mcp.Description("Schema to query (e.g. trades, ohlcv-1d, mbp-1)"),
				mcp.Enum(validSchemas...),
			),
			mcp.WithString("symbol",
				mcp.Required(),
				mcp.Description("Symbol to query (e.g. AAPL, TSLA, QQQ)"),
			),
			mcp.WithString("start",
				mcp.Required(),
				mcp.Description("Start of range, as ISO 8601 datetime (e.g. 2024-01-15 or 2024-01-15T09:30:00Z)"),
			),
			mcp.WithString("end",
				mcp.Required(),
				mcp.Description("End of range (exclusive), as ISO 8601 datetime (e.g. 2024-01-16)"),
			),
		),
		getCostHandler,
	)
	// get_range
	mcpServer.AddTool(
		mcp.NewTool("get_range",
			mcp.WithDescription("Returns all records as JSON for a dataset/schema/symbol over a date range. CAUTION: This incurs Databento billing. Call get_cost first to check the cost. The server enforces a per-query budget limit. For large results, prefer ohlcv-1d or ohlcv-1h schemas which return compact summaries."),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
			mcp.WithString("dataset",
				mcp.Required(),
				mcp.Description("Dataset to query (e.g. XNAS.ITCH, GLBX.MDP3)"),
				mcp.Enum(validDatasets...),
			),
			mcp.WithString("schema",
				mcp.Required(),
				mcp.Description("Schema to query (e.g. trades, ohlcv-1d, mbp-1)"),
				mcp.Enum(validSchemas...),
			),
			mcp.WithString("symbol",
				mcp.Required(),
				mcp.Description("Symbol to query (e.g. AAPL, TSLA, QQQ)"),
			),
			mcp.WithString("start",
				mcp.Required(),
				mcp.Description("Start of range, as ISO 8601 datetime (e.g. 2024-01-15 or 2024-01-15T09:30:00Z)"),
			),
			mcp.WithString("end",
				mcp.Required(),
				mcp.Description("End of range (exclusive), as ISO 8601 datetime (e.g. 2024-01-16)"),
			),
		),
		getRangeHandler,
	)
}

///////////////////////////////////////////////////////////////////////////////

// commonParams holds the parsed and validated parameters shared across tool handlers.
type commonParams struct {
	Dataset   string
	SchemaStr string
	Symbol    string
	StartStr  string
	EndStr    string
	StartTime time.Time
	EndTime   time.Time
}

// parseCommonParams extracts and validates dataset, schema, symbol, start, and end
// from a tool request. Returns a tool-result error (not a Go error) so the LLM
// can see and reason about validation failures.
func parseCommonParams(request mcp.CallToolRequest) (*commonParams, *mcp.CallToolResult) {
	var p commonParams
	var err error

	if p.Dataset, err = request.RequireString("dataset"); err != nil {
		return nil, mcp.NewToolResultError("dataset must be set")
	}
	p.Dataset = strings.ToUpper(p.Dataset)
	if !slices.Contains(validDatasets, p.Dataset) {
		return nil, mcp.NewToolResultErrorf("unknown dataset: %s", p.Dataset)
	}

	if p.SchemaStr, err = request.RequireString("schema"); err != nil {
		return nil, mcp.NewToolResultError("schema must be set")
	}
	p.SchemaStr = strings.ToLower(p.SchemaStr)
	if !slices.Contains(validSchemas, p.SchemaStr) {
		return nil, mcp.NewToolResultErrorf("unknown schema: %s", p.SchemaStr)
	}

	if p.Symbol, err = request.RequireString("symbol"); err != nil {
		return nil, mcp.NewToolResultError("symbol must be set")
	}

	if p.StartStr, err = request.RequireString("start"); err != nil {
		return nil, mcp.NewToolResultError("start must be set")
	}
	if p.StartTime, err = iso8601.ParseString(p.StartStr); err != nil {
		return nil, mcp.NewToolResultErrorf("start was invalid ISO 8601: %s", err)
	}

	if p.EndStr, err = request.RequireString("end"); err != nil {
		return nil, mcp.NewToolResultError("end must be set")
	}
	if p.EndTime, err = iso8601.ParseString(p.EndStr); err != nil {
		return nil, mcp.NewToolResultErrorf("end was invalid ISO 8601: %s", err)
	}

	return &p, nil
}

// metadataQueryParams builds a MetadataQueryParams from commonParams.
func (p *commonParams) metadataQueryParams() dbn_hist.MetadataQueryParams {
	return dbn_hist.MetadataQueryParams{
		Dataset:   p.Dataset,
		Schema:    p.SchemaStr,
		Symbols:   []string{p.Symbol},
		DateRange: dbn_hist.DateRange{Start: p.StartTime, End: p.EndTime},
		Mode:      dbn_hist.FeedMode_Historical,
		StypeIn:   dbn.SType_RawSymbol,
	}
}

// parseOptionalDateRange extracts optional start and end parameters into a DateRange.
// Returns a tool-result error if a provided value is not valid ISO 8601.
func parseOptionalDateRange(request mcp.CallToolRequest) (dbn_hist.DateRange, *mcp.CallToolResult) {
	var dateRange dbn_hist.DateRange

	if startStr, err := request.RequireString("start"); err == nil && startStr != "" {
		if startTime, err := iso8601.ParseString(startStr); err != nil {
			return dateRange, mcp.NewToolResultErrorf("start was invalid ISO 8601: %s", err)
		} else {
			dateRange.Start = startTime
		}
	}
	if endStr, err := request.RequireString("end"); err == nil && endStr != "" {
		if endTime, err := iso8601.ParseString(endStr); err != nil {
			return dateRange, mcp.NewToolResultErrorf("end was invalid ISO 8601: %s", err)
		} else {
			dateRange.End = endTime
		}
	}

	return dateRange, nil
}

