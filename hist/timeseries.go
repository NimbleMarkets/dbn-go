// Copyright (C) 2024 Neomantra Corp

package dbn_hist

import (
	"fmt"
	"net/url"
)

// DataBento Time Series API:
//  https://databento.com/docs/api-reference-historical/timeseries/timeseries-get-range?historical=http&live=python

///////////////////////////////////////////////////////////////////////////////

// GetRange makes a streaming request for timeseries data from Databento.
//
// This method returns the byte array of the DBN stream.
//
// # Errors
// This function returns an error when it fails to communicate with the Databento API
// or the API indicates there's an issue with the request.
func GetRange(apiKey string, jobParams SubmitJobParams) ([]byte, error) {
	apiUrl := "https://hist.databento.com/v0/timeseries.get_range"

	formData := url.Values{}
	err := jobParams.ApplyToURLValues(&formData)
	if err != nil {
		return nil, fmt.Errorf("bad params: %w", err)
	}

	body, err := databentoPostFormRequest(apiUrl, apiKey, formData, "application/octet-stream")
	if err != nil {
		return nil, fmt.Errorf("failed post request: %w", err)
	}

	return body, nil
}
