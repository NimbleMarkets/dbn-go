// Copyright (C) 2024 Neomantra Corp

package dbn_hist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbn "github.com/NimbleMarkets/dbn-go"
)

// DataBento Batch API:
//  https://databento.com/docs/api-reference-historical/batch/batch-list-files/returns?historical=http&live=python

///////////////////////////////////////////////////////////////////////////////

type JobState string

const (
	JobState_Unknown    JobState = ""
	JobState_Received   JobState = "received"
	JobState_Queued     JobState = "queued"
	JobState_Processing JobState = "processing"
	JobState_Done       JobState = "done"
	JobState_Expired    JobState = "expired"
)

// Returns the string representation of the JobState, or empty string if unknown.
func (j JobState) String() string {
	return string(j)
}

// JobStateFromString converts a string to an JobState.
// Returns an error if the string is unknown.
func JobStateFromString(str string) (JobState, error) {
	str = strings.ToLower(str)
	switch str {
	case "received":
		return JobState_Received, nil
	case "queued":
		return JobState_Queued, nil
	case "processing":
		return JobState_Processing, nil
	case "done":
		return JobState_Done, nil
	case "expired":
		return JobState_Expired, nil
	default:
		return JobState_Unknown, fmt.Errorf("unknown JobState: %s", str)
	}
}

func (j JobState) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(j))
}

func (j *JobState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	js, err := JobStateFromString(str)
	if err != nil {
		return err
	}
	*j = js
	return nil
}

// JobExpiredError is returned when an expired Job has its file listed
type JobExpiredError struct {
	JobID string
}

func (e JobExpiredError) Error() string {
	return fmt.Sprintf("Job %s is expired", e.JobID)
}

///////////////////////////////////////////////////////////////////////////////

type Packaging string

const (
	Packaging_Unknown Packaging = ""
	Packaging_None    Packaging = "none"
	Packaging_Zip     Packaging = "zip"
	Packaging_Tar     Packaging = "tar"
)

// Returns the string representation of the Packaging, or empty string if unknown.
func (p Packaging) String() string {
	return string(p)
}

// PackagingFromString converts a string to an Packaging.
// Returns an error if the string is unknown.
func PackagingFromString(str string) (Packaging, error) {
	str = strings.ToLower(str)
	switch str {
	case "none":
		return Packaging_None, nil
	case "zip":
		return Packaging_Zip, nil
	case "tar":
		return Packaging_Tar, nil
	default:
		return Packaging_Unknown, fmt.Errorf("unknown Packaging: %s", str)
	}
}

func (p Packaging) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(p))
}

func (p *Packaging) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*p = Packaging_None // default
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	js, err := PackagingFromString(str)
	if err != nil {
		return err
	}
	*p = js
	return nil
}

///////////////////////////////////////////////////////////////////////////////

type Delivery string

const (
	Delivery_Unknown  Delivery = ""
	Delivery_Download Delivery = "download"
	Delivery_S3       Delivery = "s3"
	Delivery_Disk     Delivery = "disk"
)

// Returns the string representation of the Delivery, or empty string if unknown.
func (d Delivery) String() string {
	return string(d)
}

// DeliveryFromString converts a string to an Delivery.
// Returns an error if the string is unknown.
func DeliveryFromString(str string) (Delivery, error) {
	str = strings.ToLower(str)
	switch str {
	case "download":
		return Delivery_Download, nil
	case "s3":
		return Delivery_S3, nil
	case "disk":
		return Delivery_Disk, nil
	default:
		return Delivery_Unknown, fmt.Errorf("unknown Delivery: %s", str)
	}
}

func (d Delivery) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(d))
}

