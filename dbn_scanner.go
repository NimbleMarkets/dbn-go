// Copyright (c) 2024 Neomantra Corp

package dbn

import (
	"bufio"
	"fmt"
	"io"
)

///////////////////////////////////////////////////////////////////////////////

// Default buffer size for decoding
const DEFAULT_DECODE_BUFFER_SIZE = 64 * 1024
const DEFAULT_SCRATCH_BUFFER_SIZE = 1024 // bigger than largest record size

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

// Error returns the last error from Next().  May be io.EOF.
func (s *DbnScanner) Error() error {
	return s.lastError
}

// GetLastHeader returns the RHeader of the last record read, or an error
func (s *DbnScanner) GetLastHeader() (RHeader, error) {
	var rheader RHeader
	err := rheader.Fill_Raw(s.lastRecord[0:RHeader_Size])
	return rheader, err
}

// GetLastRecord returns the raw bytes of the last record read
func (s *DbnScanner) GetLastRecord() []byte {
	return s.lastRecord
}

// GetLastSize returns the size of the last record read
func (s *DbnScanner) GetLastSize() int {
	return s.lastSize
}

/////////////////////////////////////////////////////////////////////////////

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
	// 1: because we already got the first size byte
	// :mustRead because we only want a subset of the buffer (the full record size)
	numRead, err := io.ReadFull(s.buffReader, s.lastRecord[1:mustRead])
	if err != nil {
		// we didn't read the full amount by num
		s.lastError = err
		s.lastSize = numRead + 1 // +1 for size byte
		return false
	}
	s.lastError = nil
	s.lastSize = mustRead
	return true
}

// Parses the Scanner's current record as a `Record`.
// This a plain function because receiver functions cannot be generic.
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

// DecodeSymbolMappingMsg parses the Scanner's current record as a `SymbolMappingMsg`.
// This is outside the DbnScannerDecode function because SymbolMappingMsg.Fill_Raw
// is two-argum,ent, depending on the DbnScanner's metadata's SymbolCstrLen.
func (s *DbnScanner) DecodeSymbolMappingMsg() (*SymbolMappingMsg, error) {
	// Ensure there's a record to decode
	if s.lastSize <= RHeader_Size {
		return nil, ErrNoRecord
	}
	recordLen := 4 * int(s.lastRecord[0])
	if s.lastSize < recordLen {
		return nil, ErrMalformedRecord
	}
	if s.metadata == nil { // we need valid metadata
		return nil, ErrNoMetadata
	}

	// Object to return, instantiating an R and putting it in an RP
	var rp *SymbolMappingMsg = new(SymbolMappingMsg)

	// Make sure it's the right record type
	rtype := RType(s.lastRecord[1])
	if !rtype.IsCompatibleWith(rp.RType()) {
		return nil, unexpectedRTypeError(rtype, rp.RType())
	}

	err := SymbolMappingMsgFillRaw(rp, s.lastRecord[0:s.lastSize], s.metadata.SymbolCstrLen)
	if err != nil {
		return nil, err
	}
	return rp, nil
}

// DecodeStatMsg parses the Scanner's current record as a StatMsg (V3 layout).
// V1/V2 records are automatically upgraded to V3 (sign-extending Quantity).
func (s *DbnScanner) DecodeStatMsg() (*StatMsgV3, error) {
	if s.lastSize <= RHeader_Size {
		return nil, ErrNoRecord
	}
	recordLen := 4 * int(s.lastRecord[0])
	if s.lastSize < recordLen {
		return nil, ErrMalformedRecord
	}
	if s.metadata == nil {
		return nil, ErrNoMetadata
	}
	rtype := RType(s.lastRecord[1])
	if !rtype.IsCompatibleWith(RType_Statistics) {
		return nil, unexpectedRTypeError(rtype, RType_Statistics)
	}
	return s.decodeStatMsg()
}

