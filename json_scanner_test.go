package dbn_test

import (
	"os"
	"strings"

	"github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JsonScanner", func() {
	Context("json files", func() {
		It("should read a JSON metadata test file correctly", func() {
			file, err := os.Open("./tests/data/test_data.ohlcv-1s.meta.json")
			Expect(err).To(BeNil())
			defer file.Close()

			// TODO: slurp meta
		})
		It("should read a JSON test file correctly", func() {
			file, err := os.Open("./tests/data/test_data.ohlcv-1s.json")
			Expect(err).To(BeNil())
			defer file.Close()

			records, err := dbn.ReadJsonToSlice[dbn.OhlcvMsg](file)
			Expect(err).To(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.ohlcv-1s.dbn
			// {"hd":{"ts_event":"1609160400000000000","rtype":32,"publisher_id":1,"instrument_id":5482},"open":"372025000000000","high":"372050000000000","low":"372025000000000","close":"372050000000000","volume":"57"}
			// {"hd":{"ts_event":"1609160401000000000","rtype":32,"publisher_id":1,"instrument_id":5482},"open":"372050000000000","high":"372050000000000","low":"372050000000000","close":"372050000000000","volume":"13"}

			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609160400000000000)))
			Expect(r0h.RType).To(Equal(dbn.RType(32)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r0.Open).To(Equal(int64(372025000000000)))
			Expect(r0.High).To(Equal(int64(372050000000000)))
			Expect(r0.Low).To(Equal(int64(372025000000000)))
			Expect(r0.Close).To(Equal(int64(372050000000000)))
			Expect(r0.Volume).To(Equal(uint64(57)))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609160401000000000)))
			Expect(r1h.RType).To(Equal(dbn.RType(32)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r1.Open).To(Equal(int64(372050000000000)))
			Expect(r1.High).To(Equal(int64(372050000000000)))
			Expect(r1.Low).To(Equal(int64(372050000000000)))
			Expect(r1.Close).To(Equal(int64(372050000000000)))
			Expect(r1.Volume).To(Equal(uint64(13)))
		})
	})
})

// /////////////////////////////////////////////////////////////////////////////
// Upstream JSON encodings
//
// The JSON constants below are the upstream JSON encoder's output for the V3
// fixtures and the upstream-encoded live-control streams in dbn_scanner_test.go.
// Expected values are the upstream decoder's records (upstream*V3 vars in
// messagesV3_test.go) or the upstream constructor arguments, never dbn-go output.
//
// JSON records carry no `len`, so Header.Length is cleared before comparing.
//
// mapRecords returns a modified copy of records.
func mapRecords[R any](records []R, f func(*R)) []R {
	out := make([]R, len(records))
	copy(out, records)
	for i := range out {
		f(&out[i])
	}
	return out
}

