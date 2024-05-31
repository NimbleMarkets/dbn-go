// Copyright (c) 2024 Neomantra Corp
//
// NOTE: this incurs billing, handle with care!
//

package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	dbn "github.com/NimbleMarkets/dbn-go"
	dbn_live "github.com/NimbleMarkets/dbn-go/live"
	"github.com/klauspost/compress/zstd"
	"github.com/relvacode/iso8601"
	"github.com/spf13/pflag"
)

///////////////////////////////////////////////////////////////////////////////

type Config struct {
	OutFilename string
	ApiKey      string
	Dataset     string
	Schema      string
	Symbols     []string
	StartTime   time.Time
	Verbose     bool
}

///////////////////////////////////////////////////////////////////////////////

func main() {
	var config Config
	var startTimeArg string
	var showHelp bool

	pflag.StringVarP(&config.Dataset, "dataset", "d", "", "Dataset to subscribe to ")
	pflag.StringVarP(&config.Schema, "schema", "s", "", "Schema to subscribe to")
	pflag.StringVarP(&config.ApiKey, "key", "k", "", "Databento API key (or set 'DATABENTO_API_KEY' envvar)")
	pflag.StringVarP(&config.OutFilename, "out", "o", "", "Output filename for DBN stream ('-' for stdout)")
	pflag.StringVarP(&startTimeArg, "start", "t", "", "Start time to request as ISO 8601 format (default: now)")
	pflag.BoolVarP(&config.Verbose, "verbose", "v", false, "Verbose logging")
	pflag.BoolVarP(&showHelp, "help", "h", false, "Show help")
	pflag.Parse()

	config.Symbols = pflag.Args()

	if showHelp {
		fmt.Fprintf(os.Stdout, "usage: %s -d <dataset> -s <schema> [opts] symbol1 symbol2 ...\n\n", os.Args[0])
		pflag.PrintDefaults()
		os.Exit(0)
	}

	if startTimeArg != "" {
		var err error
		config.StartTime, err = iso8601.ParseString(startTimeArg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse --start as ISO 8601 time: %s\n", err.Error())
			os.Exit(1)
		}
	}

	if config.ApiKey == "" {
		config.ApiKey = os.Getenv("DATABENTO_API_KEY")
		requireValOrExit(config.ApiKey, "missing DataBento API key, use --key or set DATABENTO_API_KEY envvar\n")
	}

	if len(config.Symbols) == 0 {
		fmt.Fprintf(os.Stderr, "requires at least one symbol argument\n")
		os.Exit(1)
	}

	requireValOrExit(config.Dataset, "missing required --dataset")
	requireValOrExit(config.Schema, "missing required --schema")
	requireValOrExit(config.OutFilename, "missing required --out")

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
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

func run(config Config) error {
	// Create output file before connecting
	outWriter, outCloser, err := makeCompressedWriter(config.OutFilename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outCloser()

	// Create and connect LiveClient
	client, err := dbn_live.NewLiveClient(dbn_live.LiveConfig{
		ApiKey:               config.ApiKey,
		Dataset:              config.Dataset,
		SendTsOut:            false,
		VersionUpgradePolicy: dbn.VersionUpgradePolicy_AsIs,
		Verbose:              config.Verbose,
	})
	if err != nil {
		return fmt.Errorf("failed to create LiveClient: %w", err)
	}
	defer client.Stop()

	// Authenticate to server
	if _, err = client.Authenticate(config.ApiKey); err != nil {
		return fmt.Errorf("failed to authenticate LiveClient: %w", err)
	}

	// Subscribe
	subRequest := dbn_live.SubscriptionRequestMsg{
		Schema:   config.Schema,
		StypeIn:  dbn.SType_RawSymbol,
		Symbols:  config.Symbols,
		Start:    config.StartTime,
		Snapshot: false,
	}
	if err = client.Subscribe(subRequest); err != nil {
		return fmt.Errorf("failed to subscribe LiveClient: %w", err)
	}

	// Start session
	if _, err = client.Start(); err != nil {
		return fmt.Errorf("failed to start LiveClient: %w", err)
	}

	// Write metadata to file
	dbnScanner := client.GetDbnScanner()
	if dbnScanner == nil {
		return fmt.Errorf("failed to get DbnScanner from LiveClient")
	}
	metadata, err := dbnScanner.Metadata()
	if err != nil {
		return fmt.Errorf("failed to get metadata from LiveClient: %w", err)
	}
	if err = metadata.Write(outWriter); err != nil {
		return fmt.Errorf("failed to write metadata from LiveClient: %w", err)
	}

	// Follow the DBN stream, writing DBN messages to the file
	for dbnScanner.Next() {
		recordBytes := dbnScanner.GetLastRecord()[:dbnScanner.GetLastSize()]
		_, err := outWriter.Write(recordBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to write record: %s\n", err.Error())
			return err
		}
	}
	if err := dbnScanner.Error(); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "scanner err: %s\n", err.Error())
		return err
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// Compression helpers
// https://gist.github.com/neomantra/691a6028cdf2ac3fc6ec97d00e8ea802

// Returns an io.Writer for the given filename, or os.Stdout if filename is "-".  Also returns a closing function to defer and any error.
// If the filename ends in ".zst" or ".zstd", the writer will zstd-compress the output.
func makeCompressedWriter(filename string) (io.Writer, func(), error) {
	var writer io.Writer
	var closer io.Closer
	fileCloser := func() {
		if closer != nil {
			closer.Close()
		}
	}
	if filename != "-" {
		if file, err := os.Create(filename); err == nil {
			writer, closer = file, file
		} else {
			return nil, nil, err
		}
	} else {
		writer, closer = os.Stdout, nil
	}

	if strings.HasSuffix(filename, ".zst") || strings.HasSuffix(filename, ".zstd") {
		zstdWriter, err := zstd.NewWriter(writer)
		if err != nil {
			fileCloser()
			return nil, nil, err
		}
		zstdCloser := func() {
			zstdWriter.Close()
			fileCloser()
		}
		return zstdWriter, zstdCloser, nil
	} else {
		return writer, fileCloser, nil
	}
}
