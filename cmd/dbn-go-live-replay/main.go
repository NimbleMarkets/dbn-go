// Copyright (c) 2026 Neomantra Corp
//
// dbn-go-live-replay is a local Databento Live gateway that replays DBN files,
// for developing and testing Live clients without a real gateway.
//

package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/NimbleMarkets/dbn-go/internal/version"
	dbn_live "github.com/NimbleMarkets/dbn-go/live"
	dbn_live_replay "github.com/NimbleMarkets/dbn-go/live/replay"
	"github.com/spf13/pflag"
)

func main() {
	var config dbn_live_replay.ServerConfig
	var listenAddr string
	var noAuth bool
	var verbose bool
	var showHelp bool
	var showVersion bool

	pflag.StringVarP(&listenAddr, "listen", "l", "127.0.0.1:13000", "Address to listen on")
	pflag.StringArrayVarP(&config.Files, "file", "f", []string{}, "DBN file to replay, .dbn or .dbn.zst (multiple allowed)")
	pflag.StringVarP(&config.ApiKey, "key", "k", "", "API key clients must use (or set 'DATABENTO_API_KEY' envvar; default: accept any key)")
	pflag.BoolVar(&noAuth, "no-auth", false, "Accept any API key, ignoring 'DATABENTO_API_KEY'")
	pflag.Float64VarP(&config.Speed, "speed", "s", 0, "Replay speed by ts_event: 1 is real time, 2 is double speed; 0 is as fast as possible")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "Verbose logging")
	pflag.BoolVarP(&showHelp, "help", "h", false, "Show help")
	pflag.BoolVar(&showVersion, "version", false, "Show version")
	pflag.Parse()

	// Positional arguments are also replay files
	config.Files = append(config.Files, pflag.Args()...)

	if showVersion {
		fmt.Println(version.String("dbn-go-live-replay"))
		os.Exit(0)
	}

	if showHelp {
		fmt.Fprintf(os.Stdout, "usage: %s [opts] file1.dbn.zst [file2.dbn ...]\n\n", os.Args[0])
		pflag.PrintDefaults()
		os.Exit(0)
	}

	if len(config.Files) == 0 {
		fmt.Fprintf(os.Stderr, "requires at least one DBN file to replay\n")
		os.Exit(1)
	}

	if noAuth {
		config.ApiKey = ""
	} else if config.ApiKey == "" {
		config.ApiKey = os.Getenv(dbn_live.DATABENTO_API_ENV_KEY)
	}

	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	config.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	if err := run(config, listenAddr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}

func run(config dbn_live_replay.ServerConfig, listenAddr string) error {
	server, err := dbn_live_replay.NewServer(config)
	if err != nil {
		return err
	}
	if err := server.Listen(listenAddr); err != nil {
		return err
	}
	auth := "any API key"
	if config.ApiKey != "" {
		auth = "the configured API key"
	}
	config.Logger.Info("replay server listening", "addr", server.Addr().String(), "files", len(config.Files), "speed", config.Speed, "auth", auth)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals
	config.Logger.Info("shutting down")
	return server.Close()
}