var _ = Describe("JsonScanner upstream JSON", func() {
	Context("market data (V3 fixtures)", func() {
		mboNoLen := func(r *dbn.MboMsg) { r.Header.Length = 0 }
		mboNoChars := func(r *dbn.MboMsg) { mboNoLen(r); r.Action, r.Side = 0, 0 }
		mbp1NoLen := func(r *dbn.Mbp1Msg) { r.Header.Length = 0 }
		mbp1NoChars := func(r *dbn.Mbp1Msg) { mbp1NoLen(r); r.Action, r.Side = 0, 0 }
		mbp10NoLen := func(r *dbn.Mbp10Msg) { r.Header.Length = 0 }
		mbp10NoChars := func(r *dbn.Mbp10Msg) { mbp10NoLen(r); r.Action, r.Side = 0, 0 }
		cmbp1NoLen := func(r *dbn.Cmbp1Msg) { r.Header.Length = 0 }
		cmbp1NoChars := func(r *dbn.Cmbp1Msg) { cmbp1NoLen(r); r.Action, r.Side = 0, 0 }
		statusNoLen := func(r *dbn.StatusMsg) { r.Header.Length = 0 }
		statusNoChars := func(r *dbn.StatusMsg) {
			statusNoLen(r)
			r.IsTrading, r.IsQuoting, r.IsShortSellRestricted = 0, 0, 0
		}

		It("should decode upstream mbo JSON (non-char fields)", func() {
			records, err := dbn.ReadJsonToSlice[dbn.MboMsg](strings.NewReader(upstreamMboV3Json))
			Expect(err).To(BeNil())
			Expect(mapRecords(records, mboNoChars)).To(Equal(mapRecords(upstreamMboV3, mboNoChars)))
		})
		It("should decode upstream mbo JSON action/side", func() {
			records, err := dbn.ReadJsonToSlice[dbn.MboMsg](strings.NewReader(upstreamMboV3Json))
			Expect(err).To(BeNil())
			Expect(records).To(Equal(mapRecords(upstreamMboV3, mboNoLen)))
		})

		It("should decode upstream mbp1 JSON (non-char fields)", func() {
			records, err := dbn.ReadJsonToSlice[dbn.Mbp1Msg](strings.NewReader(upstreamMbp1V3Json))
			Expect(err).To(BeNil())
			Expect(mapRecords(records, mbp1NoChars)).To(Equal(mapRecords(upstreamMbp1V3, mbp1NoChars)))
		})
		It("should decode upstream mbp1 JSON action/side", func() {
			records, err := dbn.ReadJsonToSlice[dbn.Mbp1Msg](strings.NewReader(upstreamMbp1V3Json))
			Expect(err).To(BeNil())
			Expect(records).To(Equal(mapRecords(upstreamMbp1V3, mbp1NoLen)))
		})

		It("should decode upstream mbp10 JSON (non-char fields)", func() {
			records, err := dbn.ReadJsonToSlice[dbn.Mbp10Msg](strings.NewReader(upstreamMbp10V3Json))
			Expect(err).To(BeNil())
			Expect(mapRecords(records, mbp10NoChars)).To(Equal(mapRecords(upstreamMbp10V3, mbp10NoChars)))
		})
		It("should decode upstream mbp10 JSON action/side", func() {
			records, err := dbn.ReadJsonToSlice[dbn.Mbp10Msg](strings.NewReader(upstreamMbp10V3Json))
			Expect(err).To(BeNil())
			Expect(records).To(Equal(mapRecords(upstreamMbp10V3, mbp10NoLen)))
		})

		It("should decode upstream cmbp1 JSON (non-char fields)", func() {
			records, err := dbn.ReadJsonToSlice[dbn.Cmbp1Msg](strings.NewReader(upstreamCmbp1V3Json))
			Expect(err).To(BeNil())
			Expect(mapRecords(records, cmbp1NoChars)).To(Equal(mapRecords(upstreamCmbp1V3, cmbp1NoChars)))
		})
		It("should decode upstream cmbp1 JSON action/side", func() {
			records, err := dbn.ReadJsonToSlice[dbn.Cmbp1Msg](strings.NewReader(upstreamCmbp1V3Json))
			Expect(err).To(BeNil())
			Expect(records).To(Equal(mapRecords(upstreamCmbp1V3, cmbp1NoLen)))
		})

		It("should decode upstream status JSON (non-char fields)", func() {
			records, err := dbn.ReadJsonToSlice[dbn.StatusMsg](strings.NewReader(upstreamStatusV3Json))
			Expect(err).To(BeNil())
			Expect(mapRecords(records, statusNoChars)).To(Equal(mapRecords(upstreamStatusV3, statusNoChars)))
		})
		It("should decode upstream status JSON is_trading/is_quoting/is_short_sell_restricted", func() {
			records, err := dbn.ReadJsonToSlice[dbn.StatusMsg](strings.NewReader(upstreamStatusV3Json))
			Expect(err).To(BeNil())
			Expect(records).To(Equal(mapRecords(upstreamStatusV3, statusNoLen)))
		})

		It("should decode upstream bbo-1s and bbo-1m JSON and dispatch via Visit", func() {
			bboNoLen := func(r *dbn.BboMsg) { r.Header.Length = 0 }
			want1s := mapRecords(readV3Fixture[dbn.BboMsg]("./tests/data/test_data.bbo-1s.v3.dbn.zst"), bboNoLen)
			got1s, err := dbn.ReadJsonToSlice[dbn.BboMsg](strings.NewReader(upstreamBbo1sJson))
			Expect(err).To(BeNil())
			Expect(got1s).To(Equal(want1s))

			want1m := mapRecords(readV3Fixture[dbn.BboMsg]("./tests/data/test_data.bbo-1m.v3.dbn.zst"), bboNoLen)
			got1m, err := dbn.ReadJsonToSlice[dbn.BboMsg](strings.NewReader(upstreamBbo1mJson))
			Expect(err).To(BeNil())
			Expect(got1m[:2]).To(Equal(want1m[:2]))

			scanner := dbn.NewJsonScanner(strings.NewReader(upstreamBbo1sJson + upstreamBbo1mJson))
			var visited []dbn.BboMsg
			v := &bboDefJsonVisitor{onBbo: func(r *dbn.BboMsg) { visited = append(visited, *r) }}
			for scanner.Next() {
				Expect(scanner.Visit(v)).To(Succeed())
			}
			Expect(scanner.Error()).To(BeNil())
			Expect(visited).To(Equal(append(want1s, want1m[:2]...)))

			_, err = dbn.ReadJsonToSlice[dbn.BboMsg](strings.NewReader(`{"hd":{"rtype":195},"levels":[]}`))
			Expect(err).To(HaveOccurred())
		})

		It("should decode upstream definition V3 and V2 JSON and dispatch via Visit", func() {
			defV3NoLen := func(r *dbn.InstrumentDefMsgV3) { r.Header.Length = 0 }
			wantV3 := mapRecords(readV3Fixture[dbn.InstrumentDefMsgV3]("./tests/data/test_data.definition.v3.dbn.zst"), defV3NoLen)
			gotV3, err := dbn.ReadJsonToSlice[dbn.InstrumentDefMsgV3](strings.NewReader(upstreamDefinitionV3Record1Json))
			Expect(err).To(BeNil())
			Expect(gotV3).To(Equal(wantV3[1:2]))

			scanner := dbn.NewJsonScanner(strings.NewReader(upstreamDefinitionV3Record1Json))
			var visited []dbn.InstrumentDefMsg
			v := &bboDefJsonVisitor{onDef: func(r *dbn.InstrumentDefMsg) { visited = append(visited, *r) }}
			Expect(scanner.Next()).To(BeTrue())
			Expect(scanner.Visit(v)).To(Succeed())
			Expect(visited).To(Equal(wantV3[1:2]))

			gotV2, err := dbn.ReadJsonToSlice[dbn.InstrumentDefMsgV2](strings.NewReader(upstreamDefinitionV3Record1Json))
			Expect(err).To(BeNil())
			Expect(gotV2).To(HaveLen(1))
			Expect(gotV2[0].InstrumentClass).To(Equal(byte('K')))
			Expect(gotV2[0].MatchAlgorithm).To(Equal(byte('F')))
			Expect(gotV2[0].SecurityUpdateAction).To(Equal(byte('A')))
			Expect(gotV2[0].UserDefinedInstrument).To(Equal(dbn.UserDefinedInstrument('N')))
		})
	})

	Context("live-control records", func() {
		It("should decode upstream V3 SymbolMappingMsg JSON", func() {
			scanner := dbn.NewJsonScanner(strings.NewReader(upstreamLiveControlStreamV3Json))
			Expect(scanner.Next()).To(BeTrue())
			sm, err := dbn.JsonScannerDecode[dbn.SymbolMappingMsg](scanner)
			Expect(err).To(BeNil())
			// SymbolMappingMsg(1, 5482, 1609160400000000006, RAW_SYMBOL, "ESH1", INSTRUMENT_ID, "5482", START, END)
			Expect(sm.Header).To(Equal(dbn.RHeader{RType: dbn.RType_SymbolMapping, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609160400000000006}))
			Expect(sm.StypeIn).To(Equal(dbn.SType_RawSymbol))
			Expect(sm.StypeInSymbol).To(Equal("ESH1"))
			Expect(sm.StypeOut).To(Equal(dbn.SType_InstrumentId))
			Expect(sm.StypeOutSymbol).To(Equal("5482"))
		})

		It("should decode upstream V3 SymbolMappingMsg JSON start_ts/end_ts", func() {
			scanner := dbn.NewJsonScanner(strings.NewReader(upstreamLiveControlStreamV3Json))
			Expect(scanner.Next()).To(BeTrue())
			sm, err := dbn.JsonScannerDecode[dbn.SymbolMappingMsg](scanner)
			Expect(err).To(BeNil())
			Expect(sm.StartTs).To(Equal(uint64(1609113600000000000)))
			Expect(sm.EndTs).To(Equal(uint64(1609200000000000000)))
		})

		It("should decode upstream V1 SymbolMappingMsg JSON start_ts/end_ts", func() {
			scanner := dbn.NewJsonScanner(strings.NewReader(upstreamLiveControlStreamV1Json))
			Expect(scanner.Next()).To(BeTrue())
			sm, err := dbn.JsonScannerDecode[dbn.SymbolMappingMsgV1](scanner)
			Expect(err).To(BeNil())
			Expect(sm.StartTs).To(Equal(uint64(1609113600000000000)))
			Expect(sm.EndTs).To(Equal(uint64(1609200000000000000)))
		})

		It("should decode upstream V3 SystemMsg JSON", func() {
			scanner := dbn.NewJsonScanner(strings.NewReader(upstreamLiveControlStreamV3Json))
			Expect(scanner.Next()).To(BeTrue()) // symbol mapping
			Expect(scanner.Next()).To(BeTrue())
			s0, err := dbn.JsonScannerDecode[dbn.SystemMsg](scanner)
			Expect(err).To(BeNil())
			// SystemMsg(1609160400000000003, "Subscription request 1 for mbo data succeeded", code=SUBSCRIPTION_ACK)
			Expect(s0.Header).To(Equal(dbn.RHeader{RType: dbn.RType_System, TsEvent: 1609160400000000003}))
			Expect(dbn.TrimNullBytes(s0.Message[:])).To(Equal("Subscription request 1 for mbo data succeeded"))
			Expect(s0.Code).To(Equal(dbn.SystemCode(1)))

			Expect(scanner.Next()).To(BeTrue())
			s1, err := dbn.JsonScannerDecode[dbn.SystemMsg](scanner)
			Expect(err).To(BeNil())
			// SystemMsg(1609160400000000004, "Heartbeat", code=HEARTBEAT)
			Expect(s1.Header.TsEvent).To(Equal(uint64(1609160400000000004)))
			Expect(dbn.TrimNullBytes(s1.Message[:])).To(Equal("Heartbeat"))
			Expect(s1.Code).To(Equal(dbn.SystemCode(0)))
			Expect(s1.IsHeartbeat()).To(BeTrue())
		})

		It("should decode upstream V3 ErrorMsg JSON", func() {
			scanner := dbn.NewJsonScanner(strings.NewReader(upstreamLiveControlStreamV3Json))
			for i := 0; i < 4; i++ {
				Expect(scanner.Next()).To(BeTrue())
			}
			e0, err := dbn.JsonScannerDecode[dbn.ErrorMsg](scanner)
			Expect(err).To(BeNil())
			// ErrorMsg(1609160400000000000, "Authentication failed: bad key", is_last=True, code=AUTH_FAILED)
			Expect(e0.Header).To(Equal(dbn.RHeader{RType: dbn.RType_Error, TsEvent: 1609160400000000000}))
			Expect(dbn.TrimNullBytes(e0.Error[:])).To(Equal("Authentication failed: bad key"))
			Expect(e0.Code).To(Equal(dbn.ErrorCode(1)))
			Expect(e0.IsLast).To(Equal(uint8(1)))
			Expect(scanner.Next()).To(BeFalse())
		})

		It("should decode upstream V1 SymbolMappingMsg, SystemMsg and ErrorMsg JSON", func() {
			scanner := dbn.NewJsonScanner(strings.NewReader(upstreamLiveControlStreamV1Json))
			Expect(scanner.Next()).To(BeTrue())
			sm, err := dbn.JsonScannerDecode[dbn.SymbolMappingMsgV1](scanner)
			Expect(err).To(BeNil())
			// SymbolMappingMsgV1(1, 5482, 1609160400000000007, "ESH1", "5482", START, END)
			Expect(sm.Header).To(Equal(dbn.RHeader{RType: dbn.RType_SymbolMapping, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609160400000000007}))
			Expect(sm.StypeInSymbol).To(Equal("ESH1"))
			Expect(sm.StypeOutSymbol).To(Equal("5482"))

			Expect(scanner.Next()).To(BeTrue())
			s0, err := dbn.JsonScannerDecode[dbn.SystemMsg](scanner)
			Expect(err).To(BeNil())
			// SystemMsgV1(1609160400000000005, "Heartbeat")
			Expect(s0.Header.TsEvent).To(Equal(uint64(1609160400000000005)))
			Expect(dbn.TrimNullBytes(s0.Message[:])).To(Equal("Heartbeat"))

			Expect(scanner.Next()).To(BeTrue())
			e0, err := dbn.JsonScannerDecode[dbn.ErrorMsgV1](scanner)
			Expect(err).To(BeNil())
			// ErrorMsgV1(1609160400000000002, "Unknown dataset")
			Expect(e0.Header.TsEvent).To(Equal(uint64(1609160400000000002)))
			Expect(dbn.TrimNullBytes(e0.Error[:])).To(Equal("Unknown dataset"))
		})
	})
})

