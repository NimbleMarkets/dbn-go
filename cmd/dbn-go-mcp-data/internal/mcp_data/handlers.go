// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NimbleMarkets/dbn-go"
	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/NimbleMarkets/dbn-go/internal/file"
	"github.com/NimbleMarkets/dbn-go/internal/mcp_meta"
	"github.com/mark3labs/mcp-go/mcp"
	"nimbleterminal.dev/ntduckbox"
)

///////////////////////////////////////////////////////////////////////////////

type fetchRangeQuery struct {
	Dataset string `json:"dataset"`
	Schema  string `json:"schema"`
	Symbols string `json:"symbols"`
	STypeIn string `json:"stype_in"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

type fetchRangeResult struct {
	Status      string          `json:"status"`
	ViewName    string          `json:"view_name"`
	ParquetPath string          `json:"parquet_path"`
	FileSize    int64           `json:"file_size"`
	RecordCount int64           `json:"record_count"`
	Cost        float64         `json:"cost"`
	Query       fetchRangeQuery `json:"query"`
}

func (s *Server) fetchRangeOnce(cacheKey string, fn func() (string, error)) (string, error) {
	result, err, _ := s.fetches.Do(cacheKey, func() (any, error) {
		return fn()
	})
	if err != nil {
		return "", err
	}

	text, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("internal error: unexpected fetch result type %T", result)
	}
	return text, nil
}

func (s *Server) countParquetRecords(parquetPath string) int64 {
	var recordCount int64
	if s.box != nil {
		out, err := s.box.Query(context.Background(), fmt.Sprintf(`SELECT count(*) AS n FROM read_parquet(%s)`, sqlLiteral(parquetPath)))
		if err == nil {
			lines := strings.Split(strings.TrimSpace(out), "\n")
			if len(lines) >= 2 {
				_, _ = fmt.Sscanf(strings.TrimSpace(lines[1]), "%d", &recordCount)
			}
		}
	}
	return recordCount
}

func (s *Server) fetchRangeResultText(status, viewName, parquetPath string, fileSize, recordCount int64, cost float64,
	p *mcp_meta.CommonParams, symbolsStr string,
) (string, error) {
	result := fetchRangeResult{
		Status:      status,
		ViewName:    viewName,
		ParquetPath: parquetPath,
		FileSize:    fileSize,
		RecordCount: recordCount,
		Cost:        cost,
		Query: fetchRangeQuery{
			Dataset: p.Dataset,
			Schema:  p.SchemaStr,
			Symbols: symbolsStr,
			STypeIn: p.StypeIn.String(),
			Start:   p.StartStr,
			End:     p.EndStr,
		},
	}

	jbytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(jbytes), nil
}

func (s *Server) cachedFetchRangeResult(parquetPath, viewName string, p *mcp_meta.CommonParams, symbolsStr string) (string, bool, error) {
	info, err := os.Stat(parquetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("failed to stat cache file: %w", err)
	}

	recordCount := s.countParquetRecords(parquetPath)
	text, err := s.fetchRangeResultText("cache_hit", viewName, parquetPath, info.Size(), recordCount, 0.0, p, symbolsStr)
	if err != nil {
		return "", false, err
	}

	s.Logger.Info("fetch_range cache_hit", "dataset", p.Dataset, "schema", p.SchemaStr,
		"symbols", len(p.Symbols), "start", p.StartStr, "end", p.EndStr, "records", recordCount)
	return text, true, nil
}

func (s *Server) fetchRangeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if errResult := s.RequireAPIKey(); errResult != nil {
		return errResult, nil
	}

	p, errResult := mcp_meta.ParseCommonParams(request)
	if errResult != nil {
		return errResult, nil
	}

	schema, err := dbn.SchemaFromString(p.SchemaStr)
	if err != nil {
		return mcp.NewToolResultErrorf("schema was invalid: %s", err), nil
	}

	if !schemaSupportsParquet(schema) {
		return mcp.NewToolResultErrorf("schema %q is not supported by fetch_range (no parquet conversion). Supported: ohlcv-1s, ohlcv-1m, ohlcv-1h, ohlcv-1d, trades, mbp-1, mbp-10, tbbo, cmbp-1, bbo-1s, bbo-1m, imbalance, statistics, status", p.SchemaStr), nil
	}

	stypeOut := dbn.SType_InstrumentId
	if stypeOutStr, err := request.RequireString("stype_out"); err == nil && stypeOutStr != "" {
		if stypeOut, err = dbn.STypeFromString(stypeOutStr); err != nil {
			return mcp.NewToolResultErrorf("invalid stype_out: %s", err), nil
		}
	}

	// Normalize symbol order for consistent cache keys
	sort.Strings(p.Symbols)
	symbolsStr := strings.Join(p.Symbols, ",")
	stypeOutStr := stypeOut.String()
	parquetPath := s.cacheParquetPath(p.Dataset, p.SchemaStr, symbolsStr, p.StypeIn.String(), stypeOutStr, p.StartStr, p.EndStr)
	viewName := p.Dataset + "/" + p.SchemaStr

	force := request.GetBool("force", false)
	resultText, err := s.fetchRangeOnce(parquetPath, func() (string, error) {
		if !force {
			if text, ok, err := s.cachedFetchRangeResult(parquetPath, viewName, p, symbolsStr); err != nil {
				return "", err
			} else if ok {
				return text, nil
			}
		}

		metaParams := p.MetadataQueryParams()
		cost, err := dbn_hist.GetCost(s.GetApiKey(), metaParams)
		if err != nil {
			return "", fmt.Errorf("failed to get cost: %w", err)
		}
		if s.maxCost > 0 && cost > s.maxCost {
			return "", fmt.Errorf("query exceeds budget: estimated cost $%.2f, budget $%.2f", cost, s.maxCost)
		}

		jobParams := dbn_hist.SubmitJobParams{
			Dataset:      p.Dataset,
			Symbols:      symbolsStr,
			Schema:       schema,
			DateRange:    dbn_hist.DateRange{Start: p.StartTime, End: p.EndTime},
			Encoding:     dbn.Encoding_Dbn,
			Compression:  dbn.Compress_ZStd,
			SplitSymbols: false,
			Delivery:     dbn_hist.Delivery_Download,
			StypeIn:      p.StypeIn,
			StypeOut:     stypeOut,
		}
		rangeData, err := dbn_hist.GetRange(s.GetApiKey(), jobParams)
		if err != nil {
			return "", fmt.Errorf("failed to get range: %w", err)
		}

		tmpFile, err := os.CreateTemp("", "dbn-*.dbn.zst")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := tmpFile.Write(rangeData); err != nil {
			_ = tmpFile.Close()
			return "", fmt.Errorf("failed to write temp file: %w", err)
		}
		if err := tmpFile.Close(); err != nil {
			return "", fmt.Errorf("failed to close temp file: %w", err)
		}

		s.mu.Lock()
		if err := os.MkdirAll(filepath.Dir(parquetPath), 0755); err != nil {
			s.mu.Unlock()
			return "", fmt.Errorf("failed to create cache dir: %w", err)
		}
		if err := file.WriteDbnFileAsParquet(tmpPath, true, parquetPath); err != nil {
			s.mu.Unlock()
			return "", fmt.Errorf("failed to write parquet: %w", err)
		}
		s.refreshViewForSchema(p.Dataset, p.SchemaStr)
		s.mu.Unlock()

		var fileSize int64
		if info, err := os.Stat(parquetPath); err == nil {
			fileSize = info.Size()
		}
		recordCount := s.countParquetRecords(parquetPath)

		if err := writeManifest(parquetPath, CacheManifest{
			Symbols:     p.Symbols,
			StypeIn:     p.StypeIn.String(),
			StypeOut:    stypeOutStr,
			Start:       p.StartStr,
			End:         p.EndStr,
			RecordCount: recordCount,
			Cost:        cost,
		}); err != nil {
			s.Logger.Warn("failed to write cache manifest", "error", err)
		}

		text, err := s.fetchRangeResultText("cached", viewName, parquetPath, fileSize, recordCount, cost, p, symbolsStr)
		if err != nil {
			return "", err
		}

		s.Logger.Info("fetch_range", "dataset", p.Dataset, "schema", p.SchemaStr, "symbols", len(p.Symbols),
			"start", p.StartStr, "end", p.EndStr, "cost", cost, "records", recordCount, "size", fileSize)
		return text, nil
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(resultText), nil
}

///////////////////////////////////////////////////////////////////////////////
// Cache tool handlers

func (s *Server) queryCacheHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sqlStr, err := request.RequireString("sql")
	if err != nil {
		return mcp.NewToolResultError("sql must be set"), nil
	}

	format := request.GetString("format", "csv")
	maxRows := request.GetInt("max_rows", 0)

	opts := queryDuckDBOptions{Format: format, MaxRows: maxRows}
	res, err := s.queryDuckDBResult(ctx, sqlStr, opts)
	if err != nil {
		return mcp.NewToolResultErrorf("query failed: %s", err), nil
	}

	s.Logger.Info("query_cache", "sql", sqlStr, "format", format, "max_rows", maxRows,
		"rows", res.Rows, "bytes", res.Bytes, "truncated", res.Truncated, "elapsed", res.Elapsed)
	toolResult := mcp.NewToolResultText(res.Text)
	if res.Truncated {
		toolResult.Content = append(toolResult.Content, mcp.NewTextContent(truncationNotice(res, opts.effectiveMaxRows())))
	}
	return toolResult, nil
}

// truncationNotice explains which limit cut the result short so the LLM takes
// the right corrective action: raising max_rows cannot help when the byte cap hit.
func truncationNotice(res ntduckbox.Result, maxRows int) string {
	if res.Rows < maxRows {
		return fmt.Sprintf("Showing %d rows (%d bytes); more exist but the result hit the %d MB size limit. Add a WHERE/LIMIT clause or select fewer columns.",
			res.Rows, res.Bytes, maxQueryChars/1_000_000)
	}
	return fmt.Sprintf("Showing %d rows, more exist. Add a SQL LIMIT/WHERE clause or increase max_rows (up to %d).", res.Rows, maxQueryRows)
}

// DescribedColumn is one column of a described view.
type DescribedColumn struct {
	Name string `json:"name"`
	Type string `json:"type"` // DuckDB type, e.g. DOUBLE, UINTEGER, TIMESTAMPTZ_NS
}

// DescribedView is one entry of describe_cache output.
type DescribedView struct {
	ViewName  string            `json:"view_name"`            // Raw Parquet view, e.g. "XNAS.ITCH/trades"
	AliasView string            `json:"alias_view,omitempty"` // Plain-identifier alias, e.g. xnas_itch_trades
	Described string            `json:"described"`            // Which of the two the columns below belong to
	Columns   []DescribedColumn `json:"columns"`
}

func describedColumns(cols []ntduckbox.Column) []DescribedColumn {
	out := make([]DescribedColumn, len(cols))
	for i, c := range cols {
		out[i] = DescribedColumn{Name: c.Name, Type: c.Type}
	}
	return out
}

func (s *Server) describeCacheHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.box == nil {
		return mcp.NewToolResultError("cache not initialized"), nil
	}
	view := request.GetString("view", "")

	var described []DescribedView
	for _, entry := range s.listCacheEntries() {
		target := entry.AliasView // the recommended view: plain name, TIMESTAMPTZ timestamps
		switch view {
		case "":
		case entry.ViewName, entry.AliasView:
			target = view
		default:
			continue
		}
		cols, err := s.box.Columns(ctx, target)
		if err != nil {
			return mcp.NewToolResultErrorf("describe %q failed: %s", target, classifyQueryError(err)), nil
		}
		described = append(described, DescribedView{
			ViewName:  entry.ViewName,
			AliasView: entry.AliasView,
			Described: target,
			Columns:   describedColumns(cols),
		})
	}
	if view != "" && len(described) == 0 {
		return mcp.NewToolResultErrorf("no cached view named %q; call list_cache or describe_cache without arguments to see the available views", view), nil
	}
	if described == nil {
		described = []DescribedView{}
	}

	jbytes, err := json.Marshal(described)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal view descriptions: %s", err), nil
	}
	s.Logger.Info("describe_cache", "view", view, "views", len(described))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func (s *Server) listCacheHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	entries := s.listCacheEntries()

	jbytes, err := json.Marshal(entries)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal cache entries: %s", err), nil
	}

	s.Logger.Info("list_cache", "entries", len(entries))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func (s *Server) clearCacheHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataset, _ := request.RequireString("dataset")
	schema, _ := request.RequireString("schema")

	if dataset != "" && !safeName.MatchString(dataset) {
		return mcp.NewToolResultError("invalid dataset name"), nil
	}
	if schema != "" && !safeName.MatchString(schema) {
		return mcp.NewToolResultError("invalid schema name"), nil
	}

	s.mu.Lock()
	removed := s.clearCache(dataset, schema)
	s.mu.Unlock()

	s.Logger.Info("clear_cache", "dataset", dataset, "schema", schema, "removed", removed)
	return mcp.NewToolResultText(fmt.Sprintf("Removed %d cached file(s)", removed)), nil
}
