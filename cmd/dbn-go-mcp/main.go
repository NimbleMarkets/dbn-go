// Copyright (c) 2025 Neomantra Corp
//
// This is a Model Context Protocol (MCP) server for Databento APIs.
// It bridges LLMs and Databento's historical and metadata APIs.
//
// NOTE: this incurs billing, handle with care!
//

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/mark3labs/mcp-go/mcp"
	mcp_server "github.com/mark3labs/mcp-go/server"
	"github.com/relvacode/iso8601"
	"github.com/spf13/pflag"
)

///////////////////////////////////////////////////////////////////////////////

const (
	mcpServerVersion = "0.0.1"

	defaultSSEHostPort = ":8889"
	defaultLogDest     = "dbn-go-mcp.log"
	defaultMaxCost     = 1.0 // $1.00
)

type Config struct {
	ApiKey string // Databento API key

	LogJSON bool // Log in JSON format instead of text

	Name    string // Service Name
	Version string // Service Version

	UseSSE      bool   // Use SSE Transport instead of STDIO
	SSEHostPort string // HostPort to use for SSE

	Verbose bool // Verbose logging

	MaxCost float64 // Max cost for a query

	// TODO:
	//   allow/deny lists for schema/dataset
}

// Global configuration state
var config Config
var logger *slog.Logger

///////////////////////////////////////////////////////////////////////////////

func main() {
	var showHelp bool
	var apikeyFilename string
	var logFilename string

	pflag.StringVarP(&config.ApiKey, "key", "k", "", "Databento API key (or set 'DATABENTO_API_KEY' envvar)")
	pflag.StringVarP(&apikeyFilename, "key-file", "f", "", "File to read Databento API key from (or set 'DATABENTO_API_KEY_FILE' envvar)")
	pflag.StringVarP(&logFilename, "log-file", "l", "", "Log file destination (or MCP_LOG_FILE envvar). Default is stderr")
	pflag.BoolVarP(&config.LogJSON, "log-json", "j", false, "Log in JSON (default is plaintext)")
	pflag.StringVarP(&config.SSEHostPort, "port", "p", "", "host:port to listen to SSE connections")
	pflag.Float64VarP(&config.MaxCost, "max-cost", "c", defaultMaxCost, "Max cost, in dollars, for a query (<=0 is unlimited)")
	pflag.BoolVarP(&config.UseSSE, "sse", "", false, "Use SSE Transport (default is STDIO transport)")
	pflag.BoolVarP(&config.Verbose, "verbose", "v", false, "Verbose logging")
	pflag.BoolVarP(&showHelp, "help", "h", false, "Show help")
	pflag.Parse()

	if showHelp {
		fmt.Fprintf(os.Stdout, "usage: %s -k <api_key> [opts]\n\n", os.Args[0])
		pflag.PrintDefaults()
		os.Exit(0)
	}

	if config.ApiKey == "" { // prefer CLI option
		// check keyfile
		if apikeyFilename == "" {
			apikeyFilename = os.Getenv("DATABENTO_API_KEY_FILE")
		}
		if apikeyFilename != "" {
			bytes, err := os.ReadFile(apikeyFilename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to read key file: %s\n", err.Error())
				os.Exit(1)
			}
			config.ApiKey = strings.TrimSpace(string(bytes))
		} else {
			config.ApiKey = os.Getenv("DATABENTO_API_KEY")
		}
		requireValOrExit(config.ApiKey, "missing Databento API key, use --key or --file or set either DATABENTO_API_KEY or DATABENTO_API_KEY_FILE envvar\n")
	}

	if config.SSEHostPort == "" {
		config.SSEHostPort = defaultSSEHostPort
	}

	config.Name = "dbn-go-mcp"
	config.Version = mcpServerVersion

	// Set up logging
	logWriter := os.Stderr // default is stderr
	if logFilename == "" { // prefer CLI option
		logFilename = os.Getenv("MCP_LOG_FILE")
	}
	if logFilename != "" {
		logFile, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file: %s\n", err.Error())
			os.Exit(1)
		}
		logWriter = logFile
		defer logFile.Close()
	}

	var logLevel = slog.LevelInfo
	if config.Verbose {
		logLevel = slog.LevelDebug
	}

	if config.LogJSON {
		logger = slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{Level: logLevel}))
	} else {
		logger = slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{Level: logLevel}))
	}

	// Run our MCP server
	if err := run(); err != nil {
		logger.Error("run loop error", "error", err.Error())
		os.Exit(1)
	}

}

// requireValOrExit exits with an error message if `val` is empty.
func requireValOrExit(val string, errstr string) {
	if val == "" {
		fmt.Fprintf(os.Stderr, "%s\n", errstr)
		os.Exit(1)
	}
}

