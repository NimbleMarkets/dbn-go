---
title: "dbn-go"
type: docs
---

# dbn-go

**dbn-go** is a Go library and set of command-line tools for working with [Databento](https://databento.com)'s APIs and DBN (Databento Binary eNcoding) format.

## Command-Line Tools

* [**dbn-go-hist**]({{< relref "/dbn-go-hist" >}}): CLI for the Databento Historical API
* [**dbn-go-file**]({{< relref "/dbn-go-file" >}}): CLI to process DBN files (parquet, split, metadata)
* **dbn-go-live**: Live API feed handler
* **dbn-go-mcp**: LLM Model Context Protocol (MCP) server for Databento
* **dbn-go-tui**: Terminal UI for your Databento account
* **dbn-go-slurp-docs**: Tool to scrape Databento docs for offline use

## Installation

Binaries are available from the [releases page](https://github.com/NimbleMarkets/dbn-go/releases).

Via Homebrew:

```sh
brew install NimbleMarkets/homebrew-tap/dbn-go
```

Via Docker:

```sh
docker run -e DATABENTO_API_KEY --rm ghcr.io/nimblemarkets/dbn-go:latest /usr/local/bin/dbn-go-hist datasets
```

Build from source (requires [Taskfile](https://taskfile.dev)):

```sh
task go-build
```

## Links

* [GitHub Repository](https://github.com/NimbleMarkets/dbn-go)
* [Go Package Documentation](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go)
* [Databento Documentation](https://databento.com/docs)