func (d *Delivery) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*d = Delivery_Download // default
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	js, err := DeliveryFromString(str)
	if err != nil {
		return err
	}
	*d = js
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// BatchJob is the description of a submitted batch job.
type BatchJob struct {
	Id             string          `json:"id"`                         // The unique job ID.
	UserID         *string         `json:"user_id"`                    // The user ID of the user who submitted the job.
	BillID         *string         `json:"bill_id"`                    // The bill ID (for internal use).
	CostUSD        *float64        `json:"cost_usd"`                   // The cost of the job in US dollars. Will be `None` until the job is processed.
	Dataset        string          `json:"dataset"`                    // The dataset code.
	Symbols        string          `json:"symbols"`                    // The CSV list of symbols specified in the request.
	StypeIn        dbn.SType       `json:"stype_in"`                   // The symbology type of the input `symbols`.
	StypeOut       dbn.SType       `json:"stype_out"`                  // The symbology type of the output `symbols`.
	Schema         dbn.Schema      `json:"schema"`                     // The data record schema.
	Start          time.Time       `json:"start"`                      // The start of the request time range (inclusive).
	End            time.Time       `json:"end"`                        // The end of the request time range (exclusive).
	Limit          uint64          `json:"limit"`                      // The maximum number of records to return.
	Encoding       dbn.Encoding    `json:"encoding"`                   // The data encoding.
	Compression    dbn.Compression `json:"compression"`                // The data compression mode.
	PrettyPx       bool            `json:"pretty_px"`                  // If prices are formatted to the correct scale (using the fixed-precision scalar 1e-9).
	PrettyTs       bool            `json:"pretty_ts"`                  // If timestamps are formatted as ISO 8601 strings.
	MapSymbols     bool            `json:"map_symbols"`                // If a symbol field is included with each text-encoded record.
	SplitSymbols   bool            `json:"split_symbols"`              // If files are split by raw symbol.
	SplitDuration  string          `json:"split_duration"`             // The maximum time interval for an individual file before splitting into multiple files.
	SplitSize      uint64          `json:"split_size,omitempty"`       // The maximum size for an individual file before splitting into multiple files.
	Packaging      Packaging       `json:"packaging,omitempty"`        // The packaging method of the batch data.
	Delivery       Delivery        `json:"delivery"`                   // The delivery mechanism of the batch data.
	RecordCount    uint64          `json:"record_count"`               // The number of data records (`None` until the job is processed).
	BilledSize     uint64          `json:"billed_size"`                // The size of the raw binary data used to process the batch job (used for billing purposes).
	ActualSize     uint64          `json:"actual_size"`                // The total size of the result of the batch job after splitting and compression.
	PackageSize    uint64          `json:"package_size"`               // The total size of the result of the batch job after any packaging (including metadata).
	State          JobState        `json:"state"`                      // The current status of the batch job.
	TsReceived     time.Time       `json:"ts_received,omitempty"`      //The timestamp of when Databento received the batch job.
	TsQueued       time.Time       `json:"ts_queued,omitempty"`        // The timestamp of when the batch job was queued.
	TsProcessStart time.Time       `json:"ts_process_start,omitempty"` // The timestamp of when the batch job began processing.
	TsProcessDone  time.Time       `json:"ts_process_done,omitempty"`  // The timestamp of when the batch job finished processing.
	TsExpiration   time.Time       `json:"ts_expiration,omitempty"`    // The timestamp of when the batch job will expire from the Download center.
}

// GetRangeParams is the parameters for [`TimeseriesClient::get_range()`].
// Use NewGetRangeParams to get an instance type with all the preset defaults.
type GetRangeParams struct {
	Dataset       string                   `json:"dataset"`         // The dataset code.
	Symbols       []string                 `json:"symbols"`         // The symbols to filter for.
	Schema        dbn.Schema               `json:"schema"`          // The data record schema.
	DateTimeRange DateRange                `json:"date_time_range"` // The request time range.
	StypeIn       dbn.SType                `json:"stype_in"`        // The symbology type of the input `symbols`. Defaults to [`RawSymbol`](dbn::enums::SType::RawSymbol).
	StypeOut      dbn.SType                `json:"stype_out"`       // The symbology type of the output `symbols`. Defaults to [`InstrumentId`](dbn::enums::SType::InstrumentId).
	Limit         uint64                   `json:"limit"`           // The optional maximum number of records to return. Defaults to no limit.
	UpgradePolicy dbn.VersionUpgradePolicy `json:"upgrade_policy"`  // How to decode DBN from prior versions. Defaults to upgrade.
}

