---
title: "dbn-go-hist"
weight: 10
bookCollapseSection: true
---

# dbn-go-hist

`dbn-go-hist` is a command-line tool for the [Databento Historical API](https://databento.com/docs/api-reference-historical).

It requires your [Databento API Key](https://databento.com/portal/keys) to be set with `--key` or via the `DATABENTO_API_KEY` environment variable.

**CAUTION:** This program may incur billing!

## Quick Start

```sh
# List available datasets
dbn-go-hist datasets

# List schemas for a dataset
dbn-go-hist schemas -d EQUS.MINI --json

# Get cost estimate
dbn-go-hist cost -d GLBX.MDP3 -s ohlcv-1d -t 2024-01-01 -e 2024-01-31 SPY

# Download data
dbn-go-hist get-range -d GLBX.MDP3 -s ohlcv-1d -t 2024-01-01 -e 2024-01-31 -o data.dbn SPY
```

## Command Reference

See the auto-generated command pages below for full details on each subcommand.
