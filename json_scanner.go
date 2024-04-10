// Copyright (c) 2024 Neomantra Corp

package dbn

import (
	"bufio"
	"io"

	"github.com/valyala/fastjson"
)

///////////////////////////////////////////////////////////////////////////////

// JsonScanner scans a series of DBN JSON values.  Delimited by whitespace (generally newlines)
type JsonScanner struct {
	scanner *bufio.Scanner
}

// NewJsonScanner creates a new dbn.JsonScanner from a byte array
func NewJsonScanner(r io.Reader) *JsonScanner {
	return &JsonScanner{
		scanner: bufio.NewScanner(r),
	}
}

// Next parses the next JSON value from the data
// Returns true on success. The parsed Envelope is available via Envelope call.
// Returns false either on error or on the end of data. Call Error() in order to determine the cause of the returned false.
func (s *JsonScanner) Next() bool {
	return s.scanner.Scan()
}

// Error returns the last error from Next().
func (s *JsonScanner) Error() error {
	return s.scanner.Err()
}

// Parses the Scanner's current record as a `Record`.
func JsonScannerDecode[R Record, RP RecordPtr[R]](s *JsonScanner) (*R, error) {
	val, header, err := s.parseWithHeader()
	if err != nil {
		return nil, err
	}

	var rp RP = new(R)

	if !header.RType.IsCompatibleWith(rp.RType()) {
		return nil, unexpectedRTypeError(header.RType, rp.RType())
	}

	if err := rp.Fill_Json(val, header); err != nil {
		return nil, err
	} else {
		return rp, nil
	}
}

// Parses the current Record and passes it to the Visitor.
func (s *JsonScanner) Visit(visitor Visitor) error {
	val, header, err := s.parseWithHeader()
	if err != nil {
		// TODO: EOF -> OnStreamEnd
		return err
	}
	return dispatchJsonVisitor(val, header, visitor)
}

///////////////////////////////////////////////////////////////////////////////

func (s *JsonScanner) parseWithHeader() (*fastjson.Value, *RHeader, error) {
	var p fastjson.Parser
	val, err := p.ParseBytes(s.scanner.Bytes())
	if err != nil {
		return nil, nil, err
	}

	var header RHeader
	err = FillRHeader_Json(val.Get("hd"), &header)
	if err != nil {
		return nil, nil, err
	}
	return val, &header, err
}

func dispatchJsonVisitor(val *fastjson.Value, header *RHeader, visitor Visitor) error {
	switch header.RType {
	case RType_Mbp0: // Trade
		record := Mbp0{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnMbp0(&record)
		}
	// Candlestick schemas
	case RType_Ohlcv1S, RType_Ohlcv1M, RType_Ohlcv1H, RType_Ohlcv1D, RType_OhlcvEod, RType_OhlcvDeprecated:
		record := Ohlcv{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnOhlcv(&record)
		}
	// Imbalance
	case RType_Imbalance:
		record := Imbalance{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnImbalance(&record)
		}

	default:
		return ErrUnknownRType
	}
	// RType_Mbp1
	// RType_Mbp10
	// RType_Status
	// RType_InstrumentDef
	// RType_Error
	// RType_SymbolMapping
	// RType_System
	// RType_Statistics
	// RType_Mbo
}

///////////////////////////////////////////////////////////////////////////////

// ReadJsonToSlice reads the entire stream from a JSONL stream of DBN records.
// It will scan for type R (for example Mbp0) and decode it into a slice of R.
// Returns the slice and any error.
// Example:
//
//	fileReader, err := os.Open(dbnFilename)
//	records, err := dbn.ReadJsonToSlice[dbn.Mbp0](fileReader)
func ReadJsonToSlice[R Record, RP RecordPtr[R]](reader io.Reader) ([]R, error) {
	records := make([]R, 0)
	scanner := NewJsonScanner(reader)
	for scanner.Next() {
		r, err := JsonScannerDecode[R, RP](scanner)
		if err != nil {
			return records, err
		}
		records = append(records, *r)
	}
	return records, scanner.Error()
}