// databento_dbn 0.70.0 Transcoder(JSON, pretty_px=False, pretty_ts=False) of ./tests/data/test_data.mbo.v3.dbn.zst
const upstreamMboV3Json = `{"ts_recv":"1609160400000704060","hd":{"ts_event":"1609160400000429831","rtype":160,"publisher_id":1,"instrument_id":5482},"action":"C","side":"A","price":"3722750000000","size":1,"channel_id":0,"order_id":"647784973705","flags":128,"ts_in_delta":22993,"sequence":1170352}
{"ts_recv":"1609160400000711344","hd":{"ts_event":"1609160400000431665","rtype":160,"publisher_id":1,"instrument_id":5482},"action":"C","side":"A","price":"3723000000000","size":1,"channel_id":0,"order_id":"647784973631","flags":128,"ts_in_delta":19621,"sequence":1170353}
`

// databento_dbn 0.70.0 Transcoder(JSON, pretty_px=False, pretty_ts=False) of ./tests/data/test_data.mbp-1.v3.dbn.zst
const upstreamMbp1V3Json = `{"ts_recv":"1609160400006136329","hd":{"ts_event":"1609160400006001487","rtype":1,"publisher_id":1,"instrument_id":5482},"action":"A","side":"A","depth":0,"price":"3720500000000","size":1,"flags":128,"ts_in_delta":17214,"sequence":1170362,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":11,"bid_ct":15,"ask_ct":9}]}
{"ts_recv":"1609160400006246513","hd":{"ts_event":"1609160400006146661","rtype":1,"publisher_id":1,"instrument_id":5482},"action":"A","side":"A","depth":0,"price":"3720500000000","size":1,"flags":128,"ts_in_delta":18858,"sequence":1170364,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":12,"bid_ct":15,"ask_ct":10}]}
`