///////////////////////////////////////////////////////////////////////////////

func run() error {
	// Create the MCP Server
	mcpServer := mcp_server.NewMCPServer(config.Name, config.Version)
	registerTools(mcpServer)

	if config.UseSSE {
		sseServer := mcp_server.NewSSEServer(mcpServer)
		logger.Info("MCP SSE server started", "hostPort", config.SSEHostPort)
		if err := sseServer.Start(config.SSEHostPort); err != nil {
			return fmt.Errorf("MCP SSE server error: %w", err)
		}
	} else {
		logger.Info("MCP STDIO server started")
		if err := mcp_server.ServeStdio(mcpServer); err != nil {
			return fmt.Errorf("MCP STDIO server error: %w", err)
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////

// Valid datasets  to query. This will serve as a enum in our MCP API
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

// Valid schemas to query. This will serve as a enum in our MCP API
// via:  dbn-go-hist schemas -d XNAS.ITCH
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

// registerTools registers all MCP tools with the server.
func registerTools(mcpServer *mcp_server.MCPServer) error {
	// list_datasets - discovery tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("list_datasets",
			mcp.WithDescription("Lists all available Databento dataset codes. Use this to discover valid dataset values before querying. Optionally filter by a date range to see which datasets have data in that period. This does not incur any billing."),
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
		),
		listPublishersHandler,
	)
	// get_dataset_range - discovery tool (no billing)
	mcpServer.AddTool(
		mcp.NewTool("get_dataset_range",
			mcp.WithDescription("Returns the available date range (start and end) for a dataset. Use this to determine what time period of data is available before querying. This does not incur any billing."),
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
			mcp.WithString("dataset",
				mcp.Required(),
				mcp.Description("Dataset code (e.g. XNAS.ITCH, GLBX.MDP3). Use list_datasets to discover valid values."),
				mcp.Enum(validDatasets...),
			),
		),
		listUnitPricesHandler,
	)
	// get_cost
	mcpServer.AddTool(
		mcp.NewTool("get_cost",
			mcp.WithDescription("Returns the estimated cost in USD, billable data size in bytes, and record count for a query. Always call this before get_range to understand cost implications. This does not incur any billing."),
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
	return nil
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

func listDatasetsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var dateRange dbn_hist.DateRange

	if startStr, err := request.RequireString("start"); err == nil && startStr != "" {
		if startTime, err := iso8601.ParseString(startStr); err != nil {
			return mcp.NewToolResultErrorf("start was invalid ISO 8601: %s", err), nil
		} else {
			dateRange.Start = startTime
		}
	}
	if endStr, err := request.RequireString("end"); err == nil && endStr != "" {
		if endTime, err := iso8601.ParseString(endStr); err != nil {
			return mcp.NewToolResultErrorf("end was invalid ISO 8601: %s", err), nil
		} else {
			dateRange.End = endTime
		}
	}

	datasets, err := dbn_hist.ListDatasets(config.ApiKey, dateRange)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list datasets: %s", err), nil
	}

	jbytes, err := json.Marshal(datasets)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("list_datasets", "count", len(datasets))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func listSchemasHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataset, err := request.RequireString("dataset")
	if err != nil {
		return mcp.NewToolResultError("dataset must be set"), nil
	}
	dataset = strings.ToUpper(dataset)

	schemas, err := dbn_hist.ListSchemas(config.ApiKey, dataset)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list schemas: %s", err), nil
	}

	jbytes, err := json.Marshal(schemas)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("list_schemas", "dataset", dataset, "count", len(schemas))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func listFieldsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	schemaStr, err := request.RequireString("schema")
	if err != nil {
		return mcp.NewToolResultError("schema must be set"), nil
	}
	schemaStr = strings.ToLower(schemaStr)

	schema, err := dbn.SchemaFromString(schemaStr)
	if err != nil {
		return mcp.NewToolResultErrorf("invalid schema: %s", err), nil
	}

	fields, err := dbn_hist.ListFields(config.ApiKey, dbn.Encoding_Json, schema)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list fields: %s", err), nil
	}

	jbytes, err := json.Marshal(fields)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("list_fields", "schema", schemaStr, "count", len(fields))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func listPublishersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	publishers, err := dbn_hist.ListPublishers(config.ApiKey)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list publishers: %s", err), nil
	}

	jbytes, err := json.Marshal(publishers)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("list_publishers", "count", len(publishers))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func getDatasetRangeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataset, err := request.RequireString("dataset")
	if err != nil {
		return mcp.NewToolResultError("dataset must be set"), nil
	}
	dataset = strings.ToUpper(dataset)

	dateRange, err := dbn_hist.GetDatasetRange(config.ApiKey, dataset)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get dataset range: %s", err), nil
	}

	jbytes, err := json.Marshal(map[string]any{
		"dataset": dataset,
		"start":   dateRange.Start.Format(time.RFC3339),
		"end":     dateRange.End.Format(time.RFC3339),
	})
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("get_dataset_range", "dataset", dataset, "start", dateRange.Start, "end", dateRange.End)
	return mcp.NewToolResultText(string(jbytes)), nil
}

func getDatasetConditionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataset, err := request.RequireString("dataset")
	if err != nil {
		return mcp.NewToolResultError("dataset must be set"), nil
	}
	dataset = strings.ToUpper(dataset)

	var dateRange dbn_hist.DateRange
	if startStr, err := request.RequireString("start"); err == nil && startStr != "" {
		if startTime, err := iso8601.ParseString(startStr); err != nil {
			return mcp.NewToolResultErrorf("start was invalid ISO 8601: %s", err), nil
		} else {
			dateRange.Start = startTime
		}
	}
	if endStr, err := request.RequireString("end"); err == nil && endStr != "" {
		if endTime, err := iso8601.ParseString(endStr); err != nil {
			return mcp.NewToolResultErrorf("end was invalid ISO 8601: %s", err), nil
		} else {
			dateRange.End = endTime
		}
	}

	conditions, err := dbn_hist.GetDatasetCondition(config.ApiKey, dataset, dateRange)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get dataset condition: %s", err), nil
	}

	jbytes, err := json.Marshal(conditions)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("get_dataset_condition", "dataset", dataset, "count", len(conditions))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func listUnitPricesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataset, err := request.RequireString("dataset")
	if err != nil {
		return mcp.NewToolResultError("dataset must be set"), nil
	}
	dataset = strings.ToUpper(dataset)

	unitPrices, err := dbn_hist.ListUnitPrices(config.ApiKey, dataset)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list unit prices: %s", err), nil
	}

	jbytes, err := json.Marshal(unitPrices)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("list_unit_prices", "dataset", dataset, "count", len(unitPrices))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func getCostHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	p, errResult := parseCommonParams(request)
	if errResult != nil {
		return errResult, nil
	}

	metaParams := p.metadataQueryParams()

	cost, err := dbn_hist.GetCost(config.ApiKey, metaParams)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get cost: %s", err), nil
	}
	dataSize, err := dbn_hist.GetBillableSize(config.ApiKey, metaParams)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get data size: %s", err), nil
	}
	recordCount, err := dbn_hist.GetRecordCount(config.ApiKey, metaParams)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get record count: %s", err), nil
	}

	jbytes, err := json.Marshal(map[string]any{
		"query":        metaParams,
		"cost":         cost,
		"data_size":    dataSize,
		"record_count": recordCount,
	})
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("get_cost", "dataset", p.Dataset, "schema", p.SchemaStr, "symbol", p.Symbol,
		"start", p.StartStr, "end", p.EndStr, "cost", cost, "size", dataSize, "count", recordCount)

	return mcp.NewToolResultText(string(jbytes)), nil
}

func getRangeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	p, errResult := parseCommonParams(request)
	if errResult != nil {
		return errResult, nil
	}

	schema, err := dbn.SchemaFromString(p.SchemaStr)
	if err != nil {
		return mcp.NewToolResultErrorf("schema was invalid: %s", err), nil
	}

	metaParams := p.metadataQueryParams()
	cost, err := dbn_hist.GetCost(config.ApiKey, metaParams)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get cost: %s", err), nil
	}
	if config.MaxCost > 0 && cost > config.MaxCost {
		return mcp.NewToolResultErrorf("query exceeds budget: estimated cost $%.2f, budget $%.2f", cost, config.MaxCost), nil
	}

	jobParams := dbn_hist.SubmitJobParams{
		Dataset:      p.Dataset,
		Symbols:      p.Symbol,
		Schema:       schema,
		DateRange:    dbn_hist.DateRange{Start: p.StartTime, End: p.EndTime},
		Encoding:     dbn.Encoding_Json,
		PrettyPx:     true,
		PrettyTs:     true,
		MapSymbols:   true,
		SplitSymbols: false,
		Delivery:     dbn_hist.Delivery_Download,
		StypeIn:      dbn.SType_RawSymbol,
		StypeOut:     dbn.SType_InstrumentId,
	}
	rangeData, err := dbn_hist.GetRange(config.ApiKey, jobParams)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get range: %s", err), nil
	}

	logger.Info("get_range", "dataset", p.Dataset, "schema", p.SchemaStr, "symbol", p.Symbol,
		"start", p.StartStr, "end", p.EndStr, "cost", cost, "size", len(rangeData))
	return mcp.NewToolResultText(string(rangeData)), nil
}
