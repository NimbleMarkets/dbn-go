# dbn-go - Golang bindings to DBN

<p>
    <a href="https://github.com/NimbleMarkets/dbn-go/tags"><img src="https://img.shields.io/github/tag/NimbleMarkets/dbn-go.svg" alt="Latest Tag"></a>
    <a href="https://pkg.go.dev/github.com/NimbleMarkets/dbn-go"><img src="https://pkg.go.dev/badge/github.com/NimbleMarkets/dbn-go.svg" alt="Go Reference"></a>
    <a href="https://github.com/NimbleMarkets/dbn-go/blob/main/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg"  alt="Code Of Conduct"></a>
</p>

**Golang tooling for Databento's APIs and DBN format**

This repository contains Golang bindings to [Databento's](https://databento.com) file format [Databento Binary Encoding (DBN)](https://databento.com/docs/knowledge-base/new-users/dbn-encoding), [Historical API](#historical-api), and [Live API](#live-api).  It also includes [tools](./cmd/README.md) to interact with these services.

 * [Library Usage](#library-usage)
 * [Reading DBN Files](#reading-dbn-files)
 * [Reading JSON Files](#reading-json-files)
 * [Historical API](#historical-api)
 * [Live API](#live-api)
 * [Tools](#tools)
   * [Installation](#tools-installation)

**NOTE:** This library is **not** affiliated with Databento.  Please be careful with commands which incur billing.  We are not responsible for any charges you incur.


## Library Usage

To use this library, import the following package, optionally with Databento Historical or Live support:

```go
import (
    dbn "github.com/NimbleMarkets/dbn-go"
    dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
    dbn_live "github.com/NimbleMarkets/dbn-go/live"
)
```

Most `dbn-go` [types](./structs.go) and [enums](./consts.go) parallel Databento's libraries.  Available messages types are:

  * [`Mbp0Msg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#Mbp0Msg)
  * [`MboMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#MboMsg)
  * [`Mbp1Msg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#Mbp1Msg)
  * [`Cmbp1Msg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#Cmbp1Msg)
  * [`Mbp10Msg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#Mbp10Msg)
  * [`OhlcvMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#OhlcvMsg)
  * [`ImbalanceMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#ImbalanceMsg)
  * [`SymbolMappingMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#SymbolMappingMsg)
  * [`ErrorMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#ErrorMsg)
  * [`SystemMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#SystemMsg)
  * [`StatMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#StatMsg)
  * [`StatusMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#StatusMsg)
  * [`InstrumentDefMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#InstrumentDefMsg)


## Reading DBN Files

If you want to read a homogeneous array of DBN records from a file, use the [`dbn.ReadDBNToSlice`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#ReadDBNToSlice) generic function. We include an `io.Reader` wrapper, [`dbn.MakeCompressedReader`]https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#MakeCompressedReader), that automatically handles `zstd`-named files.  The generic argument dicates which message type to read.

```go
file, closer, _ := dbn.MakeCompressedReader("ohlcv-1s.dbn.zstd", false)
defer closer.Close()
records, metadata, err := dbn.ReadDBNToSlice[dbn.OhlcvMsg](file)
```

Alternatively, you can use the [`DBNScanner`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#DbnScanner) to read records one-by-one.  Each record can be handled directly, or automatically dispatched to the callback method of a struct that implements the [`dbn.Visitor` interface](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#Visitor).

```go
dbnFile, _ := os.Open("ohlcv-1s.dbn")
defer dbnFile.Close()

dbnScanner := dbn.NewDbnScanner(dbnFile)
metadata, err := dbnScanner.Metadata()
if err != nil {
	return fmt.Errorf("scanner failed to read metadata: %w", err)
}
for dbnScanner.Next() {
    header := dbnScanner.GetLastHeader()
    fmt.Printf("rtype: %s  ts: %d\n", header.RType, header.TsEvent)

    // you could get the raw bytes and size:
    lastRawByteArray := dbnScanner.GetLastRecord()
    lastSize := dbnScanner.GetLastSize()

    // you probably want to use this generic helper to crack the message:
    ohlcv, err := dbn.DbnScannerDecode[dbn.OhlcvMsg](dbnScanner)

    // or if you are handing multiple message types, dispatch a Visitor:
    err = dbnScanner.Visit(visitor)
}
if err := dbnScanner.Error(); err != nil && err != io.EOF {
    return fmt.Errorf("scanner error: %w", err)
}
```


## Reading JSON Files

If you already have DBN-based JSON text files, you can use the generic [`dbn.ReadJsonToSlice`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#ReadJsonToSlice) or [`dbn.JsonScanner`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#JsonScanner) to read them in as `dbn-go` structs.  Similar to the raw DBN, you can handle records manually or use the [`dbn.Visitor` interface](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#Visitor).

```go
jsonFile, _ := os.Open("ohlcv-1s.dbn.json")
defer jsonFile.Close()
records, metadata, err := dbn.ReadJsonToSlice[dbn.OhlcvMsg](jsonFile)
```

```go
jsonFile, _ := os.Open("ohlcv-1s.dbn.json")
defer jsonFile.Close()
jsonScanner := dbn.NewJsonScanner(jsonFile)
for jsonScanner.Next() {
    if err := jsonScanner.Visit(visitor); err != nil {
        return fmt.Errorf("visitor err: %w", err)
    }
}
if err := jsonScanner.Error(); err != nil {
    fmt.Errorf("scanner err: %w", err)
}
```

Many of the `dbn-go` structs are annotated with `json` tags to facilitate JSON serialization and deserialization using `json.Marshal` and `json.Unmarshal`.  That said, `dbn-go` uses [`valyala/fastjson`](https://github.com/valyala/fastjson) and hand-written extraction code.


## Historical API

Support for the [Databento Historical API](https://databento.com/docs/api-reference-historical) is available in the [`/hist`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go/hist) folder.  Every API method is a function that takes an API key and arguments and returns a response struct and error.

```
databentoApiKey := os.Getenv("DATABENTO_API_KEY")
schemas, err := dbn_hist.ListSchemas(databentoApiKey, "EQUS.MINI")
```

The source for `dbn-go-hist` illustrates [using this `dbn_hist` module](https://github.com/NimbleMarkets/dbn-go/blob/main/cmd/dbn-go-hist/main.go#L104).


## Live API

Support for the [Databento Live API](https://databento.com/docs/api-reference-live) is available in the [`/live`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go/live) folder.

The general model is:

  1. Use [`dbn_live.NewLiveClient`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go/live#NewLiveClient) to create a [`LiveClient`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go/live#LiveClient) based on a [`LiveConfig`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go/live#LiveConfig) and connect to a DBN gateway
  2. [`client.Authenticate`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go/live#LiveClient.Authenticate) with the gateway.
  3. [`client.Subscribe`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go/live#LiveClient.Subscribe) to a stream using one or many [`SubscriptionRequestMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go/live#SubscriptionRequestMsg).
  4. Begin the stream with [`client.Start`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go/live#LiveClient.Start).
  5. Read the stream using [`client.GetDbnScanner`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go/live#LiveClient.GetDbnScanner).

The source for `dbn-go-live` illustrates [using this `dbn_live` module](https://github.com/NimbleMarkets/dbn-go/blob/main/cmd/dbn-go-live/main.go#L111).


## Tools

We include [some tools](./cmd/README.md) to make our lives easier. [Installation instructions](./cmd/README.md#installation)

 * [`dbn-go-file`](./cmd/README.md#dbn-go-file): a CLI to process DBN files
 * [`dbn-go-hist`](./cmd/README.md#dbn-go-hist): a CLI to use the Historical API
 * [`dbn-go-live`](./cmd/README.md#dbn-go-live): a simple Live API feed handler
 * [`dbn-go-mcp`](./cmd/README.md#dbn-go-mcp): a LLM Model Context Protocol (MCP) server
 * [`dbn-go-tui`](./cmd/README.md#dbn-go-tui): a TUI for your Databento account

## Open Collaboration

We welcome contributions and feedback.  Please adhere to our [Code of Conduct](./CODE_OF_CONDUCT.md) when engaging our community.

 * [GitHub Issues](https://github.com/NimbleMarkets/dbn-go/issues)
 * [GitHub Pull Requests](https://github.com/NimbleMarkets/dbn-go/pulls)


## License

Released under the [Apache License, version 2.0](https://www.apache.org/licenses/LICENSE-2.0), see [LICENSE.txt](./LICENSE.txt).

Portions adapted from [`databento/dbn`](https://github.com/databento/dbn) [`databendo/databento-rs`](https://github.com/databento/databento-rs) under the same Apache license.

Copyright (c) 2024-2025 [Neomantra Corp](https://www.neomantra.com).   

----
Made with :heart: and :fire: by the team behind [Nimble.Markets](https://nimble.markets).
