// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	"github.com/NimbleMarkets/dbn-go/internal/file"
	"nimbleterminal.dev/ntduckbox"
)

// safeNameChars matches the characters allowed in dataset names (e.g. XNAS.ITCH)
// and schema names (e.g. ohlcv-1d): alphanumeric, dot, hyphen, and underscore.
var safeNameChars = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// isSafeName reports whether name is a valid dataset or schema name that is
// safe to use as a cache directory name. "." and ".." pass the character
// check but are directory references, not names, so they are rejected.
func isSafeName(name string) bool {
	return name != "." && name != ".." && safeNameChars.MatchString(name)
}

// isCacheSchemaDir reports whether dir is exactly {cacheDir}/{dataset}/{schema}.
// Cache files are only ever removed from such directories.
func isCacheSchemaDir(dir string, cacheDir string) bool {
	rel, err := filepath.Rel(cacheDir, dir)
	if err != nil || !filepath.IsLocal(rel) {
		return false
	}
	parts := strings.Split(rel, string(filepath.Separator))
	return len(parts) == 2 && isSafeName(parts[0]) && isSafeName(parts[1])
}

// sqlLiteral escapes a string for use as a SQL string literal,
// preventing SQL injection via embedded single quotes.
func sqlLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// schemaSupportsParquet returns true if the schema can be converted to parquet.
func schemaSupportsParquet(schema dbn.Schema) bool {
	return file.ParquetGroupNodeForDbnSchema(schema) != nil
}

// normalizeDateForFilename strips hyphens and colons from date strings for filesystem-safe names.
// e.g. "2024-01-15" -> "20240115", "2024-01-15T09:30:00Z" -> "20240115T093000Z"
func normalizeDateForFilename(s string) string {
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, ":", "")
	return s
}

// cacheParquetPath returns the parquet file path for a cache entry.
// Format: {symbols}__{stype_in}__{stype_out}__{start}__{end}.parquet
// If the filename exceeds 200 chars, symbols are truncated and a hash suffix is added.
//
// symbols, start and end are caller-supplied text, so they are encoded with
// file.SafePathComponent: the result is always a single file name directly
// under {cacheDir}/{dataset}/{schema}. dataset and schema must already be
// validated names.
func (s *Server) cacheParquetPath(dataset, schema, symbols, stypeIn, stypeOut, start, end string) string {
	startNorm := file.SafePathComponent(normalizeDateForFilename(start))
	endNorm := file.SafePathComponent(normalizeDateForFilename(end))
	rawSymbols := symbols // hashed as given, so existing hashed names stay stable
	symbols = file.SafePathComponent(symbols)

	// Build filename: {symbols}__{stype_in}__{stype_out}__{start}__{end}
	name := fmt.Sprintf("%s__%s__%s__%s__%s", symbols, stypeIn, stypeOut, startNorm, endNorm)

	const maxLen = 200
	if len(name)+len(".parquet") > maxLen {
		// Truncate symbols and append hash
		h := sha256.Sum256([]byte(rawSymbols + "|" + stypeIn + "|" + stypeOut + "|" + start + "|" + end))
		hash8 := fmt.Sprintf("%x", h[:4])
		suffix := fmt.Sprintf("__%s__%s__%s__%s__%s", stypeIn, stypeOut, startNorm, endNorm, hash8)
		// Count how many symbols we can fit
		symParts := strings.Split(symbols, ",")
		budget := maxLen - len(suffix) - len(".parquet")
		var truncSymbols string
		for i := range symParts {
			candidate := strings.Join(symParts[:i+1], ",")
			extra := fmt.Sprintf("+%dmore", len(symParts)-i-1)
			if len(candidate)+len(extra) > budget && i > 0 {
				truncSymbols = strings.Join(symParts[:i], ",") + fmt.Sprintf("+%dmore", len(symParts)-i)
				break
			}
		}
		if truncSymbols == "" {
			truncSymbols = symParts[0]
			if len(symParts) > 1 {
				truncSymbols += fmt.Sprintf("+%dmore", len(symParts)-1)
			}
		}
		name = truncSymbols + suffix
	}

	return filepath.Join(s.cacheDir, dataset, schema, name+".parquet")
}

