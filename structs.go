// Copyright (c) 2024 Neomantra Corp
//
// DBN File Layout:
//   https://databento.com/docs/knowledge-base/new-users/dbn-encoding/layout
//
// Schemas:
//   https://databento.com/docs/knowledge-base/new-users/fields-by-schema/
//
// Adapted from Databento's DBN:
//   https://github.com/databento/dbn/blob/194d9006155c684e172f71fd8e66ddeb6eae092e/rust/dbn/src/record.rs
//
// DBN encoding is little-endian.
//
// NOTE: The field metadata do not round-trip between DBN <> JSON
// This is because DBN encodes uint64 as strings, while the field annotations
// know them as uint64.
//

package dbn

import (
	"bytes"
	"encoding/binary"
	"errors"

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
	RSize() uint16
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

// Decodes a fastjson.Value as int64, tolerant of both quoted strings (V3) and bare numbers (V2).
func fastjson_GetInt64Tolerant(val *fastjson.Value, key string) int64 {
	v := val.Get(key)
	if v == nil {
		return 0
	}
	if v.Type() == fastjson.TypeString {
		return fastfloat.ParseInt64BestEffort(string(v.GetStringBytes()))
	}
	return v.GetInt64()
}

// Decodes a fastjson.Value as uint64, tolerant of both quoted strings (V3) and bare numbers (V2).
func fastjson_GetUint64Tolerant(val *fastjson.Value, key string) uint64 {
	v := val.Get(key)
	if v == nil {
		return 0
	}
	if v.Type() == fastjson.TypeString {
		return fastfloat.ParseUint64BestEffort(string(v.GetStringBytes()))
	}
	return v.GetUint64()
}

func (rtype RType) IsCompatibleWith(rtype2 RType) bool {
	// If they are equal, they are compatible
	if rtype == rtype2 {
		return true
	}
	// Otherwise they are compatible if they are both candles or both BBO
	return (rtype.IsCandle() && rtype2.IsCandle()) || (rtype.IsBbo() && rtype2.IsBbo())
}

func (rtype RType) IsCandle() bool {
	switch rtype {
	case RType_Ohlcv1S, RType_Ohlcv1M, RType_Ohlcv1H, RType_Ohlcv1D, RType_OhlcvEod, RType_OhlcvDeprecated:
		return true
	default:
		return false
	}
}

func (rtype RType) IsBbo() bool {
	switch rtype {
	case RType_Bbo1S, RType_Bbo1M:
		return true
	default:
		return false
	}
}

///////////////////////////////////////////////////////////////////////////////

// Databento Normalized Record Header
// {"ts_event":"1704186000403918695","rtype":0,"publisher_id":2,"instrument_id":15144}
type RHeader struct {
	Length       uint8  `json:"len,omitempty"`                     // The length of the record in 32-bit words.
	RType        RType  `json:"rtype" csv:"rtype"`                 // Sentinel values for different DBN record types.
	PublisherID  uint16 `json:"publisher_id" csv:"publisher_id"`   // The publisher ID assigned by Databento, which denotes the dataset and venue.
	InstrumentID uint32 `json:"instrument_id" csv:"instrument_id"` // The numeric instrument ID.
	TsEvent      uint64 `json:"ts_event" csv:"ts_event"`           // The matching-engine-received timestamp expressed as the number of nanoseconds since the UNIX epoch.
}

const RHeader_Size = 16

// Minimum size of SymbolMappingMsg, the size with 0-length c-strings
// We add 2*SymbolCstrLength to it to get actual size.
const SymbolMappingMsg_MinSize = RHeader_Size + 16

func (h *RHeader) RSize() uint16 {
	return RHeader_Size
}

func (h *RHeader) Fill_Raw(b []byte) error {
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

func (h *RHeader) Fill_Json(val *fastjson.Value) error {
	h.TsEvent = fastjson_GetUint64FromString(val, "ts_event")
	h.PublisherID = uint16(val.GetUint("publisher_id"))
	h.InstrumentID = uint32(val.GetUint("instrument_id"))
	h.RType = RType(val.GetUint("rtype"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

type BidAskPair struct {
	BidPx int64  `json:"bid_px" csv:"bid_px"` // The bid price.
	AskPx int64  `json:"ask_px" csv:"ask_px"` // The ask price.
	BidSz uint32 `json:"bid_sz" csv:"bid_sz"` // The bid size.
	AskSz uint32 `json:"ask_sz" csv:"ask_sz"` // The ask size.
	BidCt uint32 `json:"bid_ct" csv:"bid_ct"` // The bid order count.
	AskCt uint32 `json:"ask_ct" csv:"ask_ct"` // The ask order count.
}

const BidAskPair_Size = 32

func (p *BidAskPair) Fill_Raw(b []byte) error {
	p.BidPx = int64(binary.LittleEndian.Uint64(b[0:8]))
	p.AskPx = int64(binary.LittleEndian.Uint64(b[8:16]))
	p.BidSz = binary.LittleEndian.Uint32(b[16:20])
	p.AskSz = binary.LittleEndian.Uint32(b[20:24])
	p.BidCt = binary.LittleEndian.Uint32(b[24:28])
	p.AskCt = binary.LittleEndian.Uint32(b[28:32])
	return nil
}

func (p *BidAskPair) Fill_Json(val *fastjson.Value) error {
	p.BidPx = fastjson_GetInt64FromString(val, "bid_px")
	p.AskPx = fastjson_GetInt64FromString(val, "ask_px")
	p.BidSz = uint32(val.GetUint("bid_sz"))
	p.AskSz = uint32(val.GetUint("ask_sz"))
	p.BidCt = uint32(val.GetUint("bid_ct"))
	p.AskCt = uint32(val.GetUint("ask_ct"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// A price level consolidated from multiple venues.
type ConsolidatedBidAskPair struct {
	BidPx     int64  `json:"bid_px" csv:"bid_px"`         // The bid price.
	AskPx     int64  `json:"ask_px" csv:"ask_px"`         // The ask price.
	BidSz     uint32 `json:"bid_sz" csv:"bid_sz"`         // The bid size.
	AskSz     uint32 `json:"ask_sz" csv:"ask_sz"`         // The ask size.
	BidPb     uint16 `json:"bid_pb" csv:"bid_pb"`         // The bid publisher ID assigned by Databento, which denotes the dataset and venue.
	Reserved1 uint16 `json:"_reserved1" csv:"_reserved1"` // Reserved for later usage
	AskPb     uint16 `json:"ask_pb" csv:"ask_pb"`         // The ask publisher ID assigned by Databento, which denotes the dataset and venue.
	Reserved2 uint16 `json:"_reserved2" csv:"_reserved2"` // Reserved for later usage
}

const ConsolidatedBidAskPair_Size = 32

func (p *ConsolidatedBidAskPair) Fill_Raw(b []byte) error {
	p.BidPx = int64(binary.LittleEndian.Uint64(b[0:8]))
	p.AskPx = int64(binary.LittleEndian.Uint64(b[8:16]))
	p.BidSz = binary.LittleEndian.Uint32(b[16:20])
	p.AskSz = binary.LittleEndian.Uint32(b[20:24])
	p.BidPb = binary.LittleEndian.Uint16(b[24:26])
	// Reserved1 26:28
	p.AskPb = binary.LittleEndian.Uint16(b[28:30])
	// Reserved2 30:32
	return nil
}

func (p *ConsolidatedBidAskPair) Fill_Json(val *fastjson.Value) error {
	p.BidPx = fastjson_GetInt64FromString(val, "bid_px")
	p.AskPx = fastjson_GetInt64FromString(val, "ask_px")
	p.BidSz = uint32(val.GetUint("bid_sz"))
	p.AskSz = uint32(val.GetUint("ask_sz"))
	p.BidPb = uint16(val.GetUint("bid_pb"))
	p.AskPb = uint16(val.GetUint("ask_pb"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// Databento Normalized Mbp0 message (Market-by-price depth0)
// {"ts_recv":"1704186000404085841","hd":{"ts_event":"1704186000403918695","rtype":0,"publisher_id":2,"instrument_id":15144},"action":"T","side":"B","depth":0,"price":"476370000000","size":40,"flags":130,"ts_in_delta":167146,"sequence":277449,"symbol":"SPY"}
type Mbp0Msg struct { // TradeMsg
	Header    RHeader `json:"hd" csv:"hd"`                   // The record header.
	Price     int64   `json:"price" csv:"price"`             // The order price where every 1 unit corresponds to 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	Size      uint32  `json:"size" csv:"size"`               // The order quantity.
	Action    uint8   `json:"action" csv:"action"`           // The event action. Always Trade in the trades schema. See Action.
	Side      uint8   `json:"side" csv:"side"`               // The side that initiates the event. Can be Ask for a sell aggressor, Bid for a buy aggressor, or None where no side is specified by the original trade.
	Flags     uint8   `json:"flags" csv:"flags"`             // A bit field indicating packet end, message characteristics, and data quality. See Flags.
	Depth     uint8   `json:"depth" csv:"depth"`             // The book level where the update event occurred.
	TsRecv    uint64  `json:"ts_recv" csv:"ts_recv"`         // The capture-server-received timestamp expressed as the number of nanoseconds since the UNIX epoch.
	TsInDelta int32   `json:"ts_in_delta" csv:"ts_in_delta"` // The matching-engine-sending timestamp expressed as the number of nanoseconds before ts_recv.
	Sequence  uint32  `json:"sequence" csv:"sequence"`       // The message sequence number assigned at the venue.
}

const Mbp0Msg_Size = RHeader_Size + 32

func (*Mbp0Msg) RType() RType {
	return RType_Mbp0
}

func (*Mbp0Msg) RSize() uint16 {
	return Mbp0Msg_Size
}

func (r *Mbp0Msg) Fill_Raw(b []byte) error {
	if len(b) < Mbp0Msg_Size {
		return unexpectedBytesError(len(b), Mbp0Msg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.Price = int64(binary.LittleEndian.Uint64(body[0:8]))
	r.Size = binary.LittleEndian.Uint32(body[8:12])
	r.Action = body[12]
	r.Side = body[13]
	r.Flags = body[14]
	r.Depth = body[15]
	r.TsRecv = binary.LittleEndian.Uint64(body[16:24])
	r.TsInDelta = int32(binary.LittleEndian.Uint32(body[24:28]))
	r.Sequence = binary.LittleEndian.Uint32(body[28:32])
	return nil
}

func (r *Mbp0Msg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.Price = fastjson_GetInt64FromString(val, "price")
	r.Size = uint32(val.GetUint("size"))
	r.Action = uint8(val.GetUint("action"))
	r.Side = uint8(val.GetUint("side"))
	r.Flags = uint8(val.GetUint("flags"))
	r.Depth = uint8(val.GetUint("depth"))
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.TsInDelta = int32(val.GetInt("ts_in_delta"))
	r.Sequence = uint32(val.GetUint("sequence"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// Databento Normalized market-by-order (MBO) message.
// The record of the [`Mbo`](crate::enums::Schema::Mbo) schema.
type MboMsg struct {
	Header    RHeader `json:"hd" csv:"hd"`                   // The record header.
	OrderID   uint64  `json:"order_id" csv:"order_id"`       // The order ID assigned at the venue.
	Price     int64   `json:"price" csv:"price"`             // The order price expressed as a signed integer where every 1 unit corresponds to 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	Size      uint32  `json:"size" csv:"size"`               // The order quantity.
	Flags     uint8   `json:"flags" csv:"flags"`             // A bit field indicating event end, message characteristics, and data quality. See [`enums::flags`](crate::enums::flags) for possible values.
	ChannelID uint8   `json:"channel_id" csv:"channel_id"`   // The channel ID assigned by Databento as an incrementing integer starting at zero.
	Action    byte    `json:"action" csv:"action"`           // The event action. Can be **A**dd, **C**ancel, **M**odify, clea**R**,  **T**rade, **F**ill, or **NN**one.
	Side      byte    `json:"side" csv:"side"`               // The side that initiates the event. Can be **A**sk for a sell order (or sell aggressor in a trade), **B**id for a buy order (or buy aggressor in a trade), or **N**one where no side is specified by the original source.
	TsRecv    uint64  `json:"ts_recv" csv:"ts_recv"`         // The capture-server-received timestamp expressed as the number of nanoseconds since the UNIX epoch.
	TsInDelta int32   `json:"ts_in_delta" csv:"ts_in_delta"` // The delta of `ts_recv - ts_exchange_send`, max 2 seconds.
	Sequence  uint32  `json:"sequence" csv:"sequence"`       // The message sequence number assigned at the venue.
}

const MboMsg_Size = RHeader_Size + 40

func (*MboMsg) RType() RType {
	return RType_Mbo
}

func (*MboMsg) RSize() uint16 {
	return MboMsg_Size
}

func (r *MboMsg) Fill_Raw(b []byte) error {
	if len(b) < MboMsg_Size {
		return unexpectedBytesError(len(b), MboMsg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.OrderID = binary.LittleEndian.Uint64(body[0:8])
	r.Price = int64(binary.LittleEndian.Uint64(body[8:16]))
	r.Size = binary.LittleEndian.Uint32(body[16:20])
	r.Flags = body[20]
	r.ChannelID = body[21]
	r.Action = body[22]
	r.Side = body[23]
	r.TsRecv = binary.LittleEndian.Uint64(body[24:32])
	r.TsInDelta = int32(binary.LittleEndian.Uint32(body[32:36]))
	r.Sequence = binary.LittleEndian.Uint32(body[36:40])
	return nil
}

func (r *MboMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.OrderID = fastjson_GetUint64FromString(val, "order_id")
	r.Price = fastjson_GetInt64FromString(val, "price")
	r.Size = uint32(val.GetUint("size"))
	r.Flags = uint8(val.GetUint("flags"))
	r.ChannelID = uint8(val.GetUint("channel_id"))
	r.Action = byte(val.GetUint("action"))
	r.Side = byte(val.GetUint("side"))
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.TsInDelta = int32(val.GetUint("ts_in_delta"))
	r.Sequence = uint32(val.GetUint("sequence"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// Databento Normalized market-by-price (MBP) implementation with a known book depth of 1. The record of the [`Mbp1`](crate::enums::Schema::Mbp1) schema.
type Mbp1Msg struct {
	Header    RHeader    `json:"hd" csv:"hd"`                   // The record header.
	Price     int64      `json:"price" csv:"price"`             // The order price expressed as a signed integer where every 1 unit corresponds to 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	Size      uint32     `json:"size" csv:"size"`               // The order quantity.
	Action    byte       `json:"action" csv:"action"`           // The event action. Can be **A**dd, **C**ancel, **M**odify, clea**R**, or **T**rade.
	Side      byte       `json:"side" csv:"side"`               // The side that initiates the event. Can be **A**sk for a sell order (or sell aggressor in a trade), **B**id for a buy order (or buy aggressor in a trade), or **N**one where no side is specified by the original source.
	Flags     uint8      `json:"flags" csv:"flags"`             // A bit field indicating event end, message characteristics, and data quality. See [`enums::flags`](crate::enums::flags) for possible values.
	Depth     uint8      `json:"depth" csv:"depth"`             // The depth of actual book change.
	TsRecv    uint64     `json:"ts_recv" csv:"ts_recv"`         // The capture-server-received timestamp expressed as number of nanoseconds since the UNIX epoch.
	TsInDelta int32      `json:"ts_in_delta" csv:"ts_in_delta"` // The delta of `ts_recv - ts_exchange_send`, max 2 seconds.
	Sequence  uint32     `json:"sequence" csv:"sequence"`       // The message sequence number assigned at the venue.
	Level     BidAskPair `json:"levels" csv:"levels"`           // The top of the order book.
}

const Mbp1Msg_Size = RHeader_Size + 64

func (*Mbp1Msg) RType() RType {
	return RType_Mbp1
}

func (*Mbp1Msg) RSize() uint16 {
	return Mbp1Msg_Size
}

func (r *Mbp1Msg) Fill_Raw(b []byte) error {
	if len(b) < Mbp1Msg_Size {
		return unexpectedBytesError(len(b), Mbp1Msg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.Price = int64(binary.LittleEndian.Uint64(body[0:8]))
	r.Size = binary.LittleEndian.Uint32(body[8:12])
	r.Action = body[12]
	r.Side = body[13]
	r.Flags = body[14]
	r.Depth = body[15]
	r.TsRecv = binary.LittleEndian.Uint64(body[16:24])
	r.TsInDelta = int32(binary.LittleEndian.Uint32(body[24:28]))
	r.Sequence = binary.LittleEndian.Uint32(body[28:32])
	r.Level.Fill_Raw(body[32 : 32+BidAskPair_Size])
	return nil
}

func (r *Mbp1Msg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.Price = fastjson_GetInt64FromString(val, "price")
	r.Size = uint32(val.GetUint("size"))
	r.Action = byte(val.GetUint("action"))
	r.Side = byte(val.GetUint("side"))
	r.Flags = uint8(val.GetUint("flags"))
	r.Depth = uint8(val.GetUint("depth"))
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.TsInDelta = int32(val.GetUint("ts_in_delta"))
	r.Sequence = uint32(val.GetUint("sequence"))
	levelsArr := val.GetArray("levels")
	if len(levelsArr) == 0 {
		return errors.New("levels array is empty")
	}
	r.Level.Fill_Json(levelsArr[0])
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// Consolidated market by price implementation with a known book depth of 1. The record of the
// [`Cmbp1`](crate::Schema::Cmbp1) schema.
type Cmbp1Msg struct {
	Header    RHeader                `json:"hd" csv:"hd"`                         // The record header.
	Price     int64                  `json:"price" csv:"price"`                   // The order price expressed as a signed integer where every 1 unit corresponds to 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	Size      uint32                 `json:"size" csv:"size"`                     // The order quantity.
	Action    byte                   `json:"action" csv:"action"`                 // The event action. Can be **A**dd, **C**ancel, **M**odify, clea**R**, or **T**rade.
	Side      byte                   `json:"side" csv:"side"`                     // The side that initiates the event. Can be **A**sk for a sell order (or sell  aggressor in a trade), **B**id for a buy order (or buy aggressor in a trade), or  **N**one where no side is specified by the original source.
	Flags     uint8                  `json:"flags" csv:"flags"`                   // A bit field indicating event end, message characteristics, and data quality. See [`enums::flags`](crate::enums::flags) for possible values.
	Reserved  byte                   `json:"_reserved,omitempty" csv:"_reserved"` // Reserved for future usage.
	TsRecv    uint64                 `json:"ts_recv" csv:"ts_recv"`               // The capture-server-received timestamp expressed as number of nanoseconds since the UNIX epoch.
	TsInDelta int32                  `json:"ts_in_delta" csv:"ts_in_delta"`       // The delta of `ts_recv - ts_exchange_send`, max 2 seconds.
	Sequence  uint32                 `json:"sequence" csv:"sequence"`             // The message sequence number assigned at the venue.
	Level     ConsolidatedBidAskPair `json:"levels" csv:"levels"`                 // The top of the order book.
}

const Cmbp1Msg_Size = RHeader_Size + 32 + ConsolidatedBidAskPair_Size

func (*Cmbp1Msg) RType() RType {
	return RType_Cmbp1 // TODO
}

func (*Cmbp1Msg) RSize() uint16 {
	return Cmbp1Msg_Size
}

func (r *Cmbp1Msg) Fill_Raw(b []byte) error {
	if len(b) < Cmbp1Msg_Size {
		return unexpectedBytesError(len(b), Cmbp1Msg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.Price = int64(binary.LittleEndian.Uint64(body[0:8]))
	r.Size = binary.LittleEndian.Uint32(body[8:12])
	r.Action = body[12]
	r.Side = body[13]
	r.Flags = body[14]
	r.TsRecv = binary.LittleEndian.Uint64(body[16:24])
	r.TsInDelta = int32(binary.LittleEndian.Uint32(body[24:28]))
	r.Sequence = binary.LittleEndian.Uint32(body[28:32])
	r.Level.Fill_Raw(body[32 : 32+ConsolidatedBidAskPair_Size])
	return nil
}

func (r *Cmbp1Msg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.Price = fastjson_GetInt64FromString(val, "price")
	r.Size = uint32(val.GetUint("size"))
	r.Action = byte(val.GetUint("action"))
	r.Side = byte(val.GetUint("side"))
	r.Flags = uint8(val.GetUint("flags"))
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.TsInDelta = int32(val.GetUint("ts_in_delta"))
	r.Sequence = uint32(val.GetUint("sequence"))
	levelsArr := val.GetArray("levels")
	if len(levelsArr) == 0 {
		return errors.New("levels array is empty")
	}
	r.Level.Fill_Json(levelsArr[0])
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// Databento Normalized market-by-price implementation with a known book depth of 10. The record of the [`Mbp10`](crate::enums::Schema::Mbp10) schema.
type Mbp10Msg struct {
	Header    RHeader        `json:"hd" csv:"hd"`                   // The record header.
	Price     int64          `json:"price" csv:"price"`             // The order price expressed as a signed integer where every 1 unit corresponds to 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	Size      uint32         `json:"size" csv:"size"`               // The order quantity.
	Action    byte           `json:"action" csv:"action"`           // The event action. Can be **A**dd, **C**ancel, **M**odify, clea**R**, or **T**rade.
	Side      byte           `json:"side" csv:"side"`               // The side that initiates the event. Can be **A**sk for a sell order (or sell aggressor in a trade), **B**id for a buy order (or buy aggressor in a trade), or **N**one where no side is specified by the original source.
	Flags     uint8          `json:"flags" csv:"flags"`             // A bit field indicating event end, message characteristics, and data quality. See [`enums::flags`](crate::enums::flags) for possible values.
	Depth     uint8          `json:"depth" csv:"depth"`             // The depth of actual book change.
	TsRecv    uint64         `json:"ts_recv" csv:"ts_recv"`         // The capture-server-received timestamp expressed as number of nanoseconds since the UNIX epoch.
	TsInDelta int32          `json:"ts_in_delta" csv:"ts_in_delta"` // The delta of `ts_recv - ts_exchange_send`, max 2 seconds.
	Sequence  uint32         `json:"sequence" csv:"sequence"`       // The message sequence number assigned at the venue.
	Levels    [10]BidAskPair `json:"levels" csv:"levels"`           // The top 10 levels of the order book.
}

const Mbp10Msg_Size = RHeader_Size + 32 + 10*BidAskPair_Size

func (*Mbp10Msg) RType() RType {
	return RType_Mbp10
}

func (*Mbp10Msg) RSize() uint16 {
	return Mbp10Msg_Size
}

func (r *Mbp10Msg) Fill_Raw(b []byte) error {
	if len(b) < Mbp10Msg_Size {
		return unexpectedBytesError(len(b), Mbp10Msg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.Price = int64(binary.LittleEndian.Uint64(body[0:8]))
	r.Size = binary.LittleEndian.Uint32(body[8:12])
	r.Action = body[12]
	r.Side = body[13]
	r.Flags = body[14]
	r.Depth = body[15]
	r.TsRecv = binary.LittleEndian.Uint64(body[16:24])
	r.TsInDelta = int32(binary.LittleEndian.Uint32(body[24:28]))
	r.Sequence = binary.LittleEndian.Uint32(body[28:32])
	for i := 0; i < 10; i++ {
		offset := 32 + i*BidAskPair_Size
		r.Levels[i].Fill_Raw(body[offset : offset+BidAskPair_Size])
	}
	return nil
}

func (r *Mbp10Msg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.Price = fastjson_GetInt64FromString(val, "price")
	r.Size = uint32(val.GetUint("size"))
	r.Action = byte(val.GetUint("action"))
	r.Side = byte(val.GetUint("side"))
	r.Flags = uint8(val.GetUint("flags"))
	r.Depth = uint8(val.GetUint("depth"))
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.TsInDelta = int32(val.GetUint("ts_in_delta"))
	r.Sequence = uint32(val.GetUint("sequence"))
	levelsArr := val.GetArray("levels")
	if len(levelsArr) < 10 {
		return errors.New("levels array is less than 10")
	}
	for i := 0; i < 10; i++ {
		r.Levels[i].Fill_Json(levelsArr[i])
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// Databento Normalized Ohlcv Message (OHLC candlestick, bar)
// {"hd":{"ts_event":"1702987922000000000","rtype":32,"publisher_id":40,"instrument_id":15144},"open":"472600000000","high":"472600000000","low":"472600000000","close":"472600000000","volume":"300"}
type OhlcvMsg struct {
	Header RHeader `json:"hd" csv:"hd"`         // The record header.
	Open   int64   `json:"open" csv:"open"`     // The open price for the bar.
	High   int64   `json:"high" csv:"high"`     // The high price for the bar.
	Low    int64   `json:"low" csv:"low"`       // The low price for the bar.
	Close  int64   `json:"close" csv:"close"`   // The close price for the bar.
	Volume uint64  `json:"volume" csv:"volume"` // The total volume traded during the aggregation period.
}

const OhlcvMsg_Size = RHeader_Size + 40

func (*OhlcvMsg) RType() RType {
	// RType was nil, return a generic Candle RTtype
	return RType_OhlcvEod
}

func (*OhlcvMsg) RSize() uint16 {
	return OhlcvMsg_Size
}

func (r *OhlcvMsg) Fill_Raw(b []byte) error {
	if len(b) < OhlcvMsg_Size {
		return unexpectedBytesError(len(b), OhlcvMsg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
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

func (r *OhlcvMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.Open = fastjson_GetInt64FromString(val, "open")
	r.High = fastjson_GetInt64FromString(val, "high")
	r.Low = fastjson_GetInt64FromString(val, "low")
	r.Close = fastjson_GetInt64FromString(val, "close")
	r.Volume = fastjson_GetUint64FromString(val, "volume")
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// Databento Normalized Imbalance Message
// {"ts_recv":"1711027500000942123","hd":{"ts_event":"1711027500000776211","rtype":20,"publisher_id":2,"instrument_id":17598},"ref_price":"0","auction_time":"0","cont_book_clr_price":"0","auct_interest_clr_price":"0","ssr_filling_price":"0","ind_match_price":"0","upper_collar":"0","lower_collar":"0","paired_qty":0,"total_imbalance_qty":0,"market_imbalance_qty":0,"unpaired_qty":0,"auction_type":"O","side":"N","auction_status":0,"freeze_status":0,"num_extensions":0,"unpaired_side":"N","significant_imbalance":"~"}
type ImbalanceMsg struct {
	Header               RHeader `json:"hd" csv:"hd"`                                          // The record header.
	TsRecv               uint64  `json:"ts_recv" csv:"ts_recv"`                                // The capture-server-received timestamp expressed as the number of nanoseconds since the UNIX epoch.
	RefPrice             int64   `json:"ref_price" csv:"ref_price"`                            // The price at which the imbalance shares are calculated, where every 1 unit corresponds to 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	AuctionTime          uint64  `json:"auction_time" csv:"auction_time"`                      // Reserved for future use.
	ContBookClrPrice     int64   `json:"cont_book_clr_price" csv:"contBook_clr_price"`         // The hypothetical auction-clearing price for both cross and continuous orders.
	AuctInterestClrPrice int64   `json:"auct_interest_clr_price" csv:"auctInterest_clr_price"` // The hypothetical auction-clearing price for cross orders only.
	SsrFillingPrice      int64   `json:"ssr_filling_price" csv:"ssr_filling_price"`            // Reserved for future use.
	IndMatchPrice        int64   `json:"ind_match_price" csv:"ind_match_price"`                // Reserved for future use.
	UpperCollar          int64   `json:"upper_collar" csv:"upper_collar"`                      // Reserved for future use.
	LowerCollar          int64   `json:"lower_collar" csv:"lower_collar"`                      // Reserved for future use.
	PairedQty            uint32  `json:"paired_qty" csv:"paired_qty"`                          // The quantity of shares that are eligible to be matched at `ref_price`.
	TotalImbalanceQty    uint32  `json:"total_imbalance_qty" csv:"total_imbalance_qty"`        // The quantity of shares that are not paired at `ref_price`.
	MarketImbalanceQty   uint32  `json:"market_imbalance_qty" csv:"market_ombalance_qty"`      // Reserved for future use.
	UnpairedQty          int32   `json:"unpaired_qty" csv:"unpaired_qty"`                      // Reserved for future use.
	AuctionType          uint8   `json:"auction_type" csv:"auction_type"`                      // Venue-specific character code indicating the auction type.
	Side                 uint8   `json:"side" csv:"side"`                                      // The market side of the `total_imbalance_qty`. Can be **A**sk, **B**id, or **N**one.
	AuctionStatus        uint8   `json:"auction_status" csv:"auction_status"`                  // Reserved for future use.
	FreezeStatus         uint8   `json:"freeze_status" csv:"freeze_status"`                    // Reserved for future use.
	NumExtensions        uint8   `json:"num_extensions" csv:"num_extensions"`                  // Reserved for future use.
	UnpairedSide         uint8   `json:"unpaired_side" csv:"unpaired_side"`                    // Reserved for future use.
	SignificantImbalance uint8   `json:"significant_imbalance" csv:"significant_imbalance"`    // Venue-specific character code. For Nasdaq, contains the raw Price Variation Indicator.
	Reserved             uint8   `json:"reserved" csv:"reserved"`                              // Filler for alignment.
}

const ImbalanceMsg_Size = RHeader_Size + 96

func (*ImbalanceMsg) RType() RType {
	return RType_Imbalance
}

func (*ImbalanceMsg) RSize() uint16 {
	return ImbalanceMsg_Size
}

func (r *ImbalanceMsg) Fill_Raw(b []byte) error {
	if len(b) < ImbalanceMsg_Size {
		return unexpectedBytesError(len(b), ImbalanceMsg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.TsRecv = binary.LittleEndian.Uint64(body[0:8])
	r.RefPrice = int64(binary.LittleEndian.Uint64(body[8:16]))
	r.AuctionTime = binary.LittleEndian.Uint64(body[16:24])
	r.ContBookClrPrice = int64(binary.LittleEndian.Uint64(body[24:32]))
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

func (r *ImbalanceMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
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

type ErrorMsg struct {
	Header RHeader                `json:"hd" csv:"hd"`           // The common header.
	Error  [ErrorMsg_ErrSize]byte `json:"err" csv:"err"`         // The error message.
	Code   ErrorCode              `json:"code" csv:"code"`       // The error code.
	IsLast uint8                  `json:"is_last" csv:"is_last"` // Sometimes multiple errors are sent together. This field will be non-zero for the last error.
}

const ErrorMsg_ErrSize = 302
const ErrorMsg_Size = RHeader_Size + ErrorMsg_ErrSize + 2

func (*ErrorMsg) RType() RType {
	return RType_Error
}

func (*ErrorMsg) RSize() uint16 {
	return ErrorMsg_Size
}

func (r *ErrorMsg) Fill_Raw(b []byte) error {
	if len(b) < ErrorMsg_Size {
		return unexpectedBytesError(len(b), ErrorMsg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	copy(r.Error[:], body[:ErrorMsg_ErrSize])
	r.Code = ErrorCode(body[ErrorMsg_ErrSize])
	r.IsLast = body[ErrorMsg_ErrSize+1]
	return nil
}

func (r *ErrorMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	copy(r.Error[:], val.GetStringBytes("err"))
	r.Code = ErrorCode(uint8(val.GetUint("code")))
	r.IsLast = uint8(val.GetUint("is_last"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

type SystemMsg struct {
	Header  RHeader                 `json:"hd" csv:"hd"`     // The common header.
	Message [SystemMsg_MsgSize]byte `json:"msg" csv:"msg"`   // The message from the Databento Live Subscription Gateway (LSG).
	Code    SystemCode              `json:"code" csv:"code"` // The type of system message.
}

const SystemMsg_MsgSize = 303
const SystemMsg_Size = RHeader_Size + SystemMsg_MsgSize + 1

func (*SystemMsg) RType() RType {
	return RType_System
}

func (*SystemMsg) RSize() uint16 {
	return SystemMsg_Size
}

func (r *SystemMsg) Fill_Raw(b []byte) error {
	if len(b) < SystemMsg_Size {
		return unexpectedBytesError(len(b), SystemMsg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	copy(r.Message[:], body[:SystemMsg_MsgSize])
	r.Code = SystemCode(body[SystemMsg_MsgSize])
	return nil
}

func (r *SystemMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	copy(r.Message[:], val.GetStringBytes("msg"))
	r.Code = SystemCode(uint8(val.GetUint("code")))
	return nil
}

// IsHeartbeat checks if the system message is a heartbeat.
// For fullest compatibility, it falls back to a string check
func (r *SystemMsg) IsHeartbeat() bool {
	if r.Code == SystemCode_Heartbeat {
		return true
	}
	// Fallback to string check for backwards compatibility
	if bytes.Equal(r.Message[:], []byte(SystemCodeString_Heartbeat)) {
		return true
	}
	return false
}

///////////////////////////////////////////////////////////////////////////////

// Databento normalized Trading Status Update message.
type StatusMsg struct {
	Header                RHeader  `json:"hd" csv:"hd"`                                             // The record header.
	TsRecv                uint64   `json:"ts_recv" csv:"ts_recv"`                                   // The capture-server-received timestamp expressed as number of nanoseconds since the UNIX epoch.
	Action                uint16   `json:"action" csv:"action"`                                     // The type of status change.
	Reason                uint16   `json:"reason" csv:"reason"`                                     // Additional details about the cause of the status change.
	TradingEvent          uint16   `json:"trading_event" csv:"trading_event"`                       // Further information about the status change and its effect on trading.
	IsTrading             uint8    `json:"is_trading" csv:"is_trading"`                             // The state of trading in the instrument.
	IsQuoting             uint8    `json:"is_quoting" csv:"is_quoting"`                             // The state of quoting in the instrument.
	IsShortSellRestricted uint8    `json:"is_short_sell_restricted" csv:"is_short_sell_restricted"` // The state of short sell restrictions for the instrument.
	Reserved              [7]uint8 // Filler for alignment.
}

const StatusMsg_Size = RHeader_Size + 24 // TODO check size, add test

func (*StatusMsg) RType() RType {
	return RType_Status
}

func (*StatusMsg) RSize() uint16 {
	return StatusMsg_Size
}

func (r *StatusMsg) Fill_Raw(b []byte) error {
	if len(b) < StatusMsg_Size {
		return unexpectedBytesError(len(b), StatusMsg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.TsRecv = binary.LittleEndian.Uint64(body[0:8])
	r.Action = binary.LittleEndian.Uint16(body[8:10])
	r.Reason = binary.LittleEndian.Uint16(body[10:12])
	r.TradingEvent = binary.LittleEndian.Uint16(body[12:14])
	r.IsTrading = body[14]
	r.IsQuoting = body[15]
	r.IsShortSellRestricted = body[16]
	return nil
}

func (r *StatusMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.Action = uint16(val.GetUint("action"))
	r.Reason = uint16(val.GetUint("reason"))
	r.TradingEvent = uint16(val.GetUint("trading_event"))
	r.IsTrading = uint8(val.GetUint("is_trading"))
	r.IsQuoting = uint8(val.GetUint("is_quoting"))
	r.IsShortSellRestricted = uint8(val.GetUint("is_short_sell_restricted"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// BboMsg is a Best Bid and Offer record subsampled on a 1-second or 1-minute interval.
// It provides the last best bid, best offer, and sale at the specified interval.
type BboMsg struct {
	Header    RHeader    `json:"hd" csv:"hd"`                 // The common header.
	Price     int64      `json:"price" csv:"price"`           // The last trade price where every 1 unit corresponds to 1e-9, i.e. 1/1,000,000,000 or 0.000000001. Will be UNDEF_PRICE if there was no last trade in the session.
	Size      uint32     `json:"size" csv:"size"`             // The last trade quantity. Will be 0 if there was no last trade in the session.
	Reserved1 byte       `json:"_reserved1" csv:"_reserved1"` // Reserved for future use.
	Side      byte       `json:"side" csv:"side"`             // The side that initiated the last trade. Can be Ask for a sell aggressor in a trade, Bid for a buy aggressor in a trade, or None where no side is specified.
	Flags     uint8      `json:"flags" csv:"flags"`           // A bit field indicating event end, message characteristics, and data quality.
	Reserved2 byte       `json:"_reserved2" csv:"_reserved2"` // Reserved for future use.
	TsRecv    uint64     `json:"ts_recv" csv:"ts_recv"`       // The end timestamp of the interval, clamped to the second/minute boundary, expressed as the number of nanoseconds since the UNIX epoch.
	Reserved3 [4]byte    `json:"_reserved3" csv:"_reserved3"` // Reserved for future use.
	Sequence  uint32     `json:"sequence" csv:"sequence"`     // The message sequence number assigned at the venue of the last update.
	Level     BidAskPair `json:"levels" csv:"levels"`         // The bid and ask prices and sizes at the top level.
}

const BboMsg_Size = RHeader_Size + 64

func (*BboMsg) RType() RType {
	// Return a generic BBO RType, similar to how OhlcvMsg returns RType_OhlcvEod
	return RType_Bbo1S
}

func (*BboMsg) RSize() uint16 {
	return BboMsg_Size
}

func (r *BboMsg) Fill_Raw(b []byte) error {
	if len(b) < BboMsg_Size {
		return unexpectedBytesError(len(b), BboMsg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:]

	r.Price = int64(binary.LittleEndian.Uint64(body[0:8]))
	r.Size = binary.LittleEndian.Uint32(body[8:12])
	// Reserved1 12
	r.Side = body[13]
	r.Flags = body[14]
	// Reserved2 15
	r.TsRecv = binary.LittleEndian.Uint64(body[16:24])
	// Reserved3 24:28
	r.Sequence = binary.LittleEndian.Uint32(body[28:32])
	r.Level.Fill_Raw(body[32 : 32+BidAskPair_Size])
	return nil
}

func (r *BboMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.Price = fastjson_GetInt64FromString(val, "price")
	r.Size = uint32(val.GetUint("size"))
	r.Side = byte(val.Get("side").String()[1]) // Get the first character from the JSON string
	r.Flags = uint8(val.GetUint("flags"))
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.Sequence = uint32(val.GetUint("sequence"))

	levels := val.Get("levels")
	if levels != nil {
		r.Level.BidPx = fastjson_GetInt64FromString(levels, "bid_px")
		r.Level.AskPx = fastjson_GetInt64FromString(levels, "ask_px")
		r.Level.BidSz = uint32(levels.GetUint("bid_sz"))
		r.Level.AskSz = uint32(levels.GetUint("ask_sz"))
		r.Level.BidCt = uint32(levels.GetUint("bid_ct"))
		r.Level.AskCt = uint32(levels.GetUint("ask_ct"))
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// Type Aliases for Backward Compatibility

// SymbolMappingMsg is a Databento Symbol Mapping message.
// This is not a strict byte-layout because StypeInSymbol and StypeOutSymbol have dynamic lengths
// that depend on metadata's SymbolCstrLen.
//
// SymbolMappingMsg is an alias for the current version (V2)
type SymbolMappingMsg = SymbolMappingMsgV2

// StatMsg is a statistics message. A catchall for various data disseminated by publishers.
// The [`stat_type`](Self::stat_type) indicates the statistic contained in the message.
//
// StatMsg is an alias for the current version (V3).
// The scanner upgrades V1/V2 records to V3 layout (low-velocity, no perf concern).
type StatMsg = StatMsgV3

const StatMsg_Size = StatMsgV3_Size

// InstrumentDefMsg is a definition of an instrument.
//
// InstrumentDefMsg is an alias for the current version (V3).
// The scanner upgrades V2 records to V3 layout (very low-velocity, no perf concern).
// NOTE: InstrumentDefMsgV1 (22-byte symbols) was never implemented; only V2+ is supported.
type InstrumentDefMsg = InstrumentDefMsgV3

const InstrumentDefMsg_Size = InstrumentDefMsgV3_Size

// /////////////////////////////////////////////////////////////////////////////
// SymbolMappingMsgFillRaw fills a SymbolMappingMsg from raw bytes based on DBN version.
// It dispatches to the appropriate version-specific implementation.
func SymbolMappingMsgFillRaw(r *SymbolMappingMsgV2, b []byte, cstrLength uint16) error {
	if cstrLength == MetadataV1_SymbolCstrLen {
		// Decode as V1, then convert
		var v1 SymbolMappingMsgV1
		if err := v1.Fill_Raw(b); err != nil {
			return err
		}
		// Copy fields from V1 to V2
		r.Header = v1.Header
		r.StypeIn = v1.StypeIn
		r.StypeInSymbol = v1.StypeInSymbol
		r.StypeOut = v1.StypeOut
		r.StypeOutSymbol = v1.StypeOutSymbol
		r.StartTs = v1.StartTs
		r.EndTs = v1.EndTs
		return nil
	} else if cstrLength == MetadataV2_SymbolCstrLen {
		// Decode as V2
		return r.Fill_Raw(b)
	} else {
		return unexpectedCStrLenError(cstrLength)
	}

}
