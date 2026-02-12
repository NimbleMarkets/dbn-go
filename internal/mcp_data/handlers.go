// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NimbleMarkets/dbn-go"
	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/NimbleMarkets/dbn-go/internal/file"
	"github.com/NimbleMarkets/dbn-go/internal/mcp_meta"
	"github.com/mark3labs/mcp-go/mcp"
)

///////////////////////////////////////////////////////////////////////////////

func (s *Server) fetchRangeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	p, errResult := mcp_meta.ParseCommonParams(request)
	if errResult != nil {
		return errResult, nil
	}

	schema, err := dbn.SchemaFromString(p.SchemaStr)
	if err != nil {
		return mcp.NewToolResultErrorf("schema was invalid: %s", err), nil
	}

	if !schemaSupportsParquet(schema) {
		return mcp.NewToolResultErrorf("schema %q is not supported by fetch_range (no parquet conversion). Supported: ohlcv-1s, ohlcv-1m, ohlcv-1h, ohlcv-1d, trades, mbp-1, tbbo, imbalance, statistics", p.SchemaStr), nil
	}

	stypeOut := dbn.SType_InstrumentId
	if stypeOutStr, err := request.RequireString("stype_out"); err == nil && stypeOutStr != "" {
		if stypeOut, err = dbn.STypeFromString(stypeOutStr); err != nil {
			return mcp.NewToolResultErrorf("invalid stype_out: %s", err), nil
		}
	}

	metaParams := p.MetadataQueryParams()
	cost, err := dbn_hist.GetCost(s.ApiKey, metaParams)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get cost: %s", err), nil
	}
	if s.MaxCost > 0 && cost > s.MaxCost {
		return mcp.NewToolResultErrorf("query exceeds budget: estimated cost $%.2f, budget $%.2f", cost, s.MaxCost), nil
	}

	// Fetch as DBN + ZStd
	jobParams := dbn_hist.SubmitJobParams{
		Dataset:      p.Dataset,
		Symbols:      strings.Join(p.Symbols, ","),
		Schema:       schema,
		DateRange:    dbn_hist.DateRange{Start: p.StartTime, End: p.EndTime},
		Encoding:     dbn.Encoding_Dbn,
		Compression:  dbn.Compress_ZStd,
		SplitSymbols: false,
		Delivery:     dbn_hist.Delivery_Download,
		StypeIn:      p.StypeIn,
		StypeOut:     stypeOut,
	}
	rangeData, err := dbn_hist.GetRange(s.ApiKey, jobParams)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get range: %s", err), nil
	}

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "dbn-*.dbn.zst")
	if err != nil {
		return mcp.NewToolResultErrorf("failed to create temp file: %s", err), nil
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(rangeData); err != nil {
		tmpFile.Close()
		return mcp.NewToolResultErrorf("failed to write temp file: %s", err), nil
	}
	tmpFile.Close()

	// Convert DBN to parquet cache file
	symbolsStr := strings.Join(p.Symbols, ",")
	parquetPath := s.cacheParquetPath(p.Dataset, p.SchemaStr, symbolsStr, p.StypeIn.String(), p.StartStr, p.EndStr)

	s.mu.Lock()
	if err := os.MkdirAll(filepath.Dir(parquetPath), 0755); err != nil {
		s.mu.Unlock()
		return mcp.NewToolResultErrorf("failed to create cache dir: %s", err), nil
	}
	if err := file.WriteDbnFileAsParquet(tmpPath, true, parquetPath); err != nil {
		s.mu.Unlock()
		return mcp.NewToolResultErrorf("failed to write parquet: %s", err), nil
	}
	s.refreshViewForSchema(p.Dataset, p.SchemaStr)
	s.mu.Unlock()

	// Gather file info for response
	viewName := p.Dataset + "/" + p.SchemaStr
	var fileSize int64
	if info, err := os.Stat(parquetPath); err == nil {
		fileSize = info.Size()
	}

	// Count records via DuckDB
	var recordCount int64
	row := s.DB.QueryRow(fmt.Sprintf(`SELECT count(*) FROM "%s"`, viewName))
	row.Scan(&recordCount)

	result := map[string]any{
		"status":       "cached",
		"view_name":    viewName,
		"parquet_path": parquetPath,
		"file_size":    fileSize,
		"record_count": recordCount,
		"cost":         cost,
		"query": map[string]any{
			"dataset":  p.Dataset,
			"schema":   p.SchemaStr,
			"symbols":  symbolsStr,
			"stype_in": p.StypeIn.String(),
			"start":    p.StartStr,
			"end":      p.EndStr,
		},
	}

	jbytes, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal result: %s", err), nil
	}

	s.Logger.Info("fetch_range", "dataset", p.Dataset, "schema", p.SchemaStr, "symbols", len(p.Symbols),
		"start", p.StartStr, "end", p.EndStr, "cost", cost, "records", recordCount, "size", fileSize)
	return mcp.NewToolResultText(string(jbytes)), nil
}

///////////////////////////////////////////////////////////////////////////////
// Cache tool handlers

func (s *Server) queryCacheHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sqlStr, err := request.RequireString("sql")
	if err != nil {
		return mcp.NewToolResultError("sql must be set"), nil
	}

	result, err := s.queryDuckDB(sqlStr)
	if err != nil {
		return mcp.NewToolResultErrorf("query failed: %s", err), nil
	}

	s.Logger.Info("query_cache", "sql", sqlStr)
	return mcp.NewToolResultText(result), nil
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

	s.mu.Lock()
	removed := s.clearCache(dataset, schema)
	s.mu.Unlock()

	s.Logger.Info("clear_cache", "dataset", dataset, "schema", schema, "removed", removed)
	return mcp.NewToolResultText(fmt.Sprintf("Removed %d cached file(s)", removed)), nil
}
