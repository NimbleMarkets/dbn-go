// Copyright (c) 2024 Neomantra Corp

package dbn_live

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test Launcher
func TestDbnLive(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dbn-go live suite")
}

var _ = Describe("DbnLive", func() {
	Context("auth", func() {
		It("should generate CRAM response properly", func() {
			// https://databento.com/docs/api-reference-live/message-flows/authentication/example?historical=http&live=raw
			apiKey := "db-89s9vCvwDDKPdQJ5Pb30Fyj9mNUM6"
			cram := "j5pwMHz6vwXruJM4cOwQrQeQE0bImIzT"
			expected := "6d3c875bb9f8cf503c3ed83ee5f476a3ad53f0c67706c51cf42d2db5ad8ff5a9-mNUM6"

			resp := generateCramReply(apiKey, cram)
			Expect(resp).To(Equal(expected))
		})
	})
})
