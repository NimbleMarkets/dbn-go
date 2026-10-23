// Copyright (c) 2024-2025 Neomantra Corp
//
// Default record constructors, mirroring the `Default` implementations of the upstream DBN Rust records:
//   https://github.com/databento/dbn/blob/main/rust/dbn/src/record/impl_default.rs
//
// Go zero values differ from upstream defaults.  For example, `0` is a valid `ChannelID`,
// so upstream defaults it to its max value to mean "unset".

package dbn

import "math"

func defaultRecordHeader(rtype RType, size uint16) RHeader {
	return RHeader{Length: uint8(size / 4), RType: rtype, TsEvent: UNDEF_TIMESTAMP}
}

// NewMboMsg returns a MboMsg with the same defaults as the upstream DBN `Default` implementation.
func NewMboMsg() MboMsg {
	return MboMsg{
		Header:    defaultRecordHeader(RType_Mbo, MboMsg_Size),
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
		Header:       defaultRecordHeader(RType_Statistics, StatMsgV3_Size),
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
		Header:       defaultRecordHeader(RType_Statistics, StatMsgV2_Size),
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
		Header:                  defaultRecordHeader(RType_InstrumentDef, InstrumentDefMsgV3_Size),
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
		Header:                  defaultRecordHeader(RType_InstrumentDef, InstrumentDefMsgV2_Size),
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

// NewBidAskPair returns a BidAskPair with the same defaults as the upstream DBN `Default` implementation.
func NewBidAskPair() BidAskPair {
	return BidAskPair{
		BidPx: UNDEF_PRICE,
		AskPx: UNDEF_PRICE,
	}
}

// NewConsolidatedBidAskPair returns a ConsolidatedBidAskPair with the same defaults as the upstream DBN `Default` implementation.
func NewConsolidatedBidAskPair() ConsolidatedBidAskPair {
	return ConsolidatedBidAskPair{
		BidPx: UNDEF_PRICE,
		AskPx: UNDEF_PRICE,
	}
}

// NewMbp0Msg returns an Mbp0Msg (upstream `TradeMsg`) with the same defaults as the upstream DBN `Default` implementation.
func NewMbp0Msg() Mbp0Msg {
	return Mbp0Msg{
		Header: defaultRecordHeader(RType_Mbp0, Mbp0Msg_Size),
		Price:  UNDEF_PRICE,
		Action: uint8(Action_None),
		Side:   uint8(Side_None),
		TsRecv: UNDEF_TIMESTAMP,
	}
}

// NewMbp1Msg returns an Mbp1Msg with the same defaults as the upstream DBN `Default` implementation.
func NewMbp1Msg() Mbp1Msg {
	return Mbp1Msg{
		Header: defaultRecordHeader(RType_Mbp1, Mbp1Msg_Size),
		Price:  UNDEF_PRICE,
		Action: byte(Action_None),
		Side:   byte(Side_None),
		TsRecv: UNDEF_TIMESTAMP,
		Level:  NewBidAskPair(),
	}
}

// NewMbp10Msg returns an Mbp10Msg with the same defaults as the upstream DBN `Default` implementation.
func NewMbp10Msg() Mbp10Msg {
	msg := Mbp10Msg{
		Header: defaultRecordHeader(RType_Mbp10, Mbp10Msg_Size),
		Price:  UNDEF_PRICE,
		Action: byte(Action_None),
		Side:   byte(Side_None),
		TsRecv: UNDEF_TIMESTAMP,
	}
	for i := range msg.Levels {
		msg.Levels[i] = NewBidAskPair()
	}
	return msg
}

// NewCmbp1Msg returns a Cmbp1Msg with the same defaults as the upstream DBN `default_for_schema`.
// Upstream takes a Schema; here `rtype` is the corresponding record type
// (`RType_Cmbp1`, `RType_Cbbo1S`, `RType_Cbbo1M`, or `RType_Tcbbo`).
func NewCmbp1Msg(rtype RType) Cmbp1Msg {
	action := byte(Action_None)
	if rtype == RType_Cbbo1S || rtype == RType_Cbbo1M {
		// Subsampled CBBO uses this byte as reserved padding, not an action.
		action = 0
	}
	return Cmbp1Msg{
		Header: defaultRecordHeader(rtype, Cmbp1Msg_Size),
		Price:  UNDEF_PRICE,
		Action: action,
		Side:   byte(Side_None),
		TsRecv: UNDEF_TIMESTAMP,
		Level:  NewConsolidatedBidAskPair(),
	}
}

// NewBboMsg returns a BboMsg with the same defaults as the upstream DBN `default_for_schema`.
// Upstream takes a Schema; here `rtype` is the corresponding record type
// (`RType_Bbo1S` or `RType_Bbo1M`).
func NewBboMsg(rtype RType) BboMsg {
	return BboMsg{
		Header: defaultRecordHeader(rtype, BboMsg_Size),
		Price:  UNDEF_PRICE,
		Side:   byte(Side_None),
		TsRecv: UNDEF_TIMESTAMP,
		Level:  NewBidAskPair(),
	}
}

// NewOhlcvMsg returns an OhlcvMsg with the same defaults as the upstream DBN `default_for_schema`.
// Upstream takes a Schema; here `rtype` is the corresponding record type
// (e.g. `RType_Ohlcv1S`, `RType_Ohlcv1M`, `RType_Ohlcv1H`, `RType_Ohlcv1D`, or `RType_OhlcvEod`).
func NewOhlcvMsg(rtype RType) OhlcvMsg {
	return OhlcvMsg{
		Header: defaultRecordHeader(rtype, OhlcvMsg_Size),
		Open:   UNDEF_PRICE,
		High:   UNDEF_PRICE,
		Low:    UNDEF_PRICE,
		Close:  UNDEF_PRICE,
	}
}

// NewStatusMsg returns a StatusMsg with the same defaults as the upstream DBN `Default` implementation.
func NewStatusMsg() StatusMsg {
	return StatusMsg{
		Header:                defaultRecordHeader(RType_Status, StatusMsg_Size),
		TsRecv:                UNDEF_TIMESTAMP,
		IsTrading:             uint8(TriState_NotAvailable),
		IsQuoting:             uint8(TriState_NotAvailable),
		IsShortSellRestricted: uint8(TriState_NotAvailable),
	}
}

// NewImbalanceMsg returns an ImbalanceMsg with the same defaults as the upstream DBN `Default` implementation.
func NewImbalanceMsg() ImbalanceMsg {
	return ImbalanceMsg{
		Header:               defaultRecordHeader(RType_Imbalance, ImbalanceMsg_Size),
		TsRecv:               UNDEF_TIMESTAMP,
		RefPrice:             UNDEF_PRICE,
		AuctionTime:          UNDEF_TIMESTAMP,
		ContBookClrPrice:     UNDEF_PRICE,
		AuctInterestClrPrice: UNDEF_PRICE,
		SsrFillingPrice:      UNDEF_PRICE,
		IndMatchPrice:        UNDEF_PRICE,
		UpperCollar:          UNDEF_PRICE,
		LowerCollar:          UNDEF_PRICE,
		PairedQty:            UNDEF_ORDER_SIZE,
		TotalImbalanceQty:    UNDEF_ORDER_SIZE,
		MarketImbalanceQty:   UNDEF_ORDER_SIZE,
		// Upstream's `unpaired_qty` is a u32; Go decodes it as an int32, so UNDEF_ORDER_SIZE is -1 here.
		UnpairedQty:          -1,
		AuctionType:          '~',
		Side:                 uint8(Side_None),
		UnpairedSide:         uint8(Side_None),
		SignificantImbalance: '~',
	}
}

// NewErrorMsgV1 returns an ErrorMsgV1 with the same defaults as the upstream DBN `Default` implementation.
func NewErrorMsgV1() ErrorMsgV1 {
	return ErrorMsgV1{
		Header: defaultRecordHeader(RType_Error, ErrorMsgV1_Size),
	}
}

// NewErrorMsgV2 returns an ErrorMsgV2 with the same defaults as the upstream DBN `Default` implementation.
func NewErrorMsgV2() ErrorMsgV2 {
	return ErrorMsgV2{
		Header: defaultRecordHeader(RType_Error, ErrorMsgV2_Size),
		Code:   ErrorCode_Unset,
		IsLast: math.MaxUint8,
	}
}

// NewSymbolMappingMsgV2 returns a SymbolMappingMsgV2 with the same defaults as the upstream DBN `Default` implementation.
func NewSymbolMappingMsgV2() SymbolMappingMsgV2 {
	return SymbolMappingMsgV2{
		Header:   defaultRecordHeader(RType_SymbolMapping, SymbolMappingMsgV2_Size),
		StypeIn:  SType(math.MaxUint8),
		StypeOut: SType(math.MaxUint8),
		StartTs:  UNDEF_TIMESTAMP,
		EndTs:    UNDEF_TIMESTAMP,
	}
}

// NewSystemMsg returns a SystemMsg with the same defaults as the upstream DBN `Default` implementation.
func NewSystemMsg() SystemMsg {
	return SystemMsg{
		Header: defaultRecordHeader(RType_System, SystemMsg_Size),
		Code:   SystemCode_Unset,
	}
}
