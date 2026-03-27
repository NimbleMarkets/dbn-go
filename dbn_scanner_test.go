package dbn_test

import (
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
})
