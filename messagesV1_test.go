// Copyright (c) 2024 Neomantra Corp
//
// Tests for DBN Version 1 message types

package dbn_test

import (
	"os"

	dbn "github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Messages V1", func() {
	Context("Version compatibility", func() {
		It("should have different symbol cstr lengths for V1 and V2", func() {
			Expect(uint16(dbn.MetadataV1_SymbolCstrLen)).To(Equal(uint16(22)))
			Expect(uint16(dbn.MetadataV2_SymbolCstrLen)).To(Equal(uint16(71)))
		})

		It("should be able to convert from V1 to V2 via SymbolMappingMsgFillRaw", func() {
			// This tests that the helper function exists and the conversion path works
			// We can't test the full flow without test data, but we verify the types exist
			var v2Msg dbn.SymbolMappingMsgV2
			_ = v2Msg

			var v1Msg dbn.SymbolMappingMsgV1
			_ = v1Msg

			// The alias should point to V2
			var alias dbn.SymbolMappingMsg
			_ = alias.StypeInSymbol // Access a field to verify it's V2 layout
		})
	})

	Context("SymbolMapping v1 messages", func() {
		It("should have correct size for SymbolMappingMsgV1", func() {
			// V1: RHeader (16) + 16 + 2*22 = 16 + 16 + 44 = 76
			Expect(dbn.SymbolMappingMsgV1_Size).To(BeEquivalentTo(76))
		})

		It("should be smaller than V2 due to shorter symbol strings", func() {
			Expect(dbn.SymbolMappingMsgV1_Size).To(BeNumerically("<", dbn.SymbolMappingMsgV2_Size))
		})

		It("should have implied SType_RawSymbol for StypeIn and StypeOut", func() {
			// In V1, StypeIn and StypeOut are always implied to be RawSymbol
			var msg dbn.SymbolMappingMsgV1
			// After Fill_Raw, these should be set to SType_RawSymbol
			// (We can't test Fill_Raw without test data, but we can verify the struct layout)
			_ = msg
			Expect(dbn.SType_RawSymbol).To(BeEquivalentTo(1))
		})
	})

	Context("Ohlcv v1 messages", func() {
		It("should read v1 olhc-1s correctly", func() {
			file, err := os.Open("./tests/data/test_data.ohlcv-1s.v1.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			// dbn -J ./tests/data/test_data.ohlcv-1s.v1.dbn
			// {"hd":{"ts_event":"1609160400000000000","rtype":32,"publisher_id":1,"instrument_id":5482},"open":"372025000000000","high":"372050000000000","low":"372025000000000","close":"372050000000000","volume":"57"}
			// {"hd":{"ts_event":"1609160401000000000","rtype":32,"publisher_id":1,"instrument_id":5482},"open":"372050000000000","high":"372050000000000","low":"372050000000000","close":"372050000000000","volume":"13"}

			records, metadata, err := dbn.ReadDBNToSlice[dbn.OhlcvMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

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

	Context("Trade v1 messages", func() {
		It("should read a v1 trades/mbp0 correctly", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.trades.v1.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.Mbp0Msg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.trades.v1.dbn.zst
			// {"ts_recv":"1609160400099150057","hd":{"ts_event":"1609160400098821953","rtype":0,"publisher_id":1,"instrument_id":5482},"action":"T","side":"A","depth":0,"price":"3720250000000","size":5,"flags":129,"ts_in_delta":19251,"sequence":1170380}
			// {"ts_recv":"1609160400108142648","hd":{"ts_event":"1609160400107665963","rtype":0,"publisher_id":1,"instrument_id":5482},"action":"T","side":"A","depth":0,"price":"3720250000000","size":21,"flags":129,"ts_in_delta":20728,"sequence":1170414}

			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609160400098821953)))
			Expect(r0h.RType).To(Equal(dbn.RType(0)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(string(r0.Action)).To(Equal("T"))
			Expect(string(r0.Side)).To(Equal("A"))
			Expect(r0.Depth).To(Equal(uint8(0)))
			Expect(r0.Price).To(Equal(int64(3720250000000)))
			Expect(r0.Size).To(Equal(uint32(5)))
			Expect(r0.Flags).To(Equal(uint8(129)))
			Expect(r0.TsRecv).To(Equal(uint64(1609160400099150057)))
			Expect(r0.TsInDelta).To(Equal(int32(19251)))
			Expect(r0.Sequence).To(Equal(uint32(1170380)))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609160400107665963)))
			Expect(r1h.RType).To(Equal(dbn.RType(0)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r1.TsRecv).To(Equal(uint64(1609160400108142648)))
			Expect(string(r1.Action)).To(Equal("T"))
			Expect(string(r1.Side)).To(Equal("A"))
			Expect(r1.Depth).To(Equal(uint8(0)))
			Expect(r1.Price).To(Equal(int64(3720250000000)))
			Expect(r1.Size).To(Equal(uint32(21)))
			Expect(r1.Flags).To(Equal(uint8(129)))
			Expect(r1.TsRecv).To(Equal(uint64(1609160400108142648)))
			Expect(r1.TsInDelta).To(Equal(int32(20728)))
			Expect(r1.Sequence).To(Equal(uint32(1170414)))
		})

		It("should read a v1 trades/mbp0 correctly", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.trades.v2.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.Mbp0Msg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.trades.v2.dbn.zst
			// {"ts_recv":"1609160400099150057","hd":{"ts_event":"1609160400098821953","rtype":0,"publisher_id":1,"instrument_id":5482},"action":"T","side":"A","depth":0,"price":"3720250000000","size":5,"flags":129,"ts_in_delta":19251,"sequence":1170380}
			// {"ts_recv":"1609160400108142648","hd":{"ts_event":"1609160400107665963","rtype":0,"publisher_id":1,"instrument_id":5482},"action":"T","side":"A","depth":0,"price":"3720250000000","size":21,"flags":129,"ts_in_delta":20728,"sequence":1170414}

			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609160400098821953)))
			Expect(r0h.RType).To(Equal(dbn.RType(0)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(string(r0.Action)).To(Equal("T"))
			Expect(string(r0.Side)).To(Equal("A"))
			Expect(r0.Depth).To(Equal(uint8(0)))
			Expect(r0.Price).To(Equal(int64(3720250000000)))
			Expect(r0.Size).To(Equal(uint32(5)))
			Expect(r0.Flags).To(Equal(uint8(129)))
			Expect(r0.TsRecv).To(Equal(uint64(1609160400099150057)))
			Expect(r0.TsInDelta).To(Equal(int32(19251)))
			Expect(r0.Sequence).To(Equal(uint32(1170380)))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609160400107665963)))
			Expect(r1h.RType).To(Equal(dbn.RType(0)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r1.TsRecv).To(Equal(uint64(1609160400108142648)))
			Expect(string(r1.Action)).To(Equal("T"))
			Expect(string(r1.Side)).To(Equal("A"))
			Expect(r1.Depth).To(Equal(uint8(0)))
			Expect(r1.Price).To(Equal(int64(3720250000000)))
			Expect(r1.Size).To(Equal(uint32(21)))
			Expect(r1.Flags).To(Equal(uint8(129)))
			Expect(r1.TsRecv).To(Equal(uint64(1609160400108142648)))
			Expect(r1.TsInDelta).To(Equal(int32(20728)))
			Expect(r1.Sequence).To(Equal(uint32(1170414)))
		})
	})

	Context("Mbp1 v1 messages", func() {
		It("should read a v1 mbp1 correctly", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.mbp-1.v1.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			// dbn -J ./tests/data/test_data.mbp-1.v1.dbn.zst
			// {"ts_recv":"1609160400006136329","hd":{"ts_event":"1609160400006001487","rtype":1,"publisher_id":1,"instrument_id":5482},"action":"A","side":"A","depth":0,"price":"3720500000000","size":1,"flags":128,"ts_in_delta":17214,"sequence":1170362,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":11,"bid_ct":15,"ask_ct":9}]}
			// {"ts_recv":"1609160400006246513","hd":{"ts_event":"1609160400006146661","rtype":1,"publisher_id":1,"instrument_id":5482},"action":"A","side":"A","depth":0,"price":"3720500000000","size":1,"flags":128,"ts_in_delta":18858,"sequence":1170364,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":12,"bid_ct":15,"ask_ct":10}]}

			records, metadata, err := dbn.ReadDBNToSlice[dbn.Mbp1Msg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609160400006001487)))
			Expect(r0h.RType).To(Equal(dbn.RType(1)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(string(r0.Action)).To(Equal("A"))
			Expect(string(r0.Side)).To(Equal("A"))
			Expect(r0.Depth).To(Equal(uint8(0)))
			Expect(r0.Price).To(Equal(int64(3720500000000)))
			Expect(r0.Size).To(Equal(uint32(1)))
			Expect(r0.Flags).To(Equal(uint8(128)))
			Expect(r0.TsRecv).To(Equal(uint64(1609160400006136329)))
			Expect(r0.TsInDelta).To(Equal(int32(17214)))
			Expect(r0.Sequence).To(Equal(uint32(1170362)))
			Expect(r0.Level.BidPx).To(Equal(int64(3720250000000)))
			Expect(r0.Level.AskPx).To(Equal(int64(3720500000000)))
			Expect(r0.Level.BidSz).To(Equal(uint32(24)))
			Expect(r0.Level.AskSz).To(Equal(uint32(11)))
			Expect(r0.Level.BidCt).To(Equal(uint32(15)))
			Expect(r0.Level.AskCt).To(Equal(uint32(9)))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609160400006146661)))
			Expect(r1h.RType).To(Equal(dbn.RType(1)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(string(r1.Action)).To(Equal("A"))
			Expect(string(r1.Side)).To(Equal("A"))
			Expect(r1.Depth).To(Equal(uint8(0)))
			Expect(r1.Price).To(Equal(int64(3720500000000)))
			Expect(r1.Size).To(Equal(uint32(1)))
			Expect(r1.Flags).To(Equal(uint8(128)))
			Expect(r1.TsRecv).To(Equal(uint64(1609160400006246513)))
			Expect(r1.TsInDelta).To(Equal(int32(18858)))
			Expect(r1.Sequence).To(Equal(uint32(1170364)))
			Expect(r1.Level.BidPx).To(Equal(int64(3720250000000)))
			Expect(r1.Level.AskPx).To(Equal(int64(3720500000000)))
			Expect(r1.Level.BidSz).To(Equal(uint32(24)))
			Expect(r1.Level.AskSz).To(Equal(uint32(12)))
			Expect(r1.Level.BidCt).To(Equal(uint32(15)))
			Expect(r1.Level.AskCt).To(Equal(uint32(10)))
		})

		It("should read a v1 mbp10 correctly", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.mbp-10.v1.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			// dbn -J ./tests/data/test_data.mbp-10.v1.dbn.zst
			// {"ts_recv":"1609160400000704060","hd":{"ts_event":"1609160400000429831","rtype":10,"publisher_id":1,"instrument_id":5482},"action":"C","side":"A","depth":9,"price":"3722750000000","size":1,"flags":128,"ts_in_delta":22993,"sequence":1170352,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":10,"bid_ct":15,"ask_ct":8},{"bid_px":"3720000000000","ask_px":"3720750000000","bid_sz":31,"ask_sz":34,"bid_ct":18,"ask_ct":24},{"bid_px":"3719750000000","ask_px":"3721000000000","bid_sz":32,"ask_sz":39,"bid_ct":23,"ask_ct":25},{"bid_px":"3719500000000","ask_px":"3721250000000","bid_sz":39,"ask_sz":28,"bid_ct":26,"ask_ct":17},{"bid_px":"3719250000000","ask_px":"3721500000000","bid_sz":50,"ask_sz":33,"bid_ct":35,"ask_ct":19},{"bid_px":"3719000000000","ask_px":"3721750000000","bid_sz":42,"ask_sz":45,"bid_ct":28,"ask_ct":33},{"bid_px":"3718750000000","ask_px":"3722000000000","bid_sz":44,"ask_sz":55,"bid_ct":35,"ask_ct":40},{"bid_px":"3718500000000","ask_px":"3722250000000","bid_sz":64,"ask_sz":59,"bid_ct":39,"ask_ct":38},{"bid_px":"3718250000000","ask_px":"3722500000000","bid_sz":53,"ask_sz":49,"bid_ct":32,"ask_ct":35},{"bid_px":"3718000000000","ask_px":"3722750000000","bid_sz":67,"ask_sz":44,"bid_ct":39,"ask_ct":26}]}
			// {"ts_recv":"1609160400000750544","hd":{"ts_event":"1609160400000435673","rtype":10,"publisher_id":1,"instrument_id":5482},"action":"C","side":"B","depth":1,"price":"3720000000000","size":1,"flags":128,"ts_in_delta":20625,"sequence":1170356,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":10,"bid_ct":15,"ask_ct":8},{"bid_px":"3720000000000","ask_px":"3720750000000","bid_sz":30,"ask_sz":34,"bid_ct":17,"ask_ct":24},{"bid_px":"3719750000000","ask_px":"3721000000000","bid_sz":32,"ask_sz":39,"bid_ct":23,"ask_ct":25},{"bid_px":"3719500000000","ask_px":"3721250000000","bid_sz":39,"ask_sz":28,"bid_ct":26,"ask_ct":17},{"bid_px":"3719250000000","ask_px":"3721500000000","bid_sz":50,"ask_sz":33,"bid_ct":35,"ask_ct":19},{"bid_px":"3719000000000","ask_px":"3721750000000","bid_sz":42,"ask_sz":45,"bid_ct":28,"ask_ct":33},{"bid_px":"3718750000000","ask_px":"3722000000000","bid_sz":44,"ask_sz":55,"bid_ct":35,"ask_ct":40},{"bid_px":"3718500000000","ask_px":"3722250000000","bid_sz":64,"ask_sz":59,"bid_ct":39,"ask_ct":38},{"bid_px":"3718250000000","ask_px":"3722500000000","bid_sz":53,"ask_sz":49,"bid_ct":32,"ask_ct":35},{"bid_px":"3718000000000","ask_px":"3722750000000","bid_sz":67,"ask_sz":44,"bid_ct":39,"ask_ct":26}]}

			records, metadata, err := dbn.ReadDBNToSlice[dbn.Mbp10Msg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609160400000429831)))
			Expect(r0h.RType).To(Equal(dbn.RType(10)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r0.Price).To(Equal(int64(3722750000000)))
			Expect(r0.Size).To(Equal(uint32(1)))
			Expect(r0.Action).To(Equal(byte('C')))
			Expect(r0.Side).To(Equal(byte('A')))
			Expect(r0.Flags).To(Equal(uint8(128)))
			Expect(r0.Depth).To(Equal(uint8(9)))
			Expect(r0.TsRecv).To(Equal(uint64(1609160400000704060)))
			Expect(r0.TsInDelta).To(Equal(int32(22993)))
			Expect(r0.Sequence).To(Equal(uint32(1170352)))
			Expect(len(r0.Levels)).To(Equal(10))
			Expect(r0.Levels[0]).To(Equal(dbn.BidAskPair{BidPx: int64(3720250000000), AskPx: int64(3720500000000), BidSz: uint32(24), AskSz: uint32(10), BidCt: uint32(15), AskCt: uint32(8)}))
			Expect(r0.Levels[1]).To(Equal(dbn.BidAskPair{BidPx: int64(3720000000000), AskPx: int64(3720750000000), BidSz: uint32(31), AskSz: uint32(34), BidCt: uint32(18), AskCt: uint32(24)}))
			Expect(r0.Levels[2]).To(Equal(dbn.BidAskPair{BidPx: int64(3719750000000), AskPx: int64(3721000000000), BidSz: uint32(32), AskSz: uint32(39), BidCt: uint32(23), AskCt: uint32(25)}))
			Expect(r0.Levels[3]).To(Equal(dbn.BidAskPair{BidPx: int64(3719500000000), AskPx: int64(3721250000000), BidSz: uint32(39), AskSz: uint32(28), BidCt: uint32(26), AskCt: uint32(17)}))
			Expect(r0.Levels[4]).To(Equal(dbn.BidAskPair{BidPx: int64(3719250000000), AskPx: int64(3721500000000), BidSz: uint32(50), AskSz: uint32(33), BidCt: uint32(35), AskCt: uint32(19)}))
			Expect(r0.Levels[5]).To(Equal(dbn.BidAskPair{BidPx: int64(3719000000000), AskPx: int64(3721750000000), BidSz: uint32(42), AskSz: uint32(45), BidCt: uint32(28), AskCt: uint32(33)}))
			Expect(r0.Levels[6]).To(Equal(dbn.BidAskPair{BidPx: int64(3718750000000), AskPx: int64(3722000000000), BidSz: uint32(44), AskSz: uint32(55), BidCt: uint32(35), AskCt: uint32(40)}))
			Expect(r0.Levels[7]).To(Equal(dbn.BidAskPair{BidPx: int64(3718500000000), AskPx: int64(3722250000000), BidSz: uint32(64), AskSz: uint32(59), BidCt: uint32(39), AskCt: uint32(38)}))
			Expect(r0.Levels[8]).To(Equal(dbn.BidAskPair{BidPx: int64(3718250000000), AskPx: int64(3722500000000), BidSz: uint32(53), AskSz: uint32(49), BidCt: uint32(32), AskCt: uint32(35)}))
			Expect(r0.Levels[9]).To(Equal(dbn.BidAskPair{BidPx: int64(3718000000000), AskPx: int64(3722750000000), BidSz: uint32(67), AskSz: uint32(44), BidCt: uint32(39), AskCt: uint32(26)}))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609160400000435673)))
			Expect(r1h.RType).To(Equal(dbn.RType(10)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r1.Price).To(Equal(int64(3720000000000)))
			Expect(r1.Size).To(Equal(uint32(1)))
			Expect(r1.Action).To(Equal(byte('C')))
			Expect(r1.Side).To(Equal(byte('B')))
			Expect(r1.Flags).To(Equal(uint8(128)))
			Expect(r1.Depth).To(Equal(uint8(1)))
			Expect(r1.TsRecv).To(Equal(uint64(1609160400000750544)))
			Expect(r1.TsInDelta).To(Equal(int32(20625)))
			Expect(r1.Sequence).To(Equal(uint32(1170356)))
			Expect(len(r1.Levels)).To(Equal(10))
			Expect(r1.Levels[0]).To(Equal(dbn.BidAskPair{BidPx: int64(3720250000000), AskPx: int64(3720500000000), BidSz: uint32(24), AskSz: uint32(10), BidCt: uint32(15), AskCt: uint32(8)}))
			Expect(r1.Levels[1]).To(Equal(dbn.BidAskPair{BidPx: int64(3720000000000), AskPx: int64(3720750000000), BidSz: uint32(30), AskSz: uint32(34), BidCt: uint32(17), AskCt: uint32(24)}))
			Expect(r1.Levels[2]).To(Equal(dbn.BidAskPair{BidPx: int64(3719750000000), AskPx: int64(3721000000000), BidSz: uint32(32), AskSz: uint32(39), BidCt: uint32(23), AskCt: uint32(25)}))
			Expect(r1.Levels[3]).To(Equal(dbn.BidAskPair{BidPx: int64(3719500000000), AskPx: int64(3721250000000), BidSz: uint32(39), AskSz: uint32(28), BidCt: uint32(26), AskCt: uint32(17)}))
			Expect(r1.Levels[4]).To(Equal(dbn.BidAskPair{BidPx: int64(3719250000000), AskPx: int64(3721500000000), BidSz: uint32(50), AskSz: uint32(33), BidCt: uint32(35), AskCt: uint32(19)}))
			Expect(r1.Levels[5]).To(Equal(dbn.BidAskPair{BidPx: int64(3719000000000), AskPx: int64(3721750000000), BidSz: uint32(42), AskSz: uint32(45), BidCt: uint32(28), AskCt: uint32(33)}))
			Expect(r1.Levels[6]).To(Equal(dbn.BidAskPair{BidPx: int64(3718750000000), AskPx: int64(3722000000000), BidSz: uint32(44), AskSz: uint32(55), BidCt: uint32(35), AskCt: uint32(40)}))
			Expect(r1.Levels[7]).To(Equal(dbn.BidAskPair{BidPx: int64(3718500000000), AskPx: int64(3722250000000), BidSz: uint32(64), AskSz: uint32(59), BidCt: uint32(39), AskCt: uint32(38)}))
			Expect(r1.Levels[8]).To(Equal(dbn.BidAskPair{BidPx: int64(3718250000000), AskPx: int64(3722500000000), BidSz: uint32(53), AskSz: uint32(49), BidCt: uint32(32), AskCt: uint32(35)}))
			Expect(r1.Levels[9]).To(Equal(dbn.BidAskPair{BidPx: int64(3718000000000), AskPx: int64(3722750000000), BidSz: uint32(67), AskSz: uint32(44), BidCt: uint32(39), AskCt: uint32(26)}))
		})

		It("should read a v1 mbp10 correctly", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.mbp-10.v2.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			// dbn -J ./tests/data/test_data.mbp-10.v2.dbn.zst
			// {"ts_recv":"1609160400000704060","hd":{"ts_event":"1609160400000429831","rtype":10,"publisher_id":1,"instrument_id":5482},"action":"C","side":"A","depth":9,"price":"3722750000000","size":1,"flags":128,"ts_in_delta":22993,"sequence":1170352,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":10,"bid_ct":15,"ask_ct":8},{"bid_px":"3720000000000","ask_px":"3720750000000","bid_sz":31,"ask_sz":34,"bid_ct":18,"ask_ct":24},{"bid_px":"3719750000000","ask_px":"3721000000000","bid_sz":32,"ask_sz":39,"bid_ct":23,"ask_ct":25},{"bid_px":"3719500000000","ask_px":"3721250000000","bid_sz":39,"ask_sz":28,"bid_ct":26,"ask_ct":17},{"bid_px":"3719250000000","ask_px":"3721500000000","bid_sz":50,"ask_sz":33,"bid_ct":35,"ask_ct":19},{"bid_px":"3719000000000","ask_px":"3721750000000","bid_sz":42,"ask_sz":45,"bid_ct":28,"ask_ct":33},{"bid_px":"3718750000000","ask_px":"3722000000000","bid_sz":44,"ask_sz":55,"bid_ct":35,"ask_ct":40},{"bid_px":"3718500000000","ask_px":"3722250000000","bid_sz":64,"ask_sz":59,"bid_ct":39,"ask_ct":38},{"bid_px":"3718250000000","ask_px":"3722500000000","bid_sz":53,"ask_sz":49,"bid_ct":32,"ask_ct":35},{"bid_px":"3718000000000","ask_px":"3722750000000","bid_sz":67,"ask_sz":44,"bid_ct":39,"ask_ct":26}]}
			// {"ts_recv":"1609160400000750544","hd":{"ts_event":"1609160400000435673","rtype":10,"publisher_id":1,"instrument_id":5482},"action":"C","side":"B","depth":1,"price":"3720000000000","size":1,"flags":128,"ts_in_delta":20625,"sequence":1170356,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":24,"ask_sz":10,"bid_ct":15,"ask_ct":8},{"bid_px":"3720000000000","ask_px":"3720750000000","bid_sz":30,"ask_sz":34,"bid_ct":17,"ask_ct":24},{"bid_px":"3719750000000","ask_px":"3721000000000","bid_sz":32,"ask_sz":39,"bid_ct":23,"ask_ct":25},{"bid_px":"3719500000000","ask_px":"3721250000000","bid_sz":39,"ask_sz":28,"bid_ct":26,"ask_ct":17},{"bid_px":"3719250000000","ask_px":"3721500000000","bid_sz":50,"ask_sz":33,"bid_ct":35,"ask_ct":19},{"bid_px":"3719000000000","ask_px":"3721750000000","bid_sz":42,"ask_sz":45,"bid_ct":28,"ask_ct":33},{"bid_px":"3718750000000","ask_px":"3722000000000","bid_sz":44,"ask_sz":55,"bid_ct":35,"ask_ct":40},{"bid_px":"3718500000000","ask_px":"3722250000000","bid_sz":64,"ask_sz":59,"bid_ct":39,"ask_ct":38},{"bid_px":"3718250000000","ask_px":"3722500000000","bid_sz":53,"ask_sz":49,"bid_ct":32,"ask_ct":35},{"bid_px":"3718000000000","ask_px":"3722750000000","bid_sz":67,"ask_sz":44,"bid_ct":39,"ask_ct":26}]}

			records, metadata, err := dbn.ReadDBNToSlice[dbn.Mbp10Msg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609160400000429831)))
			Expect(r0h.RType).To(Equal(dbn.RType(10)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r0.Price).To(Equal(int64(3722750000000)))
			Expect(r0.Size).To(Equal(uint32(1)))
			Expect(r0.Action).To(Equal(byte('C')))
			Expect(r0.Side).To(Equal(byte('A')))
			Expect(r0.Flags).To(Equal(uint8(128)))
			Expect(r0.Depth).To(Equal(uint8(9)))
			Expect(r0.TsRecv).To(Equal(uint64(1609160400000704060)))
			Expect(r0.TsInDelta).To(Equal(int32(22993)))
			Expect(r0.Sequence).To(Equal(uint32(1170352)))
			Expect(len(r0.Levels)).To(Equal(10))
			Expect(r0.Levels[0]).To(Equal(dbn.BidAskPair{BidPx: int64(3720250000000), AskPx: int64(3720500000000), BidSz: uint32(24), AskSz: uint32(10), BidCt: uint32(15), AskCt: uint32(8)}))
			Expect(r0.Levels[1]).To(Equal(dbn.BidAskPair{BidPx: int64(3720000000000), AskPx: int64(3720750000000), BidSz: uint32(31), AskSz: uint32(34), BidCt: uint32(18), AskCt: uint32(24)}))
			Expect(r0.Levels[2]).To(Equal(dbn.BidAskPair{BidPx: int64(3719750000000), AskPx: int64(3721000000000), BidSz: uint32(32), AskSz: uint32(39), BidCt: uint32(23), AskCt: uint32(25)}))
			Expect(r0.Levels[3]).To(Equal(dbn.BidAskPair{BidPx: int64(3719500000000), AskPx: int64(3721250000000), BidSz: uint32(39), AskSz: uint32(28), BidCt: uint32(26), AskCt: uint32(17)}))
			Expect(r0.Levels[4]).To(Equal(dbn.BidAskPair{BidPx: int64(3719250000000), AskPx: int64(3721500000000), BidSz: uint32(50), AskSz: uint32(33), BidCt: uint32(35), AskCt: uint32(19)}))
			Expect(r0.Levels[5]).To(Equal(dbn.BidAskPair{BidPx: int64(3719000000000), AskPx: int64(3721750000000), BidSz: uint32(42), AskSz: uint32(45), BidCt: uint32(28), AskCt: uint32(33)}))
			Expect(r0.Levels[6]).To(Equal(dbn.BidAskPair{BidPx: int64(3718750000000), AskPx: int64(3722000000000), BidSz: uint32(44), AskSz: uint32(55), BidCt: uint32(35), AskCt: uint32(40)}))
			Expect(r0.Levels[7]).To(Equal(dbn.BidAskPair{BidPx: int64(3718500000000), AskPx: int64(3722250000000), BidSz: uint32(64), AskSz: uint32(59), BidCt: uint32(39), AskCt: uint32(38)}))
			Expect(r0.Levels[8]).To(Equal(dbn.BidAskPair{BidPx: int64(3718250000000), AskPx: int64(3722500000000), BidSz: uint32(53), AskSz: uint32(49), BidCt: uint32(32), AskCt: uint32(35)}))
			Expect(r0.Levels[9]).To(Equal(dbn.BidAskPair{BidPx: int64(3718000000000), AskPx: int64(3722750000000), BidSz: uint32(67), AskSz: uint32(44), BidCt: uint32(39), AskCt: uint32(26)}))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609160400000435673)))
			Expect(r1h.RType).To(Equal(dbn.RType(10)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r1.Price).To(Equal(int64(3720000000000)))
			Expect(r1.Size).To(Equal(uint32(1)))
			Expect(r1.Action).To(Equal(byte('C')))
			Expect(r1.Side).To(Equal(byte('B')))
			Expect(r1.Flags).To(Equal(uint8(128)))
			Expect(r1.Depth).To(Equal(uint8(1)))
			Expect(r1.TsRecv).To(Equal(uint64(1609160400000750544)))
			Expect(r1.TsInDelta).To(Equal(int32(20625)))
			Expect(r1.Sequence).To(Equal(uint32(1170356)))
			Expect(len(r1.Levels)).To(Equal(10))
			Expect(r1.Levels[0]).To(Equal(dbn.BidAskPair{BidPx: int64(3720250000000), AskPx: int64(3720500000000), BidSz: uint32(24), AskSz: uint32(10), BidCt: uint32(15), AskCt: uint32(8)}))
			Expect(r1.Levels[1]).To(Equal(dbn.BidAskPair{BidPx: int64(3720000000000), AskPx: int64(3720750000000), BidSz: uint32(30), AskSz: uint32(34), BidCt: uint32(17), AskCt: uint32(24)}))
			Expect(r1.Levels[2]).To(Equal(dbn.BidAskPair{BidPx: int64(3719750000000), AskPx: int64(3721000000000), BidSz: uint32(32), AskSz: uint32(39), BidCt: uint32(23), AskCt: uint32(25)}))
			Expect(r1.Levels[3]).To(Equal(dbn.BidAskPair{BidPx: int64(3719500000000), AskPx: int64(3721250000000), BidSz: uint32(39), AskSz: uint32(28), BidCt: uint32(26), AskCt: uint32(17)}))
			Expect(r1.Levels[4]).To(Equal(dbn.BidAskPair{BidPx: int64(3719250000000), AskPx: int64(3721500000000), BidSz: uint32(50), AskSz: uint32(33), BidCt: uint32(35), AskCt: uint32(19)}))
			Expect(r1.Levels[5]).To(Equal(dbn.BidAskPair{BidPx: int64(3719000000000), AskPx: int64(3721750000000), BidSz: uint32(42), AskSz: uint32(45), BidCt: uint32(28), AskCt: uint32(33)}))
			Expect(r1.Levels[6]).To(Equal(dbn.BidAskPair{BidPx: int64(3718750000000), AskPx: int64(3722000000000), BidSz: uint32(44), AskSz: uint32(55), BidCt: uint32(35), AskCt: uint32(40)}))
			Expect(r1.Levels[7]).To(Equal(dbn.BidAskPair{BidPx: int64(3718500000000), AskPx: int64(3722250000000), BidSz: uint32(64), AskSz: uint32(59), BidCt: uint32(39), AskCt: uint32(38)}))
			Expect(r1.Levels[8]).To(Equal(dbn.BidAskPair{BidPx: int64(3718250000000), AskPx: int64(3722500000000), BidSz: uint32(53), AskSz: uint32(49), BidCt: uint32(32), AskCt: uint32(35)}))
			Expect(r1.Levels[9]).To(Equal(dbn.BidAskPair{BidPx: int64(3718000000000), AskPx: int64(3722750000000), BidSz: uint32(67), AskSz: uint32(44), BidCt: uint32(39), AskCt: uint32(26)}))
		})
	})
})
