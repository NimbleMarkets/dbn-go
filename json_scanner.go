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
// This a plain function (not a method) because methods cannot be generic.
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
	err = header.Fill_Json(val.Get("hd"))
	if err != nil {
		return nil, nil, err
	}
	return val, &header, err
}

func dispatchJsonVisitor(val *fastjson.Value, header *RHeader, visitor Visitor) error {
	switch header.RType {
	// Trade
	case RType_Mbp0:
		record := Mbp0Msg{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnMbp0(&record)
		}
	// Mbp1
	case RType_Mbp1:
		record := Mbp1Msg{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnMbp1(&record)
		}
	// Mbp10
	case RType_Mbp10:
		record := Mbp10Msg{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnMbp10(&record)
		}
	// Mbo
	case RType_Mbo:
		record := MboMsg{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnMbo(&record)
		}
	// Candlestick schemas
	case RType_Ohlcv1S, RType_Ohlcv1M, RType_Ohlcv1H, RType_Ohlcv1D, RType_OhlcvEod, RType_OhlcvDeprecated:
		record := OhlcvMsg{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnOhlcv(&record)
		}
	// CBBO schemas
	case RType_Cbbo, RType_Cbbo1S, RType_Cbbo1M, RType_Tcbbo:
		record := CbboMsg{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnCbbo(&record)
		}
	// Imbalance
	case RType_Imbalance:
		record := ImbalanceMsg{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnImbalance(&record)
		}
	// Statistics
	case RType_Statistics:
		record := StatMsg{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnStatMsg(&record)
		}
	// SymbolMapping
	case RType_SymbolMapping:
		record := SymbolMappingMsg{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnSymbolMappingMsg(&record)
		}
	// System
	case RType_System:
		record := SystemMsg{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnSystemMsg(&record)
		}
	// Error
	case RType_Error:
		record := ErrorMsg{}
		if err := record.Fill_Json(val, header); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnErrorMsg(&record)
		}
	// Unknown
	default:
		return ErrUnknownRType
	}
	// RType_Status
	// RType_InstrumentDef
}

///////////////////////////////////////////////////////////////////////////////

// ReadJsonToSlice reads the entire stream from a JSONL stream of DBN records.
// It will scan for type R (for example Mbp0) and decode it into a slice of R.
// Returns the slice and any error.
// Example:
//
//	fileReader, err := os.Open(dbnFilename)
//	records, err := dbn.ReadJsonToSlice[dbn.Mbp0Msg](fileReader)
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
