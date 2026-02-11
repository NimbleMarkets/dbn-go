---
title: "dbn-go-file"
weight: 20
bookCollapseSection: true
---

# dbn-go-file

`dbn-go-file` is a command-line tool for processing [Databento DBN](https://databento.com/docs/knowledge-base/new-users/dbn-encoding) files.

## Quick Start

```sh
# Print file metadata as JSON
dbn-go-file metadata data.ohlcv-1s.dbn

# Convert DBN to Parquet
dbn-go-file parquet data.ohlcv-1s.dbn

# Print records as JSON
dbn-go-file json data.ohlcv-1s.dbn

# Split download folders into organized structure
dbn-go-file split --dest output/ data/*.dbn
```

## Command Reference

See the auto-generated command pages below for full details on each subcommand.
