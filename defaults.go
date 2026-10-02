// Copyright (c) 2024-2025 Neomantra Corp
//
// Default record constructors, mirroring the `Default` implementations of the upstream DBN Rust records:
//   https://github.com/databento/dbn/blob/main/rust/dbn/src/record/impl_default.rs
//
// Go zero values differ from upstream defaults.  For example, `0` is a valid `ChannelID`,
// so upstream defaults it to its max value to mean "unset".

package dbn

import "math"

// NewMboMsg returns a MboMsg with the same defaults as the upstream DBN `Default` implementation.
func NewMboMsg() MboMsg {
	return MboMsg{
		Header: RHeader{
			Length:       MboMsg_Size / 4,
			RType:        RType_Mbo,
			PublisherID:  0,
			InstrumentID: 0,
			TsEvent:      UNDEF_TIMESTAMP,
		},
		Price:     UNDEF_PRICE,
		ChannelID: math.MaxUint8,
		Action:    byte(Action_None),
		Side:      byte(Side_None),
		TsRecv:    UNDEF_TIMESTAMP,
	}
}

// NewStatMsgV3 returns a StatMsgV3 with the same defaults as the upstream DBN `Default` implementation.
func NewStatMsgV3() StatMsgV3 {
	return StatMsgV3{
		Header: RHeader{
			Length:       StatMsgV3_Size / 4,
			RType:        RType_Statistics,
			PublisherID:  0,
			InstrumentID: 0,
			TsEvent:      UNDEF_TIMESTAMP,
		},
		TsRecv:       UNDEF_TIMESTAMP,
		TsRef:        UNDEF_TIMESTAMP,
		Price:        UNDEF_PRICE,
		Quantity:     StatMsgV3_UNDEF_STAT_QUANTITY,
		ChannelID:    math.MaxUint16,
		UpdateAction: uint8(StatUpdateAction_New),
	}
}

// NewStatMsgV2 returns a StatMsgV2 with the same defaults as the upstream DBN `Default` implementation.
func NewStatMsgV2() StatMsgV2 {
	return StatMsgV2{
		Header: RHeader{
			Length:       StatMsgV2_Size / 4,
			RType:        RType_Statistics,
			PublisherID:  0,
			InstrumentID: 0,
			TsEvent:      UNDEF_TIMESTAMP,
		},
		TsRecv:       UNDEF_TIMESTAMP,
		TsRef:        UNDEF_TIMESTAMP,
		Price:        UNDEF_PRICE,
		Quantity:     StatMsgV2_UNDEF_STAT_QUANTITY,
		ChannelID:    math.MaxUint16,
		UpdateAction: uint8(StatUpdateAction_New),
	}
}

// NewInstrumentDefMsgV3 returns a InstrumentDefMsgV3 with the same defaults as the upstream DBN `Default` implementation.
func NewInstrumentDefMsgV3() InstrumentDefMsgV3 {
	return InstrumentDefMsgV3{
		Header: RHeader{
			Length:       InstrumentDefMsgV3_Size / 4,
			RType:        RType_InstrumentDef,
			PublisherID:  0,
			InstrumentID: 0,
			TsEvent:      UNDEF_TIMESTAMP,
		},
		TsRecv:                  UNDEF_TIMESTAMP,
		MinPriceIncrement:       UNDEF_PRICE,
		DisplayFactor:           UNDEF_PRICE,
		Expiration:              UNDEF_TIMESTAMP,
		Activation:              UNDEF_TIMESTAMP,
		HighLimitPrice:          UNDEF_PRICE,
		LowLimitPrice:           UNDEF_PRICE,
		MaxPriceVariation:       UNDEF_PRICE,
		UnitOfMeasureQty:        UNDEF_PRICE,
		MinPriceIncrementAmount: UNDEF_PRICE,
		PriceRatio:              UNDEF_PRICE,
		StrikePrice:             UNDEF_PRICE,
		LegPrice:                UNDEF_PRICE,
		LegDelta:                UNDEF_PRICE,
		InstAttribValue:         math.MaxInt32,
		MarketDepthImplied:      math.MaxInt32,
		MarketDepth:             math.MaxInt32,
		MarketSegmentID:         math.MaxUint32,
		MaxTradeVol:             math.MaxUint32,
		MinLotSize:              math.MaxInt32,
		MinLotSizeBlock:         math.MaxInt32,
		MinLotSizeRoundLot:      math.MaxInt32,
		MinTradeVol:             math.MaxUint32,
		ContractMultiplier:      math.MaxInt32,
		DecayQuantity:           math.MaxInt32,
		OriginalContractSize:    math.MaxInt32,
		ApplID:                  math.MaxInt16,
		MaturityYear:            math.MaxUint16,
		DecayStartDate:          math.MaxUint16,
		ChannelID:               math.MaxUint16,
		MatchAlgorithm:          byte(MatchAlgorithm_Undefined),
		MainFraction:            math.MaxUint8,
		PriceDisplayFormat:      math.MaxUint8,
		SubFraction:             math.MaxUint8,
		UnderlyingProduct:       math.MaxUint8,
		SecurityUpdateAction:    byte(Add),
		UserDefinedInstrument:   UserDefinedInstrument_No,
		MaturityMonth:           math.MaxUint8,
		MaturityDay:             math.MaxUint8,
		MaturityWeek:            math.MaxUint8,
		ContractMultiplierUnit:  math.MaxInt8,
		FlowScheduleType:        math.MaxInt8,
		TickRule:                math.MaxUint8,
		LegSide:                 byte(Side_None),
	}
}

