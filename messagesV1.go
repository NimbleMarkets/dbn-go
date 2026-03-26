// Copyright (c) 2024 Neomantra Corp
//
// DBN Version 1 Message Structs
//
// These structs represent the DBN version 1 binary layout.
// V1 uses 22-byte symbol strings and has different field layouts than V2.
//
// Adapted from Databento's DBN:
//   https://github.com/databento/dbn

package dbn

import (
	"encoding/binary"

	"github.com/valyala/fastjson"
)

///////////////////////////////////////////////////////////////////////////////

// SymbolMappingMsgV1 is the DBN version 1 layout.
// V1 does not have StypeIn/StypeOut fields in the binary format.
type SymbolMappingMsgV1 struct {
	Header         RHeader `json:"hd" csv:"hd"`
	StypeIn        SType   `json:"stype_in" csv:"stype_in"`                 // Always SType_RawSymbol in V1
	StypeInSymbol  string  `json:"stype_in_symbol" csv:"stype_in_symbol"`
	StypeOut       SType   `json:"stype_out" csv:"stype_out"`               // Always SType_RawSymbol in V1
	StypeOutSymbol string  `json:"stype_out_symbol" csv:"stype_out_symbol"`
	StartTs        uint64  `json:"start_ts" csv:"start_ts"`
	EndTs          uint64  `json:"end_ts" csv:"end_ts"`
}

func (*SymbolMappingMsgV1) RType() RType {
	return RType_SymbolMapping
}

const SymbolMappingMsgV1_Size = RHeader_Size + 16 + (2 * MetadataV1_SymbolCstrLen)

func (*SymbolMappingMsgV1) RSize() uint16 {
	return SymbolMappingMsgV1_Size
}

func (r *SymbolMappingMsgV1) Fill_Raw(b []byte) error {
	rsize := r.RSize()
	if len(b) < int(rsize) {
		return unexpectedBytesError(len(b), int(rsize))
	}
	err := r.Header.Fill_Raw(b[0:RHeader_Size])
	if err != nil {
		return err
	}
	body := b[RHeader_Size:]
	// V1 has no StypeIn/StypeOut bytes, just the two symbols
	r.StypeIn = SType_RawSymbol
	r.StypeInSymbol = TrimNullBytes(body[0:MetadataV1_SymbolCstrLen])
	r.StypeOut = SType_RawSymbol
	r.StypeOutSymbol = TrimNullBytes(body[MetadataV1_SymbolCstrLen : 2*MetadataV1_SymbolCstrLen])
	pos := 2 * MetadataV1_SymbolCstrLen
	r.StartTs = binary.LittleEndian.Uint64(body[pos : pos+8])
	r.EndTs = binary.LittleEndian.Uint64(body[pos+8 : pos+16])
	return nil
}

func (r *SymbolMappingMsgV1) Fill_Json(val *fastjson.Value, header *RHeader) error {
	r.Header = *header
	r.StypeIn = SType(val.GetUint("stype_in"))
	r.StypeInSymbol = string(val.GetStringBytes("stype_in_symbol"))
	r.StypeOut = SType(val.GetUint("stype_out"))
	r.StypeOutSymbol = string(val.GetStringBytes("stype_out_symbol"))
	r.StartTs = val.GetUint64("start_ts")
	r.EndTs = val.GetUint64("end_ts")
	return nil
}
