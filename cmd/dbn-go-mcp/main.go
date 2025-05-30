// Copyright (c) 2025 Neomantra Corp
//
// This is a Nodel Context Protocol (MCP) server for Databento APIs.
// It bridges LLMs and Databento's historical and metadata APIs.
//
// NOTE: this incurs billing, handle with care!
//

package main

import (
	"context"
	"encoding/json"
	"errors"
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

// getToolDescriptions return the mcp.Tools for our MCPServer
func registerTools(mcpServer *mcp_server.MCPServer) error {
	// get_cost
	getCostTool := mcp.NewTool("get_cost",
		mcp.WithDescription("Returns the estimated cost of all records of a DBN schema for a given symbol and date range"),
		mcp.WithString("dataset",
			mcp.Required(),
			mcp.Description("Dataset to query"),
			mcp.Enum(validDatasets...),
		),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("Schema to query"),
			mcp.Enum(validSchemas...),
		),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Symbol to query"),
		),
		mcp.WithString("start",
			mcp.Required(),
			mcp.Description("start of range, as ISO 8601 datetime"),
		),
		mcp.WithString("end",
			mcp.Required(),
			mcp.Description("end of range, as ISO 8601 datetime"),
		),
	)
	mcpServer.AddTool(getCostTool, getCostHandler)
	// get_range
	getRangeTool := mcp.NewTool("get_range",
		mcp.WithDescription("Returns all records of a DBN dataset/schema for a given symbol and date range"),
		mcp.WithString("dataset",
			mcp.Required(),
			mcp.Description("Dataset to query"),
			mcp.Enum(validDatasets...),
		),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("Schema to query"),
			mcp.Enum(validSchemas...),
		),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Symbol to query"),
		),
		mcp.WithString("start",
			mcp.Required(),
			mcp.Description("start of range, as ISO 8601 datetime"),
		),
		mcp.WithString("end",
			mcp.Required(),
			mcp.Description("end of range, as ISO 8601 datetime"),
		),
	)
	mcpServer.AddTool(getRangeTool, getRangeHandler)
	return nil
}

///////////////////////////////////////////////////////////////////////////////

func getCostHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var dataset, schemaStr, symbol, startStr, endStr string
	var startTime, endTime time.Time
	var err error

	if dataset, err = request.RequireString("dataset"); err != nil {
		return nil, errors.New("dataset must be set")
	} else {
		dataset = strings.ToUpper(dataset)
		if !slices.Contains(validDatasets, dataset) {
			return nil, errors.New("unknown dataset")
		}
	}
	if schemaStr, err = request.RequireString("schema"); err != nil {
		return nil, errors.New("schema must be set")
	} else {
		schemaStr = strings.ToLower(schemaStr)
		if !slices.Contains(validSchemas, schemaStr) {
			return nil, errors.New("unknown schema")
		}
	}
	if symbol, err = request.RequireString("symbol"); err != nil {
		return nil, errors.New("symbol must be valid")
	}
	if startStr, err = request.RequireString("start"); err != nil {
		return nil, errors.New("start must be valid")
	} else {
		if startTime, err = iso8601.ParseString(startStr); err != nil {
			return nil, fmt.Errorf("start was invalid: %w", err)
		}
	}
	if endStr, err = request.RequireString("end"); err != nil {
		return nil, errors.New("end must be valid")
	} else {
		if endTime, err = iso8601.ParseString(endStr); err != nil {
			return nil, fmt.Errorf("end was invalid: %w", err)
		}
	}

	metaParams := dbn_hist.MetadataQueryParams{
		Dataset:   dataset,
		Schema:    schemaStr,
		Symbols:   []string{symbol},
		DateRange: dbn_hist.DateRange{Start: startTime, End: endTime},
		Mode:      dbn_hist.FeedMode_Historical,
		StypeIn:   dbn.SType_RawSymbol,
	}

	cost, err := dbn_hist.GetCost(config.ApiKey, metaParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost: %w", err)
	}
	dataSize, err := dbn_hist.GetBillableSize(config.ApiKey, metaParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get data size: %w", err)
	}
	recordCount, err := dbn_hist.GetRecordCount(config.ApiKey, metaParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get record count: %w", err)
	}

	jbytes, err := json.Marshal(map[string]interface{}{
		"query":        metaParams,
		"cost":         cost,
		"data_size":    dataSize,
		"record_count": recordCount,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to json.Marshal results: %w", err)
	}

	logger.Info("get_cost", "dataset", dataset, "schema", schemaStr, "symbol", symbol,
		"start", startStr, "end", endStr, "cost", cost, "size", dataSize, "count", recordCount)

	return mcp.NewToolResultText(string(jbytes)), nil
}

func getRangeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var dataset, schemaStr, symbol, startStr, endStr string
	var schema dbn.Schema
	var startTime, endTime time.Time
	var err error

	if dataset, err = request.RequireString("dataset"); err != nil {
		return nil, errors.New("dataset must be set")
	} else {
		dataset = strings.ToUpper(dataset)
		if !slices.Contains(validDatasets, dataset) {
			return nil, errors.New("unknown dataset")
		}
	}
	if schemaStr, err = request.RequireString("schema"); err != nil {
		return nil, errors.New("schema must be set")
	} else {
		schemaStr = strings.ToLower(schemaStr)
		if !slices.Contains(validSchemas, schemaStr) {
			return nil, errors.New("unknown schema")
		}
		if schema, err = dbn.SchemaFromString(schemaStr); err != nil {
			return nil, fmt.Errorf("schema was invalid: %w", err)
		}
	}
	if symbol, err = request.RequireString("symbol"); err != nil {
		return nil, errors.New("symbol must be valid")
	}
	if startStr, err = request.RequireString("start"); err != nil {
		return nil, errors.New("start must be valid")
	} else {
		if startTime, err = iso8601.ParseString(startStr); err != nil {
			return nil, fmt.Errorf("start was invalid: %w", err)
		}
	}
	if endStr, err = request.RequireString("end"); err != nil {
		return nil, errors.New("end must be valid")
	} else {
		if endTime, err = iso8601.ParseString(endStr); err != nil {
			return nil, fmt.Errorf("end was invalid: %w", err)
		}
	}

	metaParams := dbn_hist.MetadataQueryParams{
		Dataset:   dataset,
		Schema:    schemaStr,
		Symbols:   []string{symbol},
		DateRange: dbn_hist.DateRange{Start: startTime, End: endTime},
		Mode:      dbn_hist.FeedMode_Historical,
		StypeIn:   dbn.SType_RawSymbol,
	}
	cost, err := dbn_hist.GetCost(config.ApiKey, metaParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost: %w", err)
	}
	if cost > config.MaxCost { // more than budget?
		return nil, fmt.Errorf("query was above budget, cost: %0.2f budget: %0.2f", cost, config.MaxCost)
	}

	jobParams := dbn_hist.SubmitJobParams{
		Dataset:      dataset,
		Symbols:      symbol,
		Schema:       schema,
		DateRange:    dbn_hist.DateRange{Start: startTime, End: endTime},
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
		return nil, fmt.Errorf("failed to get range: %w", err)
	}

	logger.Info("get_range", "dataset", dataset, "schema", schemaStr, "symbol", symbol,
		"start", startStr, "end", endStr, "cost", cost, "size", len(rangeData))
	return mcp.NewToolResultText(string(rangeData)), nil
}