// The file details for a batch job.
type BatchFileDesc struct {
	Filename string            `json:"filename"` // The file name.
	Size     uint64            `json:"size"`     // The size of the file in bytes.
	Hash     string            `json:"hash"`     // The SHA256 hash of the file.
	Urls     map[string]string `json:"urls"`     // A map of download protocol to URL.
}

// SubmitJobParams are te parameters for [`BatchClient::submit_job()`]. Use [`SubmitJobParams::builder()`] to
// get a builder type with all the preset defaults.
type SubmitJobParams struct {
	Dataset       string          `json:"dataset"`              // The dataset code.
	Symbols       string          `json:"symbols"`              // The symbols to filter for.
	Schema        dbn.Schema      `json:"schema"`               // The data record schema.
	DateRange     DateRange       `json:"date_time_range"`      // The request time range.
	Encoding      dbn.Encoding    `json:"encoding"`             // The data encoding. Defaults to [`Dbn`](Encoding::Dbn).
	Compression   dbn.Compression `json:"compression"`          // The data compression mode. Defaults to [`ZStd`](Compression::ZStd).
	PrettyPx      bool            `json:"pretty_px"`            // If `true`, prices will be formatted to the correct scale (using the fixed-  precision scalar 1e-9). Only valid for [`Encoding::Csv`] and [`Encoding::Json`].
	PrettyTs      bool            `json:"pretty_ts"`            // If `true`, timestamps will be formatted as ISO 8601 strings. Only valid for [`Encoding::Csv`] and [`Encoding::Json`].
	MapSymbols    bool            `json:"map_symbols"`          // If `true`, a symbol field will be included with each text-encoded record, reducing the need to look at the `symbology.json`. Only valid for [`Encoding::Csv`] and [`Encoding::Json`].
	SplitSymbols  bool            `json:"split_symbols"`        // If `true`, files will be split by raw symbol. Cannot be requested with [`Symbols::All`].
	SplitDuration string          `json:"split_duration"`       // The maximum time duration before batched data is split into multiple files. Defaults to [`Day`](SplitDuration::Day).
	SplitSize     uint64          `json:"split_size,omitempty"` // The optional maximum size (in bytes) of each batched data file before being split. Defaults to `None`.
	Packaging     Packaging       `json:"packaging,omitempty"`  // The optional archive type to package all batched data files in. Defaults to `None`.
	Delivery      Delivery        `json:"delivery"`             // The delivery mechanism for the batched data files once processed. Defaults to [`Download`](Delivery::Download).
	StypeIn       dbn.SType       `json:"stype_in"`             // The symbology type of the input `symbols`. Defaults to [`RawSymbol`](dbn::enums::SType::RawSymbol).
	StypeOut      dbn.SType       `json:"stype_out"`            // The symbology type of the output `symbols`. Defaults to [`InstrumentId`](dbn::enums::SType::InstrumentId).
	Limit         uint64          `json:"limit,omitempty"`      // The optional maximum number of records to return. Defaults to no limit.
}