// DecodeInstrumentDefMsg parses the Scanner's current record as an InstrumentDefMsg (V3 layout).
// V2 records are automatically upgraded to V3. V1 is not supported.
func (s *DbnScanner) DecodeInstrumentDefMsg() (*InstrumentDefMsgV3, error) {
	if s.lastSize <= RHeader_Size {
		return nil, ErrNoRecord
	}
	recordLen := 4 * int(s.lastRecord[0])
	if s.lastSize < recordLen {
		return nil, ErrMalformedRecord
	}
	if s.metadata == nil {
		return nil, ErrNoMetadata
	}
	rtype := RType(s.lastRecord[1])
	if !rtype.IsCompatibleWith(RType_InstrumentDef) {
		return nil, unexpectedRTypeError(rtype, RType_InstrumentDef)
	}
	return s.decodeInstrumentDefMsg()
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
	// Trade
	case RType_Mbp0:
		record := Mbp0Msg{}
		if err := record.Fill_Raw(s.lastRecord[:Mbp0Msg_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnMbp0(&record)
		}
	// Market-by-price, 1 depth
	case RType_Mbp1:
		record := Mbp1Msg{}
		if err := record.Fill_Raw(s.lastRecord[:Mbp1Msg_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnMbp1(&record)
		}
	// Market-by-price, 10 depth
	case RType_Mbp10:
		record := Mbp10Msg{}
		if err := record.Fill_Raw(s.lastRecord[:Mbp10Msg_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnMbp10(&record)
		}
	// Market-by-Order
	case RType_Mbo:
		record := MboMsg{}
		if err := record.Fill_Raw(s.lastRecord[:MboMsg_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnMbo(&record)
		}
	// Candlestick schemas
	case RType_Ohlcv1S, RType_Ohlcv1M, RType_Ohlcv1H, RType_Ohlcv1D, RType_OhlcvEod, RType_OhlcvDeprecated:
		record := OhlcvMsg{}
		if err := record.Fill_Raw(s.lastRecord[:OhlcvMsg_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnOhlcv(&record)
		}
	// CBBO schemas
	case RType_Cmbp1, RType_Cbbo1S, RType_Cbbo1M, RType_Tcbbo:
		record := Cmbp1Msg{}
		if err := record.Fill_Raw(s.lastRecord[:Cmbp1Msg_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnCmbp1(&record)
		}
	// Imbalance
	case RType_Imbalance:
		record := ImbalanceMsg{}
		if err := record.Fill_Raw(s.lastRecord[:ImbalanceMsg_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnImbalance(&record)
		}
	// Error
	case RType_Error:
		record := ErrorMsg{}
		if err := record.Fill_Raw(s.lastRecord[:ErrorMsg_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnErrorMsg(&record)
		}
	// SymbolMapping
	case RType_SymbolMapping:
		if s.metadata == nil {
			return ErrNoMetadata
		}
		record := SymbolMappingMsg{}
		if err := SymbolMappingMsgFillRaw(&record, s.lastRecord[:s.lastSize], s.metadata.SymbolCstrLen); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnSymbolMappingMsg(&record)
		}
	// System
	case RType_System:
		record := SystemMsg{}
		if err := record.Fill_Raw(s.lastRecord[:SystemMsg_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnSystemMsg(&record)
		}
	// Statistics (version-aware: V1/V2 = 64 bytes, V3 = 80 bytes)
	case RType_Statistics:
		if s.metadata == nil {
			return ErrNoMetadata
		}
		record, err := s.decodeStatMsg()
		if err != nil {
			return err
		}
		return visitor.OnStatMsg(record)
	// Status
	case RType_Status:
		record := StatusMsg{}
		if err := record.Fill_Raw(s.lastRecord[:StatusMsg_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnStatusMsg(&record)
		}
	// BBO schemas
	case RType_Bbo1S, RType_Bbo1M:
		record := BboMsg{}
		if err := record.Fill_Raw(s.lastRecord[:BboMsg_Size]); err != nil {
			return err // TODO: OnError()
		} else {
			return visitor.OnBbo(&record)
		}

	// InstrumentDef (version-aware: V2 and V3 have different layouts)
	case RType_InstrumentDef:
		if s.metadata == nil {
			return ErrNoMetadata
		}
		record, err := s.decodeInstrumentDefMsg()
		if err != nil {
			return err
		}
		return visitor.OnInstrumentDefMsg(record)

	default:
		return ErrUnknownRType
	}
}

/////////////////////////////////////////////////////////////////////////////
// Version-aware decoders for records that differ across DBN versions.
// These convert V1/V2 records up to the V3 layout (the canonical type).

// decodeStatMsg decodes a StatMsg, upgrading from V1/V2 if needed.
// V1 and V2 share the same 64-byte layout (int32 Quantity).
// V3 has an 80-byte layout (int64 Quantity).
func (s *DbnScanner) decodeStatMsg() (*StatMsgV3, error) {
	switch s.metadata.VersionNum {
	case HeaderVersion1, HeaderVersion2:
		var v2 StatMsgV2
		if err := v2.Fill_Raw(s.lastRecord[:StatMsgV2_Size]); err != nil {
			return nil, err
		}
		// Upgrade V2 → V3: sign-extend Quantity from int32 to int64
		return &StatMsgV3{
			Header:       v2.Header,
			TsRecv:       v2.TsRecv,
			TsRef:        v2.TsRef,
			Price:        v2.Price,
			Quantity:     int64(v2.Quantity),
			Sequence:     v2.Sequence,
			TsInDelta:    v2.TsInDelta,
			StatType:     v2.StatType,
			ChannelID:    v2.ChannelID,
			UpdateAction: v2.UpdateAction,
			StatFlags:    v2.StatFlags,
		}, nil
	case HeaderVersion3:
		var v3 StatMsgV3
		if err := v3.Fill_Raw(s.lastRecord[:StatMsgV3_Size]); err != nil {
			return nil, err
		}
		return &v3, nil
	default:
		return nil, ErrInvalidDBNVersion
	}
}

// decodeInstrumentDefMsg decodes an InstrumentDefMsg, upgrading from V2 if needed.
// V1 instrument definitions (22-byte symbols) are not supported.
// V2 has a different field layout (uint32 RawInstrumentID, extra fields removed in V3).
// V3 has uint64 RawInstrumentID and multi-leg strategy fields.
func (s *DbnScanner) decodeInstrumentDefMsg() (*InstrumentDefMsgV3, error) {
	switch s.metadata.VersionNum {
	case HeaderVersion1:
		return nil, fmt.Errorf("InstrumentDefMsg V1 (22-byte symbols) is not supported")
	case HeaderVersion2:
		var v2 InstrumentDefMsgV2
		if err := v2.Fill_Raw(s.lastRecord[:s.lastSize]); err != nil {
			return nil, err
		}
		// Upgrade V2 → V3: zero-extend RawInstrumentID, drop removed fields, zero-fill leg fields
		v3 := InstrumentDefMsgV3{
			Header:                  v2.Header,
			TsRecv:                  v2.TsRecv,
			MinPriceIncrement:       v2.MinPriceIncrement,
			DisplayFactor:           v2.DisplayFactor,
			Expiration:              v2.Expiration,
			Activation:              v2.Activation,
			HighLimitPrice:          v2.HighLimitPrice,
			LowLimitPrice:           v2.LowLimitPrice,
			MaxPriceVariation:       v2.MaxPriceVariation,
			UnitOfMeasureQty:        v2.UnitOfMeasureQty,
			MinPriceIncrementAmount: v2.MinPriceIncrementAmount,
			PriceRatio:              v2.PriceRatio,
			StrikePrice:             v2.StrikePrice,
			RawInstrumentID:         uint64(v2.RawInstrumentID),
			InstAttribValue:         v2.InstAttribValue,
			UnderlyingID:            v2.UnderlyingID,
			MarketDepthImplied:      v2.MarketDepthImplied,
			MarketDepth:             v2.MarketDepth,
			MarketSegmentID:         v2.MarketSegmentID,
			MaxTradeVol:             v2.MaxTradeVol,
			MinLotSize:              v2.MinLotSize,
			MinLotSizeBlock:         v2.MinLotSizeBlock,
			MinLotSizeRoundLot:      v2.MinLotSizeRoundLot,
			MinTradeVol:             v2.MinTradeVol,
			ContractMultiplier:      v2.ContractMultiplier,
			DecayQuantity:           v2.DecayQuantity,
			OriginalContractSize:    v2.OriginalContractSize,
			ApplID:                  v2.ApplID,
			MaturityYear:            v2.MaturityYear,
			DecayStartDate:          v2.DecayStartDate,
			ChannelID:               v2.ChannelID,
			Currency:                v2.Currency,
			SettlCurrency:           v2.SettlCurrency,
			Secsubtype:              v2.Secsubtype,
			Group:                   v2.Group,
			Exchange:                v2.Exchange,
			Cfi:                     v2.Cfi,
			SecurityType:            v2.SecurityType,
			UnitOfMeasure:           v2.UnitOfMeasure,
			Underlying:              v2.Underlying,
			StrikePriceCurrency:     v2.StrikePriceCurrency,
			InstrumentClass:         v2.InstrumentClass,
			MatchAlgorithm:          v2.MatchAlgorithm,
			MainFraction:            v2.MainFraction,
			PriceDisplayFormat:      v2.PriceDisplayFormat,
			SubFraction:             v2.SubFraction,
			UnderlyingProduct:       v2.UnderlyingProduct,
			SecurityUpdateAction:    v2.SecurityUpdateAction,
			MaturityMonth:           v2.MaturityMonth,
			MaturityDay:             v2.MaturityDay,
			MaturityWeek:            v2.MaturityWeek,
			UserDefinedInstrument:   v2.UserDefinedInstrument,
			ContractMultiplierUnit:  v2.ContractMultiplierUnit,
			FlowScheduleType:        v2.FlowScheduleType,
			TickRule:                v2.TickRule,
			// Leg fields are zero-valued (not present in V2)
		}
		// RawSymbol is the same size in V2 and V3 (71 bytes)
		v3.RawSymbol = v2.RawSymbol
		// Asset: V2 is [7]byte, V3 is [11]byte — copy the smaller into the larger
		copy(v3.Asset[:], v2.Asset[:])
		return &v3, nil
	case HeaderVersion3:
		var v3 InstrumentDefMsgV3
		if err := v3.Fill_Raw(s.lastRecord[:s.lastSize]); err != nil {
			return nil, err
		}
		return &v3, nil
	default:
		return nil, ErrInvalidDBNVersion
	}
}

/////////////////////////////////////////////////////////////////////////////

// ReadDBNToSlice reads the entire raw DBN stream from an io.Reader.
// It will scan for type R (for example Mbp0) and decode it into a slice of R.
// Returns the slice, the stream's metadata, and any error.
// Example:
//
//	fileReader, err := os.Open(dbnFilename)
//	records, metadata, err := dbn.ReadDBNToSlice[dbn.Mbp0Msg](fileReader)
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
