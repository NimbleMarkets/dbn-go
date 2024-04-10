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

			records, metadata, err := dbn.ReadDBNToSlice[dbn.Ohlcv](file)
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

			records, metadata, err := dbn.ReadDBNToSlice[dbn.Ohlcv](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))
		})
	})
})