// CacheManifest is the sidecar JSON manifest written alongside each cached parquet file.
type CacheManifest struct {
	Symbols     []string `json:"symbols"`
	StypeIn     string   `json:"stype_in"`
	StypeOut    string   `json:"stype_out"`
	Start       string   `json:"start"`
	End         string   `json:"end"`
	FetchedAt   string   `json:"fetched_at"`
	RecordCount int64    `json:"record_count"`
	Cost        float64  `json:"cost"`
}

// manifestPath returns the .json sidecar path for a given .parquet path.
func manifestPath(parquetPath string) string {
	return strings.TrimSuffix(parquetPath, ".parquet") + ".json"
}

// writeManifest writes a sidecar JSON manifest alongside a parquet cache file.
func writeManifest(parquetPath string, m CacheManifest) error {
	m.FetchedAt = time.Now().UTC().Format(time.RFC3339)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath(parquetPath), data, 0644)
}

// readManifest reads a sidecar JSON manifest for a parquet cache file.
// Returns nil if the manifest does not exist or cannot be parsed.
func readManifest(parquetPath string) *CacheManifest {
	data, err := os.ReadFile(manifestPath(parquetPath))
	if err != nil {
		return nil
	}
	var m CacheManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	return &m
}

const (
	defaultQueryRows = 10_000
	maxQueryRows     = 100_000
	maxQueryChars    = 10_000_000
)

