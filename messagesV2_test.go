// Copyright (c) 2024 Neomantra Corp
//
// Tests for DBN Version 2 message types

package dbn_test

import (
	"strings"

	dbn "github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessagesV2", func() {
	Context("StatMsgV2", func() {
		It("should read v2 statistics correctly using StatMsgV2 type", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.statistics.v2.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.StatMsgV2](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(metadata.VersionNum).To(Equal(uint8(2)))
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

		It("should have correct size for StatMsgV2", func() {
			Expect(dbn.StatMsgV2_Size).To(BeEquivalentTo(64))
		})
	})

	Context("InstrumentDefMsgV2", func() {
		It("should read v2 definition correctly using InstrumentDefMsgV2 type", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.definition.v2.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.InstrumentDefMsgV2](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(metadata.VersionNum).To(Equal(uint8(2)))
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
			Expect(string(r1.RawSymbol[:])).To(Equal("MSFT" + strings.Repeat("\x00", dbn.MetadataV2_SymbolCstrLen-len("MSFT"))))
		})

		It("should have correct size for InstrumentDefMsgV2", func() {
			Expect(dbn.InstrumentDefMsgV2_Size).To(BeEquivalentTo(400))
		})
	})

	Context("SymbolMappingMsgV2", func() {
		It("should have correct size for SymbolMappingMsgV2", func() {
			// V2: RHeader (16) + 16 + 2*71 + 2 = 16 + 16 + 142 + 2 = 176
			Expect(dbn.SymbolMappingMsgV2_Size).To(BeEquivalentTo(176))
		})

		It("should be compatible with SymbolMappingMsg alias", func() {
			// Verify the alias works correctly
			var msg dbn.SymbolMappingMsg
			msg.StypeIn = dbn.SType_RawSymbol
			msg.StypeInSymbol = "AAPL"
			msg.StypeOut = dbn.SType_InstrumentId
			msg.StypeOutSymbol = "12345"
			msg.StartTs = 1234567890
			msg.EndTs = 9876543210

			// Should be assignable to V2 type
			var v2Msg dbn.SymbolMappingMsgV2 = msg
			Expect(v2Msg.StypeInSymbol).To(Equal("AAPL"))
			Expect(v2Msg.StypeOutSymbol).To(Equal("12345"))
		})
	})
})
