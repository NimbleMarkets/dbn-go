// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NimbleMarkets/dbn-go/internal/file"
	"github.com/mark3labs/mcp-go/mcp"
)

// testServerWithFixtureView seeds the cache with the two-record ohlcv-1s
// fixture as XNAS.ITCH/ohlcv-1s and opens the DuckDB cache over it.
func testServerWithFixtureView(t *testing.T) *Server {
	t.Helper()
	s := testServer(t)
	parquetDir := filepath.Join(s.cacheDir, "XNAS.ITCH", "ohlcv-1s")
	if err := os.MkdirAll(parquetDir, 0755); err != nil {
		t.Fatal(err)
	}
	parquetPath := filepath.Join(parquetDir, "ESH1__raw_symbol__instrument_id__20201228__20201229.parquet")
	if err := file.WriteDbnFileAsParquet("../../../../tests/data/test_data.ohlcv-1s.dbn", false, parquetPath); err != nil {
		t.Fatalf("failed to write parquet: %v", err)
	}
	if err := s.InitCache(); err != nil {
		t.Fatalf("InitCache failed: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// The engine's own statement check accepts these read-only forms, which an
// LLM reaches for constantly. The host must not reject them with its own
// narrower allowlist.
func TestQueryDuckDB_AllowsReadOnlyStatementForms(t *testing.T) {
	s := testServerWithFixtureView(t)
	queries := map[string]string{
		"describe":   `DESCRIBE "XNAS.ITCH/ohlcv-1s"`,
		"summarize":  `SUMMARIZE SELECT close, volume FROM "XNAS.ITCH/ohlcv-1s"`,
		"show":       `SHOW TABLES`,
		"from_first": `FROM "XNAS.ITCH/ohlcv-1s" SELECT count(*) AS n`,
		"values":     `VALUES (1, 'a'), (2, 'b')`,
		"cte":        `WITH q AS (SELECT instrument_id FROM "XNAS.ITCH/ohlcv-1s") SELECT * FROM q`,
	}
	for name, q := range queries {
		t.Run(name, func(t *testing.T) {
			res, err := s.queryDuckDBResult(context.Background(), q, queryDuckDBOptions{})
			if err != nil {
				t.Fatalf("query %q rejected: %v", q, err)
			}
			if res.Rows == 0 {
				t.Errorf("query %q returned no rows", q)
			}
		})
	}
}

// Rejections are classified so the LLM gets a distinct, actionable message for
// forbidden statements, blocked file access, and plain SQL mistakes.
func TestQueryDuckDB_ClassifiesRejections(t *testing.T) {
	s := testServerWithFixtureView(t)
	secret := writeSecretFile(t, "secret.csv")
	cases := []struct {
		name, sql, want string
	}{
		{"ddl", `CREATE TABLE t AS SELECT 1`, "read-only SELECT-style statement"},
		{"pragma", `PRAGMA show_tables`, "read-only SELECT-style statement"},
		{"multi", `SELECT 1; SELECT 2`, "read-only SELECT-style statement"},
		{"file_read", `SELECT * FROM read_text(` + sqlLiteral(secret) + `)`, "only read the cached views"},
		{"quoted_path", `SELECT * FROM ` + sqlLiteral(secret), "only read the cached views"},
		{"bad_column", `SELECT nope FROM "XNAS.ITCH/ohlcv-1s"`, "describe_cache"},
		{"bad_view", `SELECT * FROM "XNAS.ITCH/nope"`, "describe_cache"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.queryDuckDBResult(context.Background(), tc.sql, queryDuckDBOptions{})
			if err == nil {
				t.Fatalf("query %q should be rejected", tc.sql)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error for %q = %q, want it to mention %q", tc.sql, err.Error(), tc.want)
			}
		})
	}
}

// When the byte cap, not max_rows, stopped the result, the hint must say so
// instead of telling the LLM to raise max_rows.
func TestQueryCacheHandler_SizeTruncationHint(t *testing.T) {
	s := testServerWithDB(t)
	req := mcp.CallToolRequest{}
	req.Params.Name = "query_cache"
	req.Params.Arguments = map[string]any{
		// 200 rows x 100 KB is 20 MB, well over the 10 MB cap but under the row cap.
		"sql":    "SELECT repeat('x', 100000) AS s FROM range(200)",
		"format": "csv",
	}
	res, err := s.queryCacheHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("queryCacheHandler returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected non-error result, got: %#v", res.Content[0])
	}
	if len(res.Content) < 2 {
		t.Fatalf("expected truncation notice, got %d content blocks", len(res.Content))
	}
	notice := res.Content[1].(mcp.TextContent).Text
	if !strings.Contains(notice, "size limit") || strings.Contains(notice, "increase max_rows") {
		t.Errorf("size-truncation notice should mention the size limit and not max_rows: %q", notice)
	}
}

// Alias views give each cached view a plain identifier (no quoting needed) and
// cast nanosecond timestamps to TIMESTAMPTZ so date_trunc/extract/time_bucket
// work without a cast. The raw view keeps full precision.
func TestAliasViewName(t *testing.T) {
	cases := map[string]string{
		"XNAS.ITCH/ohlcv-1d": "xnas_itch_ohlcv_1d",
		"EQUS.MINI/bbo-1m":   "equs_mini_bbo_1m",
		"GLBX.MDP3/mbp-10":   "glbx_mdp3_mbp_10",
	}
	for in, want := range cases {
		if got := aliasViewName(in); got != want {
			t.Errorf("aliasViewName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRefreshViews_CreatesAliasViewWithTimestampCast(t *testing.T) {
	s := testServerWithFixtureView(t)

	raw, err := s.queryDuckDB(`SELECT typeof(ts_event) AS t FROM "XNAS.ITCH/ohlcv-1s" LIMIT 1`, queryDuckDBOptions{})
	if err != nil {
		t.Fatalf("raw view query failed: %v", err)
	}
	if raw != "t\nTIMESTAMPTZ_NS\n" {
		t.Errorf("raw view ts_event type: got %q, want TIMESTAMPTZ_NS", raw)
	}

	alias, err := s.queryDuckDB(`SELECT typeof(ts_event) AS t, count(*) AS n FROM xnas_itch_ohlcv_1s GROUP BY 1`, queryDuckDBOptions{})
	if err != nil {
		t.Fatalf("alias view query failed: %v", err)
	}
	if alias != "t,n\nTIMESTAMP WITH TIME ZONE,2\n" {
		t.Errorf("alias view: got %q, want TIMESTAMPTZ with 2 rows", alias)
	}

	// The whole point: date functions work on the alias without a cast.
	if _, err := s.queryDuckDB(`SELECT date_trunc('day', ts_event), extract(hour FROM ts_event) FROM xnas_itch_ohlcv_1s`, queryDuckDBOptions{}); err != nil {
		t.Errorf("date functions on alias view failed: %v", err)
	}
}

func TestListCacheEntries_IncludesAliasView(t *testing.T) {
	s := testServerWithFixtureView(t)
	entries := s.listCacheEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].AliasView != "xnas_itch_ohlcv_1s" {
		t.Errorf("AliasView = %q, want xnas_itch_ohlcv_1s", entries[0].AliasView)
	}
}

func TestClearCache_DropsAliasView(t *testing.T) {
	s := testServerWithFixtureView(t)
	s.mu.Lock()
	removed := s.clearCache("XNAS.ITCH", "ohlcv-1s")
	s.mu.Unlock()
	if removed != 1 {
		t.Fatalf("clearCache removed %d files, want 1", removed)
	}
	if _, err := s.queryDuckDB(`SELECT count(*) FROM xnas_itch_ohlcv_1s`, queryDuckDBOptions{}); err == nil {
		t.Error("alias view should be gone after clear_cache")
	}
	if _, err := s.queryDuckDB(`SELECT count(*) FROM "XNAS.ITCH/ohlcv-1s"`, queryDuckDBOptions{}); err == nil {
		t.Error("raw view should be gone after clear_cache")
	}
}
