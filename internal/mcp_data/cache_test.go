// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	cacheDir := t.TempDir()
	return &Server{
		cacheDir: cacheDir,
	}
}

func TestNormalizeDateForFilename(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"2024-01-15", "20240115"},
		{"2024-01-15T09:30:00Z", "20240115T093000Z"},
		{"2024-12-31", "20241231"},
		{"no-dashes", "nodashes"},
	}
	for _, tt := range tests {
		got := normalizeDateForFilename(tt.input)
		if got != tt.want {
			t.Errorf("normalizeDateForFilename(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCacheParquetPath_Simple(t *testing.T) {
	s := testServer(t)
	path := s.cacheParquetPath("XNAS.ITCH", "ohlcv-1d", "AAPL", "raw_symbol", "instrument_id", "2024-01-15", "2024-01-16")

	// Should be under cacheDir/dataset/schema/
	if !strings.HasPrefix(path, s.cacheDir) {
		t.Errorf("path %q does not start with cacheDir %q", path, s.cacheDir)
	}

	dir := filepath.Dir(path)
	wantDir := filepath.Join(s.cacheDir, "XNAS.ITCH", "ohlcv-1d")
	if dir != wantDir {
		t.Errorf("dir = %q, want %q", dir, wantDir)
	}

	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".parquet") {
		t.Errorf("base %q does not end with .parquet", base)
	}

	// Should contain the symbol, stype_in, stype_out, and normalized dates
	if !strings.Contains(base, "AAPL") {
		t.Errorf("base %q does not contain AAPL", base)
	}
	if !strings.Contains(base, "raw_symbol") {
		t.Errorf("base %q does not contain raw_symbol", base)
	}
	if !strings.Contains(base, "instrument_id") {
		t.Errorf("base %q does not contain instrument_id", base)
	}
	if !strings.Contains(base, "20240115") {
		t.Errorf("base %q does not contain 20240115", base)
	}
	if !strings.Contains(base, "20240116") {
		t.Errorf("base %q does not contain 20240116", base)
	}
}

func TestCacheParquetPath_MultiSymbol(t *testing.T) {
	s := testServer(t)
	path := s.cacheParquetPath("XNAS.ITCH", "trades", "AAPL,MSFT,TSLA", "raw_symbol", "instrument_id", "2024-01-01", "2024-01-31")
	base := filepath.Base(path)

	if !strings.Contains(base, "AAPL,MSFT,TSLA") {
		t.Errorf("base %q does not contain all symbols", base)
	}
}

func TestCacheParquetPath_SymbolOrderNormalized(t *testing.T) {
	s := testServer(t)
	// The handler sorts symbols before calling cacheParquetPath, so same sorted input = same path
	path1 := s.cacheParquetPath("XNAS.ITCH", "trades", "AAPL,MSFT,TSLA", "raw_symbol", "instrument_id", "2024-01-01", "2024-01-31")
	path2 := s.cacheParquetPath("XNAS.ITCH", "trades", "AAPL,MSFT,TSLA", "raw_symbol", "instrument_id", "2024-01-01", "2024-01-31")
	if path1 != path2 {
		t.Errorf("same inputs produced different paths:\n  %q\n  %q", path1, path2)
	}
}

func TestCacheParquetPath_DifferentStypeInProducesDifferentPath(t *testing.T) {
	s := testServer(t)
	path1 := s.cacheParquetPath("GLBX.MDP3", "trades", "ES.c.0", "raw_symbol", "instrument_id", "2024-01-01", "2024-01-31")
	path2 := s.cacheParquetPath("GLBX.MDP3", "trades", "ES.c.0", "continuous", "instrument_id", "2024-01-01", "2024-01-31")
	if path1 == path2 {
		t.Error("different stype_in should produce different paths")
	}
}

func TestCacheParquetPath_DifferentStypeOutProducesDifferentPath(t *testing.T) {
	s := testServer(t)
	path1 := s.cacheParquetPath("XNAS.ITCH", "trades", "AAPL", "raw_symbol", "instrument_id", "2024-01-01", "2024-01-31")
	path2 := s.cacheParquetPath("XNAS.ITCH", "trades", "AAPL", "raw_symbol", "raw_symbol", "2024-01-01", "2024-01-31")
	if path1 == path2 {
		t.Error("different stype_out should produce different paths")
	}
}

func TestCacheParquetPath_Truncation(t *testing.T) {
	s := testServer(t)

	// Build 60 symbols to exceed the 200 char limit
	var symbols []string
	for i := 0; i < 60; i++ {
		symbols = append(symbols, "SYMB"+strings.Repeat("X", 3))
	}
	symbolsStr := strings.Join(symbols, ",")

	path := s.cacheParquetPath("XNAS.ITCH", "ohlcv-1d", symbolsStr, "raw_symbol", "instrument_id", "2024-01-01", "2024-12-31")
	base := filepath.Base(path)

	if len(base) > 200+len(".parquet") {
		t.Errorf("truncated filename too long: %d chars: %q", len(base), base)
	}
	if !strings.Contains(base, "more") {
		t.Errorf("truncated filename should contain 'more': %q", base)
	}
	if !strings.HasSuffix(base, ".parquet") {
		t.Errorf("truncated filename should end with .parquet: %q", base)
	}
}

func TestCacheParquetPath_TruncationDeterministic(t *testing.T) {
	s := testServer(t)

	var symbols []string
	for i := 0; i < 60; i++ {
		symbols = append(symbols, "SYMB"+strings.Repeat("X", 3))
	}
	symbolsStr := strings.Join(symbols, ",")

	path1 := s.cacheParquetPath("XNAS.ITCH", "ohlcv-1d", symbolsStr, "raw_symbol", "instrument_id", "2024-01-01", "2024-12-31")
	path2 := s.cacheParquetPath("XNAS.ITCH", "ohlcv-1d", symbolsStr, "raw_symbol", "instrument_id", "2024-01-01", "2024-12-31")
	if path1 != path2 {
		t.Error("truncated paths should be deterministic")
	}
}

func TestManifestPath(t *testing.T) {
	got := manifestPath("/cache/XNAS.ITCH/ohlcv-1d/AAPL__raw_symbol__instrument_id__20240101__20240131.parquet")
	want := "/cache/XNAS.ITCH/ohlcv-1d/AAPL__raw_symbol__instrument_id__20240101__20240131.json"
	if got != want {
		t.Errorf("manifestPath() = %q, want %q", got, want)
	}
}

func TestWriteReadManifest(t *testing.T) {
	dir := t.TempDir()
	parquetPath := filepath.Join(dir, "test.parquet")

	original := CacheManifest{
		Symbols:     []string{"AAPL", "TSLA"},
		StypeIn:     "raw_symbol",
		StypeOut:    "instrument_id",
		Start:       "2024-01-01",
		End:         "2024-01-31",
		RecordCount: 1234,
		Cost:        0.05,
	}

	if err := writeManifest(parquetPath, original); err != nil {
		t.Fatalf("writeManifest failed: %v", err)
	}

	// Verify file exists
	jsonPath := manifestPath(parquetPath)
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("manifest file does not exist: %v", err)
	}

	// Verify it's valid JSON
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}

	// Read it back
	got := readManifest(parquetPath)
	if got == nil {
		t.Fatal("readManifest returned nil")
	}

	if len(got.Symbols) != 2 || got.Symbols[0] != "AAPL" || got.Symbols[1] != "TSLA" {
		t.Errorf("Symbols = %v, want [AAPL TSLA]", got.Symbols)
	}
	if got.StypeIn != "raw_symbol" {
		t.Errorf("StypeIn = %q, want raw_symbol", got.StypeIn)
	}
	if got.StypeOut != "instrument_id" {
		t.Errorf("StypeOut = %q, want instrument_id", got.StypeOut)
	}
	if got.Start != "2024-01-01" {
		t.Errorf("Start = %q, want 2024-01-01", got.Start)
	}
	if got.End != "2024-01-31" {
		t.Errorf("End = %q, want 2024-01-31", got.End)
	}
	if got.RecordCount != 1234 {
		t.Errorf("RecordCount = %d, want 1234", got.RecordCount)
	}
	if got.Cost != 0.05 {
		t.Errorf("Cost = %f, want 0.05", got.Cost)
	}
	if got.FetchedAt == "" {
		t.Error("FetchedAt should be set by writeManifest")
	}
}

