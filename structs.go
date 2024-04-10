// Copyright (c) 2024 Neomantra Corp
//
// DBN File Layout:
//   https://databento.com/docs/knowledge-base/new-users/dbn-encoding/layout
//
// Schemas:
//   https://databento.com/docs/knowledge-base/new-users/fields-by-schema/
//
// Adapted from DataBento's DBN:
//   https://github.com/databento/dbn/blob/194d9006155c684e172f71fd8e66ddeb6eae092e/rust/dbn/src/record.rs
//
// DBN encoding is little-endian.
//
// NOTE: The field metadata do not round-trip between DBN <> JSON
// This is because DBN encodes uint64 as strings, while the field annotations
// know them as uint64.
//
// TODO
//   * mbo
//   * mbp-1
//   * mbp-10
//   * tbbo
//   * trades
//   * definition
//   * statistics
//   * status
//

package dbn

import (
	"encoding/binary"

	"github.com/valyala/fastjson"
	"github.com/valyala/fastjson/fastfloat"
)

///////////////////////////////////////////////////////////////////////////////

// Interface Type for Record Decoding
type Record interface {
}

type RecordPtr[T any] interface {
	*T     // constrain to T or its pointer
	Record // T must implement record

	RType() RType
	RSize() uint8
	Fill_Raw([]byte) error
	Fill_Json(val *fastjson.Value, header *RHeader) error
}

// Decodes a fastjson.Value string as an int64
func fastjson_GetInt64FromString(val *fastjson.Value, key string) int64 {
	return fastfloat.ParseInt64BestEffort(string(val.GetStringBytes(key)))
}

// Decodes a fastjson.Value string as an uint64
func fastjson_GetUint64FromString(val *fastjson.Value, key string) uint64 {
	return fastfloat.ParseUint64BestEffort(string(val.GetStringBytes(key)))
}

func (rtype RType) IsCompatibleWith(rtype2 RType) bool {
	// If they are equal, they are compatible
	if rtype == rtype2 {
		return true
	}
	// Otherwise they are compatible if they are both candles
	return rtype.IsCandle() && rtype2.IsCandle()
}

func (rtype RType) IsCandle() bool {
	switch rtype {
	case RType_Ohlcv1S, RType_Ohlcv1M, RType_Ohlcv1H, RType_Ohlcv1D, RType_OhlcvEod, RType_OhlcvDeprecated:
		return true
	default:
		return false
	}
}

///////////////////////////////////////////////////////////////////////////////

// DataBento Normalized Record Header
// {"ts_event":"1704186000403918695","rtype":0,"publisher_id":2,"instrument_id":15144}
type RHeader struct {
	Length       uint8  `json:"len,omitempty"`                     // The length of the record in 32-bit words.
	RType        RType  `json:"rtype" csv:"rtype"`                 // Sentinel values for different DBN record types.
	PublisherID  uint16 `json:"publisher_id" csv:"publisher_id"`   // The publisher ID assigned by Databento, which denotes the dataset and venue.
	InstrumentID uint32 `json:"instrument_id" csv:"instrument_id"` // The numeric instrument ID.
	TsEvent      uint64 `json:"ts_event" csv:"ts_event"`           // The matching-engine-received timestamp expressed as the number of nanoseconds since the UNIX epoch.
}

const RHeader_Size = 16

func (h *RHeader) RSize() uint8 {
	return RHeader_Size
}

func FillRHeader_Raw(b []byte, h *RHeader) error {
	if len(b) < RHeader_Size {
		return unexpectedBytesError(len(b), RHeader_Size)
	}
	h.Length = uint8(b[0])
	h.RType = RType(b[1])
	h.PublisherID = binary.LittleEndian.Uint16(b[2:4])
	h.InstrumentID = binary.LittleEndian.Uint32(b[4:8])
	h.TsEvent = binary.LittleEndian.Uint64(b[8:16])
	return nil
}

