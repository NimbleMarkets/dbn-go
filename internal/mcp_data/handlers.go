// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"context"
	"strings"

	"github.com/NimbleMarkets/dbn-go"
	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/NimbleMarkets/dbn-go/internal/mcp_meta"
	"github.com/mark3labs/mcp-go/mcp"
)

///////////////////////////////////////////////////////////////////////////////

func (s *Server) getRangeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	p, errResult := mcp_meta.ParseCommonParams(request)
	if errResult != nil {
		return errResult, nil
	}

	schema, err := dbn.SchemaFromString(p.SchemaStr)
	if err != nil {
		return mcp.NewToolResultErrorf("schema was invalid: %s", err), nil
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

	jobParams := dbn_hist.SubmitJobParams{
		Dataset:      p.Dataset,
		Symbols:      strings.Join(p.Symbols, ","),
		Schema:       schema,
		DateRange:    dbn_hist.DateRange{Start: p.StartTime, End: p.EndTime},
		Encoding:     dbn.Encoding_Json,
		PrettyPx:     true,
		PrettyTs:     true,
		MapSymbols:   true,
		SplitSymbols: false,
		Delivery:     dbn_hist.Delivery_Download,
		StypeIn:      p.StypeIn,
		StypeOut:     stypeOut,
	}
	rangeData, err := dbn_hist.GetRange(s.ApiKey, jobParams)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get range: %s", err), nil
	}

	s.Logger.Info("get_range", "dataset", p.Dataset, "schema", p.SchemaStr, "symbols", len(p.Symbols),
		"start", p.StartStr, "end", p.EndStr, "cost", cost, "size", len(rangeData))
	return mcp.NewToolResultText(string(rangeData)), nil
}

///////////////////////////////////////////////////////////////////////////////
// Cache tool handlers (stubs — DuckDB integration is a follow-up)

func (s *Server) queryCacheHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError("not yet implemented: query_cache requires DuckDB integration (coming soon)"), nil
}

func (s *Server) listCacheHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError("not yet implemented: list_cache requires DuckDB integration (coming soon)"), nil
}

func (s *Server) clearCacheHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError("not yet implemented: clear_cache requires DuckDB integration (coming soon)"), nil
}
