// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/NimbleMarkets/dbn-go"
	"github.com/NimbleMarkets/dbn-go/internal/file"
	_ "github.com/duckdb/duckdb-go/v2"
)

// safeName matches valid dataset names (e.g. XNAS.ITCH) and schema names (e.g. ohlcv-1d).
// Only alphanumeric, dot, hyphen, and underscore are allowed.
var safeName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// schemaSupportsParquet returns true if the schema can be converted to parquet.
func schemaSupportsParquet(schema dbn.Schema) bool {
	return file.ParquetGroupNodeForDbnSchema(schema) != nil
}

// cacheParquetPath returns the parquet file path for a cache entry.
// Hash input: symbols|stype_in|start|end
func (s *Server) cacheParquetPath(dataset, schema, symbols, stypeIn, start, end string) string {
	h := sha256.Sum256([]byte(symbols + "|" + stypeIn + "|" + start + "|" + end))
	hash8 := fmt.Sprintf("%x", h[:4]) // 4 bytes = 8 hex chars
	return filepath.Join(s.cacheDir, dataset, schema, hash8+".parquet")
}

// InitCache opens an in-memory DuckDB database and creates views for any existing cached parquet files.
func (s *Server) InitCache() error {
	if err := os.MkdirAll(s.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}

	db, err := sql.Open("duckdb", "")
	if err != nil {
		return fmt.Errorf("failed to open DuckDB: %w", err)
	}
	// Security hardening: disable extensions and remote filesystem access.
	// We keep local file access enabled because read_parquet() needs it for views.
	// lock_configuration prevents user SQL from re-enabling these.
	for _, stmt := range []string{
		"SET allow_extensions_autoloading = false",
		"SET disabled_filesystems = 'HTTPFileSystem'",
		"SET lock_configuration = true",
	} {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return fmt.Errorf("failed to configure DuckDB (%s): %w", stmt, err)
		}
	}
	s.db = db

	return s.refreshViews()
}

// Close closes the DuckDB connection.
func (s *Server) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// refreshViews scans CacheDir for {dataset}/{schema}/*.parquet directories
// and creates or drops DuckDB views accordingly.
func (s *Server) refreshViews() error {
	if s.db == nil {
		return nil
	}

	// Collect which views should exist
	wantViews := map[string]string{} // view name -> glob path
	datasets, _ := os.ReadDir(s.cacheDir)
	for _, ds := range datasets {
		if !ds.IsDir() || !safeName.MatchString(ds.Name()) {
			continue
		}
		schemas, _ := os.ReadDir(filepath.Join(s.cacheDir, ds.Name()))
		for _, sc := range schemas {
			if !sc.IsDir() || !safeName.MatchString(sc.Name()) {
				continue
			}
			parquetGlob := filepath.Join(s.cacheDir, ds.Name(), sc.Name(), "*.parquet")
			matches, _ := filepath.Glob(parquetGlob)
			if len(matches) > 0 {
				viewName := ds.Name() + "/" + sc.Name()
				wantViews[viewName] = parquetGlob
			}
		}
	}

	// Create or replace views for existing parquet files
	for viewName, globPath := range wantViews {
		stmt := fmt.Sprintf(`CREATE OR REPLACE VIEW "%s" AS SELECT * FROM read_parquet('%s')`, viewName, globPath)
		if _, err := s.db.Exec(stmt); err != nil {
			s.Logger.Warn("failed to create view", "view", viewName, "error", err)
		}
	}

	// Drop views that no longer have backing files
	rows, err := s.db.Query("SELECT table_name FROM information_schema.tables WHERE table_type = 'VIEW'")
	if err != nil {
		return nil // non-fatal
	}
	defer rows.Close()

	var existingViews []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			existingViews = append(existingViews, name)
		}
	}
	for _, v := range existingViews {
		if _, ok := wantViews[v]; !ok {
			s.db.Exec(fmt.Sprintf(`DROP VIEW IF EXISTS "%s"`, v))
		}
	}

	return nil
}

// refreshViewForSchema creates or refreshes a single DuckDB view for the given dataset/schema.
func (s *Server) refreshViewForSchema(dataset, schema string) {
	if s.db == nil {
		return
	}
	if !safeName.MatchString(dataset) || !safeName.MatchString(schema) {
		return
	}

	viewName := dataset + "/" + schema
	parquetGlob := filepath.Join(s.cacheDir, dataset, schema, "*.parquet")
	matches, _ := filepath.Glob(parquetGlob)

	if len(matches) > 0 {
		stmt := fmt.Sprintf(`CREATE OR REPLACE VIEW "%s" AS SELECT * FROM read_parquet('%s')`, viewName, parquetGlob)
		if _, err := s.db.Exec(stmt); err != nil {
			s.Logger.Warn("failed to create view", "view", viewName, "error", err)
		}
		return
	}

	// No files, drop view if it exists
	s.db.Exec(fmt.Sprintf(`DROP VIEW IF EXISTS "%s"`, viewName))
}

