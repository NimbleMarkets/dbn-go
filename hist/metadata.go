// Copyright (c) 2024 Neomantra Corp

package dbn_hist

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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

// Returns the string representation of the FeedMode, or empty string if unknown.
func (f FeedMode) String() string {
	switch f {
	case FeedMode_Historical:
		return "historical"
	case FeedMode_HistoricalStreaming:
		return "historical-streaming"
	case FeedMode_Live:
		return "live"
	}
	return ""
}

// FeedModeFromString converts a string to an FeedMode.
// Returns an error if the string is unknown.
// Possible string values: historical, historical-streaming, live.
func FeedModeFromString(str string) (FeedMode, error) {
	str = strings.ToLower(str)
	switch str {
	case "historical":
		return FeedMode_Historical, nil
	case "historical-streaming":
		return FeedMode_HistoricalStreaming, nil
	case "live":
		return FeedMode_Live, nil
	default:
		return FeedMode_Historical, fmt.Errorf("unknown feedMode: %s", str)
	}
}

func (f FeedMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f *FeedMode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	js, err := FeedModeFromString(str)
	if err != nil {
		return err
	}
	*f = js
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// MetadataQueryParams is the common request structure for several Metadata API queries
type MetadataQueryParams struct {
	// The dataset code.
	Dataset   string    `json:"dataset,omitempty"`  // The dataset code. Must be one of the values from ListDatasets.
	Symbols   []string  `json:"symbols"`            // The product symbols to filter for. Takes up to 2,000 symbols per request. If `ALL_SYMBOLS` or not specified then will be for all symbols.
	Schema    string    `json:"schema,omitempty"`   // The data record schema. Must be one of the values from ListSchemas.
	DateRange DateRange `json:"date_range"`         // The date range of the request to get the cost for.
	Mode      FeedMode  `json:"mode,omitempty"`     // The data feed mode of the request. Must be one of 'historical', 'historical-streaming', or 'live'.
	StypeIn   dbn.SType `json:"stype_in,omitempty"` // The symbology type of input symbols. Must be one of 'raw_symbol', 'instrument_id', 'parent', or 'continuous'.
	Limit     int32     `json:"limit,omitempty"`    // The maximum number of records to return. 0 and negative (-1) means no limit.
}

// ApplyToURLValues fills a url.Values with the MetadataQueryParams per the Metadata API spec.
// Returns any error.
func (metaParams *MetadataQueryParams) ApplyToURLValues(params *url.Values) error {
	params.Add("dataset", metaParams.Dataset)
	params.Add("schema", metaParams.Schema)
	params.Add("mode", metaParams.Mode.String())
	params.Add("stype_in", metaParams.StypeIn.String())
	if metaParams.Limit > 0 {
		params.Add("limit", strconv.Itoa(int(metaParams.Limit)))
	}

	if metaParams.DateRange.Start.IsZero() {
		return errors.New("DateRange.Start must be defined")
	} else {
		params.Add("start", metaParams.DateRange.Start.Format("2006-01-02"))
	}
	if !metaParams.DateRange.End.IsZero() {
		params.Add("end", metaParams.DateRange.End.Format("2006-01-02"))
	}

	csvSymbols := strings.Join(metaParams.Symbols, ",")
	params.Add("symbols", csvSymbols)

	return nil
}

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
	Mode FeedMode `json:"mode,omitempty"`
	/// The unit prices in US dollars by data record schema.
	UnitPrices map[string]float64 `json:"unit_prices,omitempty"`
}

const (
	Condition_Available = "available"
	Condition_Degraded  = "degraded"
	Condition_Pending   = "pending"
	Condition_Missing   = "missing"
)

