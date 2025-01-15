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
	ChannelID uint8   `json:"channel_id" csv:"channel_id"`   // A channel ID within the venue.
	Action    byte    `json:"action" csv:"action"`           // The event action. Can be **A**dd, **C**ancel, **M**odify, clea**R**,  **T**rade, or **F**ill.
	Side      byte    `json:"side" csv:"side"`               // The side that initiates the event. Can be **A**sk for a sell order (or sell aggressor in a trade), **B**id for a buy order (or buy aggressor in a trade), or **N**one where no side is specified by the original source.
	TsRecv    uint64  `json:"ts_recv" csv:"ts_recv"`         // The capture-server-received timestamp expressed as number of nanoseconds since the UNIX epoch.
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

type CbboMsg struct {
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

const CbboMsg_Size = RHeader_Size + 32 + ConsolidatedBidAskPair_Size

func (*CbboMsg) RType() RType {
	return RType_Cbbo // TODO
}

func (*CbboMsg) RSize() uint16 {
	return CbboMsg_Size
}

func (r *CbboMsg) Fill_Raw(b []byte) error {
	if len(b) < CbboMsg_Size {
		return unexpectedBytesError(len(b), CbboMsg_Size)
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

func (r *CbboMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
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

///////////////////////////////////////////////////////////////////////////////

// Databento Symbol Mapping Message
// This is not a strict byte-layout because StypeInSymbol and StypeOutSymbol have dynamic lengths
// that depend on metadata's SymbolCstrLen.
type SymbolMappingMsg struct {
	Header         RHeader `json:"hd" csv:"hd"`                             // The common header.
	StypeIn        SType   `json:"stype_in" csv:"stype_in"`                 // The input symbology type of `stype_in_symbol`.
	StypeInSymbol  string  `json:"stype_in_symbol" csv:"stype_in_symbol"`   // The input symbol.
	StypeOut       SType   `json:"stype_out" csv:"stype_out"`               // The output symbology type of `stype_out_symbol`.
	StypeOutSymbol string  `json:"stype_out_symbol" csv:"stype_out_symbol"` // The output symbol.
	StartTs        uint64  `json:"start_ts" csv:"start_ts"`                 // The start of the mapping interval expressed as the number of nanoseconds since the UNIX epoch.
	EndTs          uint64  `json:"end_ts" csv:"end_ts"`                     // The end of the mapping interval expressed as the number of nanoseconds since the UNIX epoch.
}

// Minimum size of SymbolMappingMsg, the size with 0-length c-strings
// We add 2*SymbolCstrLength to it to get actual size.
const SymbolMappingMsg_MinSize = RHeader_Size + 16

func (*SymbolMappingMsg) RType() RType {
	return RType_SymbolMapping
}

func (*SymbolMappingMsg) RSize(cstrLength uint16) uint16 {
	if cstrLength == MetadataV1_SymbolCstrLen {
		// V1 message doesn't have StypeIn and StypeOut fields
		return SymbolMappingMsg_MinSize + (2 * cstrLength)
	} else {
		return SymbolMappingMsg_MinSize + (2 * cstrLength) + 2
	}
}

func (r *SymbolMappingMsg) Fill_Raw(b []byte, cstrLength uint16) error {
	rsize := r.RSize(cstrLength)
	if len(b) < int(rsize) {
		return unexpectedBytesError(len(b), int(rsize))
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	pos := uint16(0)
	if cstrLength == MetadataV1_SymbolCstrLen {
		r.StypeIn = SType_RawSymbol // DBN1 doesn't have this field
	} else {
		r.StypeIn = SType(body[pos])
		pos += 1
	}
	r.StypeInSymbol = TrimNullBytes(body[pos : pos+cstrLength])
	pos += cstrLength
	if cstrLength == MetadataV1_SymbolCstrLen {
		r.StypeIn = SType_RawSymbol // DBN1 doesn't have this field
	} else {
		r.StypeIn = SType(body[pos])
		pos += 1
	}
	r.StypeOutSymbol = TrimNullBytes(body[pos : pos+cstrLength])
	pos += +cstrLength
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

///////////////////////////////////////////////////////////////////////////////

type ErrorMsg struct {
	Header RHeader                `json:"hd" csv:"hd"`           // The common header.
	Error  [ErrorMsg_ErrSize]byte `json:"err" csv:"err"`         // The error message.
	Code   uint8                  `json:"code" csv:"code"`       // The error code. Currently unused.
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
	r.Code = body[ErrorMsg_ErrSize]
	r.IsLast = body[ErrorMsg_ErrSize+1]
	return nil
}

func (r *ErrorMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	copy(r.Error[:], val.GetStringBytes("err"))
	r.Code = uint8(val.GetUint("code"))
	r.IsLast = uint8(val.GetUint("is_last"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

type SystemMsg struct {
	Header  RHeader                 `json:"hd" csv:"hd"`     // The common header.
	Message [SystemMsg_MsgSize]byte `json:"msg" csv:"msg"`   // The message from the Databento Live Subscription Gateway (LSG).
	Code    uint8                   `json:"code" csv:"code"` // The type of system message. Currently unused.
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
	r.Code = body[SystemMsg_MsgSize]
	return nil
}

func (r *SystemMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	copy(r.Message[:], val.GetStringBytes("msg"))
	r.Code = uint8(val.GetUint("code"))
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// StatMsg is a statistics message. A catchall for various data disseminated by publishers.
// The [`stat_type`](Self::stat_type) indicates the statistic contained in the message.
type StatMsg struct {
	Header       RHeader  `json:"hd" csv:"hd"`                       // The common header.
	TsRecv       uint64   `json:"ts_recv" csv:"ts_recv"`             // The capture-server-received timestamp expressed as the number of nanoseconds since the UNIX epoch.
	TsRef        uint64   `json:"ts_ref" csv:"ts_ref"`               // The reference timestamp of the statistic value expressed as the number of nanoseconds since the UNIX epoch. Will be [`crate::UNDEF_TIMESTAMP`] when unused.
	Price        int64    `json:"price" csv:"price"`                 // The value for price statistics expressed as a signed integer where every 1 unit corresponds to 1e-9, i.e. 1/1,000,000,000 or 0.000000001. Will be [`crate::UNDEF_PRICE`] when unused.
	Quantity     int32    `json:"quantity" csv:"quantity"`           // The value for non-price statistics. Will be [`crate::UNDEF_STAT_QUANTITY`] when unused.
	Sequence     uint32   `json:"sequence" csv:"sequence"`           // The message sequence number assigned at the venue.
	TsInDelta    int32    `json:"ts_in_delta" csv:"ts_in_delta"`     // The delta of `ts_recv - ts_exchange_send`, max 2 seconds.
	StatType     uint16   `json:"stat_type" csv:"stat_type"`         // The type of statistic value contained in the message. Refer to the [`StatType`](crate::enums::StatType) for variants.
	ChannelID    uint16   `json:"channel_id" csv:"channel_id"`       // A channel ID within the venue.
	UpdateAction uint8    `json:"update_action" csv:"update_action"` // Indicates if the statistic is newly added (1) or deleted (2). (Deleted is only used with some stat types)
	StatFlags    uint8    `json:"stat_flags" csv:"stat_flags"`       // Additional flags associate with certain stat types.
	Reserved     [6]uint8 // Filler for alignment
}

const StatMsg_Size = RHeader_Size + 48

func (*StatMsg) RType() RType {
	return RType_Statistics
}

func (*StatMsg) RSize() uint16 {
	return StatMsg_Size
}

func (r *StatMsg) Fill_Raw(b []byte) error {
	if len(b) < StatMsg_Size {
		return unexpectedBytesError(len(b), StatMsg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.TsRecv = binary.LittleEndian.Uint64(body[0:8])
	r.TsRef = binary.LittleEndian.Uint64(body[8:16])
	r.Price = int64(binary.LittleEndian.Uint64(body[16:24]))
	r.Quantity = int32(binary.LittleEndian.Uint32(body[24:28]))
	r.Sequence = binary.LittleEndian.Uint32(body[28:32])
	r.TsInDelta = int32(binary.LittleEndian.Uint32(body[32:36]))
	r.StatType = binary.LittleEndian.Uint16(body[36:38])
	r.ChannelID = binary.LittleEndian.Uint16(body[39:40])
	r.UpdateAction = body[40]
	r.StatFlags = body[41]
	return nil
}

func (r *StatMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.TsRef = fastjson_GetUint64FromString(val, "ts_ref")
	r.Price = fastjson_GetInt64FromString(val, "price")
	r.Quantity = int32(val.GetUint("quantity"))
	r.Sequence = uint32(val.GetUint("sequence"))
	r.TsInDelta = int32(val.GetUint("ts_in_delta"))
	r.StatType = uint16(val.GetUint("stat_type"))
	r.ChannelID = uint16(val.GetUint("channel_id"))
	r.UpdateAction = uint8(val.GetUint("update_action"))
	r.StatFlags = uint8(val.GetUint("stat_flags"))
	return nil
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
	r.Action = binary.LittleEndian.Uint16(body[8:9])
	r.Reason = binary.LittleEndian.Uint16(body[9:10])
	r.TradingEvent = binary.LittleEndian.Uint16(body[10:11])
	r.IsTrading = body[11]
	r.IsQuoting = body[12]
	r.IsShortSellRestricted = body[13]
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

// InstrumentDefMsg is a statistics message. A catchall for various data disseminated by publishers.
// The [`stat_type`](Self::stat_type) indicates the statistic contained in the message.
// This is not a strict byte-layout because RawSymbol has dynamic length that depends on metadata's SymbolCstrLen.
type InstrumentDefMsg struct {
	Header                  RHeader                        `json:"hd" csv:"hd"`                                                 // The common header.
	TsRecv                  uint64                         `json:"ts_recv" csv:"ts_recv"`                                       // The capture-server-received timestamp expressed as the number of nanoseconds since the UNIX epoch.
	MinPriceIncrement       int64                          `json:"min_price_increment" csv:"min_price_increment"`               // Fixed price The minimum constant tick for the instrument in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	DisplayFactor           int64                          `json:"display_factor" csv:"display_factor"`                         // The multiplier to convert the venue’s display price to the conventional price, in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	Expiration              uint64                         `json:"expiration" csv:"expiration"`                                 // The last eligible trade time expressed as a number of nanoseconds since the UNIX epoch. Will be [`crate::UNDEF_TIMESTAMP`] when null, such as for equities.
	Activation              uint64                         `json:"activation" csv:"activation"`                                 // The time of instrument activation expressed as a number of nanoseconds since the UNIX epoch. Will be [`crate::UNDEF_TIMESTAMP`] when null, such as for equities.
	HighLimitPrice          int64                          `json:"high_limit_price" csv:"high_limit_price"`                     // The allowable high limit price for the trading day in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	LowLimitPrice           int64                          `json:"low_limit_price" csv:"low_limit_price"`                       // The allowable low limit price for the trading day in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	MaxPriceVariation       int64                          `json:"max_price_variation" csv:"max_price_variation"`               // The differential value for price banding in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	TradingReferencePrice   int64                          `json:"trading_reference_price" csv:"trading_reference_price"`       // The trading session settlement price on `trading_reference_date`.
	UnitOfMeasureQty        int64                          `json:"unit_of_measure_qty" csv:"unit_of_measure_qty"`               // The contract size for each instrument, in combination with `unit_of_measure`, in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	MinPriceIncrementAmount int64                          `json:"min_price_increment_amount" csv:"min_price_increment_amount"` // The value currently under development by the venue. Converted to units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	PriceRatio              int64                          `json:"price_ratio" csv:"price_ratio"`                               // The value used for price calculation in spread and leg pricing in units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	StrikePrice             int64                          `json:"strike_price" csv:"strike_price"`                             // The strike price of the option. Converted to units of 1e-9, i.e. 1/1,000,000,000 or 0.000000001.
	InstAttribValue         int32                          `json:"inst_attrib_value" csv:"inst_attrib_value"`                   // A bitmap of instrument eligibility attributes.
	UnderlyingID            uint32                         `json:"underlying_id" csv:"underlying_id"`                           // The `instrument_id` of the first underlying instrument.
	RawInstrumentID         uint32                         `json:"raw_instrument_id" csv:"raw_instrument_id"`                   // The instrument ID assigned by the publisher. May be the same as `instrument_id`.
	MarketDepthImplied      int32                          `json:"market_depth_implied" csv:"market_depth_implied"`             // The implied book depth on the price level data feed.
	MarketDepth             int32                          `json:"market_depth" csv:"market_depth"`                             // The (outright) book depth on the price level data feed.
	MarketSegmentID         uint32                         `json:"market_segment_id" csv:"market_segment_id"`                   // The market segment of the instrument.
	MaxTradeVol             uint32                         `json:"max_trade_vol" csv:"max_trade_vol"`                           // The maximum trading volume for the instrument.
	MinLotSize              int32                          `json:"min_lot_size" csv:"min_lot_size"`                             // The minimum order entry quantity for the instrument.
	MinLotSizeBlock         int32                          `json:"min_lot_size_block" csv:"min_lot_size_block"`                 // The minimum quantity required for a block trade of the instrument.
	MinLotSizeRoundLot      int32                          `json:"min_lot_size_round_lot" csv:"min_lot_size_round_lot"`         // The minimum quantity required for a round lot of the instrument. Multiples of this quantity are also round lots.
	MinTradeVol             uint32                         `json:"min_trade_vol" csv:"min_trade_vol"`                           // The minimum trading volume for the instrument.
	ContractMultiplier      int32                          `json:"contract_multiplier" csv:"contract_multiplier"`               // The number of deliverables per instrument, i.e. peak days.
	DecayQuantity           int32                          `json:"decay_quantity" csv:"decay_quantity"`                         // The quantity that a contract will decay daily, after `decay_start_date` has been reached.
	OriginalContractSize    int32                          `json:"original_contract_size" csv:"original_contract_size"`         // The fixed contract value assigned to each instrument.
	TradingReferenceDate    uint16                         `json:"trading_reference_date" csv:"trading_reference_date"`         // The trading session date corresponding to the settlement price in  `trading_reference_price`, in number of days since the UNIX epoch.
	ApplID                  int16                          `json:"appl_id" csv:"appl_id"`                                       // The channel ID assigned at the venue.
	MaturityYear            uint16                         `json:"maturity_year" csv:"maturity_year"`                           // The calendar year reflected in the instrument symbol.
	DecayStartDate          uint16                         `json:"decay_start_date" csv:"decay_start_date"`                     // The date at which a contract will begin to decay.
	ChannelID               uint16                         `json:"channel_id" csv:"channel_id"`                                 // The channel ID assigned by Databento as an incrementing integer starting at zero.
	Currency                [4]byte                        `json:"currency" csv:"currency"`                                     // The currency used for price fields.
	SettlCurrency           [4]byte                        `json:"settl_currency" csv:"settl_currency"`                         // The currency used for settlement, if different from `currency`.
	Secsubtype              [6]byte                        `json:"secsubtype" csv:"secsubtype"`                                 // The strategy type of the spread.
	RawSymbol               [MetadataV2_SymbolCstrLen]byte `json:"raw_symbol" csv:"raw_symbol"`                                 // The instrument raw symbol assigned by the publisher.
	Group                   [21]byte                       `json:"group" csv:"group"`                                           // The security group code of the instrument.
	Exchange                [5]byte                        `json:"exchange" csv:"exchange"`                                     // The exchange used to identify the instrument.
	Asset                   [7]byte                        `json:"asset" csv:"asset"`                                           // The underlying asset code (product code) of the instrument.
	Cfi                     [7]byte                        `json:"cfi" csv:"cfi"`                                               // The ISO standard instrument categorization code.
	SecurityType            [7]byte                        `json:"security_type" csv:"security_type"`                           // The type of the instrument, e.g. FUT for future or future spread.
	UnitOfMeasure           [31]byte                       `json:"unit_of_measure" csv:"unit_of_measure"`                       // The unit of measure for the instrument’s original contract size, e.g. USD or LBS.
	Underlying              [21]byte                       `json:"underlying" csv:"underlying"`                                 // The symbol of the first underlying instrument.
	StrikePriceCurrency     [4]byte                        `json:"strike_price_currency" csv:"strike_price_currency"`           // The currency of [`strike_price`](Self::strike_price).
	InstrumentClass         byte                           `json:"instrument_class" csv:"instrument_class"`                     // The classification of the instrument.
	MatchAlgorithm          byte                           `json:"match_algorithm" csv:"match_algorithm"`                       // The matching algorithm used for the instrument, typically **F**IFO.
	MdSecurityTradingStatus uint8                          `json:"md_security_trading_status" csv:"md_security_trading_status"` // The current trading state of the instrument.
	MainFraction            uint8                          `json:"main_fraction" csv:"main_fraction"`                           // The price denominator of the main fraction.
	PriceDisplayFormat      uint8                          `json:"price_display_format" csv:"price_display_format"`             // The number of digits to the right of the tick mark, to display fractional prices.
	SettlPrice_type         uint8                          `json:"settl_price_type" csv:"settl_price_type"`                     // The type indicators for the settlement price, as a bitmap.
	SubFraction             uint8                          `json:"sub_fraction" csv:"sub_fraction"`                             // The price denominator of the sub fraction.
	UnderlyingProduct       uint8                          `json:"underlying_product" csv:"underlying_product"`                 // The product complex of the instrument.
	SecurityUpdateAction    byte                           `json:"security_update_action" csv:"security_update_action"`         // Indicates if the instrument definition has been added, modified, or deleted.
	MaturityMonth           uint8                          `json:"maturity_month" csv:"maturity_month"`                         // The calendar month reflected in the instrument symbol.
	MaturityDay             uint8                          `json:"maturity_day" csv:"maturity_day"`                             // The calendar day reflected in the instrument symbol, or 0.
	MaturityWeek            uint8                          `json:"maturity_week" csv:"maturity_week"`                           // The calendar week reflected in the instrument symbol, or 0.
	UserDefinedInstrument   UserDefinedInstrument          `json:"user_defined_instrument" csv:"user_defined_instrument"`       // Indicates if the instrument is user defined: **Y**es or **N**o.
	ContractMultiplierUnit  int8                           `json:"contract_multiplier_unit" csv:"contract_multiplier_unit"`     // The type of `contract_multiplier`. Either `1` for hours, or `2` for days.
	FlowScheduleType        int8                           `json:"flow_schedule_type" csv:"flow_schedule_type"`                 // The schedule for delivering electricity.
	TickRule                uint8                          `json:"tick_rule" csv:"tick_rule"`                                   // The tick rule of the spread.
	Reserved                [10]byte                       `json:"_reserved" csv:"_reserved"`                                   // Filler for alignment.
}

// Minimum size of InstrumentDefMsg, the size with 0-length c-strings
// We add 1*SymbolCstrLength to it to get actual size.
// const InstrumentDefMsg_MinSize = RHeader_Size + 384 // TODO: count
//
//	func (*InstrumentDefMsg) RSize(cstrLength uint16) uint16 {
//		return InstrumentDefMsg_MinSize + 2*cstrLength
//	}

const InstrumentDefMsg_Size = RHeader_Size + 313 + MetadataV2_SymbolCstrLen

func (*InstrumentDefMsg) RType() RType {
	return RType_InstrumentDef
}

func (*InstrumentDefMsg) RSize() uint16 {
	return InstrumentDefMsg_Size
}

// func (r *InstrumentDefMsg) Fill_Raw(b []byte, cstrLength uint16) error {
func (r *InstrumentDefMsg) Fill_Raw(b []byte) error {
	if len(b) < StatMsg_Size {
		return unexpectedBytesError(len(b), StatMsg_Size)
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:] // slice of just the body
	r.TsRecv = binary.LittleEndian.Uint64(body[0:8])
	r.MinPriceIncrement = int64(binary.LittleEndian.Uint64(body[8:16]))
	r.DisplayFactor = int64(binary.LittleEndian.Uint64(body[16:24]))
	r.Expiration = binary.LittleEndian.Uint64(body[24:32])
	r.Activation = binary.LittleEndian.Uint64(body[32:40])
	r.HighLimitPrice = int64(binary.LittleEndian.Uint64(body[40:48]))
	r.LowLimitPrice = int64(binary.LittleEndian.Uint64(body[48:56]))
	r.MaxPriceVariation = int64(binary.LittleEndian.Uint64(body[56:64]))
	r.TradingReferencePrice = int64(binary.LittleEndian.Uint64(body[64:72]))
	r.UnitOfMeasureQty = int64(binary.LittleEndian.Uint64(body[72:80]))
	r.MinPriceIncrementAmount = int64(binary.LittleEndian.Uint64(body[80:88]))
	r.PriceRatio = int64(binary.LittleEndian.Uint64(body[88:96]))
	r.StrikePrice = int64(binary.LittleEndian.Uint64(body[96:104]))
	r.InstAttribValue = int32(binary.LittleEndian.Uint32(body[104:108]))
	r.UnderlyingID = binary.LittleEndian.Uint32(body[108:112])
	r.RawInstrumentID = binary.LittleEndian.Uint32(body[112:116])
	r.MarketDepthImplied = int32(binary.LittleEndian.Uint32(body[116:120]))
	r.MarketDepth = int32(binary.LittleEndian.Uint32(body[120:124]))
	r.MarketSegmentID = binary.LittleEndian.Uint32(body[124:128])
	r.MaxTradeVol = binary.LittleEndian.Uint32(body[128:132])
	r.MinLotSize = int32(binary.LittleEndian.Uint32(body[132:136]))
	r.MinLotSizeBlock = int32(binary.LittleEndian.Uint32(body[136:140]))
	r.MinLotSizeRoundLot = int32(binary.LittleEndian.Uint32(body[140:144]))
	r.MinTradeVol = binary.LittleEndian.Uint32(body[144:148])
	r.ContractMultiplier = int32(binary.LittleEndian.Uint32(body[148:152]))
	r.DecayQuantity = int32(binary.LittleEndian.Uint32(body[152:156]))
	r.OriginalContractSize = int32(binary.LittleEndian.Uint32(body[156:160]))
	r.TradingReferenceDate = binary.LittleEndian.Uint16(body[160:162])
	r.ApplID = int16(binary.LittleEndian.Uint16(body[162:164]))
	r.MaturityYear = binary.LittleEndian.Uint16(body[164:166])
	r.DecayStartDate = binary.LittleEndian.Uint16(body[166:168])
	r.ChannelID = binary.LittleEndian.Uint16(body[168:170])
	copy(r.Currency[:], body[170:174])            // byte[4]
	copy(r.SettlCurrency[:], body[174:178])       // byte[4]
	copy(r.Secsubtype[:], body[178:184])          // byte[6]
	copy(r.RawSymbol[:], body[184:255])           // byte[MetadataV2_SymbolCstrLen] = 71
	copy(r.Group[:], body[255:276])               // byte[21]
	copy(r.Exchange[:], body[276:281])            // byte[5]
	copy(r.Asset[:], body[281:288])               // byte[7]
	copy(r.Cfi[:], body[288:295])                 // byte[7]
	copy(r.SecurityType[:], body[295:302])        // byte[7]
	copy(r.UnitOfMeasure[:], body[302:333])       // byte[31]
	copy(r.Underlying[:], body[333:354])          // byte[21]
	copy(r.StrikePriceCurrency[:], body[354:358]) // byte[4]
	r.InstrumentClass = body[358]
	r.MatchAlgorithm = body[359]
	r.MdSecurityTradingStatus = body[360]
	r.MainFraction = body[361]
	r.PriceDisplayFormat = body[362]
	r.SettlPrice_type = body[363]
	r.SubFraction = body[364]
	r.UnderlyingProduct = body[365]
	r.SecurityUpdateAction = body[366]
	r.MaturityMonth = body[367]
	r.MaturityDay = body[368]
	r.MaturityWeek = body[369]
	r.UserDefinedInstrument = UserDefinedInstrument(body[370])
	r.ContractMultiplierUnit = int8(body[371])
	r.FlowScheduleType = int8(body[372])
	r.TickRule = body[373]
	return nil
}

func (r *InstrumentDefMsg) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.TsRecv = fastjson_GetUint64FromString(val, "ts_recv")
	r.MinPriceIncrement = fastjson_GetInt64FromString(val, "min_price_increment")
	r.DisplayFactor = fastjson_GetInt64FromString(val, "display_factor")
	r.Expiration = fastjson_GetUint64FromString(val, "expiration")
	r.Activation = fastjson_GetUint64FromString(val, "activation")
	r.HighLimitPrice = fastjson_GetInt64FromString(val, "high_limit_price")
	r.LowLimitPrice = fastjson_GetInt64FromString(val, "low_limit_price")
	r.MaxPriceVariation = fastjson_GetInt64FromString(val, "max_price_variation")
	r.TradingReferencePrice = fastjson_GetInt64FromString(val, "trading_reference_price")
	r.UnitOfMeasureQty = fastjson_GetInt64FromString(val, "unit_of_measure_qty")
	r.MinPriceIncrementAmount = fastjson_GetInt64FromString(val, "min_price_increment_amount")
	r.PriceRatio = fastjson_GetInt64FromString(val, "price_ratio")
	r.StrikePrice = fastjson_GetInt64FromString(val, "strike_price")
	r.InstAttribValue = int32(val.GetUint("inst_attrib_value"))
	r.UnderlyingID = uint32(val.GetUint("underlying_id"))
	r.RawInstrumentID = uint32(val.GetUint("raw_instrument_id"))
	r.MarketDepthImplied = int32(val.GetUint("market_depth_implied"))
	r.MarketDepth = int32(val.GetUint("market_depth"))
	r.MarketSegmentID = uint32(val.GetUint("market_segment_id"))
	r.MaxTradeVol = uint32(val.GetUint("max_trade_vol"))
	r.MinLotSize = int32(val.GetUint("min_lot_size"))
	r.MinLotSizeBlock = int32(val.GetUint("min_lot_size_block"))
	r.MinLotSizeRoundLot = int32(val.GetUint("min_lot_size_round_lot"))
	r.MinTradeVol = uint32(val.GetUint("min_trade_vol"))
	r.ContractMultiplier = int32(val.GetUint("contract_multiplier"))
	r.DecayQuantity = int32(val.GetUint("decay_quantity"))
	r.OriginalContractSize = int32(val.GetUint("original_contract_size"))
	r.TradingReferenceDate = uint16(val.GetUint("trading_reference_date"))
	r.ApplID = int16(val.GetUint("appl_id"))
	r.MaturityYear = uint16(val.GetUint("maturity_year"))
	r.DecayStartDate = uint16(val.GetUint("decay_start_date"))
	r.ChannelID = uint16(val.GetUint("channel_id"))
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
	r.InstrumentClass = byte(val.GetUint("instrument_class"))
	r.MatchAlgorithm = byte(val.GetUint("match_algorithm"))
	r.MdSecurityTradingStatus = uint8(val.GetUint("md_security_trading_status"))
	r.MainFraction = uint8(val.GetUint("main_fraction"))
	r.PriceDisplayFormat = uint8(val.GetUint("price_display_format"))
	r.SettlPrice_type = uint8(val.GetUint("settl_price_type"))
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
	return nil
}

func fillBytesFromStringValue(t []byte, val *fastjson.Value, key string) {
	b := val.GetStringBytes(key)
	for i := 0; i < len(t); i++ {
		if i < len(b) {
			t[i] = b[i]
		} else {
			t[i] = 0
		}
	}
}
