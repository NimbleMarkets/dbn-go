# CHANGELOG

## v0.6.5 (2025-10-01)

 * Add Parquet support for Statistics message `StatMsg`
 * `dbn-go-hist` now dispatches multiple smaller requests for cost estimates since
    large numbers of tickers would overrun the URI limits.

## v0.6.4 (2025-09-05)

 * Add Parquet export of publishers with `dbn-go-hist publishers --parquet <outfile>`.  Useful for some DuckDB queries.

## v0.6.3 (2025-08-11)

 * Add BBO schema support (#15)
 * Add list of datafile checksums in [`tests/data/checksums.txt`](./tests/data/checksums.txt)

## v0.6.2 (2025-06-03)

 * Fix `StatusMsg` decoding.
 * Added more message structure tests (#6)

## v0.6.1 (2025-06-02)

 * Update Go dependencies, including fixes for latest  `go-mcp`
 * Fix `StatMsg` decoding and add tests for it (#14)

## v0.6.0 (2025-05-27)

  * **BREAKING** Rename `Cbbo` to `Cmbp1` per upstream SDK
    * `RType_Cbbo` becomes `RType_Cmbp1`
    * `Schema_Cbbo` becomes `Schema_Cmbp1` which requests `cmbp-1`
    * `CbboMsg` becomes `Cmbp1Msg`
    * `Visitor.OnCbbo` becomes `Visitor.OnCmbp1`

  * Update from `DBEQ.MINI` to `EQUS.MINI`
  * Added `task test-all`
  * Added new constants and enums up to DBN `0.34.0`:
    * `UNDEF_TIMESTAMP`, `Action_None`, `InstrumentClass_CommoditySpot`, `Condition_Intraday`
    * `MatchAlgorithm_TimeProRata`, `MatchAlgorithm_InstitutionalPrioritization`
    * `StatType_Volatility`, `StatType_Delta`, `StatType_UncrossingPrice`

## v0.5.0 (2025-04-01)

 * Added `dbn-go-mcp` [MCP server](./cmd/dbn-go-mcp/README.md)

## v0.4.1 (2025-03-24)

 * Add `DbnScanner.DecodeSymbolMappingMsg`

## v0.4.0 (2025-02-20)

*“The two most powerful warriors are patience and time” – Leo Tolstoy*

 * Many Hist API calls were passing Dates, rather than full DateTimes, to the backend.   Now date+time is used where appropriate. (#8)
 * Bugfix incorrect end time in `dbn_hist.ListDatasets`.
 * **BREAKING** `dbn-go-hist` time flags (`--start`, `--end`) now uses [ISO 8601](https://www.iso.org/iso-8601-date-and-time-format.html) `YYYY-MM-DD` instead of `YYYYMMDD`.  Times and timezones may now be included: `YYYY-MM-DDTHH:MM:SS±HH:MM`.

  * Unfortunately, `arm64` Docker images are currently disabled due to infrastructure issues (?)

## v0.3.0 (2025-01-22)

 * Removed obsolete `Packaging` parameter from `dbn_hist.SubmitJob` and structs
 * `dbn-go-file`: Do not include `.zst` suffix in Parquet destination file names
 
## v0.2.2 (2025-01-21)

 * `dbn-go-tui` improvements:
   * Add `--limit` argument to control concurrent downloads
   * Add pretty progress bars and table
   * Fix queueing of duplicate files
   * Fix honoring of max active downloads

## v0.2.1 (2025-01-15)

 * Fix `dbn-go-file` was not included in relase.

## v0.2.0 (2025-01-15)

 * Add `dbn-go-file parquet` tool for processing DBN files with commands:
   * `metadata`
   * `json`
   * `spit`
   * `parquet`
 * Fix `Mbp0Msg` structure and add tests.
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
 * Add Mbp1, Mbp10, Mbo, Error, SymbolMapping, System, Statistics
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
