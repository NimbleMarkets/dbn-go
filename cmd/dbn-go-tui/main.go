// Copyright (c) 2025 Neomantra Corp

package main

import (
	"fmt"
	"os"

	dbn_tui "github.com/NimbleMarkets/dbn-go/internal/tui"
	"github.com/spf13/pflag"
)

const defaultMaxActiveDownloads = 4

///////////////////////////////////////////////////////////////////////////////

func main() {
	var config dbn_tui.Config
	var showHelp bool

	pflag.BoolVarP(&showHelp, "help", "h", false, "Show help")
	pflag.StringVarP(&config.DatabentoApiKey, "key", "k", "", "Databento API key (or set 'DATABENTO_API_KEY' envvar)")
	pflag.IntVarP(&config.MaxActiveDownloads, "limit", "l", defaultMaxActiveDownloads, "Limit maximum concurrent downloads")
	pflag.Parse()

	if showHelp {
		fmt.Fprintf(os.Stdout, "usage: %s [options]\n\n", os.Args[0])
		pflag.PrintDefaults()
		os.Exit(0)
	}

	if config.DatabentoApiKey == "" {
		config.DatabentoApiKey = os.Getenv("DATABENTO_API_KEY")
		if config.DatabentoApiKey == "" {
			fmt.Fprintf(os.Stderr, "missing Databento API key, use --key or set DATABENTO_API_KEY envvar\n")
			os.Exit(1)
		}
	}

	if config.MaxActiveDownloads < 0 {
		fmt.Fprintf(os.Stderr, "--limit cannot be negative\n")
		return
	}

	err := dbn_tui.Run(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}
