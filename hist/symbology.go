// Copyright (c) 2024 Neomantra Corp

package dbn_hist

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/NimbleMarkets/dbn-go"
)

///////////////////////////////////////////////////////////////////////////////

// https://github.com/databento/databento-rs/blob/main/src/historical/symbology.rs
// The parameters for [`SymbologyClient::resolve()`].
// Use [`ResolveParams::builder()`] to get a builder type with all the preset defaults.
type ResolveParams struct {
	// The dataset code.
	Dataset string `json:"dataset"`
	// The symbols to resolve.
	Symbols []string `json:"symbols"`
	// The symbology type of the input `symbols`. Defaults to SType_RawSymbol
	StypeIn dbn.SType `json:"stype_in"`
	// The symbology type of the output `symbols`. Defaults to SType_InstrumentId
	StypeOut dbn.SType `json:"stype_out"`
	// The date range of the resolution.
	DateRange DateRange `json:"date_range"`
}

// The resolved symbol for a date range.
type MappingInterval struct {
	StartDate string `json:"d0"` // The UTC start date of interval (inclusive).
	EndDate   string `json:"d1"` // The UTC end date of interval (exclusive).
	Symbol    string `json:"s"`  // The resolved symbol for this interval.
}

// A symbology resolution from one symbology type to another.
type Resolution struct {
	// A mapping from input symbol to a list of resolved symbols in the output symbology.
	Mappings map[string][]MappingInterval `json:"result"`
	// A list of symbols that were resolved for part, but not all of the date range from the request.
	Partial []string `json:"partial"`
	// A list of symbols that were not resolved.
	NotFound []string `json:"not_found"`
	// The input symbology type, in string form.
	StypeIn dbn.SType `json:"stype_in"`
	// The output symbology type, in string form.
	StypeOut dbn.SType `json:"stype_out"`
	// A message from the server.
	Message string `json:"message"`
	// The status code of the response.
	Status int `json:"status"`
}

///////////////////////////////////////////////////////////////////////////////

func SymbologyResolve(apiKey string, params ResolveParams) (*Resolution, error) {
	apiUrl := "https://hist.databento.com/v0/symbology.resolve"

	csvSymbols := strings.Join(params.Symbols, ",")
	formData := url.Values{
		"dataset":   {params.Dataset},
		"symbols":   {csvSymbols},
		"stype_in":  {params.StypeIn.String()},
		"stype_out": {params.StypeOut.String()},
	}
	if params.DateRange.Start.IsZero() {
		return nil, fmt.Errorf("DateRange.Start is required")
	} else {
		formData.Add("start_date", params.DateRange.Start.Format("2006-01-02"))
	}
	if !params.DateRange.End.IsZero() {
		formData.Add("end_date", params.DateRange.End.Format("2006-01-02"))
	}

	body, err := databentoPostFormRequest(apiUrl, apiKey, formData, "")
	if err != nil {
		return nil, err
	}

	var resp Resolution
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