// databento_dbn 0.70.0 Transcoder(JSON, pretty_px=False, pretty_ts=False) of ./tests/data/test_data.mbp-10.v3.dbn.zst
const upstreamMbp10V3Json = `{"ts_recv":"1609160400000704060","hd":{"ts_event":"1609160400000429831","rtype":10,"publisher_id":1,"instrument_id":5482},"action":"C","side":"A","depth":9,"price":"3722750000000","size":1,"flags":128,"ts_in_delta":22993,"sequence":1170352,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":10,"bid_ct":15,"ask_ct":8},{"bid_px":"3720000000000","ask_px":"3720750000000","bid_sz":31,"ask_sz":34,"bid_ct":18,"ask_ct":24},{"bid_px":"3719750000000","ask_px":"3721000000000","bid_sz":32,"ask_sz":39,"bid_ct":23,"ask_ct":25},{"bid_px":"3719500000000","ask_px":"3721250000000","bid_sz":39,"ask_sz":28,"bid_ct":26,"ask_ct":17},{"bid_px":"3719250000000","ask_px":"3721500000000","bid_sz":50,"ask_sz":33,"bid_ct":35,"ask_ct":19},{"bid_px":"3719000000000","ask_px":"3721750000000","bid_sz":42,"ask_sz":45,"bid_ct":28,"ask_ct":33},{"bid_px":"3718750000000","ask_px":"3722000000000","bid_sz":44,"ask_sz":55,"bid_ct":35,"ask_ct":40},{"bid_px":"3718500000000","ask_px":"3722250000000","bid_sz":64,"ask_sz":59,"bid_ct":39,"ask_ct":38},{"bid_px":"3718250000000","ask_px":"3722500000000","bid_sz":53,"ask_sz":49,"bid_ct":32,"ask_ct":35},{"bid_px":"3718000000000","ask_px":"3722750000000","bid_sz":67,"ask_sz":44,"bid_ct":39,"ask_ct":26}]}
{"ts_recv":"1609160400000750544","hd":{"ts_event":"1609160400000435673","rtype":10,"publisher_id":1,"instrument_id":5482},"action":"C","side":"B","depth":1,"price":"3720000000000","size":1,"flags":128,"ts_in_delta":20625,"sequence":1170356,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":10,"bid_ct":15,"ask_ct":8},{"bid_px":"3720000000000","ask_px":"3720750000000","bid_sz":30,"ask_sz":34,"bid_ct":17,"ask_ct":24},{"bid_px":"3719750000000","ask_px":"3721000000000","bid_sz":32,"ask_sz":39,"bid_ct":23,"ask_ct":25},{"bid_px":"3719500000000","ask_px":"3721250000000","bid_sz":39,"ask_sz":28,"bid_ct":26,"ask_ct":17},{"bid_px":"3719250000000","ask_px":"3721500000000","bid_sz":50,"ask_sz":33,"bid_ct":35,"ask_ct":19},{"bid_px":"3719000000000","ask_px":"3721750000000","bid_sz":42,"ask_sz":45,"bid_ct":28,"ask_ct":33},{"bid_px":"3718750000000","ask_px":"3722000000000","bid_sz":44,"ask_sz":55,"bid_ct":35,"ask_ct":40},{"bid_px":"3718500000000","ask_px":"3722250000000","bid_sz":64,"ask_sz":59,"bid_ct":39,"ask_ct":38},{"bid_px":"3718250000000","ask_px":"3722500000000","bid_sz":53,"ask_sz":49,"bid_ct":32,"ask_ct":35},{"bid_px":"3718000000000","ask_px":"3722750000000","bid_sz":67,"ask_sz":44,"bid_ct":39,"ask_ct":26}]}
`