// InitCache opens an in-memory ntduckbox.DB and registers views for any existing cached parquet files.
func (s *Server) InitCache() error {
	if err := os.MkdirAll(s.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}
	absCacheDir, err := filepath.Abs(s.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to resolve cache dir: %w", err)
	}
	if evalDir, err := filepath.EvalSymlinks(absCacheDir); err == nil {
		absCacheDir = evalDir
	}
	s.cacheDir = absCacheDir

	box, err := ntduckbox.Open(ntduckbox.Options{
		Path:         ":memory:",
		AllowedDirs:  []string{s.cacheDir},
		MaxRows:      maxQueryRows,
		MaxChars:     maxQueryChars,
		QueryTimeout: 30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to open ntduckbox: %w", err)
	}
	s.box = box
	s.queryGate = make(chan struct{}, 1)

	return s.refreshViews()
}

// defaultQueryWait bounds how long a query_cache call waits for an in-flight
// query to finish before telling the caller to retry.
const defaultQueryWait = 5 * time.Second

// acquireQuery takes the query gate or returns a busy error after queryWait.
// The caller must call releaseQuery on success.
func (s *Server) acquireQuery(ctx context.Context) error {
	wait := s.queryWait
	if wait <= 0 {
		wait = defaultQueryWait
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case s.queryGate <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return fmt.Errorf("another query_cache call is still running (waited %s); retry shortly, or simplify the other query", wait)
	}
}

func (s *Server) releaseQuery() { <-s.queryGate }

// Close closes the ntduckbox DuckDB connection.
func (s *Server) Close() error {
	if s.box != nil {
		err := s.box.Close()
		s.box = nil
		return err
	}
	return nil
}

// refreshViews scans CacheDir for {dataset}/{schema}/*.parquet directories
// and updates the managed views in ntduckbox.DB.
func (s *Server) refreshViews() error {
	if s.box == nil {
		return nil
	}

	wantViews := map[string]string{} // view name -> glob path
	datasets, _ := os.ReadDir(s.cacheDir)
	for _, ds := range datasets {
		if !ds.IsDir() || !isSafeName(ds.Name()) {
			continue
		}
		schemas, _ := os.ReadDir(filepath.Join(s.cacheDir, ds.Name()))
		for _, sc := range schemas {
			if !sc.IsDir() || !isSafeName(sc.Name()) {
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

	// One alias view per raw view, replaced in the same transaction.
	aliasViews := make(map[string]string, len(wantViews))
	for viewName := range wantViews {
		_, schema, _ := strings.Cut(viewName, "/")
		alias := aliasViewName(viewName)
		if _, clash := aliasViews[alias]; clash {
			if s.Logger != nil {
				s.Logger.Warn("alias view name collision; skipping alias", "view", viewName, "alias", alias)
			}
			continue
		}
		aliasViews[alias] = aliasViewSQL(viewName, schema)
	}

	err := s.box.SetAllViews(context.Background(), ntduckbox.ViewSet{Parquet: wantViews, SQL: aliasViews})
	if err != nil {
		if s.Logger != nil {
			s.Logger.Warn("failed to refresh cache views", "error", err)
		}
		return err
	}
	return nil
}

// refreshViewForSchema refreshes the DuckDB view catalog for the given dataset/schema.
func (s *Server) refreshViewForSchema(dataset, schema string) {
	if s.box == nil {
		return
	}
	if !isSafeName(dataset) || !isSafeName(schema) {
		return
	}
	_ = s.refreshViews()
}

// queryDuckDBOptions controls query execution and output formatting.
type queryDuckDBOptions struct {
	Format  string // "csv" (default), "markdown", "json_rows", "json_columnar"
	MaxRows int    // 0 = use default limit (10000); clamped to maxQueryRows (100000)
}

// queryDuckDB executes a SQL query via ntduckbox and returns the rendered text.
func (s *Server) queryDuckDB(userSQL string, opts queryDuckDBOptions) (string, error) {
	res, err := s.queryDuckDBResult(context.Background(), userSQL, opts)
	return res.Text, err
}

// queryDuckDBResult executes a SQL query via ntduckbox and returns the full Result (including Rows and Truncated).
func (s *Server) queryDuckDBResult(ctx context.Context, userSQL string, opts queryDuckDBOptions) (ntduckbox.Result, error) {
	if s.box == nil {
		return ntduckbox.Result{}, fmt.Errorf("cache not initialized")
	}

	format := ntduckbox.CSV
	if opts.Format != "" {
		format = ntduckbox.Format(opts.Format)
	}

	if err := s.acquireQuery(ctx); err != nil {
		return ntduckbox.Result{}, err
	}
	defer s.releaseQuery()

	res, err := s.box.QueryWith(ctx, userSQL, ntduckbox.QueryOptions{
		Format:   format,
		MaxRows:  opts.effectiveMaxRows(),
		Truncate: true,
	})
	if err != nil {
		return ntduckbox.Result{}, classifyQueryError(err)
	}
	return res, nil
}

// effectiveMaxRows returns the row cap actually sent to ntduckbox.
func (opts queryDuckDBOptions) effectiveMaxRows() int {
	if opts.MaxRows > 0 {
		return min(opts.MaxRows, maxQueryRows)
	}
	return defaultQueryRows
}

// classifyQueryError turns ntduckbox's sentinel errors into messages that tell
// the LLM what to do next. The engine itself enforces the statement type
// (single read-only SELECT-style statements), the file allowlist, and the
// locked configuration, so no host-side SQL allowlist is needed.
func classifyQueryError(err error) error {
	switch {
	case errors.Is(err, ntduckbox.ErrNotSelect):
		return fmt.Errorf("only a single read-only SELECT-style statement is allowed (SELECT, WITH ... SELECT, FROM-first, VALUES, DESCRIBE, SUMMARIZE, SHOW): %w", err)
	case errors.Is(err, ntduckbox.ErrPermission):
		return fmt.Errorf("query_cache can only read the cached views; file, extension, and settings access is disabled: %w", err)
	case errors.Is(err, ntduckbox.ErrBinder):
		return fmt.Errorf("%w (use describe_cache or list_cache to see view names and columns)", err)
	default:
		return err
	}
}

// CacheFileInfo describes a single cached parquet file and its manifest metadata.
type CacheFileInfo struct {
	Filename    string   `json:"filename"`
	Symbols     []string `json:"symbols,omitempty"`
	StypeIn     string   `json:"stype_in,omitempty"`
	StypeOut    string   `json:"stype_out,omitempty"`
	Start       string   `json:"start,omitempty"`
	End         string   `json:"end,omitempty"`
	FetchedAt   string   `json:"fetched_at,omitempty"`
	RecordCount int64    `json:"record_count,omitempty"`
	Cost        float64  `json:"cost,omitempty"`
	SizeBytes   int64    `json:"size_bytes"`
}

// CacheEntry describes a cached dataset/schema in the cache directory.
type CacheEntry struct {
	ViewName  string          `json:"view_name"`  // Raw view, e.g. "XNAS.ITCH/trades" (double-quote it in SQL; nanosecond timestamps)
	AliasView string          `json:"alias_view"` // Plain-identifier alias, e.g. xnas_itch_trades (TIMESTAMPTZ timestamps)
	Dataset   string          `json:"dataset"`
	Schema    string          `json:"schema"`
	Files     []CacheFileInfo `json:"files"`
	TotalSize int64           `json:"total_size_bytes"`
}

// listCacheEntries scans the cache directory and returns information about cached data.
func (s *Server) listCacheEntries() []CacheEntry {
	var entries []CacheEntry

	datasets, _ := os.ReadDir(s.cacheDir)
	for _, ds := range datasets {
		if !ds.IsDir() || !isSafeName(ds.Name()) {
			continue
		}
		schemas, _ := os.ReadDir(filepath.Join(s.cacheDir, ds.Name()))
		for _, sc := range schemas {
			if !sc.IsDir() || !isSafeName(sc.Name()) {
				continue
			}
			parquetGlob := filepath.Join(s.cacheDir, ds.Name(), sc.Name(), "*.parquet")
			matches, _ := filepath.Glob(parquetGlob)
			if len(matches) == 0 {
				continue
			}
			var totalSize int64
			var files []CacheFileInfo
			for _, m := range matches {
				var fileSize int64
				if info, err := os.Stat(m); err == nil {
					fileSize = info.Size()
				}
				totalSize += fileSize

				fi := CacheFileInfo{
					Filename:  filepath.Base(m),
					SizeBytes: fileSize,
				}
				if manifest := readManifest(m); manifest != nil {
					fi.Symbols = manifest.Symbols
					fi.StypeIn = manifest.StypeIn
					fi.StypeOut = manifest.StypeOut
					fi.Start = manifest.Start
					fi.End = manifest.End
					fi.FetchedAt = manifest.FetchedAt
					fi.RecordCount = manifest.RecordCount
					fi.Cost = manifest.Cost
				}
				files = append(files, fi)
			}
			viewName := ds.Name() + "/" + sc.Name()
			entries = append(entries, CacheEntry{
				ViewName:  viewName,
				AliasView: aliasViewName(viewName),
				Dataset:   ds.Name(),
				Schema:    sc.Name(),
				Files:     files,
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
	if (dataset != "" && !isSafeName(dataset)) || (schema != "" && !isSafeName(schema)) {
		return 0
	}
	removed := 0

	if dataset != "" && schema != "" {
		dir := filepath.Join(s.cacheDir, dataset, schema)
		removed += removeCacheFiles(dir, s.cacheDir)
	} else if dataset != "" {
		schemas, _ := os.ReadDir(filepath.Join(s.cacheDir, dataset))
		for _, sc := range schemas {
			if sc.IsDir() {
				removed += removeCacheFiles(filepath.Join(s.cacheDir, dataset, sc.Name()), s.cacheDir)
			}
		}
	} else if schema != "" {
		datasets, _ := os.ReadDir(s.cacheDir)
		for _, ds := range datasets {
			if !ds.IsDir() {
				continue
			}
			dir := filepath.Join(s.cacheDir, ds.Name(), schema)
			removed += removeCacheFiles(dir, s.cacheDir)
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
					removed += removeCacheFiles(filepath.Join(s.cacheDir, ds.Name(), sc.Name()), s.cacheDir)
				}
			}
		}
	}

	s.refreshViews()
	return removed
}

// removeCacheFiles removes all .parquet and .json sidecar files in a directory
// and cleans up empty parent directories, stopping at (and never removing) boundDir.
// It does nothing unless dir is a {dataset}/{schema} directory directly under boundDir.
func removeCacheFiles(dir string, boundDir string) int {
	if !isCacheSchemaDir(dir, boundDir) {
		return 0
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "*.parquet"))
	for _, m := range matches {
		os.Remove(m)
		os.Remove(manifestPath(m)) // remove sidecar .json if it exists
	}
	// Also remove any orphaned .json files
	jsonMatches, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	for _, m := range jsonMatches {
		os.Remove(m)
	}
	// Remove empty directories walking up, but never past boundDir
	cleanBound := filepath.Clean(boundDir)
	for d := filepath.Clean(dir); d != cleanBound && strings.HasPrefix(d, cleanBound+string(filepath.Separator)); d = filepath.Dir(d) {
		entries, err := os.ReadDir(d)
		if err != nil || len(entries) > 0 {
			break
		}
		os.Remove(d)
	}
	return len(matches)
}
