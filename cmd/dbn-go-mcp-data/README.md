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
      --cache-db string   Path to DuckDB database file (default "~/.dbn-go/cache.duckdb")
  -p, --hostport string   host:port to listen to SSE connections (default "127.0.0.1:8889")
      --sse               Use SSE Transport (default is STDIO transport)
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

Supported schemas for `fetch_range`: `ohlcv-1s`, `ohlcv-1m`, `ohlcv-1h`, `ohlcv-1d`, `trades`, `mbp-1`, `tbbo`, `imbalance`, `statistics`.

### Cache Tools (no billing)

| Tool | Params | Description |
|------|--------|-------------|
| `query_cache` | `sql` | Query cached data with DuckDB SQL. Returns results as CSV. |
| `list_cache` | *(none)* | Lists cached datasets with view names, file counts, and sizes. |
| `clear_cache` | `dataset`?, `schema`? | Removes cached data, optionally filtered by dataset/schema. |

The recommended workflow for an LLM is: `list_datasets` -> `list_schemas` -> `list_fields` -> `get_cost` -> `fetch_range` -> `query_cache`.

## How Caching Works

Data fetched via `fetch_range` is stored as Parquet files under `~/.dbn-go/cache/{dataset}/{schema}/`. An in-memory DuckDB instance creates views backed by these files via `read_parquet()`. The LLM discovers available views using `list_cache` and queries them with standard SQL through `query_cache`.

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

## See Also

- [`dbn-go-mcp-meta`](../dbn-go-mcp-meta/) — Metadata-only MCP server (no billing, no CGO)
- [Databento Documentation](https://databento.com/docs)
- [Model Context Protocol](https://modelcontextprotocol.io)