// databento_dbn 0.70.0 Transcoder(JSON, pretty_px=False, pretty_ts=False) of ./tests/data/test_data.cmbp-1.v3.dbn.zst
const upstreamCmbp1V3Json = `{"ts_recv":"1609160400006136329","hd":{"ts_event":"1609160400006001487","rtype":177,"publisher_id":1,"instrument_id":5482},"action":"A","side":"A","price":"3720500000000","size":1,"flags":128,"ts_in_delta":17214,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":11,"bid_pb":1,"ask_pb":1}]}
{"ts_recv":"1609160400006246513","hd":{"ts_event":"1609160400006146661","rtype":177,"publisher_id":1,"instrument_id":5482},"action":"A","side":"A","price":"3720500000000","size":1,"flags":128,"ts_in_delta":18858,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":12,"bid_pb":1,"ask_pb":1}]}
`

// databento_dbn 0.70.0 Transcoder(JSON, pretty_px=False, pretty_ts=False) of ./tests/data/test_data.status.v3.dbn.zst
const upstreamStatusV3Json = `{"ts_recv":"1609113600000000000","hd":{"ts_event":"1609110000000000000","rtype":18,"publisher_id":1,"instrument_id":5482},"action":7,"reason":1,"trading_event":0,"is_trading":"Y","is_quoting":"Y","is_short_sell_restricted":"~"}
{"ts_recv":"1609190100007055917","hd":{"ts_event":"1609190100000000000","rtype":18,"publisher_id":1,"instrument_id":5482},"action":1,"reason":1,"trading_event":0,"is_trading":"N","is_quoting":"Y","is_short_sell_restricted":"~"}
{"ts_recv":"1609190970068184258","hd":{"ts_event":"1609190970000000000","rtype":18,"publisher_id":1,"instrument_id":5482},"action":1,"reason":1,"trading_event":1,"is_trading":"N","is_quoting":"Y","is_short_sell_restricted":"~"}
{"ts_recv":"1609191000007282029","hd":{"ts_event":"1609191000000000000","rtype":18,"publisher_id":1,"instrument_id":5482},"action":6,"reason":1,"trading_event":0,"is_trading":"Y","is_quoting":"Y","is_short_sell_restricted":"~"}
`

