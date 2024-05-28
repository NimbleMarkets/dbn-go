// Copyright (c) 2024 Neomantra Corp

package dbn_hist

import (
	"os"
	"testing"
	"time"

	dbn "github.com/NimbleMarkets/dbn-go"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test Launcher
func TestDbnHist(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dbn-go hist suite")
}

var databentoApiKey string

var _ = BeforeSuite(func() {
	databentoApiKey = os.Getenv("DATABENTO_API_KEY")
	if databentoApiKey == "" {
		Fail("DATABENTO_API_KEY not set")
	}
})

var _ = Describe("DbnHist", func() {
	Context("metadata", func() {
		It("should ListPublishers", func() {
			publishers, err := ListPublishers(databentoApiKey)
			Expect(err).To(BeNil())
			Expect(publishers).ToNot(BeEmpty())
		})
		It("should ListDatasets, ListSchemas, ListFields, ListUnitPrices", func() {
			datasets, err := ListDatasets(databentoApiKey, DateRange{})
			Expect(err).To(BeNil())
			Expect(datasets).ToNot(BeEmpty())

			schemas, err := ListSchemas(databentoApiKey, datasets[0])
			Expect(err).To(BeNil())
			Expect(schemas).ToNot(BeEmpty())

			schema, err := dbn.SchemaFromString(schemas[0])
			Expect(err).To(BeNil())

			fields, err := ListFields(databentoApiKey, dbn.Encoding_Dbn, schema)
			Expect(err).To(BeNil())
			Expect(fields).ToNot(BeEmpty())

			unitPrices, err := ListUnitPrices(databentoApiKey, datasets[0])
			Expect(err).To(BeNil())
			Expect(unitPrices).ToNot(BeEmpty())
		})
		It("should ResolveSymbology", func() {
			resolveParams := ResolveParams{
				Dataset:  "XNAS.ITCH",
				Symbols:  []string{"AAPL"},
				StypeIn:  dbn.SType_RawSymbol,
				StypeOut: dbn.SType_InstrumentId,
				DateRange: DateRange{
					Start: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			}
			resolveResp, err := SymbologyResolve(databentoApiKey, resolveParams)
			Expect(err).To(BeNil())
			Expect(resolveResp).ToNot(BeNil())
			Expect(resolveResp.Mappings).ToNot(BeEmpty())
		})
	})
})
