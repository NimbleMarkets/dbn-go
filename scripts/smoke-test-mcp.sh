#!/usr/bin/env bash
# smoke-test-mcp.sh — launch each MCP server and verify the JSON-RPC handshake.
# Usage:
#   ./scripts/smoke-test-mcp.sh              # test both servers (handshake only)
#   ./scripts/smoke-test-mcp.sh meta         # test only dbn-go-mcp-meta
#   ./scripts/smoke-test-mcp.sh data         # test only dbn-go-mcp-data
#   ./scripts/smoke-test-mcp.sh -q           # quiet mode (no JSON output)
#   ./scripts/smoke-test-mcp.sh -a           # also call list_datasets (requires valid API key)
#   ./scripts/smoke-test-mcp.sh -q -a data   # quiet + API test + target
#
# Requires DATABENTO_API_KEY or DATABENTO_API_KEY_FILE to be set.

set -euo pipefail

TIMEOUT=10       # seconds for handshake-only tests
TIMEOUT_API=30   # seconds when testing API calls
PASS=0
FAIL=0
QUIET=false
TEST_API=false

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

# send_mcp <binary> <label> <messages_var> <checks_fn>
#   Sends line-delimited JSON-RPC messages to the server over stdio.
#   Calls checks_fn with the stdout file path to validate responses.
send_mcp() {
  local bin="$1"
  local label="$2"
  local msg_arr_name="$3"
  local checks_fn="$4"

  local tmpout; tmpout=$(mktemp)
  local tmperr; tmperr=$(mktemp)
  trap "rm -f '$tmpout' '$tmperr'" RETURN

  info "Testing $label ($bin)"

  if [[ ! -x "$bin" ]]; then
    fail "$bin not found or not executable"
    return
  fi

  # Indirect array expansion (bash 3 compatible)
  eval 'local msgs=("${'${msg_arr_name}'[@]}")'

  # Send all messages; server exits when stdin closes.
  printf '%s\n' "${msgs[@]}" \
    | timeout "$TIMEOUT" "$bin" --key "$DATABENTO_API_KEY" \
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

  # Run checks
  $checks_fn "$label" "$tmpout"
}

# --- check functions --------------------------------------------------------

check_handshake() {
  local label="$1"
  local tmpout="$2"

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

  if grep -q '"tools"' "$tmpout" 2>/dev/null; then
    local count
    count=$(grep -o '"name"' "$tmpout" | wc -l | tr -d ' ')
    ok "$label: $count tools registered"
  fi
}

check_handshake_and_api() {
  local label="$1"
  local tmpout="$2"

  # Run handshake checks first
  check_handshake "$label" "$tmpout"

  # Check list_datasets response (id:3)
  if grep -q '"id":3' "$tmpout" 2>/dev/null; then
    ok "$label: list_datasets response received"
  else
    fail "$label: no list_datasets response (id:3)"
    return
  fi

  # Verify we got actual dataset names back
  if grep -q 'GLBX.MDP3' "$tmpout" 2>/dev/null; then
    ok "$label: list_datasets returned known datasets"
  else
    fail "$label: list_datasets response missing expected datasets"
  fi
}

# --- message sets -----------------------------------------------------------

HANDSHAKE_MSGS=(
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"smoke-test","version":"1.0"}}}'
  '{"jsonrpc":"2.0","method":"notifications/initialized"}'
  '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
)

API_MSGS=(
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"smoke-test","version":"1.0"}}}'
  '{"jsonrpc":"2.0","method":"notifications/initialized"}'
  '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_datasets","arguments":{}}}'
)

# --- main -------------------------------------------------------------------

# Parse flags
while [[ $# -gt 0 && "$1" == -* ]]; do
  case "$1" in
    -q|--quiet) QUIET=true; shift ;;
    -a|--api)   TEST_API=true; shift ;;
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

# Pick message set, check function, and timeout
if [[ "$TEST_API" == true ]]; then
  msg_var=API_MSGS
  check_fn=check_handshake_and_api
  TIMEOUT=$TIMEOUT_API
else
  msg_var=HANDSHAKE_MSGS
  check_fn=check_handshake
fi

case "$what" in
  meta)
    send_mcp "$(find_bin dbn-go-mcp-meta)" "dbn-go-mcp-meta" "$msg_var" "$check_fn"
    ;;
  data)
    send_mcp "$(find_bin dbn-go-mcp-data)" "dbn-go-mcp-data" "$msg_var" "$check_fn"
    ;;
  all)
    send_mcp "$(find_bin dbn-go-mcp-meta)" "dbn-go-mcp-meta" "$msg_var" "$check_fn"
    send_mcp "$(find_bin dbn-go-mcp-data)" "dbn-go-mcp-data" "$msg_var" "$check_fn"
    ;;
  *) die "Unknown target: $what (use meta, data, or all)" ;;
esac

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[[ $FAIL -eq 0 ]] || exit 1
