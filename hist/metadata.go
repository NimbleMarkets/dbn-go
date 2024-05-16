// Copyright (c) 2024 Neomantra Corp

package dbn_hist

import (
	"encoding/json"
	"net/url"

	"github.com/NimbleMarkets/dbn-go"
)

// A type of data feed.
type FeedMode uint8

const (
	// The historical batch data feed.
	FeedMode_Historical FeedMode = 0
	/// The historical streaming data feed.
	FeedMode_HistoricalStreaming FeedMode = 1
	/// The Live data feed for real-time and intraday historical.
	FeedMode_Live FeedMode = 2
)

// The details about a publisher.
type PublisherDetail struct {
	// The publisher ID assigned by Databento, which denotes the dataset and venue.
	PublisherID uint16 `json:"publisher_id,omitempty"`
	// The dataset code for the publisher.
	Dataset string `json:"dataset,omitempty"`
	// The venue for the publisher.
	Venue string `json:"venue,omitempty"`
	// The publisher description.
	Description string `json:"description,omitempty"`
}

// The details about a field in a schema.
type FieldDetail struct {
	// The field name.
	Name string `json:"name,omitempty"`
	// The field type name.
	TypeName string `json:"type_name,omitempty"`
}

// The unit prices for a particular [`FeedMode`].
type UnitPricesForMode struct {
	/// The data feed mode.
	Mode string `json:"mode,omitempty"`
	/// The unit prices in US dollars by data record schema.
	UnitPrices map[string]float64 `json:"unit_prices,omitempty"`
}

//////////////////////////////////////////////////////////////////////////////

// Lists the details of all publishers.
//
// # Errors
// This function returns an error when it fails to communicate with the Databento API.
func ListPublishers(apiKey string) ([]PublisherDetail, error) {
	apiUrl := "https://hist.databento.com/v0/metadata.list_publishers"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		return nil, err
	}

	var publisherDetail []PublisherDetail
	err = json.Unmarshal(body, &publisherDetail)
	if err != nil {
		return nil, err
	}
	return publisherDetail, nil
}

// Lists all available dataset codes on Databento.
//
// # Errors
// This function returns an error when it fails to communicate with the Databento API
// or the API indicates there's an issue with the request.
func ListDatasets(apiKey string, dateRange DateRange) ([]string, error) {
	apiUrl := "https://hist.databento.com/v0/metadata.list_datasets"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if !dateRange.Start.IsZero() {
		params.Add("start_date", dateRange.Start.Format("2006-01-02"))
	}
	if !dateRange.Start.IsZero() {
		params.Add("end_date", dateRange.Start.Format("2006-01-02"))
	}
	baseUrl.RawQuery = params.Encode()

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		return nil, err
	}

	var res []string
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Lists all available schemas for the given `dataset`.
//
// # Errors
// This function returns an error when it fails to communicate with the Databento API
// or the API indicates there's an issue with the request.
func ListSchemas(apiKey string, dataset string) ([]string, error) {
	apiUrl := "https://hist.databento.com/v0/metadata.list_schemas"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("dataset", dataset)
	baseUrl.RawQuery = params.Encode()

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		return nil, err
	}

	var res []string
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Lists all fields for a schema and encoding.
//
// # Errors
// This function returns an error when it fails to communicate with the Databento API
// or the API indicates there's an issue with the request.
func ListFields(apiKey string, encoding dbn.Encoding, schema dbn.Schema) ([]FieldDetail, error) {
	apiUrl := "https://hist.databento.com/v0/metadata.list_fields"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("encoding", encoding.String())
	params.Add("schema", schema.String())
	baseUrl.RawQuery = params.Encode()

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		return nil, err
	}

	var fieldDetail []FieldDetail
	err = json.Unmarshal(body, &fieldDetail)
	if err != nil {
		return nil, err
	}
	return fieldDetail, nil
}

// Lists unit prices for each data schema and feed mode in US dollars per gigabyte.
//
// # Errors
// This function returns an error when it fails to communicate with the Databento API
// or the API indicates there's an issue with the request.
func ListUnitPrices(apiKey string, dataset string) ([]UnitPricesForMode, error) {
	apiUrl := "https://hist.databento.com/v0/metadata.list_unit_prices"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("dataset", dataset)
	baseUrl.RawQuery = params.Encode()

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		return nil, err
	}

	var unitPricesForModes []UnitPricesForMode
	err = json.Unmarshal(body, &unitPricesForModes)
	if err != nil {
		return nil, err
	}
	return unitPricesForModes, nil
}

// TODO:
//   get_dataset_condition
//   get_dataset_range
//   get_record_count
//   get_billable_size
//   get_cost
