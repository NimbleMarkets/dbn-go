// Copyright (c) 2026 Neomantra Corp
//
// DBN Version 3 Message Structs
//
// These structs represent the DBN version 3 binary layout.
// V3 uses 71-byte symbol strings and has additional fields compared to V2.
//
// Adapted from Databento's DBN:
//   https://github.com/databento/dbn

package dbn

import (
	"encoding/binary"
	"math"

	"github.com/valyala/fastjson"
)

///////////////////////////////////////////////////////////////////////////////

// StatMsgV3 is the DBN version 3 layout (80 bytes).
type StatMsgV3 struct {
	Header       RHeader   `json:"hd" csv:"hd"`                       // The common header.
	TsRecv       uint64    `json:"ts_recv" csv:"ts_recv"`             // The capture-server-received timestamp expressed as the number of nanoseconds since the UNIX epoch.
	TsRef        uint64    `json:"ts_ref" csv:"ts_ref"`               // The reference timestamp of the statistic value expressed as the number of nanoseconds since the UNIX epoch. Will be [`crate::UNDEF_TIMESTAMP`] when unused.
	Price        int64     `json:"price" csv:"price"`                 // The value for price statistics expressed as a signed integer where every 1 unit corresponds to 1e-9, i.e. 1/1,000,000,000 or 0.000000001. Will be [`crate::UNDEF_PRICE`] when unused.
	Quantity     int64     `json:"quantity" csv:"quantity"`           // The value for non-price statistics. Will be [`crate::UNDEF_STAT_QUANTITY`] when unused.
	Sequence     uint32    `json:"sequence" csv:"sequence"`           // The message sequence number assigned at the venue.
	TsInDelta    int32     `json:"ts_in_delta" csv:"ts_in_delta"`     // The delta of `ts_recv - ts_exchange_send`, max 2 seconds.
	StatType     uint16    `json:"stat_type" csv:"stat_type"`         // The type of statistic value contained in the message. Refer to the [`StatType`](crate::enums::StatType) for variants.
	ChannelID    uint16    `json:"channel_id" csv:"channel_id"`       // The channel ID assigned by Databento as an incrementing integer starting at zero.
	UpdateAction uint8     `json:"update_action" csv:"update_action"` // Indicates if the statistic is newly added (1) or deleted (2). (Deleted is only used with some stat types)
	StatFlags    uint8     `json:"stat_flags" csv:"stat_flags"`       // Additional flags associate with certain stat types.
	Reserved     [18]uint8 `json:"_reserved" csv:"_reserved"`         // Filler for alignment
}

const StatMsgV3_Size = RHeader_Size + 64

const StatMsgV3_UNDEF_STAT_QUANTITY = math.MaxInt64

func (*StatMsgV3) RType() RType {
	return RType_Statistics
}

func (*StatMsgV3) RSize() uint16 {
	return StatMsgV3_Size
}

