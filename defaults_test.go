// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"bytes"
	"encoding/binary"
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
		{"Mbp0Msg", dbn.NewMbp0Msg().Header, dbn.RType_Mbp0, dbn.Mbp0Msg_Size},
		{"Mbp1Msg", dbn.NewMbp1Msg().Header, dbn.RType_Mbp1, dbn.Mbp1Msg_Size},
		{"Mbp10Msg", dbn.NewMbp10Msg().Header, dbn.RType_Mbp10, dbn.Mbp10Msg_Size},
		{"Cmbp1Msg", dbn.NewCmbp1Msg(dbn.RType_Cmbp1).Header, dbn.RType_Cmbp1, dbn.Cmbp1Msg_Size},
		{"Cmbp1Msg(Cbbo1S)", dbn.NewCmbp1Msg(dbn.RType_Cbbo1S).Header, dbn.RType_Cbbo1S, dbn.Cmbp1Msg_Size},
		{"BboMsg", dbn.NewBboMsg(dbn.RType_Bbo1S).Header, dbn.RType_Bbo1S, dbn.BboMsg_Size},
		{"OhlcvMsg", dbn.NewOhlcvMsg(dbn.RType_Ohlcv1M).Header, dbn.RType_Ohlcv1M, dbn.OhlcvMsg_Size},
		{"StatusMsg", dbn.NewStatusMsg().Header, dbn.RType_Status, dbn.StatusMsg_Size},
		{"ImbalanceMsg", dbn.NewImbalanceMsg().Header, dbn.RType_Imbalance, dbn.ImbalanceMsg_Size},
		{"ErrorMsgV1", dbn.NewErrorMsgV1().Header, dbn.RType_Error, dbn.ErrorMsgV1_Size},
		{"ErrorMsgV2", dbn.NewErrorMsgV2().Header, dbn.RType_Error, dbn.ErrorMsgV2_Size},
		{"SymbolMappingMsgV2", dbn.NewSymbolMappingMsgV2().Header, dbn.RType_SymbolMapping, dbn.SymbolMappingMsgV2_Size},
		{"SystemMsg", dbn.NewSystemMsg().Header, dbn.RType_System, dbn.SystemMsg_Size},
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

func TestDefaultBookLevels(t *testing.T) {
	pair := dbn.NewBidAskPair()
	if pair.BidPx != dbn.UNDEF_PRICE || pair.AskPx != dbn.UNDEF_PRICE || pair.BidSz != 0 {
		t.Errorf("BidAskPair defaults wrong: %+v", pair)
	}
	if dbn.NewMbp1Msg().Level != pair {
		t.Errorf("Mbp1Msg.Level should be a default BidAskPair")
	}
	for i, lvl := range dbn.NewMbp10Msg().Levels {
		if lvl != pair {
			t.Errorf("Mbp10Msg.Levels[%d] should be a default BidAskPair: %+v", i, lvl)
		}
	}
	if dbn.NewBboMsg(dbn.RType_Bbo1S).Level != pair {
		t.Errorf("BboMsg.Level should be a default BidAskPair")
	}
	cpair := dbn.NewConsolidatedBidAskPair()
	if cpair.BidPx != dbn.UNDEF_PRICE || cpair.AskPx != dbn.UNDEF_PRICE || dbn.NewCmbp1Msg(dbn.RType_Cmbp1).Level != cpair {
		t.Errorf("ConsolidatedBidAskPair defaults wrong: %+v", cpair)
	}
}

func TestDefaultConsolidatedActionByte(t *testing.T) {
	for _, tt := range []struct {
		rtype dbn.RType
		want  byte
	}{
		{dbn.RType_Cbbo1S, 0},
		{dbn.RType_Cbbo1M, 0},
		{dbn.RType_Cmbp1, 'N'},
		{dbn.RType_Tcbbo, 'N'},
	} {
		t.Run(tt.rtype.String(), func(t *testing.T) {
			record := dbn.NewCmbp1Msg(tt.rtype)
			var buf bytes.Buffer
			if err := binary.Write(&buf, binary.LittleEndian, record); err != nil {
				t.Fatal(err)
			}
			// Wire offset 28 is reserved in CBBO, but an action in CMBP1/TCBBO.
			if got := buf.Bytes()[28]; got != tt.want {
				t.Fatalf("byte at offset 28: got %#x, want %#x", got, tt.want)
			}
		})
	}
}

func TestDefaultOtherRecords(t *testing.T) {
	if o := dbn.NewOhlcvMsg(dbn.RType_Ohlcv1S); o.Open != dbn.UNDEF_PRICE || o.Close != dbn.UNDEF_PRICE || o.Volume != 0 {
		t.Errorf("OhlcvMsg defaults wrong: %+v", o)
	}
	if st := dbn.NewStatusMsg(); st.IsTrading != '~' || st.TsRecv != dbn.UNDEF_TIMESTAMP {
		t.Errorf("StatusMsg defaults wrong: %+v", st)
	}
	imb := dbn.NewImbalanceMsg()
	if imb.PairedQty != dbn.UNDEF_ORDER_SIZE || imb.AuctionType != '~' || imb.SignificantImbalance != '~' || imb.UnpairedQty != -1 {
		t.Errorf("ImbalanceMsg defaults wrong: %+v", imb)
	}
	if e := dbn.NewErrorMsgV2(); e.Code != dbn.ErrorCode_Unset || e.IsLast != math.MaxUint8 {
		t.Errorf("ErrorMsgV2 defaults wrong: %+v", e)
	}
	if sm := dbn.NewSymbolMappingMsgV2(); sm.StypeIn != math.MaxUint8 || sm.StartTs != dbn.UNDEF_TIMESTAMP || sm.EndTs != dbn.UNDEF_TIMESTAMP {
		t.Errorf("SymbolMappingMsgV2 defaults wrong: %+v", sm)
	}
	if sys := dbn.NewSystemMsg(); sys.Code != dbn.SystemCode_Unset {
		t.Errorf("SystemMsg defaults wrong: %+v", sys)
	}
}
