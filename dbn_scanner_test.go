package dbn_test

import (
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
