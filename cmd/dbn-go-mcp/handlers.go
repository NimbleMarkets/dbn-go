// Copyright (c) 2025 Neomantra Corp

package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/mark3labs/mcp-go/mcp"
)

///////////////////////////////////////////////////////////////////////////////

func listDatasetsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dateRange, errResult := parseOptionalDateRange(request)
	if errResult != nil {
		return errResult, nil
	}

	datasets, err := dbn_hist.ListDatasets(config.ApiKey, dateRange)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list datasets: %s", err), nil
	}

	jbytes, err := json.Marshal(datasets)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("list_datasets", "count", len(datasets))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func listSchemasHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataset, err := request.RequireString("dataset")
	if err != nil {
		return mcp.NewToolResultError("dataset must be set"), nil
	}
	dataset = strings.ToUpper(dataset)

	schemas, err := dbn_hist.ListSchemas(config.ApiKey, dataset)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list schemas: %s", err), nil
	}

	jbytes, err := json.Marshal(schemas)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("list_schemas", "dataset", dataset, "count", len(schemas))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func listFieldsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	schemaStr, err := request.RequireString("schema")
	if err != nil {
		return mcp.NewToolResultError("schema must be set"), nil
	}
	schemaStr = strings.ToLower(schemaStr)

	schema, err := dbn.SchemaFromString(schemaStr)
	if err != nil {
		return mcp.NewToolResultErrorf("invalid schema: %s", err), nil
	}

	fields, err := dbn_hist.ListFields(config.ApiKey, dbn.Encoding_Json, schema)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list fields: %s", err), nil
	}

	jbytes, err := json.Marshal(fields)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("list_fields", "schema", schemaStr, "count", len(fields))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func listPublishersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	publishers, err := dbn_hist.ListPublishers(config.ApiKey)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list publishers: %s", err), nil
	}

	// Optional dataset filter
	if dataset, err := request.RequireString("dataset"); err == nil && dataset != "" {
		dataset = strings.ToUpper(dataset)
		filtered := publishers[:0:0]
		for _, p := range publishers {
			if p.Dataset == dataset {
				filtered = append(filtered, p)
			}
		}
		publishers = filtered
	}

	jbytes, err := json.Marshal(publishers)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("list_publishers", "count", len(publishers))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func getDatasetRangeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataset, err := request.RequireString("dataset")
	if err != nil {
		return mcp.NewToolResultError("dataset must be set"), nil
	}
	dataset = strings.ToUpper(dataset)

	dateRange, err := dbn_hist.GetDatasetRange(config.ApiKey, dataset)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get dataset range: %s", err), nil
	}

	jbytes, err := json.Marshal(map[string]any{
		"dataset": dataset,
		"start":   dateRange.Start.Format(time.RFC3339),
		"end":     dateRange.End.Format(time.RFC3339),
	})
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("get_dataset_range", "dataset", dataset, "start", dateRange.Start, "end", dateRange.End)
	return mcp.NewToolResultText(string(jbytes)), nil
}

func getDatasetConditionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataset, err := request.RequireString("dataset")
	if err != nil {
		return mcp.NewToolResultError("dataset must be set"), nil
	}
	dataset = strings.ToUpper(dataset)

	dateRange, errResult := parseOptionalDateRange(request)
	if errResult != nil {
		return errResult, nil
	}

	conditions, err := dbn_hist.GetDatasetCondition(config.ApiKey, dataset, dateRange)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get dataset condition: %s", err), nil
	}

	jbytes, err := json.Marshal(conditions)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("get_dataset_condition", "dataset", dataset, "count", len(conditions))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func listUnitPricesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataset, err := request.RequireString("dataset")
	if err != nil {
		return mcp.NewToolResultError("dataset must be set"), nil
	}
	dataset = strings.ToUpper(dataset)

	unitPrices, err := dbn_hist.ListUnitPrices(config.ApiKey, dataset)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list unit prices: %s", err), nil
	}

	jbytes, err := json.Marshal(unitPrices)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("list_unit_prices", "dataset", dataset, "count", len(unitPrices))
	return mcp.NewToolResultText(string(jbytes)), nil
}

func resolveSymbolsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataset, err := request.RequireString("dataset")
	if err != nil {
		return mcp.NewToolResultError("dataset must be set"), nil
	}
	dataset = strings.ToUpper(dataset)

	symbolsStr, err := request.RequireString("symbols")
	if err != nil {
		return mcp.NewToolResultError("symbols must be set"), nil
	}
	symbols := strings.Split(symbolsStr, ",")
	for i := range symbols {
		symbols[i] = strings.TrimSpace(symbols[i])
	}

	stypeIn := dbn.SType_RawSymbol
	if stypeInStr, err := request.RequireString("stype_in"); err == nil && stypeInStr != "" {
		if stypeIn, err = dbn.STypeFromString(stypeInStr); err != nil {
			return mcp.NewToolResultErrorf("invalid stype_in: %s", err), nil
		}
	}

	stypeOut := dbn.SType_InstrumentId
	if stypeOutStr, err := request.RequireString("stype_out"); err == nil && stypeOutStr != "" {
		if stypeOut, err = dbn.STypeFromString(stypeOutStr); err != nil {
			return mcp.NewToolResultErrorf("invalid stype_out: %s", err), nil
		}
	}

	dateRange, errResult := parseOptionalDateRange(request)
	if errResult != nil {
		return errResult, nil
	}
	if dateRange.Start.IsZero() {
		return mcp.NewToolResultError("start must be set"), nil
	}

	resolution, err := dbn_hist.SymbologyResolve(config.ApiKey, dbn_hist.ResolveParams{
		Dataset:   dataset,
		Symbols:   symbols,
		StypeIn:   stypeIn,
		StypeOut:  stypeOut,
		DateRange: dateRange,
	})
	if err != nil {
		return mcp.NewToolResultErrorf("failed to resolve symbols: %s", err), nil
	}

	jbytes, err := json.Marshal(resolution)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("resolve_symbols", "dataset", dataset, "symbols", symbolsStr,
		"stype_in", stypeIn, "stype_out", stypeOut, "start", dateRange.Start)
	return mcp.NewToolResultText(string(jbytes)), nil
}

func getCostHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	p, errResult := parseCommonParams(request)
	if errResult != nil {
		return errResult, nil
	}

	metaParams := p.metadataQueryParams()

	var (
		cost        float64
		dataSize    int
		recordCount int
		costErr     error
		sizeErr     error
		countErr    error
		wg          sync.WaitGroup
	)
	wg.Add(3)
	go func() { defer wg.Done(); cost, costErr = dbn_hist.GetCost(config.ApiKey, metaParams) }()
	go func() { defer wg.Done(); dataSize, sizeErr = dbn_hist.GetBillableSize(config.ApiKey, metaParams) }()
	go func() { defer wg.Done(); recordCount, countErr = dbn_hist.GetRecordCount(config.ApiKey, metaParams) }()
	wg.Wait()

	if costErr != nil {
		return mcp.NewToolResultErrorf("failed to get cost: %s", costErr), nil
	}
	if sizeErr != nil {
		return mcp.NewToolResultErrorf("failed to get data size: %s", sizeErr), nil
	}
	if countErr != nil {
		return mcp.NewToolResultErrorf("failed to get record count: %s", countErr), nil
	}

	jbytes, err := json.Marshal(map[string]any{
		"query":        metaParams,
		"cost":         cost,
		"data_size":    dataSize,
		"record_count": recordCount,
	})
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal results: %s", err), nil
	}

	logger.Info("get_cost", "dataset", p.Dataset, "schema", p.SchemaStr, "symbol", p.Symbol,
		"start", p.StartStr, "end", p.EndStr, "cost", cost, "size", dataSize, "count", recordCount)

	return mcp.NewToolResultText(string(jbytes)), nil
}

func getRangeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	p, errResult := parseCommonParams(request)
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

	metaParams := p.metadataQueryParams()
	cost, err := dbn_hist.GetCost(config.ApiKey, metaParams)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get cost: %s", err), nil
	}
	if config.MaxCost > 0 && cost > config.MaxCost {
		return mcp.NewToolResultErrorf("query exceeds budget: estimated cost $%.2f, budget $%.2f", cost, config.MaxCost), nil
	}

	jobParams := dbn_hist.SubmitJobParams{
		Dataset:      p.Dataset,
		Symbols:      p.Symbol,
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
	rangeData, err := dbn_hist.GetRange(config.ApiKey, jobParams)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to get range: %s", err), nil
	}

	logger.Info("get_range", "dataset", p.Dataset, "schema", p.SchemaStr, "symbol", p.Symbol,
		"start", p.StartStr, "end", p.EndStr, "cost", cost, "size", len(rangeData))
	return mcp.NewToolResultText(string(rangeData)), nil
}