// databento_dbn 0.70.0 Transcoder(JSON, pretty_px=False, pretty_ts=False) of upstreamLiveControlStreamV3
const upstreamLiveControlStreamV3Json = `{"hd":{"ts_event":"1609160400000000006","rtype":22,"publisher_id":1,"instrument_id":5482},"stype_in":1,"stype_in_symbol":"ESH1","stype_out":0,"stype_out_symbol":"5482","start_ts":"1609113600000000000","end_ts":"1609200000000000000"}
{"hd":{"ts_event":"1609160400000000003","rtype":23,"publisher_id":0,"instrument_id":0},"msg":"Subscription request 1 for mbo data succeeded","code":1}
{"hd":{"ts_event":"1609160400000000004","rtype":23,"publisher_id":0,"instrument_id":0},"msg":"Heartbeat","code":0}
{"hd":{"ts_event":"1609160400000000000","rtype":21,"publisher_id":0,"instrument_id":0},"err":"Authentication failed: bad key","code":1,"is_last":1}
`

// databento_dbn 0.70.0 Transcoder(JSON, pretty_px=False, pretty_ts=False) of upstreamLiveControlStreamV1
const upstreamLiveControlStreamV1Json = `{"hd":{"ts_event":"1609160400000000007","rtype":22,"publisher_id":1,"instrument_id":5482},"stype_in_symbol":"ESH1","stype_out_symbol":"5482","start_ts":"1609113600000000000","end_ts":"1609200000000000000"}
{"hd":{"ts_event":"1609160400000000005","rtype":23,"publisher_id":0,"instrument_id":0},"msg":"Heartbeat"}
{"hd":{"ts_event":"1609160400000000002","rtype":21,"publisher_id":0,"instrument_id":0},"err":"Unknown dataset"}
`