type ConditionDetail struct {
	Date         string // The day of the described data, as an ISO 8601 date string
	Condition    string // The condition code describing the quality and availability of the data on the given day. Possible values are ConditionKind.
	LastModified string // The date when any schema in the dataset on the given day was last generated or modified, as an ISO 8601 date string.
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

///////////////////////////////////////////////////////////////////////////////

// Calls the Metadata API to get the condition of a dataset and date range.
// Returns ConditionDetails, or an error if any.
// Passing a zero time for Start is the beginning of the Dataset.
// Passing a zero time for End is the end of the Dataset.
func GetDatasetCondition(apiKey string, dataset string, dateRange DateRange) ([]ConditionDetail, error) {
	apiUrl := "https://hist.databento.com/v0/metadata.get_dataset_condition"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("dataset", dataset)
	if !dateRange.Start.IsZero() {
		params.Add("start_date", dateRange.Start.Format("2006-01-02"))
	}
	if !dateRange.End.IsZero() {
		params.Add("end_date", dateRange.End.Format("2006-01-02"))
	}
	baseUrl.RawQuery = params.Encode()

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		return nil, err
	}

	var ConditionDetails []ConditionDetail
	err = json.Unmarshal(body, &ConditionDetails)
	if err != nil {
		return nil, err
	}
	return ConditionDetails, nil
}

///////////////////////////////////////////////////////////////////////////////

// Calls the Metadata API to get the date range of a dataset.
// Returns the DateRage, or an error if any.
func GetDatasetRange(apiKey string, dataset string) (DateRange, error) {
	apiUrl := "https://hist.databento.com/v0/metadata.get_dataset_range"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return DateRange{}, err
	}

	params := url.Values{}
	params.Add("dataset", dataset)
	baseUrl.RawQuery = params.Encode()

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		return DateRange{}, err
	}

	var dateRange DateRange
	err = json.Unmarshal(body, &dateRange)
	if err != nil {
		return DateRange{}, err
	}
	return dateRange, nil
}

///////////////////////////////////////////////////////////////////////////////

// Calls the Metadata API to get the record count of a GetRange query.
// Returns the record count or an error if any.
func GetRecordCount(apiKey string, metaParams MetadataQueryParams) (int, error) {
	apiUrl := "https://hist.databento.com/v0/metadata.get_record_count"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return -1, err
	}

	params := url.Values{}
	err = metaParams.ApplyToURLValues(&params)
	if err != nil {
		return 0, fmt.Errorf("bad params: %w", err)
	}
	baseUrl.RawQuery = params.Encode()

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		return 0, fmt.Errorf("failed get request: %w", err)
	}

	// convert from text int
	recordCount, err := strconv.Atoi(string(body))
	if err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}
	return recordCount, nil
}

// Calls the Metadata API to get the billable size of a GetRange query.
// Returns the billable size or an error if any.
func GetBillableSize(apiKey string, metaParams MetadataQueryParams) (int, error) {
	apiUrl := "https://hist.databento.com/v0/metadata.get_billable_size"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return -1, err
	}

	params := url.Values{}
	err = metaParams.ApplyToURLValues(&params)
	if err != nil {
		return 0, fmt.Errorf("bad params: %w", err)
	}
	baseUrl.RawQuery = params.Encode()

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		return 0, fmt.Errorf("failed get request: %w", err)
	}

	// convert from text int
	billableSize, err := strconv.Atoi(string(body))
	if err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}
	return billableSize, nil
}

// Calls the Metadata API to get the cost estimate of a GetRange query.
// Returns the cost or an error if any.
func GetCost(apiKey string, metaParams MetadataQueryParams) (float64, error) {
	apiUrl := "https://hist.databento.com/v0/metadata.get_cost"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return -1, err
	}

	params := url.Values{}
	err = metaParams.ApplyToURLValues(&params)
	if err != nil {
		return 0, fmt.Errorf("bad params: %w", err)
	}
	baseUrl.RawQuery = params.Encode()

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		return 0, fmt.Errorf("failed get request: %w", err)
	}

	// convert from float
	cost, err := strconv.ParseFloat(string(body), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}
	return cost, nil
}
