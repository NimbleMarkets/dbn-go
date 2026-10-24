// Copyright (c) 2026 Neomantra Corp
//
// Tests for DBN Version 3 message types

package dbn_test

import (
	"os"
	"strings"

	dbn "github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Messages V3", func() {
	Context("InstrumentDef v3 messages", func() {
		It("should have correct size for InstrumentDefMsgV3", func() {
			Expect(dbn.InstrumentDefMsgV3_Size).To(BeEquivalentTo(520))
		})

		It("should read v3 definition correctly using InstrumentDefMsgV3 type", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.definition.v3.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.InstrumentDefMsgV3](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(metadata.VersionNum).To(Equal(uint8(dbn.HeaderVersion3)))
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.definition.v3.dbn.zst
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
			Expect(r0.UnitOfMeasureQty).To(Equal(int64(9223372036854775807)))
			Expect(r0.MinPriceIncrementAmount).To(Equal(int64(9223372036854775807)))
			Expect(r0.PriceRatio).To(Equal(int64(9223372036854775807)))
			Expect(r0.StrikePrice).To(Equal(int64(9223372036854775807)))
			Expect(r0.RawInstrumentID).To(Equal(uint64(2147483647)))
			Expect(r0.LegPrice).To(Equal(int64(9223372036854775807)))
			Expect(r0.LegDelta).To(Equal(int64(9223372036854775807)))
			Expect(r0.InstAttribValue).To(Equal(int32(2147483647)))
			Expect(r0.UnderlyingID).To(Equal(uint32(0)))
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
			Expect(r0.LegInstrumentID).To(Equal(uint32(0)))
			Expect(r0.LegRatioPriceNumerator).To(Equal(int32(0)))
			Expect(r0.LegRatioPriceDenominator).To(Equal(int32(0)))
			Expect(r0.LegRatioQtyNumerator).To(Equal(int32(0)))
			Expect(r0.LegRatioQtyDenominator).To(Equal(int32(0)))
			Expect(r0.LegUnderlyingID).To(Equal(uint32(0)))
			Expect(r0.ApplID).To(Equal(int16(32767)))
			Expect(r0.MaturityYear).To(Equal(uint16(65535)))
			Expect(r0.DecayStartDate).To(Equal(uint16(65535)))
			Expect(r0.ChannelID).To(Equal(uint16(0)))
			Expect(r0.LegCount).To(Equal(uint16(0)))
			Expect(r0.LegIndex).To(Equal(uint16(0)))
			Expect(string(r0.Currency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(string(r0.SettlCurrency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(string(r0.Secsubtype[:])).To(Equal("Z " + strings.Repeat("\x00", 4)))
			Expect(string(r0.RawSymbol[:])).To(Equal("MSFT" + strings.Repeat("\x00", dbn.MetadataV2_SymbolCstrLen-len("MSFT"))))
			Expect(string(r0.Group[:])).To(Equal("pxnas-1" + strings.Repeat("\x00", 21-len("pxnas-1"))))
			Expect(string(r0.Exchange[:])).To(Equal("XNAS" + strings.Repeat("\x00", 5-len("XNAS"))))
			Expect(string(r0.Asset[:])).To(Equal(strings.Repeat("\x00", dbn.MetadataV3_AssetCStrLen)))
			Expect(string(r0.Cfi[:])).To(Equal(strings.Repeat("\x00", 7)))
			Expect(string(r0.SecurityType[:])).To(Equal(strings.Repeat("\x00", 7)))
			Expect(string(r0.UnitOfMeasure[:])).To(Equal(strings.Repeat("\x00", 31)))
			Expect(string(r0.Underlying[:])).To(Equal(strings.Repeat("\x00", 21)))
			Expect(string(r0.StrikePriceCurrency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(string(r0.LegRawSymbol[:])).To(Equal(strings.Repeat("\x00", dbn.MetadataV3_SymbolCstrLen)))
			Expect(r0.InstrumentClass).To(Equal(uint8('K')))
			Expect(r0.MatchAlgorithm).To(Equal(uint8('F')))
			Expect(r0.MainFraction).To(Equal(uint8(255)))
			Expect(r0.PriceDisplayFormat).To(Equal(uint8(255)))
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
			Expect(r0.LegInstrumentClass).To(Equal(uint8(0)))
			Expect(r0.LegSide).To(Equal(uint8('N')))

			// {"ts_recv":"1633417621703120931","hd":{"ts_event":"1633417621703109854","rtype":19,"publisher_id":2,"instrument_id":6830},"raw_symbol":"MSFT","security_update_action":"A","instrument_class":"K","min_price_increment":"9223372036854775807","display_factor":"100000000000000","expiration":"18446744073709551615","activation":"18446744073709551615","high_limit_price":"9223372036854775807","low_limit_price":"9223372036854775807","max_price_variation":"9223372036854775807","unit_of_measure_qty":"9223372036854775807","min_price_increment_amount":"9223372036854775807","price_ratio":"9223372036854775807","inst_attrib_value":2147483647,"underlying_id":0,"raw_instrument_id":"2147483647","market_depth_implied":2147483647,"market_depth":2147483647,"market_segment_id":4294967295,"max_trade_vol":4294967295,"min_lot_size":2147483647,"min_lot_size_block":2147483647,"min_lot_size_round_lot":100,"min_trade_vol":4294967295,"contract_multiplier":2147483647,"decay_quantity":2147483647,"original_contract_size":2147483647,"appl_id":32767,"maturity_year":65535,"decay_start_date":65535,"channel_id":0,"currency":"","settl_currency":"","secsubtype":"Z ","group":"pxnas-1","exchange":"XNAS","asset":"","cfi":"","security_type":"","unit_of_measure":"","underlying":"","strike_price_currency":"","strike_price":"9223372036854775807","match_algorithm":"F","main_fraction":255,"price_display_format":255,"sub_fraction":255,"underlying_product":255,"maturity_month":255,"maturity_day":255,"maturity_week":255,"user_defined_instrument":"N","contract_multiplier_unit":127,"flow_schedule_type":127,"tick_rule":255,"leg_count":0,"leg_index":0,"leg_instrument_id":0,"leg_raw_symbol":"","leg_instrument_class":null,"leg_side":"N","leg_price":"9223372036854775807","leg_delta":"9223372036854775807","leg_ratio_price_numerator":0,"leg_ratio_price_denominator":0,"leg_ratio_qty_numerator":0,"leg_ratio_qty_denominator":0,"leg_underlying_id":0}
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
			Expect(r1.UnitOfMeasureQty).To(Equal(int64(9223372036854775807)))
			Expect(r1.MinPriceIncrementAmount).To(Equal(int64(9223372036854775807)))
			Expect(r1.PriceRatio).To(Equal(int64(9223372036854775807)))
			Expect(r1.StrikePrice).To(Equal(int64(9223372036854775807)))
			Expect(r1.RawInstrumentID).To(Equal(uint64(2147483647)))
			Expect(r1.LegPrice).To(Equal(int64(9223372036854775807)))
			Expect(r1.LegDelta).To(Equal(int64(9223372036854775807)))
			Expect(r1.InstAttribValue).To(Equal(int32(2147483647)))
			Expect(r1.UnderlyingID).To(Equal(uint32(0)))
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
			Expect(r1.LegInstrumentID).To(Equal(uint32(0)))
			Expect(r1.LegRatioPriceNumerator).To(Equal(int32(0)))
			Expect(r1.LegRatioPriceDenominator).To(Equal(int32(0)))
			Expect(r1.LegRatioQtyNumerator).To(Equal(int32(0)))
			Expect(r1.LegRatioQtyDenominator).To(Equal(int32(0)))
			Expect(r1.LegUnderlyingID).To(Equal(uint32(0)))
			Expect(r1.ApplID).To(Equal(int16(32767)))
			Expect(r1.MaturityYear).To(Equal(uint16(65535)))
			Expect(r1.DecayStartDate).To(Equal(uint16(65535)))
			Expect(r1.ChannelID).To(Equal(uint16(0)))
			Expect(r1.LegCount).To(Equal(uint16(0)))
			Expect(r1.LegIndex).To(Equal(uint16(0)))
			Expect(string(r1.Currency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(string(r1.SettlCurrency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(string(r1.Secsubtype[:])).To(Equal("Z " + strings.Repeat("\x00", 4)))
			Expect(string(r1.RawSymbol[:])).To(Equal("MSFT" + strings.Repeat("\x00", dbn.MetadataV2_SymbolCstrLen-len("MSFT"))))
			Expect(string(r1.Group[:])).To(Equal("pxnas-1" + strings.Repeat("\x00", 21-len("pxnas-1"))))
			Expect(string(r1.Exchange[:])).To(Equal("XNAS" + strings.Repeat("\x00", 5-len("XNAS"))))
			Expect(string(r1.Asset[:])).To(Equal(strings.Repeat("\x00", dbn.MetadataV3_AssetCStrLen)))
			Expect(string(r1.Cfi[:])).To(Equal(strings.Repeat("\x00", 7)))
			Expect(string(r1.SecurityType[:])).To(Equal(strings.Repeat("\x00", 7)))
			Expect(string(r1.UnitOfMeasure[:])).To(Equal(strings.Repeat("\x00", 31)))
			Expect(string(r1.Underlying[:])).To(Equal(strings.Repeat("\x00", 21)))
			Expect(string(r1.StrikePriceCurrency[:])).To(Equal(strings.Repeat("\x00", 4)))
			Expect(string(r1.LegRawSymbol[:])).To(Equal(strings.Repeat("\x00", dbn.MetadataV3_SymbolCstrLen)))
			Expect(r1.InstrumentClass).To(Equal(uint8('K')))
			Expect(r1.MatchAlgorithm).To(Equal(uint8('F')))
			Expect(r1.MainFraction).To(Equal(uint8(255)))
			Expect(r1.PriceDisplayFormat).To(Equal(uint8(255)))
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
			Expect(r1.LegInstrumentClass).To(Equal(uint8(0)))
			Expect(r1.LegSide).To(Equal(uint8('N')))
		})
	})

	Context("Stats v3 messages", func() {
		It("should have correct size for StatMsgV3", func() {
			Expect(dbn.StatMsgV3_Size).To(BeEquivalentTo(80))
		})
		It("should read v3 statistics correctly using StatMsgV3 type", func() {
			file, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.statistics.v3.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.StatMsgV3](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(metadata.VersionNum).To(Equal(uint8(dbn.HeaderVersion3)))
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.statistics.v3.dbn.zst
			// {"ts_recv":"1682269536040124325","hd":{"ts_event":"1682269536030443135","rtype":24,"publisher_id":1,"instrument_id":146945},"ts_ref":"18446744073709551615","price":"100000000000","quantity":"9223372036854775807","sequence":2,"ts_in_delta":26961,"stat_type":7,"channel_id":13,"update_action":1,"stat_flags":255}
			// {"ts_recv":"1682269536121890092","hd":{"ts_event":"1682269536071497081","rtype":24,"publisher_id":1,"instrument_id":146945},"ts_ref":"18446744073709551615","price":"100000000000","quantity":"9223372036854775807","sequence":7,"ts_in_delta":28456,"stat_type":5,"channel_id":13,"update_action":1,"stat_flags":255}
			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1682269536030443135)))
			Expect(r0h.RType).To(Equal(dbn.RType(24)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(146945)))
			Expect(r0.TsRecv).To(Equal(uint64(1682269536040124325)))
			Expect(r0.TsRef).To(Equal(uint64(18446744073709551615)))
			Expect(r0.Price).To(Equal(int64(100000000000)))
			Expect(r0.Quantity).To(Equal(int64(9223372036854775807)))
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
			Expect(r1.Quantity).To(Equal(int64(9223372036854775807)))
			Expect(r1.Sequence).To(Equal(uint32(7)))
			Expect(r1.TsInDelta).To(Equal(int32(28456)))
			Expect(r1.StatType).To(Equal(uint16(5)))
			Expect(r1.ChannelID).To(Equal(uint16(13)))
			Expect(r1.UpdateAction).To(Equal(uint8(1)))
			Expect(r1.StatFlags).To(Equal(uint8(255)))
		})
	})
})

///////////////////////////////////////////////////////////////////////////////
// V3 fixtures for MBO, MBP-1, MBP-10, CMBP-1 and Status
//
// Expected records below are what the upstream decoder produces for the
// matching ./tests/data/test_data.<schema>.v3.dbn.zst fixture:
// databento_dbn 0.70.0 (Python bindings over the upstream Rust `dbn` crate),
// DBNDecoder(upgrade_policy=AS_IS), one record per line, Length = record_size()/4.
// They agree with the `dbn -J` output quoted in the V1/V2 tests for the same data.
// They are NOT derived from dbn-go's decoder. They are reused by the JSON tests
// in json_scanner_test.go, which decode the upstream JSON encoding of the same files.

// readV3Fixture opens a V3 fixture and decodes every record as R, asserting the
// metadata is DBN version 3.
func readV3Fixture[R dbn.Record, RP dbn.RecordPtr[R]](path string) []R {
	file, closer, err := dbn.MakeCompressedReader(path, false)
	Expect(err).To(BeNil())
	defer closer.Close()

	records, metadata, err := dbn.ReadDBNToSlice[R, RP](file)
	Expect(err).To(BeNil())
	Expect(metadata).ToNot(BeNil())
	Expect(metadata.VersionNum).To(Equal(uint8(3)))
	return records
}

var _ = Describe("Messages V3 market data", func() {
	It("should read v3 mbo correctly", func() {
		Expect(readV3Fixture[dbn.MboMsg]("./tests/data/test_data.mbo.v3.dbn.zst")).To(Equal(upstreamMboV3))
	})

	It("should read uncompressed v3 mbo correctly", func() {
		file, err := os.Open("./tests/data/test_data.mbo.v3.dbn")
		Expect(err).To(BeNil())
		defer file.Close()
		records, metadata, err := dbn.ReadDBNToSlice[dbn.MboMsg](file)
		Expect(err).To(BeNil())
		Expect(metadata.VersionNum).To(Equal(uint8(3)))
		Expect(records).To(Equal(upstreamMboV3))
	})

	It("should read v3 mbp1 correctly", func() {
		Expect(readV3Fixture[dbn.Mbp1Msg]("./tests/data/test_data.mbp-1.v3.dbn.zst")).To(Equal(upstreamMbp1V3))
	})

	It("should read v3 mbp10 correctly", func() {
		records := readV3Fixture[dbn.Mbp10Msg]("./tests/data/test_data.mbp-10.v3.dbn.zst")
		Expect(records).To(Equal(upstreamMbp10V3))
		// spot-check the discriminating values called out by upstream
		Expect(records[0].Depth).To(Equal(uint8(9)))
		Expect(records[1].Depth).To(Equal(uint8(1)))
		Expect(records[1].Side).To(Equal(byte('B')))
		Expect(records[0].Levels[9].BidPx).To(Equal(int64(3718000000000)))
	})

	It("should read v3 cmbp1 correctly", func() {
		Expect(readV3Fixture[dbn.Cmbp1Msg]("./tests/data/test_data.cmbp-1.v3.dbn.zst")).To(Equal(upstreamCmbp1V3))
	})

	It("should read v3 status correctly", func() {
		Expect(readV3Fixture[dbn.StatusMsg]("./tests/data/test_data.status.v3.dbn.zst")).To(Equal(upstreamStatusV3))
	})
})

var upstreamMboV3 = []dbn.MboMsg{
	{Header: dbn.RHeader{Length: 14, RType: 160, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609160400000429831}, OrderID: 647784973705, Price: 3722750000000, Size: 1, Action: 'C', Side: 'A', Flags: 128, TsRecv: 1609160400000704060, TsInDelta: 22993, ChannelID: 0, Sequence: 1170352},
	{Header: dbn.RHeader{Length: 14, RType: 160, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609160400000431665}, OrderID: 647784973631, Price: 3723000000000, Size: 1, Action: 'C', Side: 'A', Flags: 128, TsRecv: 1609160400000711344, TsInDelta: 19621, ChannelID: 0, Sequence: 1170353},
}

var upstreamMbp1V3 = []dbn.Mbp1Msg{
	{Header: dbn.RHeader{Length: 20, RType: 1, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609160400006001487}, Price: 3720500000000, Size: 1, Action: 'A', Side: 'A', Flags: 128, TsRecv: 1609160400006136329, TsInDelta: 17214, Depth: 0, Sequence: 1170362, Level: dbn.BidAskPair{BidPx: 3720250000000, AskPx: 3720500000000, BidSz: 24, AskSz: 11, BidCt: 15, AskCt: 9}},
	{Header: dbn.RHeader{Length: 20, RType: 1, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609160400006146661}, Price: 3720500000000, Size: 1, Action: 'A', Side: 'A', Flags: 128, TsRecv: 1609160400006246513, TsInDelta: 18858, Depth: 0, Sequence: 1170364, Level: dbn.BidAskPair{BidPx: 3720250000000, AskPx: 3720500000000, BidSz: 24, AskSz: 12, BidCt: 15, AskCt: 10}},
}

var upstreamMbp10V3 = []dbn.Mbp10Msg{
	{Header: dbn.RHeader{Length: 92, RType: 10, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609160400000429831}, Price: 3722750000000, Size: 1, Action: 'C', Side: 'A', Flags: 128, TsRecv: 1609160400000704060, TsInDelta: 22993, Depth: 9, Sequence: 1170352, Levels: [10]dbn.BidAskPair{
		{BidPx: 3720250000000, AskPx: 3720500000000, BidSz: 24, AskSz: 10, BidCt: 15, AskCt: 8},
		{BidPx: 3720000000000, AskPx: 3720750000000, BidSz: 31, AskSz: 34, BidCt: 18, AskCt: 24},
		{BidPx: 3719750000000, AskPx: 3721000000000, BidSz: 32, AskSz: 39, BidCt: 23, AskCt: 25},
		{BidPx: 3719500000000, AskPx: 3721250000000, BidSz: 39, AskSz: 28, BidCt: 26, AskCt: 17},
		{BidPx: 3719250000000, AskPx: 3721500000000, BidSz: 50, AskSz: 33, BidCt: 35, AskCt: 19},
		{BidPx: 3719000000000, AskPx: 3721750000000, BidSz: 42, AskSz: 45, BidCt: 28, AskCt: 33},
		{BidPx: 3718750000000, AskPx: 3722000000000, BidSz: 44, AskSz: 55, BidCt: 35, AskCt: 40},
		{BidPx: 3718500000000, AskPx: 3722250000000, BidSz: 64, AskSz: 59, BidCt: 39, AskCt: 38},
		{BidPx: 3718250000000, AskPx: 3722500000000, BidSz: 53, AskSz: 49, BidCt: 32, AskCt: 35},
		{BidPx: 3718000000000, AskPx: 3722750000000, BidSz: 67, AskSz: 44, BidCt: 39, AskCt: 26}}},
	{Header: dbn.RHeader{Length: 92, RType: 10, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609160400000435673}, Price: 3720000000000, Size: 1, Action: 'C', Side: 'B', Flags: 128, TsRecv: 1609160400000750544, TsInDelta: 20625, Depth: 1, Sequence: 1170356, Levels: [10]dbn.BidAskPair{
		{BidPx: 3720250000000, AskPx: 3720500000000, BidSz: 24, AskSz: 10, BidCt: 15, AskCt: 8},
		{BidPx: 3720000000000, AskPx: 3720750000000, BidSz: 30, AskSz: 34, BidCt: 17, AskCt: 24},
		{BidPx: 3719750000000, AskPx: 3721000000000, BidSz: 32, AskSz: 39, BidCt: 23, AskCt: 25},
		{BidPx: 3719500000000, AskPx: 3721250000000, BidSz: 39, AskSz: 28, BidCt: 26, AskCt: 17},
		{BidPx: 3719250000000, AskPx: 3721500000000, BidSz: 50, AskSz: 33, BidCt: 35, AskCt: 19},
		{BidPx: 3719000000000, AskPx: 3721750000000, BidSz: 42, AskSz: 45, BidCt: 28, AskCt: 33},
		{BidPx: 3718750000000, AskPx: 3722000000000, BidSz: 44, AskSz: 55, BidCt: 35, AskCt: 40},
		{BidPx: 3718500000000, AskPx: 3722250000000, BidSz: 64, AskSz: 59, BidCt: 39, AskCt: 38},
		{BidPx: 3718250000000, AskPx: 3722500000000, BidSz: 53, AskSz: 49, BidCt: 32, AskCt: 35},
		{BidPx: 3718000000000, AskPx: 3722750000000, BidSz: 67, AskSz: 44, BidCt: 39, AskCt: 26}}},
}

var upstreamCmbp1V3 = []dbn.Cmbp1Msg{
	{Header: dbn.RHeader{Length: 20, RType: 177, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609160400006001487}, Price: 3720500000000, Size: 1, Action: 'A', Side: 'A', Flags: 128, TsRecv: 1609160400006136329, TsInDelta: 17214, Level: dbn.ConsolidatedBidAskPair{BidPx: 3720250000000, AskPx: 3720500000000, BidSz: 24, AskSz: 11, BidPb: 1, AskPb: 1}},
	{Header: dbn.RHeader{Length: 20, RType: 177, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609160400006146661}, Price: 3720500000000, Size: 1, Action: 'A', Side: 'A', Flags: 128, TsRecv: 1609160400006246513, TsInDelta: 18858, Level: dbn.ConsolidatedBidAskPair{BidPx: 3720250000000, AskPx: 3720500000000, BidSz: 24, AskSz: 12, BidPb: 1, AskPb: 1}},
}

var upstreamStatusV3 = []dbn.StatusMsg{
	{Header: dbn.RHeader{Length: 10, RType: 18, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609110000000000000}, TsRecv: 1609113600000000000, Action: 7, Reason: 1, TradingEvent: 0, IsTrading: 'Y', IsQuoting: 'Y', IsShortSellRestricted: '~'},
	{Header: dbn.RHeader{Length: 10, RType: 18, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609190100000000000}, TsRecv: 1609190100007055917, Action: 1, Reason: 1, TradingEvent: 0, IsTrading: 'N', IsQuoting: 'Y', IsShortSellRestricted: '~'},
	{Header: dbn.RHeader{Length: 10, RType: 18, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609190970000000000}, TsRecv: 1609190970068184258, Action: 1, Reason: 1, TradingEvent: 1, IsTrading: 'N', IsQuoting: 'Y', IsShortSellRestricted: '~'},
	{Header: dbn.RHeader{Length: 10, RType: 18, PublisherID: 1, InstrumentID: 5482, TsEvent: 1609191000000000000}, TsRecv: 1609191000007282029, Action: 6, Reason: 1, TradingEvent: 0, IsTrading: 'Y', IsQuoting: 'Y', IsShortSellRestricted: '~'},
}
