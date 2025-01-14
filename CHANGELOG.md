# CHANGELOG

## v0.2.0 (unreleased)

 * Add `MakeCompressedReader` and `MakeCompressedWriter` zstd helpers

## v0.1.0 (2025-01-13)

 * Add `dbn-go-tui` text user interface :fire:
 * Add `LiveClient` now has `GetDBNScanner` and `GetJsonScannr`, depending on encoding.
 * Add `InstrumenDefMsg` and `StatusMsg`
 * Add `OnStatusMsg` and `OnInstrumentDefMsg` to `Visitor` interface
 * Add `JsonScanner.GetLastRecord`
 * Add `ListFiles` may now return `JobExpiredError`.
 * Add `FeedMode` JSON marshalling
 * Add new reference data STypes from DBN v0.20.0
 * Fix `dbn-go-live` JSON mode.
 * Fix `ListJobs`: typo in `states` and `since` is UTC
 * Fix `Compression` error message and convert-from-null
 
## v0.0.13 (2024-06-13)

 * Add `--json` flags to many `dbn-go-hist` commands
 * Add `--sin` and `--sout` to tools
 * Add `resolve` command to `dbn-go-hist`

## v0.0.12 (2024-06-07)

 * Add `GetRange`, `SubmitJobs`, `ListFiles`, and `ListJobs` to Hist API
 * Add some JSON marshallers
 * `dbn-go-hist` now has `get-range`, `submit`, `jobs` and `files` subcommands
 * `dbn-go-hist` now supports `--file` to supply lists of tickers

## v0.0.11 (2024-06-03)

* Expand `dbn-go-hist` tool and add `tests/exercise_dbn-go-hist.sh` example.
* `dbn-go-live` supports multiple schemas and more args
* Add the rest of the Historical Metadata API
* Add custom slog Logger to `LiveClient` and cleanup logging
* Updated for DBN `0.18.0` API changes

## v0.0.10 (2024-05-30)

 * Live intraday replay fixes
 * `dbn-go-live` now support `--start <iso8601>`
 * Add `dbn-go-hist` for some historical queries

## v0.0.9 (2024-05-29)

 * Fix DBN v1 compatibility issues for writing Metadata and reading SymbolMappingMsg

## v0.0.8 (2024-05-28)
 
 * Add initial Live API support
 * Add Mpb1, Mbp10, Mbo, Error, SymbolMapping, System, Statistics
 * Add Dockerfile
 * Minor interface tweaks and bug fixes

## v0.0.7 (2024-05-21)

 * Add more getters for DbnScanner
 * Add TsSymbolMap for historical symbology
 * Add Metadata.Write and YMDToTime
 * Fix broken IsInverseMapping

## v0.0.6 (2024-05-16)

 * Add initial Hist API

## v0.0.1 (2024-04-10)

 * Initial version
