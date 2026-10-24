// Copyright (c) 2024 Neomantra Corp
//
// Common Message descriptions
//
// intended to allow versionless to the business objeects
//

package dbn

import "fmt"

/////////////////////////////////////////////////////////////////////////////
// Version-aware decoders for records that differ across DBN versions.
// These convert V1/V2 records up to the V3 layout (the canonical type).

// DecodeErrorMsg decodes an ErrorMsg, upgrading from V1 if needed.
// V1 only has a short error message
func DecodeErrorMsg(metadata *Metadata, body []byte) (*ErrorMsgV2, error) {
	if metadata == nil {
		return nil, ErrNoMetadata
	}
	switch metadata.VersionNum {
	case HeaderVersion1:
		var v1 ErrorMsgV1
		if err := v1.Fill_Raw(body); err != nil {
			return nil, err
		}
		v2 := NewErrorMsgV2()
		v2.Header = v1.Header
		v2.Header.Length = ErrorMsgV2_Size / 4
		copy(v2.Error[:], v1.Error[:])
		return &v2, nil
	case HeaderVersion2, HeaderVersion3:
		var v2 ErrorMsgV2
		if err := v2.Fill_Raw(body); err != nil {
			return nil, err
		}
		return &v2, nil
	default:
		return nil, ErrInvalidDBNVersion
	}
}

// DecodeSystemMsg decodes a SystemMsg, upgrading the 80-byte V1 layout when needed.
func DecodeSystemMsg(metadata *Metadata, body []byte) (*SystemMsg, error) {
	if metadata == nil {
		return nil, ErrNoMetadata
	}
	r := NewSystemMsg()
	switch metadata.VersionNum {
	case HeaderVersion1:
		const v1Size = RHeader_Size + 64
		if len(body) < v1Size {
			return nil, unexpectedBytesError(len(body), v1Size)
		}
		if err := r.Header.Fill_Raw(body[:RHeader_Size]); err != nil {
			return nil, err
		}
		r.Header.Length = SystemMsg_Size / 4
		copy(r.Message[:], body[RHeader_Size:v1Size])
		if r.IsHeartbeat() {
			r.Code = SystemCode_Heartbeat
		}
	case HeaderVersion2, HeaderVersion3:
		if err := r.Fill_Raw(body); err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidDBNVersion
	}
	return &r, nil
}

// decodeStatMsg decodes a StatMsg, upgrading from V1/V2 if needed.
// V1 and V2 share the same 64-byte layout (int32 Quantity).
// V3 has an 80-byte layout (int64 Quantity).
func DecodeStatMsg(metadata *Metadata, body []byte) (*StatMsgV3, error) {
	switch metadata.VersionNum {
	case HeaderVersion1, HeaderVersion2:
		var v2 StatMsgV2
		if err := v2.Fill_Raw(body[:StatMsgV2_Size]); err != nil {
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
		if err := v3.Fill_Raw(body[:StatMsgV3_Size]); err != nil {
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
func DecodeInstrumentDefMsg(metadata *Metadata, body []byte, size int) (*InstrumentDefMsgV3, error) {
	switch metadata.VersionNum {
	case HeaderVersion1:
		return nil, fmt.Errorf("InstrumentDefMsg V1 (22-byte symbols) is not supported")
	case HeaderVersion2:
		var v2 InstrumentDefMsgV2
		if err := v2.Fill_Raw(body[:size]); err != nil {
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
		if err := v3.Fill_Raw(body[:size]); err != nil {
			return nil, err
		}
		return &v3, nil
	default:
		return nil, ErrInvalidDBNVersion
	}
}