func TestReadManifest_Missing(t *testing.T) {
	got := readManifest("/nonexistent/path/test.parquet")
	if got != nil {
		t.Error("readManifest should return nil for missing file")
	}
}

func TestReadManifest_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "test.json")
	os.WriteFile(jsonPath, []byte("not json"), 0644)

	got := readManifest(filepath.Join(dir, "test.parquet"))
	if got != nil {
		t.Error("readManifest should return nil for invalid JSON")
	}
}

func TestRemoveCacheFiles(t *testing.T) {
	dir := t.TempDir()
	boundDir := dir

	// Create a subdirectory structure
	subDir := filepath.Join(dir, "XNAS.ITCH", "ohlcv-1d")
	os.MkdirAll(subDir, 0755)

	// Create parquet and json files
	os.WriteFile(filepath.Join(subDir, "AAPL__raw_symbol__instrument_id__20240101__20240131.parquet"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(subDir, "AAPL__raw_symbol__instrument_id__20240101__20240131.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(subDir, "orphan.json"), []byte("{}"), 0644)

	removed := removeCacheFiles(subDir, boundDir)

	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	// All files should be gone
	entries, _ := os.ReadDir(subDir)
	// subDir should have been cleaned up (empty dir removal)
	if _, err := os.Stat(subDir); err == nil {
		if len(entries) > 0 {
			t.Errorf("directory should be empty or removed, has %d entries", len(entries))
		}
	}
}

func TestRemoveCacheFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "XNAS.ITCH", "ohlcv-1d")
	os.MkdirAll(subDir, 0755)

	removed := removeCacheFiles(subDir, dir)
	if removed != 0 {
		t.Errorf("removed = %d, want 0", removed)
	}

	// Empty dirs should be cleaned up
	if _, err := os.Stat(subDir); err == nil {
		t.Error("empty subDir should have been removed")
	}
}

func TestRemoveCacheFiles_PreservesBoundDir(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "sub")
	os.MkdirAll(subDir, 0755)

	removeCacheFiles(subDir, dir)

	// boundDir should still exist
	if _, err := os.Stat(dir); err != nil {
		t.Error("boundDir should not be removed")
	}
}

