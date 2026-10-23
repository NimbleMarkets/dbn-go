// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"strconv"
	"strings"
	"testing"

	dbn "github.com/NimbleMarkets/dbn-go"
)

type wireDecoder interface{ Fill_Raw([]byte) error }
type wireField struct {
	path          string
	offset, width int
}

// Literal sizes and offsets pin the DBN v1-v3 wire contract independently of
// dbn-go's size constants and Go's memory alignment. Offsets include the header.
// Reference: databento/dbn v0.72.0 rust/dbn/src/{record.rs,compat.rs}.
func TestFixedWireLayouts(t *testing.T) {
	for _, tt := range []struct {
		record wireDecoder
		size   int
		fields []wireField
	}{
		{&dbn.RHeader{}, 16, []wireField{{"Length", 0, 1}, {"RType", 1, 1}, {"PublisherID", 2, 2}, {"InstrumentID", 4, 4}, {"TsEvent", 8, 8}}},
		{&dbn.BidAskPair{}, 32, []wireField{{"BidPx", 0, 8}, {"AskPx", 8, 8}, {"BidSz", 16, 4}, {"AskSz", 20, 4}, {"BidCt", 24, 4}, {"AskCt", 28, 4}}},
		{&dbn.ConsolidatedBidAskPair{}, 32, []wireField{{"BidPx", 0, 8}, {"AskPx", 8, 8}, {"BidSz", 16, 4}, {"AskSz", 20, 4}, {"BidPb", 24, 2}, {"AskPb", 28, 2}}},
		{&dbn.Mbp0Msg{}, 48, []wireField{{"Price", 16, 8}, {"Size", 24, 4}, {"Action", 28, 1}, {"TsRecv", 32, 8}, {"TsInDelta", 40, 4}, {"Sequence", 44, 4}}},
		{&dbn.MboMsg{}, 56, []wireField{{"OrderID", 16, 8}, {"Price", 24, 8}, {"ChannelID", 37, 1}, {"TsRecv", 40, 8}, {"Sequence", 52, 4}}},
		{&dbn.Mbp1Msg{}, 80, []wireField{{"Price", 16, 8}, {"Sequence", 44, 4}, {"Level.BidPx", 48, 8}, {"Level.AskCt", 76, 4}}},
		{&dbn.Mbp10Msg{}, 368, []wireField{{"Price", 16, 8}, {"Sequence", 44, 4}, {"Levels.0.BidPx", 48, 8}, {"Levels.9.BidPx", 336, 8}, {"Levels.9.AskCt", 364, 4}}},
		{&dbn.Cmbp1Msg{}, 80, []wireField{{"Price", 16, 8}, {"TsRecv", 32, 8}, {"Level.BidPx", 48, 8}, {"Level.BidPb", 72, 2}, {"Level.AskPb", 76, 2}}},
		{&dbn.BboMsg{}, 80, []wireField{{"Price", 16, 8}, {"Side", 29, 1}, {"TsRecv", 32, 8}, {"Sequence", 44, 4}, {"Level.BidPx", 48, 8}, {"Level.AskCt", 76, 4}}},
		{&dbn.OhlcvMsg{}, 56, []wireField{{"Open", 16, 8}, {"High", 24, 8}, {"Low", 32, 8}, {"Close", 40, 8}, {"Volume", 48, 8}}},
		{&dbn.ImbalanceMsg{}, 112, []wireField{{"TsRecv", 16, 8}, {"AuctionTime", 32, 8}, {"LowerCollar", 80, 8}, {"PairedQty", 88, 4}, {"UnpairedQty", 100, 4}, {"SignificantImbalance", 110, 1}}},
		{&dbn.StatusMsg{}, 40, []wireField{{"TsRecv", 16, 8}, {"Action", 24, 2}, {"Reason", 26, 2}, {"TradingEvent", 28, 2}, {"IsShortSellRestricted", 32, 1}}},
		{&dbn.StatMsgV2{}, 64, []wireField{{"Price", 32, 8}, {"Quantity", 40, 4}, {"Sequence", 44, 4}, {"TsInDelta", 48, 4}, {"StatType", 52, 2}, {"ChannelID", 54, 2}, {"StatFlags", 57, 1}}},
		{&dbn.StatMsgV3{}, 80, []wireField{{"Price", 32, 8}, {"Quantity", 40, 8}, {"Sequence", 48, 4}, {"TsInDelta", 52, 4}, {"StatType", 56, 2}, {"ChannelID", 58, 2}, {"StatFlags", 61, 1}}},
		{&dbn.InstrumentDefMsgV2{}, 400, []wireField{{"TsRecv", 16, 8}, {"RawInstrumentID", 128, 4}, {"ChannelID", 184, 2}, {"RawSymbol", 200, 71}, {"Asset", 297, 7}, {"InstrumentClass", 374, 1}, {"TickRule", 389, 1}}},
		{&dbn.InstrumentDefMsgV3{}, 520, []wireField{{"TsRecv", 16, 8}, {"RawInstrumentID", 112, 8}, {"LegPrice", 120, 8}, {"LegInstrumentID", 188, 4}, {"LegCount", 220, 2}, {"LegIndex", 222, 2}, {"RawSymbol", 238, 71}, {"Asset", 335, 11}, {"LegRawSymbol", 416, 71}, {"LegSide", 502, 1}}},
		{&dbn.ErrorMsgV1{}, 80, []wireField{{"Error", 16, 64}}},
		{&dbn.ErrorMsgV2{}, 320, []wireField{{"Error", 16, 302}, {"Code", 318, 1}, {"IsLast", 319, 1}}},
		{&dbn.SystemMsg{}, 320, []wireField{{"Message", 16, 303}, {"Code", 319, 1}}},
	} {
		typ := reflect.TypeOf(tt.record).Elem()
		t.Run(typ.Name(), func(t *testing.T) {
			if got := binary.Size(tt.record); got != tt.size {
				t.Fatalf("binary.Size = %d, want %d", got, tt.size)
			}
			var encoded bytes.Buffer
			if err := binary.Write(&encoded, binary.LittleEndian, tt.record); err != nil {
				t.Fatal(err)
			}
			if encoded.Len() != tt.size {
				t.Fatalf("encoded %d bytes, want %d", encoded.Len(), tt.size)
			}
			if sized, ok := tt.record.(interface{ RSize() uint16 }); ok {
				if got := int(sized.RSize()); got != tt.size {
					t.Fatalf("RSize = %d, want %d", got, tt.size)
				}
				if err := tt.record.Fill_Raw(make([]byte, tt.size-1)); err == nil {
					t.Fatal("accepted truncated record")
				}
			}
			fields := append([]wireField(nil), tt.fields...)
			if _, ok := typ.FieldByName("Header"); ok {
				fields = append(fields, wireField{"Header.Length", 0, 1}, wireField{"Header.RType", 1, 1}, wireField{"Header.PublisherID", 2, 2}, wireField{"Header.InstrumentID", 4, 4}, wireField{"Header.TsEvent", 8, 8})
			}
			for _, field := range fields {
				t.Run(field.path, func(t *testing.T) {
					// Distinct nonzero bytes detect endian errors, narrowing, and shifts.
					pattern := make([]byte, field.width)
					for i := range pattern {
						pattern[i] = byte(0x81 + i%0x70)
					}
					expected := make([]byte, tt.size)
					copy(expected[field.offset:field.offset+field.width], pattern)
					value := reflect.New(typ)
					target := wireFieldValue(t, value.Elem(), field.path)
					if got := binary.Size(target.Interface()); got != field.width {
						t.Fatalf("field width %d, want %d", got, field.width)
					}
					if err := binary.Read(bytes.NewReader(pattern), binary.LittleEndian, target.Addr().Interface()); err != nil {
						t.Fatal(err)
					}
					var out bytes.Buffer
					if err := binary.Write(&out, binary.LittleEndian, value.Interface()); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(out.Bytes(), expected) {
						t.Fatalf("encoded field does not occupy bytes [%d:%d]", field.offset, field.offset+field.width)
					}
					decoded := reflect.New(typ)
					if err := decoded.Interface().(wireDecoder).Fill_Raw(expected); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(decoded.Interface(), value.Interface()) {
						t.Fatalf("decoded field differs: got %+v, want %+v", decoded.Interface(), value.Interface())
					}
				})
			}
		})
	}
}