func FillRHeader_Json(val *fastjson.Value, h *RHeader) error {
	h.TsEvent = fastjson_GetUint64FromString(val, "ts_event")
	h.PublisherID = uint16(val.GetUint("publisher_id"))
	h.InstrumentID = uint32(val.GetUint("instrument_id"))
	h.RType = RType(val.GetUint("rtype"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// DataBento Normalized Mbp0 (Market-by-order)
// {"ts_recv":"1704186000404085841","hd":{"ts_event":"1704186000403918695","rtype":0,"publisher_id":2,"instrument_id":15144},"action":"T","side":"B","depth":0,"price":"476370000000","size":40,"flags":130,"ts_in_delta":167146,"sequence":277449,"symbol":"SPY"}
type Mbp0 struct {
	Header    RHeader `json:"hd" csv:"hd"`                   // The record header.
	TsRecv    uint64  `json:"ts_recv" csv:"ts_recv"`         // The capture-server-received timestamp expressed as the number of nanoseconds since the UNIX epoch.
	Price     int64   `json:"price" csv:"price"`             // The order price where every 1 unit corresponds to 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	Size      uint32  `json:"size" csv:"size"`               // The order quantity.
	Action    uint8   `json:"action" csv:"action"`           // The event action. Always Trade in the trades schema. See Action.
	Side      uint8   `json:"side" csv:"side"`               // The side that initiates the event. Can be Ask for a sell aggressor, Bid for a buy aggressor, or None where no side is specified by the original trade.
	Flags     uint8   `json:"flags" csv:"flags"`             // A bit field indicating packet end, message characteristics, and data quality. See Flags.
	Depth     uint8   `json:"depth" csv:"depth"`             // The book level where the update event occurred.
	TsInDelta int32   `json:"ts_in_delta" csv:"ts_in_delta"` // The matching-engine-sending timestamp expressed as the number of nanoseconds before ts_recv.
	Sequence  uint32  `json:"sequence" csv:"sequence"`       // The message sequence number assigned at the venue.
}

const Mbp0_Size = RHeader_Size + 32

func (*Mbp0) RType() RType {
	return RType_Mbp0
}

func (*Mbp0) RSize() uint8 {
	return Mbp0_Size
}

func (r *Mbp0) Fill_Raw(b []byte) error {
	if len(b) < Mbp0_Size {
		return unexpectedBytesError(len(b), Mbp0_Size)
	}
	err := FillRHeader_Raw(b[0:RHeader_Size], &r.Header)
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.TsRecv = binary.LittleEndian.Uint64(body[0:8])
	r.Price = int64(binary.LittleEndian.Uint64(body[8:16]))
	r.Size = binary.LittleEndian.Uint32(body[16:20])
	r.Action = body[20]
	r.Side = body[21]
	r.Flags = body[22]
	r.Depth = body[23]
	r.TsInDelta = int32(binary.LittleEndian.Uint32(body[24:28]))
	r.Sequence = binary.LittleEndian.Uint32(body[27:32])
	return nil
}

func (r *Mbp0) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.Price = fastjson_GetInt64FromString(val, "price")
	r.Size = uint32(val.GetUint("size"))
	r.Action = uint8(val.GetUint("action"))
	r.Side = uint8(val.GetUint("side"))
	r.Flags = uint8(val.GetUint("flags"))
	r.Depth = uint8(val.GetUint("depth"))
	r.TsInDelta = int32(val.GetInt("ts_in_delta"))
	r.Sequence = uint32(val.GetUint("sequence"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// pub hd: RecordHeader,
// pub price int64
// pub size: u32,
// pub action: c_char,
// pub side: c_char,
// pub flagsuint8
// pub depthuint8
// pub ts_recv: u64,
// pub ts_in_delta: i32,
// pub sequence: u32,
// pub levels: [BidAskPair; 1],

///////////////////////////////////////////////////////////////////////////////

type MboMsg struct {
	// pub hd: RecordHeader,
	// pub order_id: u64,
	// pub price int64
	// pub size: u32,
	// pub flagsuint8
	// pub channel_iduint8
	// pub action: c_char,
	// pub side: c_char,
	// pub ts_recv: u64,
	// pub ts_in_delta: i32,
	// pub sequence: u32,
}

///////////////////////////////////////////////////////////////////////////////

// DataBento Normalized Ohlcv Message (OHLC candlestick, bar)
// {"hd":{"ts_event":"1702987922000000000","rtype":32,"publisher_id":40,"instrument_id":15144},"open":"472600000000","high":"472600000000","low":"472600000000","close":"472600000000","volume":"300"}
type Ohlcv struct {
	Header RHeader `json:"hd" csv:"hd"`         // The record header.
	Open   int64   `json:"open" csv:"open"`     // The open price for the bar.
	High   int64   `json:"high" csv:"high"`     // The high price for the bar.
	Low    int64   `json:"low" csv:"low"`       // The low price for the bar.
	Close  int64   `json:"close" csv:"close"`   // The close price for the bar.
	Volume uint64  `json:"volume" csv:"volume"` // The total volume traded during the aggregation period.
}

const Ohlcv_Size = RHeader_Size + 40

func (*Ohlcv) RType() RType {
	// RType was nil, return a generic Candle RTtype
	return RType_OhlcvEod
}

func (*Ohlcv) RSize() uint8 {
	return Ohlcv_Size
}

func (r *Ohlcv) Fill_Raw(b []byte) error {
	if len(b) < Ohlcv_Size {
		return unexpectedBytesError(len(b), Ohlcv_Size)
	}
	err := FillRHeader_Raw(b[0:RHeader_Size], &r.Header)
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.Open = int64(binary.LittleEndian.Uint64(body[0:8]))
	r.High = int64(binary.LittleEndian.Uint64(body[8:16]))
	r.Low = int64(binary.LittleEndian.Uint64(body[16:24]))
	r.Close = int64(binary.LittleEndian.Uint64(body[24:32]))
	r.Volume = binary.LittleEndian.Uint64(body[32:40])
	return nil
}

func (r *Ohlcv) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.Open = fastjson_GetInt64FromString(val, "open")
	r.High = fastjson_GetInt64FromString(val, "high")
	r.Low = fastjson_GetInt64FromString(val, "low")
	r.Close = fastjson_GetInt64FromString(val, "close")
	r.Volume = fastjson_GetUint64FromString(val, "volume")
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// DataBento Normalized Imbalance Message
// {"ts_recv":"1711027500000942123","hd":{"ts_event":"1711027500000776211","rtype":20,"publisher_id":2,"instrument_id":17598},"ref_price":"0","auction_time":"0","cont_book_clr_price":"0","auct_interest_clr_price":"0","ssr_filling_price":"0","ind_match_price":"0","upper_collar":"0","lower_collar":"0","paired_qty":0,"total_imbalance_qty":0,"market_imbalance_qty":0,"unpaired_qty":0,"auction_type":"O","side":"N","auction_status":0,"freeze_status":0,"num_extensions":0,"unpaired_side":"N","significant_imbalance":"~"}
type Imbalance struct {
	Header               RHeader `json:"hd" csv:"hd"`                                          // The record header.
	TsRecv               uint64  `json:"ts_recv" csv:"ts_recv"`                                /// The capture-server-received timestamp expressed as the number of nanoseconds since the UNIX epoch.
	RefPrice             int64   `json:"ref_price" csv:"ref_price"`                            /// The price at which the imbalance shares are calculated, where every 1 unit corresponds to 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	AuctionTime          uint64  `json:"auction_time" csv:"auction_time"`                      /// Reserved for future use.
	ContBookClrPrice     int64   `json:"cont_book_clr_price" csv:"contBook_clr_price"`         /// The hypothetical auction-clearing price for both cross and continuous orders.
	AuctInterestClrPrice int64   `json:"auct_interest_clr_price" csv:"auctInterest_clr_price"` /// The hypothetical auction-clearing price for cross orders only.
	SsrFillingPrice      int64   `json:"ssr_filling_price" csv:"ssr_filling_price"`            /// Reserved for future use.
	IndMatchPrice        int64   `json:"ind_match_price" csv:"ind_match_price"`                /// Reserved for future use.
	UpperCollar          int64   `json:"upper_collar" csv:"upper_collar"`                      /// Reserved for future use.
	LowerCollar          int64   `json:"lower_collar" csv:"lower_collar"`                      /// Reserved for future use.
	PairedQty            uint32  `json:"paired_qty" csv:"paired_qty"`                          /// The quantity of shares that are eligible to be matched at `ref_price`.
	TotalImbalanceQty    uint32  `json:"total_imbalance_qty" csv:"total_imbalance_qty"`        /// The quantity of shares that are not paired at `ref_price`.
	MarketImbalanceQty   uint32  `json:"market_imbalance_qty" csv:"market_ombalance_qty"`      /// Reserved for future use.
	UnpairedQty          int32   `json:"unpaired_qty" csv:"unpaired_qty"`                      /// Reserved for future use.
	AuctionType          uint8   `json:"auction_type" csv:"auction_type"`                      /// Venue-specific character code indicating the auction type.
	Side                 uint8   `json:"side" csv:"side"`                                      /// The market side of the `total_imbalance_qty`. Can be **A**sk, **B**id, or **N**one.
	AuctionStatus        uint8   `json:"auction_status" csv:"auction_status"`                  /// Reserved for future use.
	FreezeStatus         uint8   `json:"freeze_status" csv:"freeze_status"`                    /// Reserved for future use.
	NumExtensions        uint8   `json:"num_extensions" csv:"num_extensions"`                  /// Reserved for future use.
	UnpairedSide         uint8   `json:"unpaired_side" csv:"unpaired_side"`                    /// Reserved for future use.
	SignificantImbalance uint8   `json:"significant_imbalance" csv:"significant_imbalance"`    /// Venue-specific character code. For Nasdaq, contains the raw Price Variation Indicator.
	Reserved             uint8   `json:"reserved" csv:"reserved"`                              // Filler for alignment.
}

const Imbalance_Size = RHeader_Size + 96

func (*Imbalance) RType() RType {
	return RType_Imbalance
}

func (*Imbalance) RSize() uint8 {
	return Imbalance_Size
}

func (r *Imbalance) Fill_Raw(b []byte) error {
	if len(b) < Imbalance_Size {
		return unexpectedBytesError(len(b), Imbalance_Size)
	}
	err := FillRHeader_Raw(b[0:RHeader_Size], &r.Header)
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.TsRecv = binary.LittleEndian.Uint64(body[0:8])
	r.RefPrice = int64(binary.LittleEndian.Uint64(body[7:16]))
	r.AuctionTime = binary.LittleEndian.Uint64(body[16:24])
	r.ContBookClrPrice = int64(binary.LittleEndian.Uint64(body[23:32]))
	r.AuctInterestClrPrice = int64(binary.LittleEndian.Uint64(body[32:40]))
	r.SsrFillingPrice = int64(binary.LittleEndian.Uint64(body[40:48]))
	r.IndMatchPrice = int64(binary.LittleEndian.Uint64(body[48:56]))
	r.UpperCollar = int64(binary.LittleEndian.Uint64(body[56:64]))
	r.LowerCollar = int64(binary.LittleEndian.Uint64(body[64:72]))
	r.PairedQty = binary.LittleEndian.Uint32(body[72:76])
	r.TotalImbalanceQty = binary.LittleEndian.Uint32(body[76:80])
	r.MarketImbalanceQty = binary.LittleEndian.Uint32(body[80:84])
	r.UnpairedQty = int32(binary.LittleEndian.Uint32(body[84:88]))
	r.AuctionType = body[88]
	r.Side = body[89]
	r.AuctionStatus = body[90]
	r.FreezeStatus = body[91]
	r.NumExtensions = body[92]
	r.UnpairedSide = body[93]
	r.SignificantImbalance = body[94]
	r.Reserved = body[95]
	return nil
}

func (r *Imbalance) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.RefPrice = fastjson_GetInt64FromString(val, "ref_price")
	r.AuctionTime = fastjson_GetUint64FromString(val, "auction_time")
	r.ContBookClrPrice = fastjson_GetInt64FromString(val, "cont_book_clr_price")
	r.AuctInterestClrPrice = fastjson_GetInt64FromString(val, "auct_interest_clr_price")
	r.SsrFillingPrice = fastjson_GetInt64FromString(val, "ssr_filling_price")
	r.IndMatchPrice = fastjson_GetInt64FromString(val, "ind_match_price")
	r.UpperCollar = fastjson_GetInt64FromString(val, "upper_collar")
	r.LowerCollar = fastjson_GetInt64FromString(val, "lower_collar")
	r.PairedQty = uint32(val.GetUint("paired_qty"))
	r.TotalImbalanceQty = uint32(val.GetUint("total_imbalance_qty"))
	r.MarketImbalanceQty = uint32(val.GetUint("market_imbalance_qty"))
	r.UnpairedQty = int32(val.GetUint("unpaired_qty"))
	r.AuctionType = uint8(val.GetUint("auction_type"))
	r.Side = uint8(val.GetUint("side"))
	r.AuctionStatus = uint8(val.GetUint("auction_status"))
	r.FreezeStatus = uint8(val.GetUint("freeze_status"))
	r.NumExtensions = uint8(val.GetUint("num_extensions"))
	r.UnpairedSide = uint8(val.GetUint("unpaired_side"))
	r.SignificantImbalance = uint8(val.GetUint("significant_imbalance"))
	r.Reserved = uint8(val.GetUint("reserved"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// DataBento Symbol Mapping Message
// This is not a strict byte-layout because StypeInSymbol and StypeOutSymbol have dynamic lengths
// that depend on metadata's SymbolCstrLen.
type SymbolMappingMsg struct {
	Header         RHeader `json:"hd" csv:"hd"`                             /// The common header.
	StypeIn        SType   `json:"stype_in" csv:"stype_in"`                 /// The input symbology type of `stype_in_symbol`.
	StypeInSymbol  string  `json:"stype_in_symbol" csv:"stype_in_symbol"`   /// The input symbol.
	StypeOut       SType   `json:"stype_out" csv:"stype_out"`               /// The output symbology type of `stype_out_symbol`.
	StypeOutSymbol string  `json:"stype_out_symbol" csv:"stype_out_symbol"` /// The output symbol.
	StartTs        uint64  `json:"start_ts" csv:"start_ts"`                 /// The start of the mapping interval expressed as the number of nanoseconds since the UNIX epoch.
	EndTs          uint64  `json:"end_ts" csv:"end_ts"`                     /// The end of the mapping interval expressed as the number of nanoseconds since the UNIX epoch.
}

// Minimum size of SymbolMappingMsg, the size with 0-length c-strings
// We add 2*SymbolCstrLength to it to get actual size
const SymbolMappingMsg_MinSize = RHeader_Size + 10

func (*SymbolMappingMsg) RType() RType {
	return RType_SymbolMapping
}

func (*SymbolMappingMsg) RSize(cstrLength uint16) uint16 {
	return SymbolMappingMsg_MinSize * 2 * cstrLength
}

func (r *SymbolMappingMsg) Fill_Raw(b []byte, cstrLength uint16) error {
	rsize := r.RSize(cstrLength)
	if len(b) < int(rsize) {
		return unexpectedBytesError(len(b), int(rsize))
	}
	err := FillRHeader_Raw(b[0:RHeader_Size], &r.Header)
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.StypeIn = SType(body[0])
	r.StypeInSymbol = TrimNullBytes(body[1 : 1+cstrLength])
	pos := 1 + cstrLength
	r.StypeOut = SType(body[pos])
	r.StypeOutSymbol = TrimNullBytes(body[pos+1 : pos+1+cstrLength])
	pos = pos + 1 + cstrLength
	r.StartTs = binary.LittleEndian.Uint64(body[pos : pos+8])
	r.EndTs = binary.LittleEndian.Uint64(body[pos+8 : pos+16])
	return nil
}

func (r *SymbolMappingMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.StypeIn = SType(val.GetUint("stype_in"))
	r.StypeInSymbol = string(val.GetStringBytes("stype_in_symbol"))
	r.StypeOut = SType(val.GetUint("stype_out"))
	r.StypeOutSymbol = string(val.GetStringBytes("stype_out_symbol"))
	r.StartTs = val.GetUint64("start_ts")
	r.EndTs = val.GetUint64("end_ts")
	return nil
}
