# `dbn-go-mcp-meta` Databento Metadata MCP Server

`dbn-go-mcp-meta` is a metadata-only [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server for [Databento](https://databento.com) services. It provides discovery and metadata tools for exploring datasets, schemas, pricing, and symbology.

**No tools in this server incur Databento billing.** For data download and caching, use [`dbn-go-mcp-data`](../dbn-go-mcp-data/).

It requires your [Databento API Key](https://databento.com/portal/keys). Since this program is typically executed by a host program (like Claude Desktop), the preferred method is to store the API key in a file and use `--key-file` or the `DATABENTO_API_KEY_FILE` environment variable. It supports both `STDIO` and `SSE` MCP transports.

```
$ dbn-go-mcp-meta --help
usage: dbn-go-mcp-meta -k <api_key> [opts]

  -h, --help              Show help
  -k, --key string        Databento API key (or set 'DATABENTO_API_KEY' envvar)
  -f, --key-file string   File to read Databento API key from (or set 'DATABENTO_API_KEY_FILE' envvar)
  -l, --log-file string   Log file destination (or MCP_LOG_FILE envvar). Default is stderr
  -j, --log-json          Log in JSON (default is plaintext)
  -p, --port string       host:port to listen to SSE connections (default "127.0.0.1:8889")
      --sse               Use SSE Transport (default is STDIO transport)
  -v, --verbose           Verbose logging
```

## Available Tools

### Discovery Tools (no billing)

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

The recommended workflow for an LLM is: `list_datasets` -> `list_schemas` -> `list_fields` -> `get_cost`.

## MCP Host Configuration

### Claude Desktop

Add to your Claude Desktop configuration file (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "dbn": {
      "command": "dbn-go-mcp-meta",
      "args": [
        "--key-file", "/path/to/databento_api_key.txt"
      ]
    }
  }
}
```

### Claude Code

```bash
claude mcp add --transport stdio dbn -- dbn-go-mcp-meta --key-file /path/to/databento_api_key.txt
```

### SSE Mode

For networked deployments, run over HTTP with Server-Sent Events:

```bash
dbn-go-mcp-meta --sse                              # localhost only (default)
dbn-go-mcp-meta --sse --port 0.0.0.0:8889          # listen on all interfaces
```

## See Also

- [`dbn-go-mcp-data`](../dbn-go-mcp-data/) — Full MCP server with data download and DuckDB cache
- [Databento Documentation](https://databento.com/docs)
- [Model Context Protocol](https://modelcontextprotocol.io)
