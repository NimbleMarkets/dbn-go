// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"math"
	"testing"

	dbn "github.com/NimbleMarkets/dbn-go"
)

func TestDefaultRecordHeaders(t *testing.T) {
	tests := []struct {
		name  string
		hd    dbn.RHeader
		rtype dbn.RType
		size  int
	}{
		{"MboMsg", dbn.NewMboMsg().Header, dbn.RType_Mbo, dbn.MboMsg_Size},
		{"StatMsgV2", dbn.NewStatMsgV2().Header, dbn.RType_Statistics, dbn.StatMsgV2_Size},
		{"StatMsgV3", dbn.NewStatMsgV3().Header, dbn.RType_Statistics, dbn.StatMsgV3_Size},
		{"InstrumentDefMsgV2", dbn.NewInstrumentDefMsgV2().Header, dbn.RType_InstrumentDef, dbn.InstrumentDefMsgV2_Size},
		{"InstrumentDefMsgV3", dbn.NewInstrumentDefMsgV3().Header, dbn.RType_InstrumentDef, dbn.InstrumentDefMsgV3_Size},
	}
	for _, tt := range tests {
		if tt.hd.RType != tt.rtype {
			t.Errorf("%s: RType got %v, want %v", tt.name, tt.hd.RType, tt.rtype)
		}
		if int(tt.hd.Length)*4 != tt.size {
			t.Errorf("%s: Length*4 got %d, want %d", tt.name, int(tt.hd.Length)*4, tt.size)
		}
		if tt.hd.TsEvent != dbn.UNDEF_TIMESTAMP {
			t.Errorf("%s: TsEvent got %d, want UNDEF_TIMESTAMP", tt.name, tt.hd.TsEvent)
		}
	}
}

func TestDefaultChannelID(t *testing.T) {
	if got := dbn.NewMboMsg().ChannelID; got != math.MaxUint8 {
		t.Errorf("MboMsg.ChannelID: got %d, want %d", got, math.MaxUint8)
	}
	ids := map[string]uint16{
		"StatMsgV2":          dbn.NewStatMsgV2().ChannelID,
		"StatMsgV3":          dbn.NewStatMsgV3().ChannelID,
		"InstrumentDefMsgV2": dbn.NewInstrumentDefMsgV2().ChannelID,
		"InstrumentDefMsgV3": dbn.NewInstrumentDefMsgV3().ChannelID,
	}
	for name, got := range ids {
		if got != math.MaxUint16 {
			t.Errorf("%s.ChannelID: got %d, want %d", name, got, math.MaxUint16)
		}
	}
}

func TestDefaultFieldValues(t *testing.T) {
	mbo := dbn.NewMboMsg()
	if mbo.Price != dbn.UNDEF_PRICE || mbo.TsRecv != dbn.UNDEF_TIMESTAMP || mbo.Size != 0 {
		t.Errorf("MboMsg defaults wrong: %+v", mbo)
	}
	if mbo.Action != byte(dbn.Action_None) || mbo.Side != byte(dbn.Side_None) {
		t.Errorf("MboMsg Action/Side: got %c/%c, want N/N", mbo.Action, mbo.Side)
	}
	stat := dbn.NewStatMsgV3()
	if stat.Quantity != math.MaxInt64 || stat.UpdateAction != uint8(dbn.StatUpdateAction_New) {
		t.Errorf("StatMsgV3 defaults wrong: %+v", stat)
	}
	if dbn.NewStatMsgV2().Quantity != math.MaxInt32 {
		t.Errorf("StatMsgV2.Quantity: got %d, want MaxInt32", dbn.NewStatMsgV2().Quantity)
	}
	def := dbn.NewInstrumentDefMsgV3()
	if def.MatchAlgorithm != byte(dbn.MatchAlgorithm_Undefined) || def.UserDefinedInstrument != dbn.UserDefinedInstrument_No {
		t.Errorf("InstrumentDefMsgV3 enum defaults wrong: %c %c", def.MatchAlgorithm, def.UserDefinedInstrument)
	}
	if def.StrikePrice != dbn.UNDEF_PRICE || def.MaturityYear != math.MaxUint16 || def.InstAttribValue != math.MaxInt32 {
		t.Errorf("InstrumentDefMsgV3 sentinel defaults wrong")
	}
}
