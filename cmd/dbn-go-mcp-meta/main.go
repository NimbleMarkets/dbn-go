// Copyright (c) 2025 Neomantra Corp
//
// This is a metadata-only Model Context Protocol (MCP) server for Databento APIs.
// It bridges LLMs and Databento's metadata discovery APIs.
// No data download tools are available — use dbn-go-mcp-data for that.
//

package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/NimbleMarkets/dbn-go/internal/mcp_meta"
	mcp_server "github.com/mark3labs/mcp-go/server"
	"github.com/spf13/pflag"
)

///////////////////////////////////////////////////////////////////////////////

const (
	mcpServerVersion = "0.0.1"

	defaultSSEHostPort = ":8889"
	defaultMaxCost     = 1.0 // $1.00

	// serverInstructions is sent to LLM clients during MCP initialization.
	serverInstructions = `dbn-go-mcp-meta provides metadata-only access to Databento's historical market data APIs. No data download tools are available — use dbn-go-mcp-data for get_range and cached queries.

None of these tools incur Databento billing charges.

Recommended workflow:
1. Use list_datasets to discover available datasets.
2. Use list_schemas to see which schemas a dataset supports.
3. Use list_fields to understand the record structure of a schema.
4. Use get_dataset_range and get_dataset_condition to verify data availability.
5. Use get_cost to estimate the cost of a potential query.

Additional discovery tools: list_publishers (venue/source info), list_unit_prices (cost per schema), resolve_symbols (symbol mapping across symbologies).`
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

	config.Name = "dbn-go-mcp-meta"
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
	mcpServer := mcp_server.NewMCPServer(config.Name, config.Version,
		mcp_server.WithRecovery(),
		mcp_server.WithInstructions(serverInstructions),
	)

	srv := mcp_meta.NewServer(config.ApiKey, config.MaxCost, logger)
	srv.RegisterMetaTools(mcpServer)

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