type bboDefJsonVisitor struct {
	dbn.NullVisitor
	onBbo func(*dbn.BboMsg)
	onDef func(*dbn.InstrumentDefMsg)
}

func (v *bboDefJsonVisitor) OnBbo(r *dbn.BboMsg) error {
	if v.onBbo != nil {
		v.onBbo(r)
	}
	return nil
}

func (v *bboDefJsonVisitor) OnInstrumentDefMsg(r *dbn.InstrumentDefMsg) error {
	if v.onDef != nil {
		v.onDef(r)
	}
	return nil
}

// dbn -J ./tests/data/test_data.bbo-1s.dbn
const upstreamBbo1sJson = `{"ts_recv":"1609113600000000000","hd":{"ts_event":"1609113599045849637","rtype":195,"publisher_id":1,"instrument_id":5482},"side":"A","price":"3702500000000","size":2,"flags":168,"sequence":145799,"levels":[{"bid_px":"3702250000000","ask_px":"3702750000000","bid_sz":18,"ask_sz":13,"bid_ct":10,"ask_ct":13}]}
{"ts_recv":"1609113601000000000","hd":{"ts_event":"1609113600986911551","rtype":195,"publisher_id":1,"instrument_id":5482},"side":"B","price":"3702500000000","size":2,"flags":130,"sequence":145998,"levels":[{"bid_px":"3702500000000","ask_px":"3702750000000","bid_sz":2,"ask_sz":10,"bid_ct":1,"ask_ct":10}]}
{"ts_recv":"1609113602000000000","hd":{"ts_event":"1609113601149205775","rtype":195,"publisher_id":1,"instrument_id":5482},"side":"A","price":"3702500000000","size":1,"flags":130,"sequence":146034,"levels":[{"bid_px":"3702250000000","ask_px":"3702750000000","bid_sz":20,"ask_sz":11,"bid_ct":12,"ask_ct":11}]}
{"ts_recv":"1609113603000000000","hd":{"ts_event":"1609113602738022089","rtype":195,"publisher_id":1,"instrument_id":5482},"side":"B","price":"3702500000000","size":1,"flags":130,"sequence":146167,"levels":[{"bid_px":"3702500000000","ask_px":"3702750000000","bid_sz":2,"ask_sz":11,"bid_ct":2,"ask_ct":11}]}
`

