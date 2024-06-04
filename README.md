# dbn-go - Golang bindings to DBN

<p>
    <a href="https://github.com/NimbleMarkets/dbn-go/tags"><img src="https://img.shields.io/github/tag/NimbleMarkets/dbn-go.svg" alt="Latest Tag"></a>
    <a href="https://pkg.go.dev/github.com/NimbleMarkets/dbn-go"><img src="https://pkg.go.dev/badge/github.com/NimbleMarkets/dbn-go.svg" alt="Go Reference"></a>
    <a href="https://github.com/NimbleMarkets/dbn-go/blob/main/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg"  alt="Code Of Conduct"></a>
</p>

**Golang bindings to Databento's DBN**

This repository contains Golang bindings to [Databento's](https://databento.com) file format [Databento Binary Encoding (DBN)](https://databento.com/docs/knowledge-base/new-users/dbn-encoding).

 * [Library Usage](#library-usage)
 * [Reading DBN Files](#reading-dbn-files)
 * [Reading JSON Files](#reading-json-files)
 * [Historical API](#historical-api)
 * [Live API](#live-api)
 * [Tools](#tools)
   * [`dbn-go-hist`](#dbn-go-hist)
   * [`dbn-go-live`](#dbn-go-live)

NOTE: This is a new library and is under active development.  It is not affiliated with Databento.  Please be careful with commands which incur billing.  We are not responsible for any charges you incur.


## Library Usage

To use this library, import the following package, optionally with Databento Historical or Live support:

```go
import (
    dbn "github.com/NimbleMarkets/dbn-go"
    dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
    dbn_live "github.com/NimbleMarkets/dbn-go/live"
)
```

Most `dbn-go` [types](./structs.go) and [enums](./consts.go) parallel DataBento's libraries.  Available messages types are:

  * [`Mbp0Msg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#Mbp0Msg)
  * [`MboMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#MboMsg)
  * [`Mbp1Msg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#Mbp1Msg)
  * [`CbboMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#CbboMsg)
  * [`Mbp10Msg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#Mbp10Msg)
  * [`OhlcvMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#OhlcvMsg)
  * [`ImbalanceMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#ImbalanceMsg)
  * [`SymbolMappingMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#SymbolMappingMsg)
  * [`ErrorMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#ErrorMsg)
  * [`SystemMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#SystemMsg)
  * [`StatMsg`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#StatMsg)


## Reading DBN Files

If you want to read a homogeneous array of DBN records from a file, use the [`dbn.ReadDBNToSlice`](https://pkg.go.dev/github.com/NimbleMarkets/dbn-go#ReadDBNToSlice) generic function. The generic argument dicates which message type to read.

```go
file, _ := os.Open("ohlcv-1s.dbn")
defer file.Close()
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
schemas, err := dbn_hist.ListSchemas(databentoApiKey, "DBEQ.BASIC")
```

The source for `dbn-go-hist` illustrates [using this `dbn_hist` module](https://github.com/NimbleMarkets/dbn-go/blob/main/cmd/dbn-go-hist/main.go#L104).

 * TODO: implement `get_range`, `submit_job`, `list_job`, `list_files`


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

We include some tools to make our lives easier.  They can be built-from-source to the `./bin` folder with the command `task go-build` (install [Taskfile](https://taskfile.dev)).  They are also available as Docker multi-architecture images on [GitHub's Container Registry](https://github.com/NimbleMarkets/dbn-go/pkgs/container/dbn-go) at `ghcr.io/nimblemarkets/dbn-go`.

TODO: goreleaser of binaries

### `dbn-go-hist`

`dbn-go-hist` is a command-line tool to interact with the Databento Historical API.  You can see an example of exercising it in [this script file](./tests/exercise_dbn-go-hist.sh).  It requires your [Databento API Key](https://databento.com/portal/keys) to be set with `--key` or preferably via the `DATABENTO_API_KEY` environment variable.

```
$ ./bin/dbn-go-hist --help
dbn-go-hist queries the DataBento Historical API.

Usage:
  dbn-go-hist [command]

Available Commands:
  billable-size     Queries DataBento Hist for the billable size of a GetRange query
  completion        Generate the autocompletion script for the specified shell
  cost              Queries DataBento Hist for the cost for a GetRange query
  dataset-condition Queries DataBento Hist for condition of a dataset
  dataset-range     Queries DataBento Hist for date range of a dataset
  datasets          Queries DataBento Hist for datasets and prints them
  fields            Queries DataBento Hist for fields of a schema/encoding and prints them
  help              Help about any command
  publishers        Queries DataBento Hist for publishers and prints them
  record-count      Queries DataBento Hist for record count for a GetRange query.  Args are in symbols
  schemas           Queries DataBento Hist for publishers and prints them
  unit-prices       Queries DataBento Hist for unit prices of a dataset

Flags:
  -h, --help         help for dbn-go-hist
  -k, --key string   DataBento API key (or use DATABENT_API_KEY envvar)

Use "dbn-go-hist [command] --help" for more information about a command.
```

Simple invocation:

```sh 
$ ./bin/dbn-go-hist datasets
DBEQ.BASIC
GLBX.MDP3
IFEU.IMPACT
NDEX.IMPACT
OPRA.PILLAR
XNAS.ITCH
```
 TODO: we only support Symbology and Metadata right now.  No jobs or data download.

### `dbn-go-live`

`dbn-go-live` is a command-line tool to subscribe to a Live DataBento stream and write it to a file.   It requires your [Databento API Key](https://databento.com/portal/keys) to be set with `--key` or preferably via the `DATABENTO_API_KEY` environment variable.

*This program incurs billing!*

```
$ ./bin/dbn-go-live --help
usage: ./bin/dbn-go-live -d <dataset> -s <schema> [opts] symbol1 symbol2 ...

  -d, --dataset string       Dataset to subscribe to 
  -e, --encoding string      Encoding of the output (default "dbn")
  -h, --help                 Show help
  -k, --key string           Databento API key (or set 'DATABENTO_API_KEY' envvar)
  -o, --out string           Output filename for DBN stream ('-' for stdout)
  -s, --schema stringArray   Schema to subscribe to (multiple allowed)
  -t, --start string         Start time to request as ISO 8601 format (default: now)
  -i, --stype string         SType of the symbols (default "raw")
  -v, --verbose              Verbose logging
```

Simple invocation:
```
$ ./bin/dbn-go-live -d DBEQ.BASIC -s ohlcv-1h -o foo.dbn -v -t QQQ SPY 
```

Simple Docker invocation:

```
$ docker run -it --rm \
    -e DATABENTO_API_KEY \
    -v ${pwd}/dbn \
    ghcr.io/nimblemarkets/dbn-go:0.0.11 \
    /usr/local/bin/dbn-go-live -d DBEQ.BASIC -s ohlcv-1h -o /dbn/foo.dbn -v -t QQQ SPY 
```


## Open Collaboration

We welcome contributions and feedback.  Please adhere to our [Code of Conduct](./CODE_OF_CONDUCT.md) when engaging our community.

 * [GitHub Issues](https://github.com/NimbleMarkets/dbn-go/issues)
 * [GitHub Pull Requests](https://github.com/NimbleMarkets/dbn-go/pulls)


## License

Released under the [Apache License, version 2.0](https://www.apache.org/licenses/LICENSE-2.0), see [LICENSE.txt](./LICENSE.txt).

Portions adapted from [`databento/dbn`](https://github.com/databento/dbn) [`databendo/databento-rs`](https://github.com/databento/databento-rs) under the same Apache license.

Copyright (c) 2024 [Neomantra Corp](https://www.neomantra.com).   

----
Made with :heart: and :fire: by the team behind [Nimble.Markets](https://nimble.markets).
