// Copyright (c) 2024 Neomantra Corp

package dbn

import (
	"bufio"
	"io"
)

///////////////////////////////////////////////////////////////////////////////

// Default buffer size for decoding
const DEFAULT_DECODE_BUFFER_SIZE = 16 * 1024
const DEFAULT_SCRATCH_BUFFER_SIZE = 512 // bigger than largest record size

// DbnScanner scans a raw DBN stream
type DbnScanner struct {
	srcReader  io.Reader     // the source we pull data from
	buffReader *bufio.Reader // the buffer reader we scan over
	metadata   *Metadata     // the metadata for the stream
	lastError  error         // the last error encountered
	lastRecord []byte        // last record read, waiting for decode
	lastSize   int           // the size of the last record read
}

// NewDbnScanner creates a new dbn.DbnScanner
func NewDbnScanner(sourceReader io.Reader) *DbnScanner {
	return &DbnScanner{
		srcReader:  sourceReader,
		buffReader: bufio.NewReaderSize(sourceReader, DEFAULT_DECODE_BUFFER_SIZE),
		metadata:   nil,
		lastError:  nil,
		lastRecord: make([]byte, DEFAULT_SCRATCH_BUFFER_SIZE),
		lastSize:   0,
	}
}

/////////////////////////////////////////////////////////////////////////////

// Metadata returns the metadata for the stream, or nil if none.
// May try to read the metadata, which may result in an error.
func (s *DbnScanner) Metadata() (*Metadata, error) {
	if s.metadata != nil {
		return s.metadata, nil
	}
	err := s.readMetadata()
	return s.metadata, err
}

// readMetadata is an internal method to read metadata from the stream.
func (s *DbnScanner) readMetadata() error {
	if s.metadata != nil {
		return nil
	}
	m, err := ReadMetadata(s.buffReader)
	if err != nil {
		s.lastError = err
		s.lastSize = 0
		return err
	}
	s.lastError = nil
	s.lastSize = 0
	s.metadata = m
	return nil
}

// Next parses the next record from the stream
func (s *DbnScanner) Next() bool {
	// Read the metadata if we haven't already
	if s.metadata == nil {
		if err := s.readMetadata(); err != nil {
			s.lastError = err
			s.lastSize = 0
			return false
		}
	}

	// Read the next record's header's first byte
	// That stores the record's Length IN WORDS, including Header itself
	recordLen, err := s.buffReader.ReadByte()
	if err != nil {
		s.lastError = err
		s.lastSize = 0
		return false
	}
	s.lastRecord[0] = recordLen
	mustRead := 4 * int(recordLen)

	// Read the header and record
	// 1: because we already got the first byte
	// :mustRead+1 because we need to read the full record, so specify one beyond last
	numRead, err := io.ReadFull(s.buffReader, s.lastRecord[1:mustRead])
	if err != nil {
		// we didn't read the full amount by num
		s.lastError = err
		s.lastSize = numRead
		return false
	}
	s.lastError = nil
	s.lastSize = mustRead
	return true
}

// Error returns the last error from Next().  May be io.EOF.
func (s *DbnScanner) Error() error {
	return s.lastError
}

// Parses the Scanner's current record as a `Record`.
func DbnScannerDecode[R Record, RP RecordPtr[R]](s *DbnScanner) (*R, error) {
	// Ensure there's a record to decode
	if s.lastSize <= RHeader_Size {
		return nil, ErrNoRecord
	}
	recordLen := 4 * int(s.lastRecord[0])
	if s.lastSize < recordLen {
		return nil, ErrMalformedRecord
	}

	// Object to return, instantiating an R and putting it in an RP
	var rp RP = new(R)

	// Make sure it's the right record type
	rtype := RType(s.lastRecord[1])
	if !rtype.IsCompatibleWith(rp.RType()) {
		return nil, unexpectedRTypeError(rtype, rp.RType())
	}

	err := rp.Fill_Raw(s.lastRecord[0:s.lastSize])
	if err != nil {
		return nil, err
	}
	return rp, nil
}

// Parses the current Record and passes it to the Visitor.
func (s *DbnScanner) Visit(visitor Visitor) error {
	// Ensure there's a record to decode
	if s.lastSize <= RHeader_Size {
		return ErrNoRecord
	}
	recordLen := 4 * int(s.lastRecord[0])
	if s.lastSize < recordLen {
		return ErrMalformedRecord
	}

	// Dispatch based on RType Make sure it's the right record type
	switch rtype := RType(s.lastRecord[1]); rtype {
	case RType_Mbp0: // Trade
		record := Mbp0{}
		if err := record.Fill_Raw(s.lastRecord[:Mbp0_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnMbp0(&record)
		}
	// Candlestick schemas
	case RType_Ohlcv1S, RType_Ohlcv1M, RType_Ohlcv1H, RType_Ohlcv1D, RType_OhlcvEod, RType_OhlcvDeprecated:
		record := Ohlcv{}
		if err := record.Fill_Raw(s.lastRecord[:Ohlcv_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnOhlcv(&record)
		}
	// Imbalance
	case RType_Imbalance:
		record := Imbalance{}
		if err := record.Fill_Raw(s.lastRecord[:Imbalance_Size]); err != nil {
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

/////////////////////////////////////////////////////////////////////////////

// ReadDBNToSlice reads the entire raw DBN stream from an io.Reader.
// It will scan for type R (for example Mbp0) and decode it into a slice of R.
// Returns the slice, the stream's metadata, and any error.
// Example:
//
//	fileReader, err := os.Open(dbnFilename)
//	records, metadata, err := dbn.ReadDBNToSlice[dbn.Mbp0](fileReader)
func ReadDBNToSlice[R Record, RP RecordPtr[R]](reader io.Reader) ([]R, *Metadata, error) {
	records := make([]R, 0)
	scanner := NewDbnScanner(reader)
	for scanner.Next() {
		r, err := DbnScannerDecode[R, RP](scanner)
		if err != nil {
			return records, scanner.metadata, err
		}
		records = append(records, *r)
	}
	err := scanner.Error()
	if err == io.EOF {
		// In this function, EOF is not propagated as an error
		err = nil
	}

	return records, scanner.metadata, err
}
