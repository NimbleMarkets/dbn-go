# `dbn-go-mcp-data` Databento MCP Server with DuckDB Cache

`dbn-go-mcp-data` is a full-featured [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server for [Databento](https://databento.com) services. It includes all the metadata discovery tools from [`dbn-go-mcp-meta`](../dbn-go-mcp-meta/), plus data download (`fetch_range`) and a local DuckDB-backed Parquet cache for efficient re-queries.

**CAUTION:** The `fetch_range` tool incurs Databento billing! All other tools are free.

It requires your [Databento API Key](https://databento.com/portal/keys). Since this program is typically executed by a host program (like Claude Desktop), the preferred method is to store the API key in a file and use `--key-file` or the `DATABENTO_API_KEY_FILE` environment variable. It supports both `STDIO` and `SSE` MCP transports.

```
$ dbn-go-mcp-data --help
usage: dbn-go-mcp-data -k <api_key> [opts]

  -h, --help              Show help
  -k, --key string        Databento API key (or set 'DATABENTO_API_KEY' envvar)
  -f, --key-file string   File to read Databento API key from (or set 'DATABENTO_API_KEY_FILE' envvar)
  -l, --log-file string   Log file destination (or MCP_LOG_FILE envvar). Default is stderr
  -j, --log-json          Log in JSON (default is plaintext)
  -c, --max-cost float    Max cost, in dollars, for a query (<=0 is unlimited) (default 1)
      --cache-dir string  Directory for cached data files (default "~/.dbn-go/cache/")
      --read-only         Disable fetch_range tool (no billing possible)
  -p, --hostport string   host:port to listen to SSE connections (default "127.0.0.1:8889")
      --sse               Use SSE Transport (default is STDIO transport)
      --sse-allow-origin strings   Browser origin allowed to use the SSE transport, e.g. http://localhost:6274 (repeatable; default none)
      --sse-token-file string      File holding a bearer token SSE clients must present (or set 'DBN_GO_MCP_SSE_TOKEN' envvar)
  -v, --verbose           Verbose logging
```

## Available Tools

### Discovery Tools (no billing)

These are identical to those in `dbn-go-mcp-meta`:

| Tool | Params | Description |
|------|--------|-------------|
| `list_datasets` | `start`?, `end`? | Lists all available Databento dataset codes, optionally filtered by date range. |
| `list_schemas` | `dataset` | Lists all available data record schemas for a given dataset. |
| `list_fields` | `schema` | Lists all field names and types for a given schema (JSON encoding). |
| `list_publishers` | `dataset`? | Lists all publishers with their publisher_id, dataset, venue, and description. |
| `get_dataset_range` | `dataset` | Returns the available date range (start and end) for a dataset. |
| `get_dataset_condition` | `dataset`, `start`?, `end`? | Returns data quality/availability condition per day. |
| `list_unit_prices` | `dataset` | Lists unit prices in USD per gigabyte for each schema and feed mode. |
| `resolve_symbols` | `dataset`, `symbols`, `start`, `stype_in`?, `stype_out`?, `end`? | Resolves symbols from one symbology type to another. |
| `get_cost` | `dataset`, `schema`, `symbols`, `start`, `end`, `stype_in`? | Returns estimated cost in USD, data size, and record count for a query. |

### Data Tools

| Tool | Params | Description |
|------|--------|-------------|
| `fetch_range` | `dataset`, `schema`, `symbols`, `start`, `end`, `stype_in`?, `stype_out`? | Fetches market data and caches as Parquet. Returns metadata (view name, record count, size). **Incurs billing.** |

Supported schemas for `fetch_range`: `ohlcv-1s`, `ohlcv-1m`, `ohlcv-1h`, `ohlcv-1d`, `trades`, `mbp-1`, `mbp-10`, `tbbo`, `cmbp-1`, `bbo-1s`, `bbo-1m`, `imbalance`, `statistics`, `status`.

### Cache Tools (no billing)

| Tool | Params | Description |
|------|--------|-------------|
| `query_cache` | `sql`, `format`?, `max_rows`? | Query cached data with DuckDB SQL. `format` is `csv` (default), `markdown`, `json_rows`, or `json_columnar`. |
| `describe_cache` | `view`? | Lists cached views with their alias and exact column names/types. Pass `view` (raw or alias name) to describe one view. |
| `list_cache` | *(none)* | Lists cached datasets with view names, aliases, file manifests, and sizes. |
| `clear_cache` | `dataset`?, `schema`? | Removes cached data, optionally filtered by dataset/schema. |

The recommended workflow for an LLM is: `list_datasets` -> `list_schemas` -> `list_fields` -> `get_cost` -> `fetch_range` -> `describe_cache` -> `query_cache`.

### Views and SQL

Each cached `{dataset}/{schema}` is exposed as two DuckDB views:

| View | Example | Timestamps | Notes |
|------|---------|------------|-------|
| Raw | `"XNAS.ITCH/trades"` | `TIMESTAMPTZ_NS` (nanosecond) | Reads the Parquet files directly. The name contains `.` and `/`, so it must be double-quoted. Most DuckDB date functions (`date_trunc`, `extract`, `time_bucket`) reject `TIMESTAMPTZ_NS` without a `::TIMESTAMPTZ` cast. |
| Alias | `xnas_itch_trades` | `TIMESTAMPTZ` (microsecond) | A plain identifier over the raw view with timestamp columns cast, so date functions work directly. Prefer this one unless nanosecond precision matters. |

Timestamps are rendered in UTC (`2026-09-15 08:00:00+00` in CSV/Markdown, RFC 3339 `2026-09-15T08:00:00Z` in the JSON formats). A US equities session, 04:00 to 20:00 New York time, spans 08:00 to 00:00 UTC, so a UTC calendar day splits the session; group or filter by `(ts_event AT TIME ZONE 'America/New_York')::DATE` for trading-day questions. `SUMMARIZE` works on alias views but not on raw views, because DuckDB cannot average nanosecond timestamps.

Prices are doubles (for example `3720.25`). Databento's undefined sentinels are `NULL`: an undefined price (one-sided book, no last trade) and an undefined timestamp (`ts_event` of a `bbo-1m` interval with no trade) are both `NULL`, so aggregates ignore them.

`query_cache` accepts a single read-only SELECT-style statement: `SELECT`, `WITH ... SELECT`, FROM-first (`FROM v SELECT ...`), `VALUES`, `DESCRIBE`, `SUMMARIZE`, and `SHOW TABLES`. It rejects DDL/DML, `PRAGMA`/`SET`, `EXPLAIN`, `PIVOT`, multiple statements, and any file access. Rejections say which category they fall in (forbidden statement, blocked file access, or a binding error such as an unknown column), so the next attempt can be corrected.

Results are capped at 100,000 rows (default 10,000, adjustable with `max_rows`) and 10 MB of text. When a result is cut short, a second content block says whether the row cap or the size cap applied. One `query_cache` call runs at a time; a call that waits more than 5 seconds for an in-flight query returns a "still running, retry shortly" error rather than queueing into the 30-second query timeout.

## Read-Only Mode

Use `--read-only` to disable the `fetch_range` tool entirely. In this mode, the server only exposes metadata discovery tools and cache tools — no new data can be fetched and no Databento billing can be incurred. This is useful for environments where you want to query previously cached data without risk of accidental charges.

```bash
dbn-go-mcp-data --read-only --key-file /path/to/databento_api_key.txt
```

Metadata tools (`list_datasets`, `get_cost`, etc.) still require an API key and remain fully functional. Cache tools (`query_cache`, `list_cache`, `clear_cache`) work as normal against any previously fetched data.

## How Caching Works

Data fetched via `fetch_range` is stored as Parquet files under `~/.dbn-go/cache/{dataset}/{schema}/`. An in-memory DuckDB instance (via [ntduckbox](https://github.com/NimbleTerminal/ntduckbox)) creates a raw view backed by these files via `read_parquet()` and an alias view over it; both are replaced atomically whenever the cache changes. The LLM discovers available views using `list_cache` or `describe_cache` and queries them with standard SQL through `query_cache`.

DuckDB is locked down so `query_cache` can only read files inside the cache directory: extensions and remote filesystems are disabled, file access is restricted via `allowed_directories`, the configuration is locked against changes from user SQL, and every statement is checked by the engine to be a single read-only SELECT-style statement before it runs. There is no host-side SQL keyword filter.

## MCP Host Configuration

### Claude Desktop

Add to your Claude Desktop configuration file (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "dbn": {
      "command": "dbn-go-mcp-data",
      "args": [
        "--key-file", "/path/to/databento_api_key.txt",
        "--max-cost", "1.50"
      ]
    }
  }
}
```

### Claude Code

```bash
claude mcp add --transport stdio dbn -- dbn-go-mcp-data --key-file /path/to/databento_api_key.txt
```

### SSE Mode

For networked deployments, run over HTTP with Server-Sent Events:

```bash
dbn-go-mcp-data --sse                                  # localhost only (default)
dbn-go-mcp-data --sse --hostport 0.0.0.0:8889          # listen on all interfaces
```

The SSE endpoint refuses requests made by web pages, so a site you visit cannot drive the server. A browser-based client must be allowed by its origin:

```bash
dbn-go-mcp-data --sse --sse-allow-origin http://localhost:6274
```

When listening beyond localhost, require a bearer token. Clients send it as `Authorization: Bearer <token>`:

```bash
dbn-go-mcp-data --sse --hostport 0.0.0.0:8889 --sse-token-file /path/to/sse_token.txt
```

The token can also be set with the `DBN_GO_MCP_SSE_TOKEN` environment variable. SSE is plain HTTP; put it behind TLS if it leaves the machine.

## See Also

- [`dbn-go-mcp-meta`](../dbn-go-mcp-meta/) — Metadata-only MCP server (no billing, no CGO)
- [Databento Documentation](https://databento.com/docs)
- [Model Context Protocol](https://modelcontextprotocol.io)