func wireFieldValue(t *testing.T, value reflect.Value, path string) reflect.Value {
	t.Helper()
	for _, part := range strings.Split(path, ".") {
		if value.Kind() == reflect.Array {
			index, err := strconv.Atoi(part)
			if err != nil {
				t.Fatal(err)
			}
			value = value.Index(index)
		} else {
			value = value.FieldByName(part)
		}
		if !value.IsValid() {
			t.Fatalf("unknown field %q", path)
		}
	}
	return value
}

// Symbol mappings contain Go strings and cannot use binary.Write. Construct
// their fixed-width wire representation directly, including V1 alignment padding.
func TestSymbolMappingWireLayouts(t *testing.T) {
	for _, tt := range []struct {
		name                                                           string
		size, inOffset, outOffset, stringWidth, startOffset, endOffset int
	}{
		{"V1", 80, 16, 38, 22, 64, 72},
		{"V2", 176, 17, 89, 71, 160, 168},
	} {
		t.Run(tt.name, func(t *testing.T) {
			raw := make([]byte, tt.size)
			raw[0], raw[1] = byte(tt.size/4), 0x16
			binary.LittleEndian.PutUint32(raw[4:8], 0x12345678)
			in, out := strings.Repeat("I", tt.stringWidth-1), strings.Repeat("O", tt.stringWidth-1)
			copy(raw[tt.inOffset:], in)
			copy(raw[tt.outOffset:], out)
			const start, end uint64 = 0x123456789abcdef0, 0x23456789abcdef01
			binary.LittleEndian.PutUint64(raw[tt.startOffset:], start)
			binary.LittleEndian.PutUint64(raw[tt.endOffset:], end)
			var record wireDecoder
			if tt.name == "V1" {
				copy(raw[60:64], []byte{0xde, 0xad, 0xbe, 0xef})
				record = &dbn.SymbolMappingMsgV1{}
			} else {
				raw[16], raw[88] = byte(dbn.SType_Continuous), byte(dbn.SType_RawSymbol)
				record = &dbn.SymbolMappingMsgV2{}
			}
			if got := int(record.(interface{ RSize() uint16 }).RSize()); got != tt.size {
				t.Fatalf("RSize = %d, want %d", got, tt.size)
			}
			if err := record.Fill_Raw(raw[:tt.size-1]); err == nil {
				t.Fatal("accepted truncated mapping")
			}
			if err := record.Fill_Raw(raw); err != nil {
				t.Fatal(err)
			}
			got := reflect.ValueOf(record).Elem()
			for field, want := range map[string]any{"StypeInSymbol": in, "StypeOutSymbol": out, "StartTs": start, "EndTs": end, "Header.InstrumentID": uint32(0x12345678), "Header.Length": uint8(tt.size / 4)} {
				if value := wireFieldValue(t, got, field).Interface(); !reflect.DeepEqual(value, want) {
					t.Errorf("%s = %v, want %v", field, value, want)
				}
			}
			wantIn := dbn.SType_RawSymbol
			if tt.name == "V2" {
				wantIn = dbn.SType_Continuous
			}
			if got.FieldByName("StypeIn").Interface() != wantIn || got.FieldByName("StypeOut").Interface() != dbn.SType_RawSymbol {
				t.Fatal("incorrect symbology types")
			}
		})
	}
}