// ApplyToURLValues fills a url.Values with the SubmitJobParams per the Batch API spec.
// Returns any error.
func (jobParams *SubmitJobParams) ApplyToURLValues(params *url.Values) error {
	if jobParams.Dataset == "" {
		return fmt.Errorf("missing required Dataset")
	} else {
		params.Add("dataset", jobParams.Dataset)
	}
	if jobParams.Symbols == "" {
		return fmt.Errorf("missing required Symbols CSV or specify 'ALL_SYMBOLS'")
	} else {
		params.Add("symbols", jobParams.Symbols)
	}
	if schemaStr := jobParams.Schema.String(); schemaStr == "" {
		return fmt.Errorf("missing required Schema")
	} else {
		params.Add("schema", schemaStr)
	}
	if jobParams.DateRange.Start.IsZero() {
		return errors.New("DateRange.Start is required")
	} else {
		params.Add("start", jobParams.DateRange.Start.Format("2006-01-02"))
	}
	if !jobParams.DateRange.End.IsZero() {
		params.Add("end", jobParams.DateRange.End.Format("2006-01-02"))
	}
	if encodingStr := jobParams.Encoding.String(); encodingStr == "" {
		return fmt.Errorf("missing require Encoding")
	} else {
		params.Add("encoding", encodingStr)
	}
	if compressStr := jobParams.Compression.String(); compressStr != "" {
		params.Add("compression", compressStr)
	}
	if jobParams.PrettyPx {
		params.Add("pretty_px", "true")
	}
	if jobParams.PrettyTs {
		params.Add("pretty_ts", "true")
	}
	if jobParams.MapSymbols {
		params.Add("map_symbols", "true")
	}
	if jobParams.SplitSymbols {
		params.Add("split_symbols", "true")
	}
	if jobParams.SplitDuration != "" {
		params.Add("split_duration", jobParams.SplitDuration)
	}
	if jobParams.SplitSize > 0 {
		params.Add("split_size", strconv.Itoa(int(jobParams.SplitSize)))
	}
	if packagingStr := jobParams.Packaging.String(); packagingStr != "" {
		params.Add("packaging", packagingStr)
	}
	if deliveryStr := jobParams.Delivery.String(); deliveryStr != "" {
		params.Add("delivery", deliveryStr)
	}
	if stypeStr := jobParams.StypeIn.String(); stypeStr != "" {
		params.Add("stype_in", stypeStr)
	}
	if stypeStr := jobParams.StypeOut.String(); stypeStr != "" {
		params.Add("stype_out", stypeStr)
	}
	if jobParams.Limit > 0 {
		params.Add("limit", strconv.Itoa(int(jobParams.Limit)))
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// Lists all jobs associated with the given state filter and 'since' date.
//
// # Errors
// This function returns an error when it fails to communicate with the Databento API
// or the API indicates there's an issue with the request.
func ListJobs(apiKey string, stateFilter string, sinceYMD time.Time) ([]BatchJob, error) {
	apiUrl := "https://hist.databento.com/v0/batch.list_jobs"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if stateFilter != "" {
		params.Add("states", stateFilter)
	}
	if !sinceYMD.IsZero() {
		params.Add("since", sinceYMD.UTC().Format("2006-01-02"))
	}
	baseUrl.RawQuery = params.Encode()

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		return nil, err
	}

	var batchJobs []BatchJob
	err = json.Unmarshal(body, &batchJobs)
	if err != nil {
		return nil, err
	}
	return batchJobs, nil
}

// Lists all files associated with the batch job with ID `jobID`.
// Returns JobExpiredError if the response indicates the job has expired.
//
// # Errors
// This function returns an error when it fails to communicate with the Databento API
// or the API indicates there's an issue with the request.
func ListFiles(apiKey string, jobID string) ([]BatchFileDesc, error) {
	apiUrl := "https://hist.databento.com/v0/batch.list_files"
	baseUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("job_id", jobID)
	baseUrl.RawQuery = params.Encode()

	body, err := databentoGetRequest(baseUrl.String(), apiKey)
	if err != nil {
		if errStr := err.Error(); strings.HasPrefix(errStr, "HTTP 410") {
			return nil, JobExpiredError{JobID: jobID}
		}
		return nil, err
	}

	var batchFiles []BatchFileDesc
	err = json.Unmarshal(body, &batchFiles)
	if err != nil {
		return nil, err
	}
	return batchFiles, nil
}

///////////////////////////////////////////////////////////////////////////////

// SubmitJob submits a new batch job and returns a description and identifiers for the job.
//
// # Errors
// This function returns an error when it fails to communicate with the Databento API
// or the API indicates there's an issue with the request.
func SubmitJob(apiKey string, jobParams SubmitJobParams) (*BatchJob, error) {
	apiUrl := "https://hist.databento.com/v0/batch.submit_job"

	formData := url.Values{}
	err := jobParams.ApplyToURLValues(&formData)
	if err != nil {
		return nil, fmt.Errorf("bad params: %w", err)
	}

	body, err := databentoPostFormRequest(apiUrl, apiKey, formData, "")
	if err != nil {
		return nil, fmt.Errorf("failed post request: %w", err)
	}

	var batchJob BatchJob
	err = json.Unmarshal(body, &batchJob)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &batchJob, nil
}
