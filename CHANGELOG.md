# CHANGELOG

## v0.0.11 (unreleased)

 * Add custom slog Logger to `LiveClient` and cleanup logging

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
