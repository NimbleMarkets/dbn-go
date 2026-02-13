#!/usr/bin/env bash
# smoke-test-mcp.sh — launch each MCP server and verify the JSON-RPC handshake.
# Usage:
#   ./scripts/smoke-test-mcp.sh              # test both servers
#   ./scripts/smoke-test-mcp.sh meta         # test only dbn-go-mcp-meta
#   ./scripts/smoke-test-mcp.sh data         # test only dbn-go-mcp-data
#   ./scripts/smoke-test-mcp.sh -q           # quiet mode (no JSON output)
#   ./scripts/smoke-test-mcp.sh -q data      # quiet + target
#
# Requires DATABENTO_API_KEY or DATABENTO_API_KEY_FILE to be set.

set -euo pipefail

TIMEOUT=10  # seconds to wait for server responses
PASS=0
FAIL=0
QUIET=false

# --- helpers ----------------------------------------------------------------

die()  { echo "FATAL: $*" >&2; exit 1; }
info() { echo "--- $*"; }
ok()   { echo "  PASS: $*"; PASS=$((PASS + 1)); }
fail() { echo "  FAIL: $*" >&2; FAIL=$((FAIL + 1)); }

resolve_key() {
  if [[ -n "${DATABENTO_API_KEY:-}" ]]; then
    return
  fi
  if [[ -n "${DATABENTO_API_KEY_FILE:-}" ]]; then
    DATABENTO_API_KEY="$(cat "$DATABENTO_API_KEY_FILE" | tr -d '[:space:]')"
    export DATABENTO_API_KEY
    return
  fi
  die "Set DATABENTO_API_KEY or DATABENTO_API_KEY_FILE"
}

# send_mcp <binary> <label> [extra_args...]
#   Sends initialize → initialized → tools/list over stdio, captures stdout+stderr.
#   Checks that tools/list returned a result.
send_mcp() {
  local bin="$1"; shift
  local label="$1"; shift
  local extra_args=("$@")

  local tmpout; tmpout=$(mktemp)
  local tmperr; tmperr=$(mktemp)
  trap "rm -f '$tmpout' '$tmperr'" RETURN

  info "Testing $label ($bin)"

  if [[ ! -x "$bin" ]]; then
    fail "$bin not found or not executable"
    return
  fi

  # MCP JSON-RPC messages (line-delimited JSON)
  local init_req='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"smoke-test","version":"1.0"}}}'
  local initialized='{"jsonrpc":"2.0","method":"notifications/initialized"}'
  local tools_list='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'

  # Send all three messages; server exits when stdin closes.
  printf '%s\n%s\n%s\n' "$init_req" "$initialized" "$tools_list" \
    | timeout "$TIMEOUT" "$bin" --key "$DATABENTO_API_KEY" ${extra_args[@]+"${extra_args[@]}"} \
    >"$tmpout" 2>"$tmperr" || true

  # Show stderr/stdout only in verbose mode
  if [[ "$QUIET" == false ]]; then
    if [[ -s "$tmperr" ]]; then
      echo "  stderr:"
      sed 's/^/    /' "$tmperr"
    fi
    if [[ -s "$tmpout" ]]; then
      echo "  stdout:"
      sed 's/^/    /' "$tmpout"
    fi
  fi

  # Check stdout for responses
  if [[ ! -s "$tmpout" ]]; then
    fail "$label: no stdout (server may have crashed on startup)"
    if [[ "$QUIET" == true && -s "$tmperr" ]]; then
      echo "  stderr:"
      sed 's/^/    /' "$tmperr"
    fi
    return
  fi

  # Validate: expect id:1 response (initialize) and id:2 response (tools/list)
  if grep -q '"id":1' "$tmpout" 2>/dev/null; then
    ok "$label: initialize response received"
  else
    fail "$label: no initialize response (id:1)"
  fi

  if grep -q '"id":2' "$tmpout" 2>/dev/null; then
    ok "$label: tools/list response received"
  else
    fail "$label: no tools/list response (id:2)"
  fi

  # Check for tools in the tools/list response
  if grep -q '"tools"' "$tmpout" 2>/dev/null; then
    local count
    count=$(grep -o '"name"' "$tmpout" | wc -l | tr -d ' ')
    ok "$label: $count tools registered"
  fi
}

# --- main -------------------------------------------------------------------

# Parse flags
while [[ $# -gt 0 && "$1" == -* ]]; do
  case "$1" in
    -q|--quiet) QUIET=true; shift ;;
    *) die "Unknown flag: $1" ;;
  esac
done

resolve_key

what="${1:-all}"

# Resolve binary path: check ./bin/, ./, then $PATH
find_bin() {
  local name="$1"
  if [[ -x "./bin/$name" ]]; then echo "./bin/$name"
  elif [[ -x "./$name" ]]; then echo "./$name"
  else echo "$name"  # fall back to $PATH
  fi
}

case "$what" in
  meta)
    send_mcp "$(find_bin dbn-go-mcp-meta)" "dbn-go-mcp-meta"
    ;;
  data)
    send_mcp "$(find_bin dbn-go-mcp-data)" "dbn-go-mcp-data"
    ;;
  all)
    send_mcp "$(find_bin dbn-go-mcp-meta)" "dbn-go-mcp-meta"
    send_mcp "$(find_bin dbn-go-mcp-data)" "dbn-go-mcp-data"
    ;;
  *) die "Unknown target: $what (use meta, data, or all)" ;;
esac

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[[ $FAIL -eq 0 ]] || exit 1