func (r *StatMsgV3) Fill_Raw(b []byte) error {
	if len(b) < StatMsgV3_Size {
		return unexpectedBytesError(len(b), StatMsgV3_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:]
	r.TsRecv = binary.LittleEndian.Uint64(body[0:8])
	r.TsRef = binary.LittleEndian.Uint64(body[8:16])
	r.Price = int64(binary.LittleEndian.Uint64(body[16:24]))
	r.Quantity = int64(binary.LittleEndian.Uint64(body[24:32]))
	r.Sequence = binary.LittleEndian.Uint32(body[32:36])
	r.TsInDelta = int32(binary.LittleEndian.Uint32(body[36:40]))
	r.StatType = binary.LittleEndian.Uint16(body[40:42])
	r.ChannelID = binary.LittleEndian.Uint16(body[42:44])
	r.UpdateAction = body[44]
	r.StatFlags = body[45]
	return nil
}

func (r *StatMsgV3) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.TsRef = fastjson_GetUint64FromString(val, "ts_ref")
	r.Price = fastjson_GetInt64FromString(val, "price")
	r.Quantity = fastjson_GetInt64Tolerant(val, "quantity") // V2=number, V3=quoted string
	r.Sequence = uint32(val.GetUint("sequence"))
	r.TsInDelta = int32(val.GetUint("ts_in_delta"))
	r.StatType = uint16(val.GetUint("stat_type"))
	r.ChannelID = uint16(val.GetUint("channel_id"))
	r.UpdateAction = uint8(val.GetUint("update_action"))
	r.StatFlags = uint8(val.GetUint("stat_flags"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// InstrumentDefMsgV3 is the DBN version 3 layout (520 bytes).
type InstrumentDefMsgV3 struct {
	Header                   RHeader                        `json:"hd" csv:"hd"`                                                   // The common header.
	TsRecv                   uint64                         `json:"ts_recv" csv:"ts_recv"`                                         // The capture-server-received timestamp expressed as the number of nanoseconds since the UNIX epoch.
	MinPriceIncrement        int64                          `json:"min_price_increment" csv:"min_price_increment"`                 // Fixed price The minimum constant tick for the instrument in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	DisplayFactor            int64                          `json:"display_factor" csv:"display_factor"`                           // The multiplier to convert the venue's display price to the conventional price, in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	Expiration               uint64                         `json:"expiration" csv:"expiration"`                                   // The last eligible trade time expressed as a number of nanoseconds since the UNIX epoch. Will be [`crate::UNDEF_TIMESTAMP`] when null, such as for equities.  Some publishers only provide date-level granularity.
	Activation               uint64                         `json:"activation" csv:"activation"`                                   // The time of instrument activation expressed as a number of nanoseconds since the UNIX epoch. Will be [`crate::UNDEF_TIMESTAMP`] when null, such as for equities.  Some publishers only provide date-level granularity.
	HighLimitPrice           int64                          `json:"high_limit_price" csv:"high_limit_price"`                       // The allowable high limit price for the trading day in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	LowLimitPrice            int64                          `json:"low_limit_price" csv:"low_limit_price"`                         // The allowable low limit price for the trading day in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	MaxPriceVariation        int64                          `json:"max_price_variation" csv:"max_price_variation"`                 // The differential value for price banding in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	UnitOfMeasureQty         int64                          `json:"unit_of_measure_qty" csv:"unit_of_measure_qty"`                 // The contract size for each instrument, in combination with `unit_of_measure`, in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	MinPriceIncrementAmount  int64                          `json:"min_price_increment_amount" csv:"min_price_increment_amount"`   // The value currently under development by the venue. Converted to units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	PriceRatio               int64                          `json:"price_ratio" csv:"price_ratio"`                                 // The value used for price calculation in spread and leg pricing in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	StrikePrice              int64                          `json:"strike_price" csv:"strike_price"`                               // The strike price of the option. Converted to units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	RawInstrumentID          uint64                         `json:"raw_instrument_id" csv:"raw_instrument_id"`                     // The instrument ID assigned by the publisher. May be the same as `instrument_id`.
	LegPrice                 int64                          `json:"leg_price" csv:"leg_price"`                                     // The tied price (if any) of the leg.
	LegDelta                 int64                          `json:"leg_delta" csv:"leg_delta"`                                     // The associated delta (if any) of the leg.
	InstAttribValue          int32                          `json:"inst_attrib_value" csv:"inst_attrib_value"`                     // A bitmap of instrument eligibility attributes.
	UnderlyingID             uint32                         `json:"underlying_id" csv:"underlying_id"`                             // The `instrument_id` of the first underlying instrument.
	MarketDepthImplied       int32                          `json:"market_depth_implied" csv:"market_depth_implied"`               // The implied book depth on the price level data feed.
	MarketDepth              int32                          `json:"market_depth" csv:"market_depth"`                               // The (outright) book depth on the price level data feed.
	MarketSegmentID          uint32                         `json:"market_segment_id" csv:"market_segment_id"`                     // The market segment of the instrument.
	MaxTradeVol              uint32                         `json:"max_trade_vol" csv:"max_trade_vol"`                             // The maximum trading volume for the instrument.
	MinLotSize               int32                          `json:"min_lot_size" csv:"min_lot_size"`                               // The minimum order entry quantity for the instrument.
	MinLotSizeBlock          int32                          `json:"min_lot_size_block" csv:"min_lot_size_block"`                   // The minimum quantity required for a block trade of the instrument.
	MinLotSizeRoundLot       int32                          `json:"min_lot_size_round_lot" csv:"min_lot_size_round_lot"`           // The minimum quantity required for a round lot of the instrument. Multiples of this quantity are also round lots.
	MinTradeVol              uint32                         `json:"min_trade_vol" csv:"min_trade_vol"`                             // The minimum trading volume for the instrument.
	ContractMultiplier       int32                          `json:"contract_multiplier" csv:"contract_multiplier"`                 // The number of deliverables per instrument, i.e. peak days.
	DecayQuantity            int32                          `json:"decay_quantity" csv:"decay_quantity"`                           // The quantity that a contract will decay daily, after `decay_start_date` has been reached.
	OriginalContractSize     int32                          `json:"original_contract_size" csv:"original_contract_size"`           // The fixed contract value assigned to each instrument.
	LegInstrumentID          uint32                         `json:"leg_instrument_id" csv:"leg_instrument_id"`                     /// The numeric ID assigned to the leg instrument. See [Instrument identifiers](https://databento.com/docs/standards-and-conventions/common-fields-enums-types#instrument-identifiers).
	LegRatioPriceNumerator   int32                          `json:"leg_ratio_price_numerator" csv:"leg_ratio_price_numerator"`     /// The numerator of the price ratio of the leg within the spread.
	LegRatioPriceDenominator int32                          `json:"leg_ratio_price_denominator" csv:"leg_ratio_price_denominator"` /// The denominator of the price ratio of the leg within the spread.
	LegRatioQtyNumerator     int32                          `json:"leg_ratio_qty_numerator" csv:"leg_ratio_qty_numerator"`         /// The numerator of the quantity ratio of the leg within the spread.
	LegRatioQtyDenominator   int32                          `json:"leg_ratio_qty_denominator" csv:"leg_ratio_qty_denominator"`     /// The denominator of the quantity ratio of the leg within the spread.
	LegUnderlyingID          uint32                         `json:"leg_underlying_id" csv:"leg_underlying_id"`                     /// The numeric ID of the leg instrument's underlying instrument. See [Instrument identifiers](https://databento.com/docs/standards-and-conventions/common-fields-enums-types#instrument-identifiers).
	ApplID                   int16                          `json:"appl_id" csv:"appl_id"`                                         // The channel ID assigned at the venue.
	MaturityYear             uint16                         `json:"maturity_year" csv:"maturity_year"`                             // The calendar year reflected in the instrument symbol.
	DecayStartDate           uint16                         `json:"decay_start_date" csv:"decay_start_date"`                       // The date at which a contract will begin to decay.
	ChannelID                uint16                         `json:"channel_id" csv:"channel_id"`                                   // The channel ID assigned by Databento as an incrementing integer starting at zero.
	LegCount                 uint16                         `json:"leg_count" csv:"leg_count"`                                     /// The number of legs in the strategy or spread. Will be 0 for outrights.
	LegIndex                 uint16                         `json:"leg_index" csv:"leg_index"`                                     /// The 0-based index of the leg.
	Currency                 [4]byte                        `json:"currency" csv:"currency"`                                       // The currency used for price fields.
	SettlCurrency            [4]byte                        `json:"settl_currency" csv:"settl_currency"`                           // The currency used for settlement, if different from `currency`.
	Secsubtype               [6]byte                        `json:"secsubtype" csv:"secsubtype"`                                   // The strategy type of the spread.
	RawSymbol                [MetadataV3_SymbolCstrLen]byte `json:"raw_symbol" csv:"raw_symbol"`                                   // The instrument raw symbol assigned by the publisher.
	Group                    [21]byte                       `json:"group" csv:"group"`                                             // The security group code of the instrument.
	Exchange                 [5]byte                        `json:"exchange" csv:"exchange"`                                       // The exchange used to identify the instrument.
	Asset                    [MetadataV3_AssetCStrLen]byte  `json:"asset" csv:"asset"`                                             // The underlying asset code (product code) of the instrument.
	Cfi                      [7]byte                        `json:"cfi" csv:"cfi"`                                                 // The ISO standard instrument categorization code.
	SecurityType             [7]byte                        `json:"security_type" csv:"security_type"`                             // The [Security type](https://databento.com/docs/schemas-and-data-formats/instrument-definitions#security-type) of the instrument, e.g. FUT for future or future spread.
	UnitOfMeasure            [31]byte                       `json:"unit_of_measure" csv:"unit_of_measure"`                         // The unit of measure for the instrument's original contract size, e.g. USD or LBS.
	Underlying               [21]byte                       `json:"underlying" csv:"underlying"`                                   // The symbol of the first underlying instrument.
	StrikePriceCurrency      [4]byte                        `json:"strike_price_currency" csv:"strike_price_currency"`             // The currency of [`strike_price`](Self::strike_price).
	LegRawSymbol             [MetadataV3_SymbolCstrLen]byte `json:"leg_raw_symbol" csv:"leg_raw_symbol"`                           // The leg instrument's raw symbol assigned by the publisher.
	InstrumentClass          byte                           `json:"instrument_class" csv:"instrument_class"`                       // The classification of the instrument.
	MatchAlgorithm           byte                           `json:"match_algorithm" csv:"match_algorithm"`                         // The matching algorithm used for the instrument, typically **F**IFO.
	MainFraction             uint8                          `json:"main_fraction" csv:"main_fraction"`                             // The price denominator of the main fraction.
	PriceDisplayFormat       uint8                          `json:"price_display_format" csv:"price_display_format"`               // The number of digits to the right of the tick mark, to display fractional prices.
	SubFraction              uint8                          `json:"sub_fraction" csv:"sub_fraction"`                               // The price denominator of the sub fraction.
	UnderlyingProduct        uint8                          `json:"underlying_product" csv:"underlying_product"`                   // The product complex of the instrument.
	SecurityUpdateAction     byte                           `json:"security_update_action" csv:"security_update_action"`           // Indicates if the instrument definition has been added, modified, or deleted.
	MaturityMonth            uint8                          `json:"maturity_month" csv:"maturity_month"`                           // The calendar month reflected in the instrument symbol.
	MaturityDay              uint8                          `json:"maturity_day" csv:"maturity_day"`                               // The calendar day reflected in the instrument symbol, or 0.
	MaturityWeek             uint8                          `json:"maturity_week" csv:"maturity_week"`                             // The calendar week reflected in the instrument symbol, or 0.
	UserDefinedInstrument    UserDefinedInstrument          `json:"user_defined_instrument" csv:"user_defined_instrument"`         // Indicates if the instrument is user defined: **Y**es or **N**o.
	ContractMultiplierUnit   int8                           `json:"contract_multiplier_unit" csv:"contract_multiplier_unit"`       // The type of `contract_multiplier`. Either `1` for hours, or `2` for days.
	FlowScheduleType         int8                           `json:"flow_schedule_type" csv:"flow_schedule_type"`                   // The schedule for delivering electricity.
	TickRule                 uint8                          `json:"tick_rule" csv:"tick_rule"`                                     // The tick rule of the spread.
	LegInstrumentClass       byte                           `json:"leg_instrument_class" csv:"leg_instrument_class"`               /// The classification of the leg instrument.
	LegSide                  byte                           `json:"leg_side" csv:"leg_side"`                                       /// The side taken for the leg when purchasing the spread.
	Reserved                 [17]byte                       `json:"_reserved" csv:"_reserved"`                                     // Filler for alignment.
}

const InstrumentDefMsgV3_Size = RHeader_Size + 362 + (2 * MetadataV3_SymbolCstrLen)

func (*InstrumentDefMsgV3) RType() RType {
	return RType_InstrumentDef
}

func (*InstrumentDefMsgV3) RSize() uint16 {
	return InstrumentDefMsgV3_Size
}

func (r *InstrumentDefMsgV3) Fill_Raw(b []byte) error {
	if len(b) < InstrumentDefMsgV3_Size {
		return unexpectedBytesError(len(b), InstrumentDefMsgV3_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:]
	r.TsRecv = binary.LittleEndian.Uint64(body[0:8])
	r.MinPriceIncrement = int64(binary.LittleEndian.Uint64(body[8:16]))
	r.DisplayFactor = int64(binary.LittleEndian.Uint64(body[16:24]))
	r.Expiration = binary.LittleEndian.Uint64(body[24:32])
	r.Activation = binary.LittleEndian.Uint64(body[32:40])
	r.HighLimitPrice = int64(binary.LittleEndian.Uint64(body[40:48]))
	r.LowLimitPrice = int64(binary.LittleEndian.Uint64(body[48:56]))
	r.MaxPriceVariation = int64(binary.LittleEndian.Uint64(body[56:64]))
	r.UnitOfMeasureQty = int64(binary.LittleEndian.Uint64(body[64:72]))
	r.MinPriceIncrementAmount = int64(binary.LittleEndian.Uint64(body[72:80]))
	r.PriceRatio = int64(binary.LittleEndian.Uint64(body[80:88]))
	r.StrikePrice = int64(binary.LittleEndian.Uint64(body[88:96]))
	r.RawInstrumentID = binary.LittleEndian.Uint64(body[96:104])
	r.LegPrice = int64(binary.LittleEndian.Uint64(body[104:112]))
	r.LegDelta = int64(binary.LittleEndian.Uint64(body[112:120]))
	r.InstAttribValue = int32(binary.LittleEndian.Uint32(body[120:124]))
	r.UnderlyingID = binary.LittleEndian.Uint32(body[124:128])
	r.MarketDepthImplied = int32(binary.LittleEndian.Uint32(body[128:132]))
	r.MarketDepth = int32(binary.LittleEndian.Uint32(body[132:136]))
	r.MarketSegmentID = binary.LittleEndian.Uint32(body[136:140])
	r.MaxTradeVol = binary.LittleEndian.Uint32(body[140:144])
	r.MinLotSize = int32(binary.LittleEndian.Uint32(body[144:148]))
	r.MinLotSizeBlock = int32(binary.LittleEndian.Uint32(body[148:152]))
	r.MinLotSizeRoundLot = int32(binary.LittleEndian.Uint32(body[152:156]))
	r.MinTradeVol = binary.LittleEndian.Uint32(body[156:160])
	r.ContractMultiplier = int32(binary.LittleEndian.Uint32(body[160:164]))
	r.DecayQuantity = int32(binary.LittleEndian.Uint32(body[164:168]))
	r.OriginalContractSize = int32(binary.LittleEndian.Uint32(body[168:172]))
	r.LegInstrumentID = binary.LittleEndian.Uint32(body[172:176])
	r.LegRatioPriceNumerator = int32(binary.LittleEndian.Uint32(body[176:180]))
	r.LegRatioPriceDenominator = int32(binary.LittleEndian.Uint32(body[180:184]))
	r.LegRatioQtyNumerator = int32(binary.LittleEndian.Uint32(body[184:188]))
	r.LegRatioQtyDenominator = int32(binary.LittleEndian.Uint32(body[188:192]))
	r.LegUnderlyingID = binary.LittleEndian.Uint32(body[192:196])
	r.ApplID = int16(binary.LittleEndian.Uint16(body[196:198]))
	r.MaturityYear = binary.LittleEndian.Uint16(body[198:200])
	r.DecayStartDate = binary.LittleEndian.Uint16(body[200:202])
	r.ChannelID = binary.LittleEndian.Uint16(body[202:204])
	r.LegCount = binary.LittleEndian.Uint16(body[204:206])
	r.LegIndex = binary.LittleEndian.Uint16(body[206:208])
	copy(r.Currency[:], body[208:212])
	copy(r.SettlCurrency[:], body[212:216])
	copy(r.Secsubtype[:], body[216:222])
	copy(r.RawSymbol[:], body[222:222+MetadataV3_SymbolCstrLen])
	copy(r.Group[:], body[293:314])
	copy(r.Exchange[:], body[314:319])
	copy(r.Asset[:], body[319:319+MetadataV3_AssetCStrLen])
	copy(r.Cfi[:], body[330:337])
	copy(r.SecurityType[:], body[337:344])
	copy(r.UnitOfMeasure[:], body[344:375])
	copy(r.Underlying[:], body[375:396])
	copy(r.StrikePriceCurrency[:], body[396:400])
	copy(r.LegRawSymbol[:], body[400:400+MetadataV3_SymbolCstrLen])
	r.InstrumentClass = body[471]
	r.MatchAlgorithm = body[472]
	r.MainFraction = body[473]
	r.PriceDisplayFormat = body[474]
	r.SubFraction = body[475]
	r.UnderlyingProduct = body[476]
	r.SecurityUpdateAction = body[477]
	r.MaturityMonth = body[478]
	r.MaturityDay = body[479]
	r.MaturityWeek = body[480]
	r.UserDefinedInstrument = UserDefinedInstrument(body[481])
	r.ContractMultiplierUnit = int8(body[482])
	r.FlowScheduleType = int8(body[483])
	r.TickRule = body[484]
	r.LegInstrumentClass = body[485]
	r.LegSide = body[486]
	return nil
}

func (r *InstrumentDefMsgV3) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.MinPriceIncrement = fastjson_GetInt64FromString(val, "min_price_increment")
	r.DisplayFactor = fastjson_GetInt64FromString(val, "display_factor")
	r.Expiration = fastjson_GetUint64FromString(val, "expiration")
	r.Activation = fastjson_GetUint64FromString(val, "activation")
	r.HighLimitPrice = fastjson_GetInt64FromString(val, "high_limit_price")
	r.LowLimitPrice = fastjson_GetInt64FromString(val, "low_limit_price")
	r.MaxPriceVariation = fastjson_GetInt64FromString(val, "max_price_variation")
	r.UnitOfMeasureQty = fastjson_GetInt64FromString(val, "unit_of_measure_qty")
	r.MinPriceIncrementAmount = fastjson_GetInt64FromString(val, "min_price_increment_amount")
	r.PriceRatio = fastjson_GetInt64FromString(val, "price_ratio")
	r.StrikePrice = fastjson_GetInt64FromString(val, "strike_price")
	r.RawInstrumentID = fastjson_GetUint64Tolerant(val, "raw_instrument_id") // V2=number, V3=quoted string
	r.LegPrice = fastjson_GetInt64FromString(val, "leg_price")
	r.LegDelta = fastjson_GetInt64FromString(val, "leg_delta")
	r.InstAttribValue = int32(val.GetInt("inst_attrib_value"))
	r.UnderlyingID = uint32(val.GetUint("underlying_id"))
	r.MarketDepthImplied = int32(val.GetInt("market_depth_implied"))
	r.MarketDepth = int32(val.GetInt("market_depth"))
	r.MarketSegmentID = uint32(val.GetUint("market_segment_id"))
	r.MaxTradeVol = uint32(val.GetUint("max_trade_vol"))
	r.MinLotSize = int32(val.GetInt("min_lot_size"))
	r.MinLotSizeBlock = int32(val.GetInt("min_lot_size_block"))
	r.MinLotSizeRoundLot = int32(val.GetInt("min_lot_size_round_lot"))
	r.MinTradeVol = uint32(val.GetUint("min_trade_vol"))
	r.ContractMultiplier = int32(val.GetInt("contract_multiplier"))
	r.DecayQuantity = int32(val.GetInt("decay_quantity"))
	r.OriginalContractSize = int32(val.GetInt("original_contract_size"))
	r.LegInstrumentID = uint32(val.GetUint("leg_instrument_id"))
	r.LegRatioPriceNumerator = int32(val.GetInt("leg_ratio_price_numerator"))
	r.LegRatioPriceDenominator = int32(val.GetInt("leg_ratio_price_denominator"))
	r.LegRatioQtyNumerator = int32(val.GetInt("leg_ratio_qty_numerator"))
	r.LegRatioQtyDenominator = int32(val.GetInt("leg_ratio_qty_denominator"))
	r.LegUnderlyingID = uint32(val.GetUint("leg_underlying_id"))
	r.ApplID = int16(val.GetInt("appl_id"))
	r.MaturityYear = uint16(val.GetUint("maturity_year"))
	r.DecayStartDate = uint16(val.GetUint("decay_start_date"))
	r.ChannelID = uint16(val.GetUint("channel_id"))
	r.LegCount = uint16(val.GetUint("leg_count"))
	r.LegIndex = uint16(val.GetUint("leg_index"))
	copy(r.Currency[:], val.GetStringBytes("currency"))
	copy(r.SettlCurrency[:], val.GetStringBytes("settl_currency"))
	copy(r.Secsubtype[:], val.GetStringBytes("secsubtype"))
	copy(r.RawSymbol[:], val.GetStringBytes("raw_symbol"))
	copy(r.Group[:], val.GetStringBytes("group"))
	copy(r.Exchange[:], val.GetStringBytes("exchange"))
	copy(r.Asset[:], val.GetStringBytes("asset"))
	copy(r.Cfi[:], val.GetStringBytes("cfi"))
	copy(r.SecurityType[:], val.GetStringBytes("security_type"))
	copy(r.UnitOfMeasure[:], val.GetStringBytes("unit_of_measure"))
	copy(r.Underlying[:], val.GetStringBytes("underlying"))
	copy(r.StrikePriceCurrency[:], val.GetStringBytes("strike_price_currency"))
	copy(r.LegRawSymbol[:], val.GetStringBytes("leg_raw_symbol"))
	r.InstrumentClass = byte(val.GetUint("instrument_class"))
	r.MatchAlgorithm = byte(val.GetUint("match_algorithm"))
	r.MainFraction = uint8(val.GetUint("main_fraction"))
	r.PriceDisplayFormat = uint8(val.GetUint("price_display_format"))
	r.SubFraction = uint8(val.GetUint("sub_fraction"))
	r.UnderlyingProduct = uint8(val.GetUint("underlying_product"))
	r.SecurityUpdateAction = byte(val.GetUint("security_update_action"))
	r.MaturityMonth = uint8(val.GetUint("maturity_month"))
	r.MaturityDay = uint8(val.GetUint("maturity_day"))
	r.MaturityWeek = uint8(val.GetUint("maturity_week"))
	r.UserDefinedInstrument = UserDefinedInstrument(val.GetUint("user_defined_instrument"))
	r.ContractMultiplierUnit = int8(val.GetUint("contract_multiplier_unit"))
	r.FlowScheduleType = int8(val.GetUint("flow_schedule_type"))
	r.TickRule = uint8(val.GetUint("tick_rule"))
	r.LegInstrumentClass = uint8(val.GetUint("leg_instrument_class"))
	r.LegSide = uint8(val.GetUint("leg_side"))
	return nil
}
