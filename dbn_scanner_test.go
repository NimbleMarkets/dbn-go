package dbn_test

import (
	"bytes"
	"encoding/hex"
	"io"
	"math"
	"os"
	"testing"

	"github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test Launcher
func TestDbn(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dbn-go suite")
}

// capturingVisitor captures records delivered via Visit() for assertions.
type capturingVisitor struct {
	dbn.NullVisitor
	Stats []dbn.StatMsg
	Defs  []dbn.InstrumentDefMsg
}

func (v *capturingVisitor) OnStatMsg(r *dbn.StatMsg) error {
	v.Stats = append(v.Stats, *r)
	return nil
}

func (v *capturingVisitor) OnInstrumentDefMsg(r *dbn.InstrumentDefMsg) error {
	v.Defs = append(v.Defs, *r)
	return nil
}

// visitAll runs scanner.Next()+Visit() until EOF, returning the visitor and any non-EOF error.
func visitAll(scanner *dbn.DbnScanner, visitor dbn.Visitor) error {
	for scanner.Next() {
		if err := scanner.Visit(visitor); err != nil {
			return err
		}
	}
	if err := scanner.Error(); err != io.EOF {
		return err
	}
	return nil
}

var _ = Describe("DbnScanner", func() {
	Context("v1 files", func() {
		It("should read a v1 test file correctly", func() {
			file, err := os.Open("./tests/data/test_data.ohlcv-1s.v1.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.OhlcvMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))
		})
	})

	Context("v2 files", func() {
		It("should read a v2 test file correctly", func() {
			file, err := os.Open("./tests/data/test_data.ohlcv-1s.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.OhlcvMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))
		})
	})

	Context("implementation correctness", func() {
		It("should size DEFAULT_SCRATCH_BUFFER_SIZE large enough for records", func() {
			Expect(int((&dbn.RHeader{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.Mbp0Msg{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.Mbp1Msg{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.Mbp10Msg{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.Cmbp1Msg{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.OhlcvMsg{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.ImbalanceMsg{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.ErrorMsg{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.ErrorMsgV1{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.ErrorMsgV2{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.StatMsg{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.StatMsgV2{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.StatMsgV3{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.StatusMsg{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.BboMsg{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.InstrumentDefMsg{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.InstrumentDefMsgV2{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
			Expect(int((&dbn.InstrumentDefMsgV3{}).RSize())).Should(BeNumerically("<", dbn.DEFAULT_SCRATCH_BUFFER_SIZE))
		})
	})

	// TODO: we should make a test for Error message upgrades, be we have no test file

	// Version-aware Visit() tests: verify that the scanner upgrades V1/V2 records to V3.
	Context("version-aware Visit for StatMsg", func() {
		It("should upgrade V1 statistics to V3 via Visit", func() {
			reader, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.statistics.v1.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			scanner := dbn.NewDbnScanner(reader)
			visitor := &capturingVisitor{}
			Expect(visitAll(scanner, visitor)).To(Succeed())

			metadata, _ := scanner.Metadata()
			Expect(metadata.VersionNum).To(Equal(uint8(dbn.HeaderVersion1)))
			Expect(visitor.Stats).To(HaveLen(2))

			// V1 Quantity was int32 max (2147483647), should be sign-extended to int64
			r0 := visitor.Stats[0]
			Expect(r0.Header.RType).To(Equal(dbn.RType_Statistics))
			Expect(r0.TsRecv).To(Equal(uint64(1682269536040124325)))
			Expect(r0.Price).To(Equal(int64(100000000000)))
			Expect(r0.Quantity).To(Equal(int64(math.MaxInt32)))
			Expect(r0.Sequence).To(Equal(uint32(2)))
			Expect(r0.StatType).To(Equal(uint16(7)))

			r1 := visitor.Stats[1]
			Expect(r1.TsRecv).To(Equal(uint64(1682269536121890092)))
			Expect(r1.Quantity).To(Equal(int64(math.MaxInt32)))
			Expect(r1.Sequence).To(Equal(uint32(7)))
			Expect(r1.StatType).To(Equal(uint16(5)))
		})

		It("should upgrade V2 statistics to V3 via Visit", func() {
			reader, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.statistics.v2.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			scanner := dbn.NewDbnScanner(reader)
			visitor := &capturingVisitor{}
			Expect(visitAll(scanner, visitor)).To(Succeed())

			metadata, _ := scanner.Metadata()
			Expect(metadata.VersionNum).To(Equal(uint8(dbn.HeaderVersion2)))
			Expect(visitor.Stats).To(HaveLen(2))

			r0 := visitor.Stats[0]
			Expect(r0.TsRecv).To(Equal(uint64(1682269536040124325)))
			Expect(r0.Price).To(Equal(int64(100000000000)))
			Expect(r0.Quantity).To(Equal(int64(math.MaxInt32)))
		})

		It("should read V3 statistics natively via Visit", func() {
			reader, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.statistics.v3.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			scanner := dbn.NewDbnScanner(reader)
			visitor := &capturingVisitor{}
			Expect(visitAll(scanner, visitor)).To(Succeed())

			metadata, _ := scanner.Metadata()
			Expect(metadata.VersionNum).To(Equal(uint8(dbn.HeaderVersion3)))
			Expect(visitor.Stats).To(HaveLen(2))

			// V3 Quantity is native int64, test data has MaxInt64
			r0 := visitor.Stats[0]
			Expect(r0.TsRecv).To(Equal(uint64(1682269536040124325)))
			Expect(r0.Price).To(Equal(int64(100000000000)))
			Expect(r0.Quantity).To(Equal(int64(math.MaxInt64)))
			Expect(r0.Sequence).To(Equal(uint32(2)))

			r1 := visitor.Stats[1]
			Expect(r1.TsRecv).To(Equal(uint64(1682269536121890092)))
			Expect(r1.Quantity).To(Equal(int64(math.MaxInt64)))
		})
	})

	Context("version-aware Visit for InstrumentDefMsg", func() {
		It("should upgrade V2 definition to V3 via Visit", func() {
			reader, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.definition.v2.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			scanner := dbn.NewDbnScanner(reader)
			visitor := &capturingVisitor{}
			Expect(visitAll(scanner, visitor)).To(Succeed())

			metadata, _ := scanner.Metadata()
			Expect(metadata.VersionNum).To(Equal(uint8(dbn.HeaderVersion2)))
			Expect(visitor.Defs).To(HaveLen(2))

			r0 := visitor.Defs[0]
			Expect(r0.Header.InstrumentID).To(Equal(uint32(6819)))
			Expect(r0.TsRecv).To(Equal(uint64(1633331241618029519)))
			// RawInstrumentID: V2 was uint32 max/2, zero-extended to uint64
			Expect(r0.RawInstrumentID).To(Equal(uint64(2147483647)))
			Expect(r0.DisplayFactor).To(Equal(int64(100000000000000)))
			Expect(r0.InstrumentClass).To(Equal(uint8('K')))
			Expect(r0.MatchAlgorithm).To(Equal(uint8('F')))
			// V3-only leg fields should be zero-valued after upgrade
			Expect(r0.LegPrice).To(Equal(int64(0)))
			Expect(r0.LegDelta).To(Equal(int64(0)))
			Expect(r0.LegInstrumentID).To(Equal(uint32(0)))
			Expect(r0.LegCount).To(Equal(uint16(0)))
			Expect(r0.LegIndex).To(Equal(uint16(0)))
			// Asset: V2 was [7]byte, V3 is [11]byte — extra bytes should be zero
			Expect(r0.Asset[7]).To(Equal(byte(0)))
			Expect(r0.Asset[10]).To(Equal(byte(0)))
			// RawSymbol should be copied correctly
			Expect(dbn.TrimNullBytes(r0.RawSymbol[:])).To(Equal("MSFT"))

			r1 := visitor.Defs[1]
			Expect(r1.Header.InstrumentID).To(Equal(uint32(6830)))
			Expect(r1.RawInstrumentID).To(Equal(uint64(2147483647)))
			Expect(dbn.TrimNullBytes(r1.RawSymbol[:])).To(Equal("MSFT"))
		})

		It("should read V3 definition natively via Visit", func() {
			reader, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.definition.v3.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			scanner := dbn.NewDbnScanner(reader)
			visitor := &capturingVisitor{}
			Expect(visitAll(scanner, visitor)).To(Succeed())

			metadata, _ := scanner.Metadata()
			Expect(metadata.VersionNum).To(Equal(uint8(dbn.HeaderVersion3)))
			Expect(visitor.Defs).To(HaveLen(2))

			r0 := visitor.Defs[0]
			Expect(r0.Header.InstrumentID).To(Equal(uint32(6819)))
			Expect(r0.RawInstrumentID).To(Equal(uint64(2147483647)))
			Expect(r0.LegPrice).To(Equal(int64(math.MaxInt64)))
			Expect(r0.LegDelta).To(Equal(int64(math.MaxInt64)))
			Expect(dbn.TrimNullBytes(r0.RawSymbol[:])).To(Equal("MSFT"))
		})

		It("should error on V1 definition via Visit", func() {
			reader, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.definition.v1.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			scanner := dbn.NewDbnScanner(reader)
			visitor := &capturingVisitor{}
			err = visitAll(scanner, visitor)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("V1"))
		})
	})

	Context("version-aware DecodeStatMsg and DecodeInstrumentDefMsg", func() {
		It("should upgrade V2 statistics via DecodeStatMsg", func() {
			reader, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.statistics.v2.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			scanner := dbn.NewDbnScanner(reader)
			Expect(scanner.Next()).To(BeTrue())

			r, err := scanner.DecodeStatMsg()
			Expect(err).To(BeNil())
			Expect(r.Quantity).To(Equal(int64(math.MaxInt32)))
			Expect(r.TsRecv).To(Equal(uint64(1682269536040124325)))
		})

		It("should decode V3 statistics via DecodeStatMsg", func() {
			reader, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.statistics.v3.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			scanner := dbn.NewDbnScanner(reader)
			Expect(scanner.Next()).To(BeTrue())

			r, err := scanner.DecodeStatMsg()
			Expect(err).To(BeNil())
			Expect(r.Quantity).To(Equal(int64(math.MaxInt64)))
		})

		It("should upgrade V2 definition via DecodeInstrumentDefMsg", func() {
			reader, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.definition.v2.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			scanner := dbn.NewDbnScanner(reader)
			Expect(scanner.Next()).To(BeTrue())

			r, err := scanner.DecodeInstrumentDefMsg()
			Expect(err).To(BeNil())
			Expect(r.RawInstrumentID).To(Equal(uint64(2147483647)))
			Expect(r.LegPrice).To(Equal(int64(0)))
			Expect(dbn.TrimNullBytes(r.RawSymbol[:])).To(Equal("MSFT"))
		})

		It("should decode V3 definition via DecodeInstrumentDefMsg", func() {
			reader, closer, err := dbn.MakeCompressedReader("./tests/data/test_data.definition.v3.dbn.zst", false)
			Expect(err).To(BeNil())
			defer closer.Close()

			scanner := dbn.NewDbnScanner(reader)
			Expect(scanner.Next()).To(BeTrue())

			r, err := scanner.DecodeInstrumentDefMsg()
			Expect(err).To(BeNil())
			Expect(r.RawInstrumentID).To(Equal(uint64(2147483647)))
			Expect(r.LegPrice).To(Equal(int64(math.MaxInt64)))
		})
	})
})

///////////////////////////////////////////////////////////////////////////////
// Live-control records: SymbolMappingMsg, SystemMsg, ErrorMsg
//
// Databento's test corpus has no fixture files for these record types (they
// only appear in live streams), so the byte streams below were produced by the
// upstream encoder: databento_dbn 0.70.0 (Python bindings over the upstream
// Rust `dbn` crate). The expected field values are the constructor arguments
// below, and were cross-checked against upstream DBNDecoder output. They are
// NOT derived from dbn-go's decoder.
//
//	START, END = 1609113600000000000, 1609200000000000000
//	meta = lambda v: databento_dbn.Metadata("GLBX.MDP3", START, SType.RAW_SYMBOL,
//	    SType.INSTRUMENT_ID, Schema.MBO, symbols=["ESH1"], end=END, version=v)
//
//	upstreamLiveControlStreamV3 = bytes(meta(3)) + bytes of:
//	  SymbolMappingMsg(1, 5482, 1609160400000000006, SType.RAW_SYMBOL, "ESH1",
//	      SType.INSTRUMENT_ID, "5482", START, END)
//	  SystemMsg(1609160400000000003, "Subscription request 1 for mbo data succeeded",
//	      code=SystemCode.SUBSCRIPTION_ACK)
//	  SystemMsg(1609160400000000004, "Heartbeat", code=SystemCode.HEARTBEAT)
//	  ErrorMsg(1609160400000000000, "Authentication failed: bad key", is_last=True,
//	      code=ErrorCode.AUTH_FAILED)
//
//	upstreamLiveControlStreamV1 = bytes(meta(1)) + bytes of:
//	  SymbolMappingMsgV1(1, 5482, 1609160400000000007, "ESH1", "5482", START, END)
//	  SystemMsgV1(1609160400000000005, "Heartbeat")
//	  ErrorMsgV1(1609160400000000002, "Unknown dataset")

func mustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// controlVisitor captures live-control records delivered via Visit().
type controlVisitor struct {
	dbn.NullVisitor
	SymbolMappings []dbn.SymbolMappingMsg
	Systems        []dbn.SystemMsg
	Errors         []dbn.ErrorMsg
}

func (v *controlVisitor) OnSymbolMappingMsg(r *dbn.SymbolMappingMsg) error {
	v.SymbolMappings = append(v.SymbolMappings, *r)
	return nil
}

func (v *controlVisitor) OnSystemMsg(r *dbn.SystemMsg) error {
	v.Systems = append(v.Systems, *r)
	return nil
}

func (v *controlVisitor) OnErrorMsg(r *dbn.ErrorMsg) error {
	v.Errors = append(v.Errors, *r)
	return nil
}

var _ = Describe("DbnScanner live-control records", func() {
	Context("v3 stream (V2/V3 layouts)", func() {
		It("should Visit SymbolMappingMsg, SystemMsg and ErrorMsg with upstream field values", func() {
			scanner := dbn.NewDbnScanner(bytes.NewReader(upstreamLiveControlStreamV3))
			visitor := &controlVisitor{}
			Expect(visitAll(scanner, visitor)).To(Succeed())

			metadata, err := scanner.Metadata()
			Expect(err).To(BeNil())
			Expect(metadata.VersionNum).To(Equal(uint8(3)))
			Expect(metadata.SymbolCstrLen).To(Equal(uint16(dbn.MetadataV2_SymbolCstrLen)))

			// SymbolMappingMsg(1, 5482, 1609160400000000006, RAW_SYMBOL, "ESH1", INSTRUMENT_ID, "5482", START, END)
			Expect(visitor.SymbolMappings).To(HaveLen(1))
			sm := visitor.SymbolMappings[0]
			Expect(sm.Header.Length).To(Equal(uint8(dbn.SymbolMappingMsgV2_Size / 4)))
			Expect(sm.Header.RType).To(Equal(dbn.RType_SymbolMapping))
			Expect(sm.Header.PublisherID).To(Equal(uint16(1)))
			Expect(sm.Header.InstrumentID).To(Equal(uint32(5482)))
			Expect(sm.Header.TsEvent).To(Equal(uint64(1609160400000000006)))
			Expect(sm.StypeIn).To(Equal(dbn.SType_RawSymbol))
			Expect(sm.StypeInSymbol).To(Equal("ESH1"))
			Expect(sm.StypeOut).To(Equal(dbn.SType_InstrumentId))
			Expect(sm.StypeOutSymbol).To(Equal("5482"))
			Expect(sm.StartTs).To(Equal(uint64(1609113600000000000)))
			Expect(sm.EndTs).To(Equal(uint64(1609200000000000000)))

			// SystemMsg(..., code=SUBSCRIPTION_ACK) and SystemMsg("Heartbeat", code=HEARTBEAT)
			Expect(visitor.Systems).To(HaveLen(2))
			s0 := visitor.Systems[0]
			Expect(s0.Header.Length).To(Equal(uint8(dbn.SystemMsg_Size / 4)))
			Expect(s0.Header.RType).To(Equal(dbn.RType_System))
			Expect(s0.Header.PublisherID).To(Equal(uint16(0)))
			Expect(s0.Header.InstrumentID).To(Equal(uint32(0)))
			Expect(s0.Header.TsEvent).To(Equal(uint64(1609160400000000003)))
			Expect(dbn.TrimNullBytes(s0.Message[:])).To(Equal("Subscription request 1 for mbo data succeeded"))
			Expect(s0.Code).To(Equal(dbn.SystemCode(1))) // SystemCode.SUBSCRIPTION_ACK
			Expect(s0.IsHeartbeat()).To(BeFalse())
			s1 := visitor.Systems[1]
			Expect(s1.Header.TsEvent).To(Equal(uint64(1609160400000000004)))
			Expect(dbn.TrimNullBytes(s1.Message[:])).To(Equal("Heartbeat"))
			Expect(s1.Code).To(Equal(dbn.SystemCode(0))) // SystemCode.HEARTBEAT
			Expect(s1.IsHeartbeat()).To(BeTrue())

			// ErrorMsg("Authentication failed: bad key", is_last=True, code=AUTH_FAILED)
			Expect(visitor.Errors).To(HaveLen(1))
			e0 := visitor.Errors[0]
			Expect(e0.Header.Length).To(Equal(uint8(dbn.ErrorMsgV2_Size / 4)))
			Expect(e0.Header.RType).To(Equal(dbn.RType_Error))
			Expect(e0.Header.TsEvent).To(Equal(uint64(1609160400000000000)))
			Expect(dbn.TrimNullBytes(e0.Error[:])).To(Equal("Authentication failed: bad key"))
			Expect(e0.Code).To(Equal(dbn.ErrorCode(1))) // ErrorCode.AUTH_FAILED
			Expect(e0.IsLast).To(Equal(uint8(1)))
		})

		It("should decode the same records via the Decode* helpers", func() {
			scanner := dbn.NewDbnScanner(bytes.NewReader(upstreamLiveControlStreamV3))

			Expect(scanner.Next()).To(BeTrue())
			sm, err := scanner.DecodeSymbolMappingMsg()
			Expect(err).To(BeNil())
			Expect(sm.StypeIn).To(Equal(dbn.SType_RawSymbol))
			Expect(sm.StypeInSymbol).To(Equal("ESH1"))
			Expect(sm.StypeOut).To(Equal(dbn.SType_InstrumentId))
			Expect(sm.StypeOutSymbol).To(Equal("5482"))
			Expect(sm.StartTs).To(Equal(uint64(1609113600000000000)))
			Expect(sm.EndTs).To(Equal(uint64(1609200000000000000)))

			Expect(scanner.Next()).To(BeTrue())
			sys, err := dbn.DbnScannerDecode[dbn.SystemMsg](scanner)
			Expect(err).To(BeNil())
			Expect(sys.Header.TsEvent).To(Equal(uint64(1609160400000000003)))
			Expect(dbn.TrimNullBytes(sys.Message[:])).To(Equal("Subscription request 1 for mbo data succeeded"))
			Expect(sys.Code).To(Equal(dbn.SystemCode(1)))

			Expect(scanner.Next()).To(BeTrue()) // heartbeat
			Expect(scanner.Next()).To(BeTrue())
			em, err := scanner.DecodeErrorMsg()
			Expect(err).To(BeNil())
			Expect(em.Header.TsEvent).To(Equal(uint64(1609160400000000000)))
			Expect(dbn.TrimNullBytes(em.Error[:])).To(Equal("Authentication failed: bad key"))
			Expect(em.Code).To(Equal(dbn.ErrorCode(1)))
			Expect(em.IsLast).To(Equal(uint8(1)))

			Expect(scanner.Next()).To(BeFalse())
			Expect(scanner.Error()).To(Equal(io.EOF))
		})
	})

	Context("v1 stream (V1 layouts)", func() {
		It("should Visit a V1 SymbolMappingMsg with upstream field values", func() {
			scanner := dbn.NewDbnScanner(bytes.NewReader(upstreamLiveControlStreamV1))
			Expect(scanner.Next()).To(BeTrue())
			metadata, err := scanner.Metadata()
			Expect(err).To(BeNil())
			Expect(metadata.VersionNum).To(Equal(uint8(1)))
			Expect(metadata.SymbolCstrLen).To(Equal(uint16(dbn.MetadataV1_SymbolCstrLen)))

			visitor := &controlVisitor{}
			Expect(scanner.Visit(visitor)).To(Succeed())
			Expect(visitor.SymbolMappings).To(HaveLen(1))
			// SymbolMappingMsgV1(1, 5482, 1609160400000000007, "ESH1", "5482", START, END)
			sm := visitor.SymbolMappings[0]
			Expect(sm.Header.Length).To(Equal(uint8(dbn.SymbolMappingMsgV1_Size / 4)))
			Expect(sm.Header.RType).To(Equal(dbn.RType_SymbolMapping))
			Expect(sm.Header.PublisherID).To(Equal(uint16(1)))
			Expect(sm.Header.InstrumentID).To(Equal(uint32(5482)))
			Expect(sm.Header.TsEvent).To(Equal(uint64(1609160400000000007)))
			Expect(sm.StypeInSymbol).To(Equal("ESH1"))
			Expect(sm.StypeOutSymbol).To(Equal("5482"))
			Expect(sm.StartTs).To(Equal(uint64(1609113600000000000)))
			Expect(sm.EndTs).To(Equal(uint64(1609200000000000000)))
			// NOTE: V1 has no stype bytes. dbn-go fills SType_RawSymbol (see messagesV1_test.go);
			// upstream's V1->V3 upgrade leaves them unset (255). Not asserted here.
		})

		It("should Visit a V1 SystemMsg with upstream field values", func() {
			scanner := dbn.NewDbnScanner(bytes.NewReader(upstreamLiveControlStreamV1))
			Expect(scanner.Next()).To(BeTrue()) // symbol mapping
			Expect(scanner.Next()).To(BeTrue())

			visitor := &controlVisitor{}
			Expect(scanner.Visit(visitor)).To(Succeed())
			Expect(visitor.Systems).To(HaveLen(1))
			// SystemMsgV1(1609160400000000005, "Heartbeat")
			s0 := visitor.Systems[0]
			Expect(s0.Header.RType).To(Equal(dbn.RType_System))
			Expect(s0.Header.TsEvent).To(Equal(uint64(1609160400000000005)))
			Expect(dbn.TrimNullBytes(s0.Message[:])).To(Equal("Heartbeat"))
			Expect(s0.IsHeartbeat()).To(BeTrue())
		})

		// BUG (decoder, not fixed in this branch): DbnScanner.Visit decodes every SystemMsg
		// with the V2 layout via lastRecord[:SystemMsg_Size] (320 bytes), even when the stream
		// is V1 and the record is only 80 bytes. Bytes 80..319 of the scratch buffer are stale
		// data left by an earlier, larger record (here an MBP-10), so they leak into
		// Message[64:] and Code. Upstream's V1->V3 upgrade yields msg="Heartbeat" (zero padded)
		// and code=SystemCode.HEARTBEAT (0) for this stream.
		PIt("should Visit a V1 SystemMsg after a larger record without reading stale bytes", func() {
			scanner := dbn.NewDbnScanner(bytes.NewReader(upstreamMbp10ThenHeartbeatStreamV1))
			Expect(scanner.Next()).To(BeTrue()) // MBP-10 (368 bytes)
			Expect(scanner.Next()).To(BeTrue()) // SystemMsgV1 (80 bytes)

			visitor := &controlVisitor{}
			Expect(scanner.Visit(visitor)).To(Succeed())
			Expect(visitor.Systems).To(HaveLen(1))
			s0 := visitor.Systems[0]
			Expect(s0.Header.TsEvent).To(Equal(uint64(1609160400000000010)))
			Expect(dbn.TrimNullBytes(s0.Message[:])).To(Equal("Heartbeat"))
			Expect(s0.Code).To(Equal(dbn.SystemCode(0)))
		})

		It("should Visit a V1 ErrorMsg and upgrade its error text", func() {
			scanner := dbn.NewDbnScanner(bytes.NewReader(upstreamLiveControlStreamV1))
			Expect(scanner.Next()).To(BeTrue()) // symbol mapping
			Expect(scanner.Next()).To(BeTrue()) // system
			Expect(scanner.Next()).To(BeTrue())

			visitor := &controlVisitor{}
			Expect(scanner.Visit(visitor)).To(Succeed())
			Expect(visitor.Errors).To(HaveLen(1))
			// ErrorMsgV1(1609160400000000002, "Unknown dataset")
			Expect(dbn.TrimNullBytes(visitor.Errors[0].Error[:])).To(Equal("Unknown dataset"))
		})

		// BUG (decoder, not fixed in this branch): DecodeErrorMsg's V1->V2 upgrade
		// (messages.go) copies only the error text into a fresh ErrorMsgV2 and drops the
		// record header, so RType/TsEvent/Length come back as zero. Upstream's V1->V3
		// upgrade keeps the header (rtype=21, ts_event=1609160400000000002) and sets
		// code=ErrorCode.UNSET (255) and is_last=255 (code/is_last not asserted here).
		PIt("should keep the record header when upgrading a V1 ErrorMsg", func() {
			scanner := dbn.NewDbnScanner(bytes.NewReader(upstreamLiveControlStreamV1))
			Expect(scanner.Next()).To(BeTrue()) // symbol mapping
			Expect(scanner.Next()).To(BeTrue()) // system
			Expect(scanner.Next()).To(BeTrue())

			visitor := &controlVisitor{}
			Expect(scanner.Visit(visitor)).To(Succeed())
			Expect(visitor.Errors).To(HaveLen(1))
			// ErrorMsgV1(1609160400000000002, "Unknown dataset")
			e0 := visitor.Errors[0]
			Expect(dbn.TrimNullBytes(e0.Error[:])).To(Equal("Unknown dataset"))
			Expect(e0.Header.RType).To(Equal(dbn.RType_Error))
			Expect(e0.Header.TsEvent).To(Equal(uint64(1609160400000000002)))
		})
	})
})

var upstreamLiveControlStreamV3 = mustDecodeHex("" +
	"44424e03c0000000474c42582e4d445033000000000000000000000040334bb8541600008fc4df065516000000000000" +
	"000001000047000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000100000045534831000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"00000000000000002c1601006a1500000620a0acdbe25416014553483100000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"003534383200000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000040334bb8541600008fc4df0655165017000000000000" +
	"0320a0acdbe25416537562736372697074696f6e2072657175657374203120666f72206d626f20646174612073756363" +
	"656564656400000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"00000000000000000000000000000000000000000000000150170000000000000420a0acdbe254164865617274626561" +
	"740000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000050150000000000000020a0acdbe2541641757468656e7469636174696f6e206661696c65643a2062" +
	"6164206b6579000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"00000000000000000000000000000000000000000000000000000000000000000000000000000101")

var upstreamLiveControlStreamV1 = mustDecodeHex("" +
	"44424e018e000000474c42582e4d445033000000000000000000000040334bb8541600008fc4df065516000000000000" +
	"0000ffffffffffffffff0100000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000100000045534831000000000000000000000000000000000000000000000000" +
	"000000000000141601006a1500000720a0acdbe254164553483100000000000000000000000000000000000035343832" +
	"00000000000000000000000000000000000000000000000040334bb8541600008fc4df06551614170000000000000520" +
	"a0acdbe25416486561727462656174000000000000000000000000000000000000000000000000000000000000000000" +
	"0000000000000000000000000000000000000000000014150000000000000220a0acdbe25416556e6b6e6f776e206461" +
	"746173657400000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000")

// upstreamMbp10ThenHeartbeatStreamV1 = bytes(meta(1)) + bytes of:
//
//	MBP10Msg(1, 5482, 1609160400000000008, 3720500000000, 1, Action.ADD, Side.ASK, 0,
//	    1609160400000000009, flags=128, levels=[BidAskPair(bid_px=3720250000000 - i*250000000,
//	    ask_px=3720500000000 + i*250000000, bid_sz=10+i, ask_sz=20+i, bid_ct=1+i, ask_ct=2+i)
//	    for i in range(10)])
//	SystemMsgV1(1609160400000000010, "Heartbeat")
var upstreamMbp10ThenHeartbeatStreamV1 = mustDecodeHex("" +
	"44424e018e000000474c42582e4d445033000000000000000000000040334bb8541600008fc4df065516000000000000" +
	"0000ffffffffffffffff0100000000000000000000000000000000000000000000000000000000000000000000000000" +
	"000000000000000000000000000000000100000045534831000000000000000000000000000000000000000000000000" +
	"0000000000005c0a01006a1500000820a0acdbe2541600b5143f6203000001000000414180000920a0acdbe254160000" +
	"00000000000080022e306203000000b5143f620300000a00000014000000010000000200000000504721620300008067" +
	"fb4d620300000b000000150000000200000003000000809d601262030000001ae25c620300000c000000160000000300" +
	"00000400000000eb79036203000080ccc86b620300000d000000170000000400000005000000803893f461030000007f" +
	"af7a620300000e0000001800000005000000060000000086ace56103000080319689620300000f000000190000000600" +
	"00000700000080d3c5d66103000000e47c9862030000100000001a00000007000000080000000021dfc7610300008096" +
	"63a762030000110000001b0000000800000009000000806ef8b86103000000494ab662030000120000001c0000000900" +
	"00000a00000000bc11aa6103000080fb30c562030000130000001d0000000a0000000b00000014170000000000000a20" +
	"a0acdbe25416486561727462656174000000000000000000000000000000000000000000000000000000000000000000" +
	"00000000000000000000000000000000000000000000")