// queryDuckDB executes a SQL query against DuckDB and returns results as CSV.
func (s *Server) queryDuckDB(userSQL string) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("cache not initialized")
	}

	wrappedSQL := fmt.Sprintf("SELECT * FROM (%s) LIMIT 10000", userSQL)

	rows, err := s.db.Query(wrappedSQL)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	w := csv.NewWriter(&buf)

	w.Write(columns)
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return "", err
		}
		record := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				record[i] = ""
			} else if b, ok := val.([]byte); ok {
				record[i] = string(b)
			} else {
				record[i] = fmt.Sprintf("%v", val)
			}
		}
		w.Write(record)
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// CacheEntry describes a cached dataset/schema in the cache directory.
type CacheEntry struct {
	ViewName  string `json:"view_name"`
	Dataset   string `json:"dataset"`
	Schema    string `json:"schema"`
	FileCount int    `json:"file_count"`
	TotalSize int64  `json:"total_size_bytes"`
}

// listCacheEntries scans the cache directory and returns information about cached data.
func (s *Server) listCacheEntries() []CacheEntry {
	var entries []CacheEntry

	datasets, _ := os.ReadDir(s.cacheDir)
	for _, ds := range datasets {
		if !ds.IsDir() || !safeName.MatchString(ds.Name()) {
			continue
		}
		schemas, _ := os.ReadDir(filepath.Join(s.cacheDir, ds.Name()))
		for _, sc := range schemas {
			if !sc.IsDir() || !safeName.MatchString(sc.Name()) {
				continue
			}
			parquetGlob := filepath.Join(s.cacheDir, ds.Name(), sc.Name(), "*.parquet")
			matches, _ := filepath.Glob(parquetGlob)
			if len(matches) == 0 {
				continue
			}
			var totalSize int64
			for _, m := range matches {
				if info, err := os.Stat(m); err == nil {
					totalSize += info.Size()
				}
			}
			entries = append(entries, CacheEntry{
				ViewName:  ds.Name() + "/" + sc.Name(),
				Dataset:   ds.Name(),
				Schema:    sc.Name(),
				FileCount: len(matches),
				TotalSize: totalSize,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ViewName < entries[j].ViewName
	})
	return entries
}

// clearCache removes cached parquet files matching the optional dataset and schema filters.
// Caller must hold s.mu.
func (s *Server) clearCache(dataset, schema string) int {
	removed := 0

	if dataset != "" && schema != "" {
		dir := filepath.Join(s.cacheDir, dataset, schema)
		removed += removeParquetFiles(dir, s.cacheDir)
	} else if dataset != "" {
		schemas, _ := os.ReadDir(filepath.Join(s.cacheDir, dataset))
		for _, sc := range schemas {
			if sc.IsDir() {
				removed += removeParquetFiles(filepath.Join(s.cacheDir, dataset, sc.Name()), s.cacheDir)
			}
		}
	} else {
		datasets, _ := os.ReadDir(s.cacheDir)
		for _, ds := range datasets {
			if !ds.IsDir() {
				continue
			}
			schemas, _ := os.ReadDir(filepath.Join(s.cacheDir, ds.Name()))
			for _, sc := range schemas {
				if sc.IsDir() {
					removed += removeParquetFiles(filepath.Join(s.cacheDir, ds.Name(), sc.Name()), s.cacheDir)
				}
			}
		}
	}

	s.refreshViews()
	return removed
}

// removeParquetFiles removes all .parquet files in a directory and cleans up
// empty parent directories, stopping at (and never removing) boundDir.
func removeParquetFiles(dir string, boundDir string) int {
	matches, _ := filepath.Glob(filepath.Join(dir, "*.parquet"))
	for _, m := range matches {
		os.Remove(m)
	}
	// Remove empty directories walking up, but never past boundDir
	cleanBound := filepath.Clean(boundDir)
	for d := dir; d != cleanBound && strings.HasPrefix(d, cleanBound); d = filepath.Dir(d) {
		entries, err := os.ReadDir(d)
		if err != nil || len(entries) > 0 {
			break
		}
		os.Remove(d)
	}
	return len(matches)
}
