// Copyright (c) 2024 Neomantra Corp

package dbn_test

import (
	"os"
	"strings"
	"unsafe"

	dbn "github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Struct", func() {
	Context("correctness", func() {
		It("struct consts should be correct", func() {
			Expect(unsafe.Sizeof(dbn.RHeader{})).To(Equal(uintptr(dbn.RHeader_Size)))
			Expect(unsafe.Sizeof(dbn.BidAskPair{})).To(Equal(uintptr(dbn.BidAskPair_Size)))
			Expect(unsafe.Sizeof(dbn.Mbp0Msg{})).To(Equal(uintptr(dbn.Mbp0Msg_Size)))
			Expect(unsafe.Sizeof(dbn.Mbp1Msg{})).To(Equal(uintptr(dbn.Mbp1Msg_Size)))
			Expect(unsafe.Sizeof(dbn.Mbp10Msg{})).To(Equal(uintptr(dbn.Mbp10Msg_Size)))
			Expect(unsafe.Sizeof(dbn.Cmbp1Msg{})).To(Equal(uintptr(dbn.Cmbp1Msg_Size)))
			Expect(unsafe.Sizeof(dbn.OhlcvMsg{})).To(Equal(uintptr(dbn.OhlcvMsg_Size)))
			Expect(unsafe.Sizeof(dbn.ImbalanceMsg{})).To(Equal(uintptr(dbn.ImbalanceMsg_Size)))
			Expect(unsafe.Sizeof(dbn.ErrorMsg{})).To(Equal(uintptr(dbn.ErrorMsg_Size)))
			Expect(unsafe.Sizeof(dbn.SystemMsg{})).To(Equal(uintptr(dbn.SystemMsg_Size)))
			Expect(unsafe.Sizeof(dbn.StatMsg{})).To(Equal(uintptr(dbn.StatMsg_Size)))
			Expect(unsafe.Sizeof(dbn.StatusMsg{})).To(Equal(uintptr(dbn.StatusMsg_Size)))
			Expect(unsafe.Sizeof(dbn.InstrumentDefMsg{})).To(Equal(uintptr(dbn.InstrumentDefMsg_Size)))
			Expect(int((&dbn.RHeader{}).RSize())).To(Equal(dbn.RHeader_Size))
			Expect(int((&dbn.Mbp0Msg{}).RSize())).To(Equal(dbn.Mbp0Msg_Size))
			Expect(int((&dbn.Mbp1Msg{}).RSize())).To(Equal(dbn.Mbp1Msg_Size))
			Expect(int((&dbn.Mbp10Msg{}).RSize())).To(Equal(dbn.Mbp10Msg_Size))
			Expect(int((&dbn.Cmbp1Msg{}).RSize())).To(Equal(dbn.Cmbp1Msg_Size))
			Expect(int((&dbn.OhlcvMsg{}).RSize())).To(Equal(dbn.OhlcvMsg_Size))
			Expect(int((&dbn.ImbalanceMsg{}).RSize())).To(Equal(dbn.ImbalanceMsg_Size))
			Expect(int((&dbn.ErrorMsg{}).RSize())).To(Equal(dbn.ErrorMsg_Size))
			Expect(int((&dbn.StatMsg{}).RSize())).To(Equal(dbn.StatMsg_Size))
			Expect(int((&dbn.StatusMsg{}).RSize())).To(Equal(dbn.StatusMsg_Size))
			Expect(int((&dbn.InstrumentDefMsg{}).RSize())).To(Equal(dbn.InstrumentDefMsg_Size))
		})
	})

	Context("v1 file testing", func() {
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
		// TODO: we do not currently support v1
		// It("should read v1 definition correctly", func() {
		// 	file, err := os.Open("./tests/data/test_data.definition.v1.dbn.zst")
		// 	Expect(err).To(BeNil())
		// 	defer file.Close()
		//
		// 	records, metadata, err := dbn.ReadDBNToSlice[dbn.InstrumentDefMsg](file)
		// 	Expect(err).To(BeNil())
		// 	Expect(metadata).ToNot(BeNil())
		// 	Expect(len(records)).To(Equal(2))
		// })
	})

	Context("v2 files", func() {
		It("should read a v2 ohlc-1s correctly", func() {
			file, err := os.Open("./tests/data/test_data.ohlcv-1s.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.OhlcvMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
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
		It("should read a v2 trades/mbp0 correctly", func() {
			file, err := os.Open("./tests/data/test_data.trades.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.Mbp0Msg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.trades.dbn
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
		It("should read a v2 mbp1 correctly", func() {
			file, err := os.Open("./tests/data/test_data.mbp-1.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			// dbn -J ./tests/data/test_data.mbp-1.dbn
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
		It("should read a v2 imbalance correctly", func() {
			file, err := os.Open("./tests/data/test_data.imbalance.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.ImbalanceMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.imbalance.dbn
			// {"ts_recv":"1633353900633864350","hd":{"ts_event":"1633353900633854579","rtype":20,"publisher_id":2,"instrument_id":9439},"ref_price":"229430000000","auction_time":"0","cont_book_clr_price":"0","auct_interest_clr_price":"0","ssr_filling_price":"0","ind_match_price":"0","upper_collar":"0","lower_collar":"0","paired_qty":0,"total_imbalance_qty":2000,"market_imbalance_qty":0,"unpaired_qty":0,"auction_type":"O","side":"B","auction_status":0,"freeze_status":0,"num_extensions":0,"unpaired_side":"N","significant_imbalance":"~"}
			// {"ts_recv":"1633353910208124734","hd":{"ts_event":"1633353910208114778","rtype":20,"publisher_id":2,"instrument_id":9439},"ref_price":"229990000000","auction_time":"0","cont_book_clr_price":"0","auct_interest_clr_price":"0","ssr_filling_price":"0","ind_match_price":"0","upper_collar":"0","lower_collar":"0","paired_qty":1719,"total_imbalance_qty":281,"market_imbalance_qty":0,"unpaired_qty":0,"auction_type":"O","side":"B","auction_status":0,"freeze_status":0,"num_extensions":0,"unpaired_side":"N","significant_imbalance":"~"}

			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1633353900633854579)))
			Expect(r0h.RType).To(Equal(dbn.RType(20)))
			Expect(r0h.PublisherID).To(Equal(uint16(2)))
			Expect(r0h.InstrumentID).To(Equal(uint32(9439)))
			Expect(r0.TsRecv).To(Equal(uint64(1633353900633864350)))
			Expect(r0.RefPrice).To(Equal(int64(229430000000)))
			Expect(r0.AuctionTime).To(Equal(uint64(0)))
			Expect(r0.ContBookClrPrice).To(Equal(int64(0)))
			Expect(r0.AuctInterestClrPrice).To(Equal(int64(0)))
			Expect(r0.SsrFillingPrice).To(Equal(int64(0)))
			Expect(r0.IndMatchPrice).To(Equal(int64(0)))
			Expect(r0.UpperCollar).To(Equal(int64(0)))
			Expect(r0.LowerCollar).To(Equal(int64(0)))
			Expect(r0.PairedQty).To(Equal(uint32(0)))
			Expect(r0.TotalImbalanceQty).To(Equal(uint32(2000)))
			Expect(r0.MarketImbalanceQty).To(Equal(uint32(0)))
			Expect(r0.UnpairedQty).To(Equal(int32(0)))
			Expect(string(r0.AuctionType)).To(Equal("O"))
			Expect(string(r0.Side)).To(Equal("B"))
			Expect(r0.AuctionStatus).To(Equal(uint8(0)))
			Expect(r0.FreezeStatus).To(Equal(uint8(0)))
			Expect(r0.NumExtensions).To(Equal(uint8(0)))
			Expect(string(r0.UnpairedSide)).To(Equal("N"))
			Expect(string(r0.SignificantImbalance)).To(Equal("~"))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1633353910208114778)))
			Expect(r1h.RType).To(Equal(dbn.RType(20)))
			Expect(r1h.PublisherID).To(Equal(uint16(2)))
			Expect(r1h.InstrumentID).To(Equal(uint32(9439)))
			Expect(r1.TsRecv).To(Equal(uint64(1633353910208124734)))
			Expect(r1.RefPrice).To(Equal(int64(229990000000)))
			Expect(r1.AuctionTime).To(Equal(uint64(0)))
			Expect(r1.ContBookClrPrice).To(Equal(int64(0)))
			Expect(r1.AuctInterestClrPrice).To(Equal(int64(0)))
			Expect(r1.SsrFillingPrice).To(Equal(int64(0)))
			Expect(r1.IndMatchPrice).To(Equal(int64(0)))
			Expect(r1.UpperCollar).To(Equal(int64(0)))
			Expect(r1.LowerCollar).To(Equal(int64(0)))
			Expect(r1.PairedQty).To(Equal(uint32(1719)))
			Expect(r1.TotalImbalanceQty).To(Equal(uint32(281)))
			Expect(r1.MarketImbalanceQty).To(Equal(uint32(0)))
			Expect(r1.UnpairedQty).To(Equal(int32(0)))
			Expect(string(r1.AuctionType)).To(Equal("O"))
			Expect(string(r1.Side)).To(Equal("B"))
			Expect(r1.AuctionStatus).To(Equal(uint8(0)))
			Expect(r1.FreezeStatus).To(Equal(uint8(0)))
			Expect(r1.NumExtensions).To(Equal(uint8(0)))
			Expect(string(r1.UnpairedSide)).To(Equal("N"))
			Expect(string(r1.SignificantImbalance)).To(Equal("~"))
		})
		It("should read v2 definition correctly", func() {
			file, err := os.Open("./tests/data/test_data.definition.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.InstrumentDefMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.definition.dbn
			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1633331241618018154)))
			Expect(r0h.RType).To(Equal(dbn.RType(19)))
			Expect(r0h.PublisherID).To(Equal(uint16(2)))
			Expect(r0h.InstrumentID).To(Equal(uint32(6819)))
			Expect(r0.TsRecv).To(Equal(uint64(1633331241618029519)))
			Expect(r0.MinPriceIncrement).To(Equal(int64(9223372036854775807)))
			Expect(r0.DisplayFactor).To(Equal(int64(100000000000000)))
			Expect(r0.Expiration).To(Equal(uint64(18446744073709551615)))
			Expect(r0.Activation).To(Equal(uint64(18446744073709551615)))
			Expect(r0.HighLimitPrice).To(Equal(int64(9223372036854775807)))
			Expect(r0.LowLimitPrice).To(Equal(int64(9223372036854775807)))
			Expect(r0.MaxPriceVariation).To(Equal(int64(9223372036854775807)))
			Expect(r0.TradingReferencePrice).To(Equal(int64(9223372036854775807)))
			Expect(r0.UnitOfMeasureQty).To(Equal(int64(9223372036854775807)))
			Expect(r0.MinPriceIncrementAmount).To(Equal(int64(9223372036854775807)))
			Expect(r0.PriceRatio).To(Equal(int64(9223372036854775807)))
			Expect(r0.StrikePrice).To(Equal(int64(9223372036854775807)))
			Expect(r0.InstAttribValue).To(Equal(int32(2147483647)))
			Expect(r0.UnderlyingID).To(Equal(uint32(0)))
			Expect(r0.RawInstrumentID).To(Equal(uint32(2147483647)))
			Expect(r0.MarketDepthImplied).To(Equal(int32(2147483647)))
			Expect(r0.MarketDepth).To(Equal(int32(2147483647)))
			Expect(r0.MarketSegmentID).To(Equal(uint32(4294967295)))
			Expect(r0.MaxTradeVol).To(Equal(uint32(4294967295)))
			Expect(r0.MinLotSize).To(Equal(int32(2147483647)))
			Expect(r0.MinLotSizeBlock).To(Equal(int32(2147483647)))
			Expect(r0.MinLotSizeRoundLot).To(Equal(int32(100)))
			Expect(r0.MinTradeVol).To(Equal(uint32(4294967295)))
			Expect(r0.ContractMultiplier).To(Equal(int32(2147483647)))
			Expect(r0.DecayQuantity).To(Equal(int32(2147483647)))
			Expect(r0.OriginalContractSize).To(Equal(int32(2147483647)))
			Expect(r0.TradingReferenceDate).To(Equal(uint16(65535)))
			Expect(r0.ApplID).To(Equal(int16(32767)))
			Expect(r0.MaturityYear).To(Equal(uint16(65535)))
			Expect(r0.DecayStartDate).To(Equal(uint16(65535)))
			Expect(r0.ChannelID).To(Equal(uint16(0)))
			Expect(string(r0.Currency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(string(r0.SettlCurrency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(string(r0.Secsubtype[:])).To(Equal("Z " + strings.Repeat("\x00", 4)))
			Expect(string(r0.RawSymbol[:])).To(Equal("MSFT" + strings.Repeat("\x00", dbn.MetadataV2_SymbolCstrLen-len("MSFT"))))
			Expect(string(r0.Group[:])).To(Equal("pxnas-1" + strings.Repeat("\x00", 21-len("pxnas-1"))))
			Expect(string(r0.Exchange[:])).To(Equal("XNAS" + strings.Repeat("\x00", 5-len("XNAS"))))
			Expect(string(r0.Asset[:])).To(Equal(strings.Repeat("\x00", 7)))
			Expect(string(r0.Cfi[:])).To(Equal(strings.Repeat("\x00", 7)))
			Expect(string(r0.SecurityType[:])).To(Equal(strings.Repeat("\x00", 7)))
			Expect(string(r0.UnitOfMeasure[:])).To(Equal(strings.Repeat("\x00", 31)))
			Expect(string(r0.Underlying[:])).To(Equal(strings.Repeat("\x00", 21)))
			Expect(string(r0.StrikePriceCurrency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(r0.InstrumentClass).To(Equal(uint8('K')))
			Expect(r0.MatchAlgorithm).To(Equal(uint8('F')))
			Expect(r0.MdSecurityTradingStatus).To(Equal(uint8(78)))
			Expect(r0.MainFraction).To(Equal(uint8(255)))
			Expect(r0.PriceDisplayFormat).To(Equal(uint8(255)))
			Expect(r0.SettlPrice_type).To(Equal(uint8(255)))
			Expect(r0.SubFraction).To(Equal(uint8(255)))
			Expect(r0.UnderlyingProduct).To(Equal(uint8(255)))
			Expect(r0.SecurityUpdateAction).To(Equal(uint8('A')))
			Expect(r0.MaturityMonth).To(Equal(uint8(255)))
			Expect(r0.MaturityDay).To(Equal(uint8(255)))
			Expect(r0.MaturityWeek).To(Equal(uint8(255)))
			Expect(r0.UserDefinedInstrument).To(Equal(dbn.UserDefinedInstrument('N')))
			Expect(r0.ContractMultiplierUnit).To(Equal(int8(127)))
			Expect(r0.FlowScheduleType).To(Equal(int8(127)))
			Expect(r0.TickRule).To(Equal(uint8(255)))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1633417621703109854)))
			Expect(r1h.RType).To(Equal(dbn.RType(19)))
			Expect(r1h.PublisherID).To(Equal(uint16(2)))
			Expect(r1h.InstrumentID).To(Equal(uint32(6830)))
			Expect(r1.TsRecv).To(Equal(uint64(1633417621703120931)))
			Expect(r1.MinPriceIncrement).To(Equal(int64(9223372036854775807)))
			Expect(r1.DisplayFactor).To(Equal(int64(100000000000000)))
			Expect(r1.Expiration).To(Equal(uint64(18446744073709551615)))
			Expect(r1.Activation).To(Equal(uint64(18446744073709551615)))
			Expect(r1.HighLimitPrice).To(Equal(int64(9223372036854775807)))
			Expect(r1.LowLimitPrice).To(Equal(int64(9223372036854775807)))
			Expect(r1.MaxPriceVariation).To(Equal(int64(9223372036854775807)))
			Expect(r1.TradingReferencePrice).To(Equal(int64(9223372036854775807)))
			Expect(r1.UnitOfMeasureQty).To(Equal(int64(9223372036854775807)))
			Expect(r1.MinPriceIncrementAmount).To(Equal(int64(9223372036854775807)))
			Expect(r1.PriceRatio).To(Equal(int64(9223372036854775807)))
			Expect(r1.StrikePrice).To(Equal(int64(9223372036854775807)))
			Expect(r1.InstAttribValue).To(Equal(int32(2147483647)))
			Expect(r1.UnderlyingID).To(Equal(uint32(0)))
			Expect(r1.RawInstrumentID).To(Equal(uint32(2147483647)))
			Expect(r1.MarketDepthImplied).To(Equal(int32(2147483647)))
			Expect(r1.MarketDepth).To(Equal(int32(2147483647)))
			Expect(r1.MarketSegmentID).To(Equal(uint32(4294967295)))
			Expect(r1.MaxTradeVol).To(Equal(uint32(4294967295)))
			Expect(r1.MinLotSize).To(Equal(int32(2147483647)))
			Expect(r1.MinLotSizeBlock).To(Equal(int32(2147483647)))
			Expect(r1.MinLotSizeRoundLot).To(Equal(int32(100)))
			Expect(r1.MinTradeVol).To(Equal(uint32(4294967295)))
			Expect(r1.ContractMultiplier).To(Equal(int32(2147483647)))
			Expect(r1.DecayQuantity).To(Equal(int32(2147483647)))
			Expect(r1.OriginalContractSize).To(Equal(int32(2147483647)))
			Expect(r1.TradingReferenceDate).To(Equal(uint16(65535)))
			Expect(r1.ApplID).To(Equal(int16(32767)))
			Expect(r1.MaturityYear).To(Equal(uint16(65535)))
			Expect(r1.DecayStartDate).To(Equal(uint16(65535)))
			Expect(r1.ChannelID).To(Equal(uint16(0)))
			Expect(string(r1.Currency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(string(r1.SettlCurrency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(string(r1.Secsubtype[:])).To(Equal("Z " + strings.Repeat("\x00", 4)))
			Expect(string(r1.RawSymbol[:])).To(Equal("MSFT" + strings.Repeat("\x00", dbn.MetadataV2_SymbolCstrLen-len("MSFT"))))
			Expect(string(r1.Group[:])).To(Equal("pxnas-1" + strings.Repeat("\x00", 21-len("pxnas-1"))))
			Expect(string(r1.Exchange[:])).To(Equal("XNAS" + strings.Repeat("\x00", 5-len("XNAS"))))
			Expect(string(r1.Asset[:])).To(Equal(strings.Repeat("\x00", 7)))
			Expect(string(r1.Cfi[:])).To(Equal(strings.Repeat("\x00", 7)))
			Expect(string(r1.SecurityType[:])).To(Equal(strings.Repeat("\x00", 7)))
			Expect(string(r1.UnitOfMeasure[:])).To(Equal(strings.Repeat("\x00", 31)))
			Expect(string(r1.Underlying[:])).To(Equal(strings.Repeat("\x00", 21)))
			Expect(string(r1.StrikePriceCurrency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(r1.InstrumentClass).To(Equal(uint8('K')))
			Expect(r1.MatchAlgorithm).To(Equal(uint8('F')))
			Expect(r1.MdSecurityTradingStatus).To(Equal(uint8(78)))
			Expect(r1.MainFraction).To(Equal(uint8(255)))
			Expect(r1.PriceDisplayFormat).To(Equal(uint8(255)))
			Expect(r1.SettlPrice_type).To(Equal(uint8(255)))
			Expect(r1.SubFraction).To(Equal(uint8(255)))
			Expect(r1.UnderlyingProduct).To(Equal(uint8(255)))
			Expect(r1.SecurityUpdateAction).To(Equal(uint8('A')))
			Expect(r1.MaturityMonth).To(Equal(uint8(255)))
			Expect(r1.MaturityDay).To(Equal(uint8(255)))
			Expect(r1.MaturityWeek).To(Equal(uint8(255)))
			Expect(r1.UserDefinedInstrument).To(Equal(dbn.UserDefinedInstrument('N')))
			Expect(r1.ContractMultiplierUnit).To(Equal(int8(127)))
			Expect(r1.FlowScheduleType).To(Equal(int8(127)))
			Expect(r1.TickRule).To(Equal(uint8(255)))
		})
	})

	Context("StatMsg testing", func() {
		It("should read a v1 statistics correctly", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.statistics.v1.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.StatMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.statistics.v1.dbn.zst
			// {"ts_recv":"1682269536040124325","hd":{"ts_event":"1682269536030443135","rtype":24,"publisher_id":1,"instrument_id":146945},"ts_ref":"18446744073709551615","price":"100000000000","quantity":2147483647,"sequence":2,"ts_in_delta":26961,"stat_type":7,"channel_id":13,"update_action":1,"stat_flags":255}
			// {"ts_recv":"1682269536121890092","hd":{"ts_event":"1682269536071497081","rtype":24,"publisher_id":1,"instrument_id":146945},"ts_ref":"18446744073709551615","price":"100000000000","quantity":2147483647,"sequence":7,"ts_in_delta":28456,"stat_type":5,"channel_id":13,"update_action":1,"stat_flags":255}
			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1682269536030443135)))
			Expect(r0h.RType).To(Equal(dbn.RType(24)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(146945)))
			Expect(r0.TsRecv).To(Equal(uint64(1682269536040124325)))
			Expect(r0.TsRef).To(Equal(uint64(18446744073709551615)))
			Expect(r0.Price).To(Equal(int64(100000000000)))
			Expect(r0.Quantity).To(Equal(int32(2147483647)))
			Expect(r0.Sequence).To(Equal(uint32(2)))
			Expect(r0.TsInDelta).To(Equal(int32(26961)))
			Expect(r0.StatType).To(Equal(uint16(7)))
			Expect(r0.ChannelID).To(Equal(uint16(13)))
			Expect(r0.UpdateAction).To(Equal(uint8(1)))
			Expect(r0.StatFlags).To(Equal(uint8(255)))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1682269536071497081)))
			Expect(r1h.RType).To(Equal(dbn.RType(24)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(146945)))
			Expect(r1.TsRecv).To(Equal(uint64(1682269536121890092)))
			Expect(r1.TsRef).To(Equal(uint64(18446744073709551615)))
			Expect(r1.Price).To(Equal(int64(100000000000)))
			Expect(r1.Quantity).To(Equal(int32(2147483647)))
			Expect(r1.Sequence).To(Equal(uint32(7)))
			Expect(r1.TsInDelta).To(Equal(int32(28456)))
			Expect(r1.StatType).To(Equal(uint16(5)))
			Expect(r1.ChannelID).To(Equal(uint16(13)))
			Expect(r1.UpdateAction).To(Equal(uint8(1)))
			Expect(r1.StatFlags).To(Equal(uint8(255)))
		})

		It("should read a v2 statistics correctly", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.statistics.v2.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.StatMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.statistics.v2.dbn.zst
			// {"ts_recv":"1682269536040124325","hd":{"ts_event":"1682269536030443135","rtype":24,"publisher_id":1,"instrument_id":146945},"ts_ref":"18446744073709551615","price":"100000000000","quantity":2147483647,"sequence":2,"ts_in_delta":26961,"stat_type":7,"channel_id":13,"update_action":1,"stat_flags":255}
			// {"ts_recv":"1682269536121890092","hd":{"ts_event":"1682269536071497081","rtype":24,"publisher_id":1,"instrument_id":146945},"ts_ref":"18446744073709551615","price":"100000000000","quantity":2147483647,"sequence":7,"ts_in_delta":28456,"stat_type":5,"channel_id":13,"update_action":1,"stat_flags":255}
			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1682269536030443135)))
			Expect(r0h.RType).To(Equal(dbn.RType(24)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(146945)))
			Expect(r0.TsRecv).To(Equal(uint64(1682269536040124325)))
			Expect(r0.TsRef).To(Equal(uint64(18446744073709551615)))
			Expect(r0.Price).To(Equal(int64(100000000000)))
			Expect(r0.Quantity).To(Equal(int32(2147483647)))
			Expect(r0.Sequence).To(Equal(uint32(2)))
			Expect(r0.TsInDelta).To(Equal(int32(26961)))
			Expect(r0.StatType).To(Equal(uint16(7)))
			Expect(r0.ChannelID).To(Equal(uint16(13)))
			Expect(r0.UpdateAction).To(Equal(uint8(1)))
			Expect(r0.StatFlags).To(Equal(uint8(255)))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1682269536071497081)))
			Expect(r1h.RType).To(Equal(dbn.RType(24)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(146945)))
			Expect(r1.TsRecv).To(Equal(uint64(1682269536121890092)))
			Expect(r1.TsRef).To(Equal(uint64(18446744073709551615)))
			Expect(r1.Price).To(Equal(int64(100000000000)))
			Expect(r1.Quantity).To(Equal(int32(2147483647)))
			Expect(r1.Sequence).To(Equal(uint32(7)))
			Expect(r1.TsInDelta).To(Equal(int32(28456)))
			Expect(r1.StatType).To(Equal(uint16(5)))
			Expect(r1.ChannelID).To(Equal(uint16(13)))
			Expect(r1.UpdateAction).To(Equal(uint8(1)))
			Expect(r1.StatFlags).To(Equal(uint8(255)))
		})
	})

	Context("BBO message testing", func() {
		It("should read a v1 mbo correctly", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.mbo.v1.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.MboMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.mbo.v1.dbn.zst
			// {"ts_recv":"1609160400000704060","hd":{"ts_event":"1609160400000429831","rtype":160,"publisher_id":1,"instrument_id":5482},"action":"C","side":"A","price":"3722750000000","size":1,"channel_id":0,"order_id":"647784973705","flags":128,"ts_in_delta":22993,"sequence":1170352}
			// {"ts_recv":"1609160400000711344","hd":{"ts_event":"1609160400000431665","rtype":160,"publisher_id":1,"instrument_id":5482},"action":"C","side":"A","price":"3723000000000","size":1,"channel_id":0,"order_id":"647784973631","flags":128,"ts_in_delta":19621,"sequence":1170353}
			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609160400000429831)))
			Expect(r0h.RType).To(Equal(dbn.RType(160)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r0.OrderID).To(Equal(uint64(647784973705)))
			Expect(r0.Price).To(Equal(int64(3722750000000)))
			Expect(r0.Size).To(Equal(uint32(1)))
			Expect(r0.Flags).To(Equal(uint8(128)))
			Expect(r0.ChannelID).To(Equal(uint8(0)))
			Expect(r0.Action).To(Equal(byte('C')))
			Expect(r0.Side).To(Equal(byte('A')))
			Expect(r0.TsRecv).To(Equal(uint64(1609160400000704060)))
			Expect(r0.TsInDelta).To(Equal(int32(22993)))
			Expect(r0.Sequence).To(Equal(uint32(1170352)))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609160400000431665)))
			Expect(r1h.RType).To(Equal(dbn.RType(160)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r1.OrderID).To(Equal(uint64(647784973631)))
			Expect(r1.Price).To(Equal(int64(3723000000000)))
			Expect(r1.Size).To(Equal(uint32(1)))
			Expect(r1.Flags).To(Equal(uint8(128)))
			Expect(r1.ChannelID).To(Equal(uint8(0)))
			Expect(r1.Action).To(Equal(byte('C')))
			Expect(r1.Side).To(Equal(byte('A')))
			Expect(r1.TsRecv).To(Equal(uint64(1609160400000711344)))
			Expect(r1.TsInDelta).To(Equal(int32(19621)))
			Expect(r1.Sequence).To(Equal(uint32(1170353)))
		})

		It("should read a v2 mbo correctly", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.mbo.v2.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.MboMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.mbo.v2.dbn.zst
			// {"ts_recv":"1609160400000704060","hd":{"ts_event":"1609160400000429831","rtype":160,"publisher_id":1,"instrument_id":5482},"action":"C","side":"A","price":"3722750000000","size":1,"channel_id":0,"order_id":"647784973705","flags":128,"ts_in_delta":22993,"sequence":1170352}
			// {"ts_recv":"1609160400000711344","hd":{"ts_event":"1609160400000431665","rtype":160,"publisher_id":1,"instrument_id":5482},"action":"C","side":"A","price":"3723000000000","size":1,"channel_id":0,"order_id":"647784973631","flags":128,"ts_in_delta":19621,"sequence":1170353}
			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609160400000429831)))
			Expect(r0h.RType).To(Equal(dbn.RType(160)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r0.OrderID).To(Equal(uint64(647784973705)))
			Expect(r0.Price).To(Equal(int64(3722750000000)))
			Expect(r0.Size).To(Equal(uint32(1)))
			Expect(r0.Flags).To(Equal(uint8(128)))
			Expect(r0.ChannelID).To(Equal(uint8(0)))
			Expect(r0.Action).To(Equal(byte('C')))
			Expect(r0.Side).To(Equal(byte('A')))
			Expect(r0.TsRecv).To(Equal(uint64(1609160400000704060)))
			Expect(r0.TsInDelta).To(Equal(int32(22993)))
			Expect(r0.Sequence).To(Equal(uint32(1170352)))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609160400000431665)))
			Expect(r1h.RType).To(Equal(dbn.RType(160)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r1.OrderID).To(Equal(uint64(647784973631)))
			Expect(r1.Price).To(Equal(int64(3723000000000)))
			Expect(r1.Size).To(Equal(uint32(1)))
			Expect(r1.Flags).To(Equal(uint8(128)))
			Expect(r1.ChannelID).To(Equal(uint8(0)))
			Expect(r1.Action).To(Equal(byte('C')))
			Expect(r1.Side).To(Equal(byte('A')))
			Expect(r1.TsRecv).To(Equal(uint64(1609160400000711344)))
			Expect(r1.TsInDelta).To(Equal(int32(19621)))
			Expect(r1.Sequence).To(Equal(uint32(1170353)))
		})

		It("should read a v1 tbbo correctly", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.tbbo.v1.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.Mbp1Msg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.tbbo.v1.dbn.zst
			// {"ts_recv":"1609160400099150057","hd":{"ts_event":"1609160400098821953","rtype":1,"publisher_id":1,"instrument_id":5482},"action":"T","side":"A","depth":0,"price":"3720250000000","size":5,"flags":129,"ts_in_delta":19251,"sequence":1170380,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":26,"ask_sz":7,"bid_ct":16,"ask_ct":6}]}
			// {"ts_recv":"1609160400108142648","hd":{"ts_event":"1609160400107665963","rtype":1,"publisher_id":1,"instrument_id":5482},"action":"T","side":"A","depth":0,"price":"3720250000000","size":21,"flags":129,"ts_in_delta":20728,"sequence":1170414,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":21,"ask_sz":22,"bid_ct":13,"ask_ct":15}]}
			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609160400098821953)))
			Expect(r0h.RType).To(Equal(dbn.RType(1)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r0.Price).To(Equal(int64(3720250000000)))
			Expect(r0.Size).To(Equal(uint32(5)))
			Expect(r0.Action).To(Equal(byte('T')))
			Expect(r0.Side).To(Equal(byte('A')))
			Expect(r0.Flags).To(Equal(uint8(129)))
			Expect(r0.Depth).To(Equal(uint8(0)))
			Expect(r0.TsRecv).To(Equal(uint64(1609160400099150057)))
			Expect(r0.TsInDelta).To(Equal(int32(19251)))
			Expect(r0.Sequence).To(Equal(uint32(1170380)))
			Expect(r0.Level).To(Equal(dbn.BidAskPair{
				BidPx: int64(3720250000000),
				AskPx: int64(3720500000000),
				BidSz: uint32(26),
				AskSz: uint32(7),
				BidCt: uint32(16),
				AskCt: uint32(6),
			}))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609160400107665963)))
			Expect(r1h.RType).To(Equal(dbn.RType(1)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r1.Price).To(Equal(int64(3720250000000)))
			Expect(r1.Size).To(Equal(uint32(21)))
			Expect(r1.Action).To(Equal(byte('T')))
			Expect(r1.Side).To(Equal(byte('A')))
			Expect(r1.Flags).To(Equal(uint8(129)))
			Expect(r1.Depth).To(Equal(uint8(0)))
			Expect(r1.TsRecv).To(Equal(uint64(1609160400108142648)))
			Expect(r1.TsInDelta).To(Equal(int32(20728)))
			Expect(r1.Sequence).To(Equal(uint32(1170414)))
			Expect(r1.Level).To(Equal(dbn.BidAskPair{
				BidPx: int64(3720250000000),
				AskPx: int64(3720500000000),
				BidSz: uint32(21),
				AskSz: uint32(22),
				BidCt: uint32(13),
				AskCt: uint32(15),
			}))
		})
	})

	It("should read a v2 tbbo correctly", func() {
		file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.tbbo.v2.dbn.zst", false)
		Expect(err).To(BeNil())
		defer closer.Close()

		records, metadata, err := dbn.ReadDBNToSlice[dbn.Mbp1Msg](file)
		Expect(err).To(BeNil())
		Expect(metadata).ToNot(BeNil())
		Expect(len(records)).To(Equal(2))

		// dbn -J ./tests/data/test_data.tbbo.v2.dbn.zst
		// {"ts_recv":"1609160400099150057","hd":{"ts_event":"1609160400098821953","rtype":1,"publisher_id":1,"instrument_id":5482},"action":"T","side":"A","depth":0,"price":"3720250000000","size":5,"flags":129,"ts_in_delta":19251,"sequence":1170380,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":26,"ask_sz":7,"bid_ct":16,"ask_ct":6}]}
		// {"ts_recv":"1609160400108142648","hd":{"ts_event":"1609160400107665963","rtype":1,"publisher_id":1,"instrument_id":5482},"action":"T","side":"A","depth":0,"price":"3720250000000","size":21,"flags":129,"ts_in_delta":20728,"sequence":1170414,"levels":[{"bid_px":"3720250000000","ask_px":"3720500000000","bid_sz":21,"ask_sz":22,"bid_ct":13,"ask_ct":15}]}
		r0, r0h := records[0], records[0].Header
		Expect(r0h.TsEvent).To(Equal(uint64(1609160400098821953)))
		Expect(r0h.RType).To(Equal(dbn.RType(1)))
		Expect(r0h.PublisherID).To(Equal(uint16(1)))
		Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
		Expect(r0.Price).To(Equal(int64(3720250000000)))
		Expect(r0.Size).To(Equal(uint32(5)))
		Expect(r0.Action).To(Equal(byte('T')))
		Expect(r0.Side).To(Equal(byte('A')))
		Expect(r0.Flags).To(Equal(uint8(129)))
		Expect(r0.Depth).To(Equal(uint8(0)))
		Expect(r0.TsRecv).To(Equal(uint64(1609160400099150057)))
		Expect(r0.TsInDelta).To(Equal(int32(19251)))
		Expect(r0.Sequence).To(Equal(uint32(1170380)))
		Expect(r0.Level).To(Equal(dbn.BidAskPair{
			BidPx: int64(3720250000000),
			AskPx: int64(3720500000000),
			BidSz: uint32(26),
			AskSz: uint32(7),
			BidCt: uint32(16),
			AskCt: uint32(6),
		}))

		r1, r1h := records[1], records[1].Header
		Expect(r1h.TsEvent).To(Equal(uint64(1609160400107665963)))
		Expect(r1h.RType).To(Equal(dbn.RType(1)))
		Expect(r1h.PublisherID).To(Equal(uint16(1)))
		Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
		Expect(r1.Price).To(Equal(int64(3720250000000)))
		Expect(r1.Size).To(Equal(uint32(21)))
		Expect(r1.Action).To(Equal(byte('T')))
		Expect(r1.Side).To(Equal(byte('A')))
		Expect(r1.Flags).To(Equal(uint8(129)))
		Expect(r1.Depth).To(Equal(uint8(0)))
		Expect(r1.TsRecv).To(Equal(uint64(1609160400108142648)))
		Expect(r1.TsInDelta).To(Equal(int32(20728)))
		Expect(r1.Sequence).To(Equal(uint32(1170414)))
		Expect(r1.Level).To(Equal(dbn.BidAskPair{
			BidPx: int64(3720250000000),
			AskPx: int64(3720500000000),
			BidSz: uint32(21),
			AskSz: uint32(22),
			BidCt: uint32(13),
			AskCt: uint32(15),
		}))
	})

	Context("Misc message testing", func() {
		It("should read a v2 status correctly", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.status.v2.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.StatusMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(4))

			//dbn -J tests/data/test_data.status.v2.dbn.zst
			// {"ts_recv":"1609113600000000000","hd":{"ts_event":"1609110000000000000","rtype":18,"publisher_id":1,"instrument_id":5482},"action":7,"reason":1,"trading_event":0,"is_trading":"Y","is_quoting":"Y","is_short_sell_restricted":"~"}
			// {"ts_recv":"1609190100007055917","hd":{"ts_event":"1609190100000000000","rtype":18,"publisher_id":1,"instrument_id":5482},"action":1,"reason":1,"trading_event":0,"is_trading":"N","is_quoting":"Y","is_short_sell_restricted":"~"}
			// {"ts_recv":"1609190970068184258","hd":{"ts_event":"1609190970000000000","rtype":18,"publisher_id":1,"instrument_id":5482},"action":1,"reason":1,"trading_event":1,"is_trading":"N","is_quoting":"Y","is_short_sell_restricted":"~"}
			// {"ts_recv":"1609191000007282029","hd":{"ts_event":"1609191000000000000","rtype":18,"publisher_id":1,"instrument_id":5482},"action":6,"reason":1,"trading_event":0,"is_trading":"Y","is_quoting":"Y","is_short_sell_restricted":"~"}
			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609110000000000000)))
			Expect(r0h.RType).To(Equal(dbn.RType(18)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r0.TsRecv).To(Equal(uint64(1609113600000000000)))
			Expect(r0.Action).To(Equal(uint16(7)))
			Expect(r0.Reason).To(Equal(uint16(1)))
			Expect(r0.TradingEvent).To(Equal(uint16(0)))
			Expect(r0.IsTrading).To(Equal(uint8('Y')))
			Expect(r0.IsQuoting).To(Equal(uint8('Y')))
			Expect(r0.IsShortSellRestricted).To(Equal(uint8('~')))
			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609190100000000000)))
			Expect(r1h.RType).To(Equal(dbn.RType(18)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r1.TsRecv).To(Equal(uint64(1609190100007055917)))
			Expect(r1.Action).To(Equal(uint16(1)))
			Expect(r1.Reason).To(Equal(uint16(1)))
			Expect(r1.TradingEvent).To(Equal(uint16(0)))
			Expect(r1.IsTrading).To(Equal(uint8('N')))
			Expect(r1.IsQuoting).To(Equal(uint8('Y')))
			Expect(r1.IsShortSellRestricted).To(Equal(uint8('~')))
			r2, r2h := records[2], records[2].Header
			Expect(r2h.TsEvent).To(Equal(uint64(1609190970000000000)))
			Expect(r2h.RType).To(Equal(dbn.RType(18)))
			Expect(r2h.PublisherID).To(Equal(uint16(1)))
			Expect(r2h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r2.TsRecv).To(Equal(uint64(1609190970068184258)))
			Expect(r2.Action).To(Equal(uint16(1)))
			Expect(r2.Reason).To(Equal(uint16(1)))
			Expect(r2.TradingEvent).To(Equal(uint16(1)))
			Expect(r2.IsTrading).To(Equal(uint8('N')))
			Expect(r2.IsQuoting).To(Equal(uint8('Y')))
			Expect(r2.IsShortSellRestricted).To(Equal(uint8('~')))
			r3, r3h := records[3], records[3].Header
			Expect(r3h.TsEvent).To(Equal(uint64(1609191000000000000)))
			Expect(r3h.RType).To(Equal(dbn.RType(18)))
			Expect(r3h.PublisherID).To(Equal(uint16(1)))
			Expect(r3h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r3.TsRecv).To(Equal(uint64(1609191000007282029)))
			Expect(r3.Action).To(Equal(uint16(6)))
			Expect(r3.Reason).To(Equal(uint16(1)))
			Expect(r3.TradingEvent).To(Equal(uint16(0)))
			Expect(r3.IsTrading).To(Equal(uint8('Y')))
			Expect(r3.IsQuoting).To(Equal(uint8('Y')))
			Expect(r3.IsShortSellRestricted).To(Equal(uint8('~')))
		})
	})
})

/*
   #[test]
   fn test_bbo_alignment_matches_mbp1() {
       assert_eq!(offset_of!(BboMsg, hd), offset_of!(Mbp1Msg, hd));
       assert_eq!(offset_of!(BboMsg, price), offset_of!(Mbp1Msg, price));
       assert_eq!(offset_of!(BboMsg, size), offset_of!(Mbp1Msg, size));
       assert_eq!(offset_of!(BboMsg, side), offset_of!(Mbp1Msg, side));
       assert_eq!(offset_of!(BboMsg, flags), offset_of!(Mbp1Msg, flags));
       assert_eq!(offset_of!(BboMsg, ts_recv), offset_of!(Mbp1Msg, ts_recv));
       assert_eq!(offset_of!(BboMsg, sequence), offset_of!(Mbp1Msg, sequence));
       assert_eq!(offset_of!(BboMsg, levels), offset_of!(Mbp1Msg, levels));
   }
*/
