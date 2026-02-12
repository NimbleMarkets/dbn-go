---
title: "dbn-go-mcp-meta"
weight: 30
bookCollapseSection: true
---

# dbn-go-mcp-meta

`dbn-go-mcp-meta` is a metadata-only [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server for [Databento](https://databento.com) services. It bridges LLMs and Databento's metadata discovery APIs, allowing AI assistants to explore datasets, schemas, and pricing without downloading data.

**No tools in this server incur Databento billing.** For data download and caching, use [`dbn-go-mcp-data`]({{< relref "/dbn-go-mcp-data" >}}).

It requires your [Databento API Key](https://databento.com/portal/keys). Since this program is typically executed by a host program (like Claude Desktop), the preferred method is to store the API key in a file and use `--key-file` or the `DATABENTO_API_KEY_FILE` environment variable.

## Overview

```bash
dbn-go-mcp-meta                          # MCP server on stdio (default)
dbn-go-mcp-meta --sse --port :8889       # MCP server over HTTP (SSE)
```

The MCP server supports two transport modes:
- **stdio** (default): For direct integration with Claude Desktop, Claude Code, and other stdio-based clients
- **HTTP/SSE**: For web-based or networked clients using Server-Sent Events

### Command-Line Flags

```
  -h, --help              Show help
  -k, --key string        Databento API key (or set 'DATABENTO_API_KEY' envvar)
  -f, --key-file string   File to read Databento API key from (or set 'DATABENTO_API_KEY_FILE' envvar)
  -l, --log-file string   Log file destination (or MCP_LOG_FILE envvar). Default is stderr
  -j, --log-json          Log in JSON (default is plaintext)
  -c, --max-cost float    Max cost, in dollars, for a query (<=0 is unlimited) (default 1)
      --port string       host:port to listen to SSE connections
      --sse               Use SSE Transport (default is STDIO transport)
  -v, --verbose           Verbose logging
```

## Available Tools

### Discovery Tools (no billing)

These tools query Databento metadata and do not incur any billing charges.

---

#### list_datasets

Lists all available Databento dataset codes. Optionally filter by date range.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `start` | string | No | Start of date range filter (ISO 8601) |
| `end` | string | No | End of date range filter (ISO 8601) |

---

#### list_schemas

Lists all available data record schemas for a given dataset.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dataset` | string | Yes | Dataset code (e.g. `XNAS.ITCH`, `GLBX.MDP3`) |

---

#### list_fields

Lists all field names and types for a given schema (JSON encoding).

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `schema` | string | Yes | Schema to inspect (e.g. `trades`, `ohlcv-1d`, `mbp-1`) |

---

#### list_publishers

Lists all publishers with their publisher_id, dataset, venue, and description. Optionally filter by dataset.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dataset` | string | No | Dataset code to filter publishers |

---

#### get_dataset_range

Returns the available date range (start and end) for a dataset.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dataset` | string | Yes | Dataset code |

---

#### get_dataset_condition

Returns data quality/availability condition per day (available, degraded, pending, missing, intraday).

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dataset` | string | Yes | Dataset code |
| `start` | string | No | Start of date range filter (ISO 8601) |
| `end` | string | No | End of date range filter (ISO 8601) |

---

#### list_unit_prices

Lists unit prices in USD per gigabyte for each schema and feed mode in a dataset.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dataset` | string | Yes | Dataset code |

---

#### resolve_symbols

Resolves symbols from one symbology type to another for a given dataset and date range. For example, convert a raw symbol like `AAPL` to its `instrument_id`, or resolve a continuous futures contract.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dataset` | string | Yes | Dataset code |
| `symbols` | string | Yes | Comma-separated symbols (e.g. `AAPL,TSLA,QQQ`) |
| `start` | string | Yes | Start of date range (ISO 8601) |
| `stype_in` | string | No | Input symbology type (default: `raw_symbol`). Use `continuous` for futures like `ES.c.0` |
| `stype_out` | string | No | Output symbology type (default: `instrument_id`) |
| `end` | string | No | End of date range (ISO 8601) |

---

#### get_cost

Returns the estimated cost in USD, billable data size in bytes, and record count for a query.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dataset` | string | Yes | Dataset code |
| `schema` | string | Yes | Schema (e.g. `trades`, `ohlcv-1d`) |
| `symbols` | string | Yes | Comma-separated symbols |
| `stype_in` | string | No | Input symbology type (default: `raw_symbol`) |
| `start` | string | Yes | Start of range (ISO 8601) |
| `end` | string | Yes | End of range, exclusive (ISO 8601) |

---

## Recommended Workflow

The recommended workflow for an LLM is:

1. `list_datasets` — discover available datasets
2. `list_schemas` — see which schemas a dataset supports
3. `list_fields` — understand the record structure of a schema
4. `get_dataset_range` / `get_dataset_condition` — verify data availability
5. `get_cost` — estimate the cost of a potential query

## Installation

### Claude Desktop

Add `dbn-go-mcp-meta` to your Claude Desktop configuration file:

{{< tabs "claude-desktop" >}}
{{< tab "macOS" >}}
Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:
{{< /tab >}}
{{< tab "Windows" >}}
Edit `%APPDATA%\Claude\claude_desktop_config.json`:
{{< /tab >}}
{{< /tabs >}}

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

Restart Claude Desktop to load the new MCP server.

---

### Claude Code

```bash
claude mcp add --transport stdio dbn -- dbn-go-mcp-meta --key-file /path/to/databento_api_key.txt
claude mcp list
```

---

### Gemini CLI

```bash
gemini mcp add --transport stdio dbn -- dbn-go-mcp-meta --key-file /path/to/databento_api_key.txt
gemini mcp list
```

---

### GitHub Copilot CLI

{{< tabs "copilot-cli" >}}
{{< tab "macOS/Linux" >}}
Edit `~/.config/github-copilot/mcp.json`:
{{< /tab >}}
{{< tab "Windows" >}}
Edit `%USERPROFILE%\.config\github-copilot\mcp.json`:
{{< /tab >}}
{{< /tabs >}}

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

---

## HTTP/SSE Mode

For networked deployments, run the MCP server over HTTP with Server-Sent Events:

```bash
dbn-go-mcp-meta --sse --port :8889
dbn-go-mcp-meta --sse --port 0.0.0.0:8889   # Listen on all interfaces
```

## See Also

- [`dbn-go-mcp-data`]({{< relref "/dbn-go-mcp-data" >}}) — Full MCP server with data download and DuckDB cache
- [Databento Documentation](https://databento.com/docs)
- [Model Context Protocol](https://modelcontextprotocol.io) — MCP specification
- [GitHub Repository](https://github.com/NimbleMarkets/dbn-go)