// NewInstrumentDefMsgV2 returns a InstrumentDefMsgV2 with the same defaults as the upstream DBN `Default` implementation.
func NewInstrumentDefMsgV2() InstrumentDefMsgV2 {
	return InstrumentDefMsgV2{
		Header: RHeader{
			Length:       InstrumentDefMsgV2_Size / 4,
			RType:        RType_InstrumentDef,
			PublisherID:  0,
			InstrumentID: 0,
			TsEvent:      UNDEF_TIMESTAMP,
		},
		TsRecv:                  UNDEF_TIMESTAMP,
		MinPriceIncrement:       UNDEF_PRICE,
		DisplayFactor:           UNDEF_PRICE,
		Expiration:              UNDEF_TIMESTAMP,
		Activation:              UNDEF_TIMESTAMP,
		HighLimitPrice:          UNDEF_PRICE,
		LowLimitPrice:           UNDEF_PRICE,
		MaxPriceVariation:       UNDEF_PRICE,
		TradingReferencePrice:   UNDEF_PRICE,
		UnitOfMeasureQty:        UNDEF_PRICE,
		MinPriceIncrementAmount: UNDEF_PRICE,
		PriceRatio:              UNDEF_PRICE,
		StrikePrice:             UNDEF_PRICE,
		InstAttribValue:         math.MaxInt32,
		MarketDepthImplied:      math.MaxInt32,
		MarketDepth:             math.MaxInt32,
		MarketSegmentID:         math.MaxUint32,
		MaxTradeVol:             math.MaxUint32,
		MinLotSize:              math.MaxInt32,
		MinLotSizeBlock:         math.MaxInt32,
		MinLotSizeRoundLot:      math.MaxInt32,
		MinTradeVol:             math.MaxUint32,
		ContractMultiplier:      math.MaxInt32,
		DecayQuantity:           math.MaxInt32,
		OriginalContractSize:    math.MaxInt32,
		TradingReferenceDate:    math.MaxUint16,
		ApplID:                  math.MaxInt16,
		MaturityYear:            math.MaxUint16,
		DecayStartDate:          math.MaxUint16,
		ChannelID:               math.MaxUint16,
		MatchAlgorithm:          byte(MatchAlgorithm_Undefined),
		MdSecurityTradingStatus: math.MaxUint8,
		MainFraction:            math.MaxUint8,
		PriceDisplayFormat:      math.MaxUint8,
		SettlPrice_type:         math.MaxUint8,
		SubFraction:             math.MaxUint8,
		UnderlyingProduct:       math.MaxUint8,
		SecurityUpdateAction:    byte(Add),
		MaturityMonth:           math.MaxUint8,
		MaturityDay:             math.MaxUint8,
		MaturityWeek:            math.MaxUint8,
		UserDefinedInstrument:   UserDefinedInstrument_No,
		ContractMultiplierUnit:  math.MaxInt8,
		FlowScheduleType:        math.MaxInt8,
		TickRule:                math.MaxUint8,
	}
}
