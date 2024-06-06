// Copyright (C) 2024 Neomantra Corp

package dbn_hist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
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
		*p = Packaging_None
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
		params.Add("state", stateFilter)
	}
	if !sinceYMD.IsZero() {
		params.Add("since", sinceYMD.Format("2006-01-02"))
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
		return nil, err
	}

	var batchFiles []BatchFileDesc
	err = json.Unmarshal(body, &batchFiles)
	if err != nil {
		return nil, err
	}
	return batchFiles, nil
}