func TestListCacheEntries_Empty(t *testing.T) {
	s := testServer(t)
	entries := s.listCacheEntries()
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestListCacheEntries_WithManifest(t *testing.T) {
	s := testServer(t)

	// Create directory structure
	subDir := filepath.Join(s.cacheDir, "XNAS.ITCH", "ohlcv-1d")
	os.MkdirAll(subDir, 0755)

	// Write a parquet file (just needs to exist)
	pqPath := filepath.Join(subDir, "AAPL__raw_symbol__instrument_id__20240101__20240131.parquet")
	os.WriteFile(pqPath, []byte("fake parquet data"), 0644)

	// Write a manifest
	writeManifest(pqPath, CacheManifest{
		Symbols:     []string{"AAPL"},
		StypeIn:     "raw_symbol",
		StypeOut:    "instrument_id",
		Start:       "2024-01-01",
		End:         "2024-01-31",
		RecordCount: 100,
		Cost:        0.01,
	})

	entries := s.listCacheEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	e := entries[0]
	if e.ViewName != "XNAS.ITCH/ohlcv-1d" {
		t.Errorf("ViewName = %q, want XNAS.ITCH/ohlcv-1d", e.ViewName)
	}
	if e.Dataset != "XNAS.ITCH" {
		t.Errorf("Dataset = %q, want XNAS.ITCH", e.Dataset)
	}
	if e.Schema != "ohlcv-1d" {
		t.Errorf("Schema = %q, want ohlcv-1d", e.Schema)
	}
	if len(e.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(e.Files))
	}

	f := e.Files[0]
	if f.Start != "2024-01-01" {
		t.Errorf("Start = %q, want 2024-01-01", f.Start)
	}
	if f.StypeOut != "instrument_id" {
		t.Errorf("StypeOut = %q, want instrument_id", f.StypeOut)
	}
	if f.RecordCount != 100 {
		t.Errorf("RecordCount = %d, want 100", f.RecordCount)
	}
}