// dbn -J ./tests/data/test_data.bbo-1m.dbn (first 2 records)
const upstreamBbo1mJson = `{"ts_recv":"1609113600000000000","hd":{"ts_event":"1609113599045849637","rtype":196,"publisher_id":1,"instrument_id":5482},"side":"A","price":"3702500000000","size":2,"flags":168,"sequence":145799,"levels":[{"bid_px":"3702250000000","ask_px":"3702750000000","bid_sz":18,"ask_sz":13,"bid_ct":10,"ask_ct":13}]}
{"ts_recv":"1609113660000000000","hd":{"ts_event":"1609113659578979253","rtype":196,"publisher_id":1,"instrument_id":5482},"side":"B","price":"3704750000000","size":1,"flags":130,"sequence":149903,"levels":[{"bid_px":"3704500000000","ask_px":"3705000000000","bid_sz":15,"ask_sz":22,"bid_ct":10,"ask_ct":22}]}
`

// dbn -J ./tests/data/test_data.definition.v3.dbn.zst (record 1)
const upstreamDefinitionV3Record1Json = `{"ts_recv":"1633417621703120931","hd":{"ts_event":"1633417621703109854","rtype":19,"publisher_id":2,"instrument_id":6830},"raw_symbol":"MSFT","security_update_action":"A","instrument_class":"K","min_price_increment":"9223372036854775807","display_factor":"100000000000000","expiration":"18446744073709551615","activation":"18446744073709551615","high_limit_price":"9223372036854775807","low_limit_price":"9223372036854775807","max_price_variation":"9223372036854775807","unit_of_measure_qty":"9223372036854775807","min_price_increment_amount":"9223372036854775807","price_ratio":"9223372036854775807","inst_attrib_value":2147483647,"underlying_id":0,"raw_instrument_id":"2147483647","market_depth_implied":2147483647,"market_depth":2147483647,"market_segment_id":4294967295,"max_trade_vol":4294967295,"min_lot_size":2147483647,"min_lot_size_block":2147483647,"min_lot_size_round_lot":100,"min_trade_vol":4294967295,"contract_multiplier":2147483647,"decay_quantity":2147483647,"original_contract_size":2147483647,"appl_id":32767,"maturity_year":65535,"decay_start_date":65535,"channel_id":0,"currency":"","settl_currency":"","secsubtype":"Z ","group":"pxnas-1","exchange":"XNAS","asset":"","cfi":"","security_type":"","unit_of_measure":"","underlying":"","strike_price_currency":"","strike_price":"9223372036854775807","match_algorithm":"F","main_fraction":255,"price_display_format":255,"sub_fraction":255,"underlying_product":255,"maturity_month":255,"maturity_day":255,"maturity_week":255,"user_defined_instrument":"N","contract_multiplier_unit":127,"flow_schedule_type":127,"tick_rule":255,"leg_count":0,"leg_index":0,"leg_instrument_id":0,"leg_raw_symbol":"","leg_instrument_class":null,"leg_side":"N","leg_price":"9223372036854775807","leg_delta":"9223372036854775807","leg_ratio_price_numerator":0,"leg_ratio_price_denominator":0,"leg_ratio_qty_numerator":0,"leg_ratio_qty_denominator":0,"leg_underlying_id":0}
`
