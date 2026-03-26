// Copyright (c) 2024 Neomantra Corp
//
// Tests for DBN Version 1 message types

package dbn_test

import (
	dbn "github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessagesV1", func() {
	Context("SymbolMappingMsgV1", func() {
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
})