func TestListCacheEntries_WithoutManifest(t *testing.T) {
	s := testServer(t)

	// Create directory structure with parquet but no manifest (legacy file)
	subDir := filepath.Join(s.cacheDir, "XNAS.ITCH", "trades")
	os.MkdirAll(subDir, 0755)
	pqPath := filepath.Join(subDir, "abcd1234.parquet")
	os.WriteFile(pqPath, []byte("old data"), 0644)

	entries := s.listCacheEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	f := entries[0].Files[0]
	if f.Filename != "abcd1234.parquet" {
		t.Errorf("Filename = %q, want abcd1234.parquet", f.Filename)
	}
	// Manifest fields should be empty/zero
	if len(f.Symbols) != 0 {
		t.Errorf("Symbols should be empty for legacy file, got %v", f.Symbols)
	}
	if f.Start != "" {
		t.Errorf("Start should be empty for legacy file, got %q", f.Start)
	}
}

func TestClearCache(t *testing.T) {
	s := testServer(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	_ = logger

	// Create files in two dataset/schema combos
	dir1 := filepath.Join(s.cacheDir, "XNAS.ITCH", "ohlcv-1d")
	dir2 := filepath.Join(s.cacheDir, "GLBX.MDP3", "trades")
	os.MkdirAll(dir1, 0755)
	os.MkdirAll(dir2, 0755)

	pq1 := filepath.Join(dir1, "test1.parquet")
	pq2 := filepath.Join(dir2, "test2.parquet")
	os.WriteFile(pq1, []byte("data"), 0644)
	os.WriteFile(pq2, []byte("data"), 0644)
	writeManifest(pq1, CacheManifest{Symbols: []string{"AAPL"}})
	writeManifest(pq2, CacheManifest{Symbols: []string{"ES"}})

	// Clear only XNAS.ITCH/ohlcv-1d
	removed := s.clearCache("XNAS.ITCH", "ohlcv-1d")
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	// dir1 files should be gone
	if _, err := os.Stat(pq1); err == nil {
		t.Error("pq1 should have been removed")
	}
	if _, err := os.Stat(manifestPath(pq1)); err == nil {
		t.Error("pq1 manifest should have been removed")
	}

	// dir2 files should still exist
	if _, err := os.Stat(pq2); err != nil {
		t.Error("pq2 should still exist")
	}
}

func TestClearCache_ByDataset(t *testing.T) {
	s := testServer(t)

	dir1 := filepath.Join(s.cacheDir, "XNAS.ITCH", "ohlcv-1d")
	dir2 := filepath.Join(s.cacheDir, "XNAS.ITCH", "trades")
	os.MkdirAll(dir1, 0755)
	os.MkdirAll(dir2, 0755)

	os.WriteFile(filepath.Join(dir1, "a.parquet"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(dir2, "b.parquet"), []byte("data"), 0644)

	removed := s.clearCache("XNAS.ITCH", "")
	if removed != 2 {
		t.Errorf("removed = %d, want 2", removed)
	}
}

func TestClearCache_BySchema(t *testing.T) {
	s := testServer(t)

	dir1 := filepath.Join(s.cacheDir, "XNAS.ITCH", "trades")
	dir2 := filepath.Join(s.cacheDir, "GLBX.MDP3", "trades")
	dir3 := filepath.Join(s.cacheDir, "XNAS.ITCH", "ohlcv-1d")
	os.MkdirAll(dir1, 0755)
	os.MkdirAll(dir2, 0755)
	os.MkdirAll(dir3, 0755)

	pq1 := filepath.Join(dir1, "a.parquet")
	pq2 := filepath.Join(dir2, "b.parquet")
	pq3 := filepath.Join(dir3, "c.parquet")
	os.WriteFile(pq1, []byte("data"), 0644)
	os.WriteFile(pq2, []byte("data"), 0644)
	os.WriteFile(pq3, []byte("data"), 0644)

	removed := s.clearCache("", "trades")
	if removed != 2 {
		t.Fatalf("removed = %d, want 2", removed)
	}

	if _, err := os.Stat(pq1); err == nil {
		t.Error("XNAS.ITCH/trades should have been removed")
	}
	if _, err := os.Stat(pq2); err == nil {
		t.Error("GLBX.MDP3/trades should have been removed")
	}
	if _, err := os.Stat(pq3); err != nil {
		t.Error("XNAS.ITCH/ohlcv-1d should still exist")
	}
}

func TestClearCache_All(t *testing.T) {
	s := testServer(t)

	dir1 := filepath.Join(s.cacheDir, "XNAS.ITCH", "ohlcv-1d")
	dir2 := filepath.Join(s.cacheDir, "GLBX.MDP3", "trades")
	os.MkdirAll(dir1, 0755)
	os.MkdirAll(dir2, 0755)

	os.WriteFile(filepath.Join(dir1, "a.parquet"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(dir1, "a.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir2, "b.parquet"), []byte("data"), 0644)

	removed := s.clearCache("", "")
	if removed != 2 {
		t.Errorf("removed = %d, want 2", removed)
	}

	// JSON sidecar should also be gone
	if _, err := os.Stat(filepath.Join(dir1, "a.json")); err == nil {
		t.Error("sidecar json should have been removed")
	}
}

func TestFetchRangeOnce_DeduplicatesConcurrentCalls(t *testing.T) {
	s := testServer(t)

	start := make(chan struct{})
	entered := make(chan struct{}, 1)
	release := make(chan struct{})

	var calls atomic.Int32
	var wg sync.WaitGroup
	results := make([]string, 8)
	errs := make([]error, 8)

	for i := 0; i < len(results); i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			results[idx], errs[idx] = s.fetchRangeOnce("same-cache-key", func() (string, error) {
				if calls.Add(1) == 1 {
					entered <- struct{}{}
				}
				<-release
				return "ok", nil
			})
		}(i)
	}

	close(start)
	<-entered
	time.Sleep(25 * time.Millisecond)
	close(release)
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("fetchRangeOnce invoked callback %d times, want 1", got)
	}
	for i := range results {
		if errs[i] != nil {
			t.Fatalf("call %d returned error: %v", i, errs[i])
		}
		if results[i] != "ok" {
			t.Fatalf("call %d result = %q, want ok", i, results[i])
		}
	}
}

func TestValidateQueryCacheSQL_AllowsReadOnlySelects(t *testing.T) {
	queries := []string{
		`SELECT * FROM "XNAS.ITCH/trades" LIMIT 10`,
		`WITH q AS (SELECT instrument_id FROM "XNAS.ITCH/trades") SELECT * FROM q`,
		`-- comment
SELECT * FROM "GLBX.MDP3/ohlcv-1d"`,
	}
	for _, q := range queries {
		if err := validateQueryCacheSQL(q); err != nil {
			t.Fatalf("validateQueryCacheSQL unexpectedly rejected %q: %v", q, err)
		}
	}
}

func testServerWithDB(t *testing.T) *Server {
	t.Helper()
	s := testServer(t)
	if err := s.InitCache(); err != nil {
		t.Fatalf("InitCache failed: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestQueryDuckDB_CSV_Default(t *testing.T) {
	s := testServerWithDB(t)
	result, err := s.queryDuckDB("SELECT 1 AS a, 'hello' AS b", queryDuckDBOptions{})
	if err != nil {
		t.Fatalf("queryDuckDB failed: %v", err)
	}
	if !strings.Contains(result, "a,b") {
		t.Errorf("expected CSV header, got: %s", result)
	}
	if !strings.Contains(result, "1,hello") {
		t.Errorf("expected CSV data row, got: %s", result)
	}
}

func TestQueryDuckDB_CSV_Explicit(t *testing.T) {
	s := testServerWithDB(t)
	result, err := s.queryDuckDB("SELECT 42 AS val", queryDuckDBOptions{Format: "csv"})
	if err != nil {
		t.Fatalf("queryDuckDB failed: %v", err)
	}
	if !strings.Contains(result, "val") || !strings.Contains(result, "42") {
		t.Errorf("expected CSV output, got: %s", result)
	}
}

func TestValidateQueryCacheSQL_BlocksExternalReadsAndDDL(t *testing.T) {
	blocked := []string{
		`SELECT * FROM read_csv_auto('/etc/hosts')`,
		`SELECT * FROM read_parquet('/tmp/data.parquet')`,
		`SELECT * FROM parquet_scan('/tmp/data.parquet')`,
		`SELECT * FROM '/etc/hosts'`,
		`PRAGMA show_tables`,
		`SELECT 1; SELECT 2`,
		`CREATE TABLE t AS SELECT 1`,
	}
	for _, q := range blocked {
		if err := validateQueryCacheSQL(q); err == nil {
			t.Fatalf("validateQueryCacheSQL should reject %q", q)
		}
	}
}

func TestQueryDuckDB_JSONRows(t *testing.T) {
	s := testServerWithDB(t)
	result, err := s.queryDuckDB(
		"SELECT 1 AS id, 'alice' AS name UNION ALL SELECT 2, 'bob'",
		queryDuckDBOptions{Format: "json_rows"},
	)
	if err != nil {
		t.Fatalf("queryDuckDB failed: %v", err)
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(result), &rows); err != nil {
		t.Fatalf("failed to parse JSON: %v\nraw: %s", err, result)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["name"] != "alice" {
		t.Errorf("row 0 name = %v, want alice", rows[0]["name"])
	}
}

func TestQueryDuckDB_JSONRows_Empty(t *testing.T) {
	s := testServerWithDB(t)
	result, err := s.queryDuckDB(
		"SELECT 1 AS id WHERE false",
		queryDuckDBOptions{Format: "json_rows"},
	)
	if err != nil {
		t.Fatalf("queryDuckDB failed: %v", err)
	}
	if result != "[]" {
		t.Errorf("expected empty array, got: %s", result)
	}
}

func TestQueryDuckDB_JSONRows_NullValues(t *testing.T) {
	s := testServerWithDB(t)
	result, err := s.queryDuckDB(
		"SELECT 1 AS id, NULL AS val",
		queryDuckDBOptions{Format: "json_rows"},
	)
	if err != nil {
		t.Fatalf("queryDuckDB failed: %v", err)
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(result), &rows); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if rows[0]["val"] != nil {
		t.Errorf("expected null val, got: %v", rows[0]["val"])
	}
}

func TestQueryDuckDB_JSONColumnar(t *testing.T) {
	s := testServerWithDB(t)
	result, err := s.queryDuckDB(
		"SELECT 1 AS id, 'alice' AS name UNION ALL SELECT 2, 'bob'",
		queryDuckDBOptions{Format: "json_columnar"},
	)
	if err != nil {
		t.Fatalf("queryDuckDB failed: %v", err)
	}

	var cols map[string][]any
	if err := json.Unmarshal([]byte(result), &cols); err != nil {
		t.Fatalf("failed to parse JSON: %v\nraw: %s", err, result)
	}
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	if len(cols["id"]) != 2 {
		t.Fatalf("expected 2 values in 'id' column, got %d", len(cols["id"]))
	}
	if len(cols["name"]) != 2 {
		t.Fatalf("expected 2 values in 'name' column, got %d", len(cols["name"]))
	}
	if cols["name"][0] != "alice" || cols["name"][1] != "bob" {
		t.Errorf("name column = %v, want [alice, bob]", cols["name"])
	}
}

func TestQueryDuckDB_JSONColumnar_Empty(t *testing.T) {
	s := testServerWithDB(t)
	result, err := s.queryDuckDB(
		"SELECT 1 AS id, 'x' AS name WHERE false",
		queryDuckDBOptions{Format: "json_columnar"},
	)
	if err != nil {
		t.Fatalf("queryDuckDB failed: %v", err)
	}

	var cols map[string][]any
	if err := json.Unmarshal([]byte(result), &cols); err != nil {
		t.Fatalf("failed to parse JSON: %v\nraw: %s", err, result)
	}
	if len(cols["id"]) != 0 {
		t.Errorf("expected empty id column, got %d values", len(cols["id"]))
	}
}

func TestQueryDuckDB_JSONColumnar_PreservesColumnOrder(t *testing.T) {
	s := testServerWithDB(t)
	result, err := s.queryDuckDB(
		"SELECT 'z' AS z_col, 'a' AS a_col, 'm' AS m_col",
		queryDuckDBOptions{Format: "json_columnar"},
	)
	if err != nil {
		t.Fatalf("queryDuckDB failed: %v", err)
	}

	zIdx := strings.Index(result, `"z_col"`)
	aIdx := strings.Index(result, `"a_col"`)
	mIdx := strings.Index(result, `"m_col"`)
	if zIdx < 0 || aIdx < 0 || mIdx < 0 {
		t.Fatalf("missing column keys in output: %s", result)
	}
	if !(zIdx < aIdx && aIdx < mIdx) {
		t.Errorf("column order not preserved: z=%d, a=%d, m=%d\nraw: %s", zIdx, aIdx, mIdx, result)
	}
}

func TestQueryDuckDB_MaxRows(t *testing.T) {
	s := testServerWithDB(t)

	result, err := s.queryDuckDB(
		"SELECT unnest(range(100)) AS n",
		queryDuckDBOptions{Format: "json_rows", MaxRows: 5},
	)
	if err != nil {
		t.Fatalf("queryDuckDB failed: %v", err)
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(result), &rows); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(rows) != 5 {
		t.Errorf("expected 5 rows, got %d", len(rows))
	}
}

func TestQueryDuckDB_MaxRows_Zero_UsesDefault(t *testing.T) {
	s := testServerWithDB(t)
	_, err := s.queryDuckDB(
		"SELECT 1 AS val",
		queryDuckDBOptions{MaxRows: 0},
	)
	if err != nil {
		t.Fatalf("queryDuckDB with MaxRows=0 failed: %v", err)
	}
}

func TestQueryDuckDB_MaxRows_CSV(t *testing.T) {
	s := testServerWithDB(t)

	result, err := s.queryDuckDB(
		"SELECT unnest(range(100)) AS n",
		queryDuckDBOptions{Format: "csv", MaxRows: 3},
	)
	if err != nil {
		t.Fatalf("queryDuckDB failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines (header + 3 rows), got %d:\n%s", len(lines), result)
	}
}

func TestQueryDuckDB_InvalidFormat(t *testing.T) {
	s := testServerWithDB(t)
	_, err := s.queryDuckDB(
		"SELECT 1 AS val",
		queryDuckDBOptions{Format: "xml"},
	)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestQueryDuckDB_JSONRows_NumericTypes(t *testing.T) {
	s := testServerWithDB(t)
	result, err := s.queryDuckDB(
		"SELECT 42 AS int_val, CAST(3.14 AS DOUBLE) AS float_val, true AS bool_val",
		queryDuckDBOptions{Format: "json_rows"},
	)
	if err != nil {
		t.Fatalf("queryDuckDB failed: %v", err)
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(result), &rows); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	intVal, ok := rows[0]["int_val"].(float64)
	if !ok {
		t.Errorf("int_val should be a number, got %T: %v", rows[0]["int_val"], rows[0]["int_val"])
	} else if intVal != 42 {
		t.Errorf("int_val = %v, want 42", intVal)
	}

	floatVal, ok := rows[0]["float_val"].(float64)
	if !ok {
		t.Errorf("float_val should be a number, got %T: %v", rows[0]["float_val"], rows[0]["float_val"])
	} else if floatVal != 3.14 {
		t.Errorf("float_val = %v, want 3.14", floatVal)
	}
}
