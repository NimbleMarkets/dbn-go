// Copyright (c) 2024 Neomantra Corp

package dbn_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"unsafe"

	"github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metadata", func() {
	Context("correctness", func() {
		It("metadata sizes should be correct", func() {
			Expect(unsafe.Sizeof(dbn.RType_Error)).To(Equal(uintptr(1)))
			Expect(unsafe.Sizeof(dbn.SType_RawSymbol)).To(Equal(uintptr(1)))
			Expect(unsafe.Sizeof(dbn.Schema_Mixed)).To(Equal(uintptr(2)))
			Expect(unsafe.Sizeof(dbn.MetadataPrefix{})).To(Equal(uintptr(dbn.Metadata_PrefixSize)))
			Expect(unsafe.Sizeof(dbn.MetadataHeaderV1{})).To(Equal(uintptr(dbn.MetadataHeaderV1_Size + dbn.MetadataHeaderV1_SizeFuzz)))
			Expect(unsafe.Sizeof(dbn.MetadataHeaderV2{})).To(Equal(uintptr(dbn.MetadataHeaderV2_Size + dbn.MetadataHeaderV2_SizeFuzz)))

			// If this changes, we need to update offsets in metadata.go
			Expect(dbn.Metadata_DatasetCstrLen).To(Equal(16))
		})
	})
	Context("reading", func() {
		It("we should decode v1 metadata properly", func() {
			file, err := os.Open("./tests/data/test_data.ohlcv-1s.v1.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			// dbn -J -m ./tests/data/test_data.ohlcv-1s.v1.dbn
			// "version":1,"dataset":"GLBX.MDP3","schema":"ohlcv-1s","start":"1609160400000000000",
			// "end":"1609200000000000000","limit":"2","stype_in":"raw_symbol","stype_out":"instrument_id",
			// "ts_out":false,"symbol_cstr_len":22,"symbols":["ESH1"],"partial":[],"not_found":[],"mappings":[{"raw_symbol":"ESH1",
			// "intervals":[{"start_date":20201228,"end_date":20201229,"symbol":"5482"}]}]}
			m1, err := dbn.ReadMetadata(file)
			Expect(err).To(BeNil())
			Expect(m1).ToNot(BeNil())
			Expect(m1.VersionNum).To(Equal(uint8(dbn.HeaderVersion1)))
			Expect(m1.Schema).To(Equal(dbn.Schema_Ohlcv1S))
			Expect(m1.Start).To(Equal(uint64(1609160400000000000)))
			Expect(m1.End).To(Equal(uint64(1609200000000000000)))
			Expect(m1.Limit).To(Equal(uint64(2)))
			Expect(m1.StypeIn).To(Equal(dbn.SType_RawSymbol))
			Expect(m1.StypeOut).To(Equal(dbn.SType_InstrumentId))
			Expect(m1.TsOut).To(Equal(uint8(0)))
			Expect(m1.Dataset).To(Equal("GLBX.MDP3"))
			Expect(m1.SymbolCstrLen).To(Equal(uint16(dbn.MetadataV1_SymbolCstrLen)))
			Expect(len(m1.Symbols)).To(Equal(1))
			Expect(m1.Symbols[0]).To(Equal("ESH1"))
			Expect(len(m1.Partial)).To(Equal(0))
			Expect(len(m1.NotFound)).To(Equal(0))
			Expect(len(m1.Mappings)).To(Equal(1))
			Expect(m1.Mappings[0].RawSymbol).To(Equal("ESH1"))
			intervals := m1.Mappings[0].Intervals
			Expect(len(intervals)).To(Equal(1))
			Expect(intervals[0].StartDate).To(Equal(uint32(20201228)))
			Expect(intervals[0].EndDate).To(Equal(uint32(20201229)))
			Expect(intervals[0].Symbol).To(Equal("5482"))
		})
		It("we should decode v2 metadata properly", func() {
			file, err := os.Open("./tests/data/test_data.ohlcv-1s.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			// dbn -J -m ./tests/data/test_data.ohlcv-1s.dbn
			// {"version":2,"dataset":"GLBX.MDP3","schema":"ohlcv-1s","start":"1609160400000000000",
			// "end":"1609200000000000000","limit":"2","stype_in":"raw_symbol","stype_out":"instrument_id",
			// "ts_out":false,"symbol_cstr_len":71,"symbols":["ESH1"],"partial":[],"not_found":[],"mappings":[{"raw_symbol":"ESH1",
			// "intervals":[{"start_date":20201228,"end_date":20201229,"symbol":"5482"}]}]
			m2, err := dbn.ReadMetadata(file)
			Expect(err).To(BeNil())
			Expect(m2).ToNot(BeNil())
			Expect(m2.VersionNum).To(Equal(uint8(dbn.HeaderVersion2)))
			Expect(m2.Schema).To(Equal(dbn.Schema_Ohlcv1S))
			Expect(m2.Start).To(Equal(uint64(1609160400000000000)))
			Expect(m2.End).To(Equal(uint64(1609200000000000000)))
			Expect(m2.Limit).To(Equal(uint64(2)))
			Expect(m2.StypeIn).To(Equal(dbn.SType_RawSymbol))
			Expect(m2.StypeOut).To(Equal(dbn.SType_InstrumentId))
			Expect(m2.TsOut).To(Equal(uint8(0)))
			Expect(m2.Dataset).To(Equal("GLBX.MDP3"))
			Expect(m2.SymbolCstrLen).To(Equal(uint16(dbn.MetadataV2_SymbolCstrLen)))
			Expect(len(m2.Symbols)).To(Equal(1))
			Expect(m2.Symbols[0]).To(Equal("ESH1"))
			Expect(len(m2.Partial)).To(Equal(0))
			Expect(len(m2.NotFound)).To(Equal(0))
			Expect(len(m2.Mappings)).To(Equal(1))
			Expect(m2.Mappings[0].RawSymbol).To(Equal("ESH1"))
			intervals := m2.Mappings[0].Intervals
			Expect(len(intervals)).To(Equal(1))
			Expect(intervals[0].StartDate).To(Equal(uint32(20201228)))
			Expect(intervals[0].EndDate).To(Equal(uint32(20201229)))
			Expect(intervals[0].Symbol).To(Equal("5482"))
		})
	})
	Context("writing", func() {
		It("should round-trip metadata and records", func() {
			// Create a buffer to write to
			buf := &bytes.Buffer{}

			// Create metadata for the DBN file
			originalMetadata := &dbn.Metadata{
				VersionNum:    dbn.HeaderVersion2,
				Schema:        dbn.Schema_Ohlcv1S,
				Start:         1609160400000000000,
				End:           1609200000000000000,
				Limit:         0,
				StypeIn:       dbn.SType_RawSymbol,
				StypeOut:      dbn.SType_InstrumentId,
				TsOut:         0,
				SymbolCstrLen: dbn.MetadataV2_SymbolCstrLen,
				Dataset:       "TEST.DATA",
				Symbols:       []string{"AAPL"},
				Partial:       []string{},
				NotFound:      []string{},
			}

			// Write the metadata header
			err := originalMetadata.Write(buf)
			Expect(err).To(BeNil())

			// Write a test record
			record := &dbn.OhlcvMsg{
				Header: dbn.RHeader{
					Length:       dbn.OhlcvMsg_Size / 4,  // Size in 32-bit words
					RType:        dbn.RType_OhlcvEod,
					PublisherID:  1,
					InstrumentID: 12345,
					TsEvent:      1609160400000000000,
				},
				Open:   10000,
				High:   10100,
				Low:    9900,
				Close:  10050,
				Volume: 1000000,
			}

			// Write the record (just write the entire struct with binary.Write)
			err = binary.Write(buf, binary.LittleEndian, record)
			Expect(err).To(BeNil())

			// Now read it back using a scanner which handles both metadata and records
			reader := bytes.NewReader(buf.Bytes())
			scanner := dbn.NewDbnScanner(reader)

			// Read metadata from scanner
			readMetadata, err := scanner.Metadata()
			Expect(err).To(BeNil())
			Expect(readMetadata).ToNot(BeNil())

			// Verify metadata matches
			Expect(readMetadata.VersionNum).To(Equal(originalMetadata.VersionNum))
			Expect(readMetadata.Schema).To(Equal(originalMetadata.Schema))
			Expect(readMetadata.Start).To(Equal(originalMetadata.Start))
			Expect(readMetadata.End).To(Equal(originalMetadata.End))
			Expect(readMetadata.Limit).To(Equal(originalMetadata.Limit))
			Expect(readMetadata.StypeIn).To(Equal(originalMetadata.StypeIn))
			Expect(readMetadata.StypeOut).To(Equal(originalMetadata.StypeOut))
			Expect(readMetadata.TsOut).To(Equal(originalMetadata.TsOut))
			Expect(readMetadata.Dataset).To(Equal(originalMetadata.Dataset))
			Expect(readMetadata.Symbols).To(Equal(originalMetadata.Symbols))

			// Read the record back
			Expect(scanner.Next()).To(BeTrue())

			// Decode the record
			decodedRecord, err := dbn.DbnScannerDecode[dbn.OhlcvMsg](scanner)
			Expect(err).To(BeNil())
			Expect(decodedRecord).ToNot(BeNil())

			// Verify record matches
			Expect(decodedRecord.Header.RType).To(Equal(record.Header.RType))
			Expect(decodedRecord.Header.PublisherID).To(Equal(record.Header.PublisherID))
			Expect(decodedRecord.Header.InstrumentID).To(Equal(record.Header.InstrumentID))
			Expect(decodedRecord.Header.TsEvent).To(Equal(record.Header.TsEvent))
			Expect(decodedRecord.Open).To(Equal(record.Open))
			Expect(decodedRecord.High).To(Equal(record.High))
			Expect(decodedRecord.Low).To(Equal(record.Low))
			Expect(decodedRecord.Close).To(Equal(record.Close))
			Expect(decodedRecord.Volume).To(Equal(record.Volume))
		})
	})
})
