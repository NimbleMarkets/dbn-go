// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"bytes"

	"github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// shortRecordStream returns a V3 DBN stream holding a full-size record of
// rtype filled with 0xAB, followed by a record of the same rtype whose header
// says it is only shortLen bytes. The scanner reuses one buffer for every
// record, so the short record's missing bytes are leftovers from the first.
func shortRecordStream(rtype dbn.RType, fullLen int, shortLen int) *bytes.Buffer {
	var buf bytes.Buffer
	meta := dbn.Metadata{VersionNum: dbn.HeaderVersion3, Dataset: "TEST.DATA",
		StypeIn: dbn.SType_RawSymbol, StypeOut: dbn.SType_InstrumentId}
	Expect(meta.Write(&buf)).To(Succeed())

	writeRecord := func(length int) {
		record := bytes.Repeat([]byte{0xAB}, length)
		record[0] = byte(length / 4)
		record[1] = byte(rtype)
		buf.Write(record)
	}
	writeRecord(fullLen)
	writeRecord(shortLen)
	return &buf
}

var _ = Describe("DbnScanner short records", func() {
	DescribeTable("Visit should reject a record shorter than its type, not fill it from stale bytes",
		func(rtype dbn.RType, fullLen int) {
			const shortLen = 32
			Expect(shortLen).To(BeNumerically("<", fullLen))
			scanner := dbn.NewDbnScanner(shortRecordStream(rtype, fullLen, shortLen))
			visitor := &dbn.NullVisitor{}

			Expect(scanner.Next()).To(BeTrue())
			Expect(scanner.Visit(visitor)).To(Succeed())

			Expect(scanner.Next()).To(BeTrue())
			Expect(scanner.GetLastSize()).To(Equal(shortLen))
			Expect(func() {
				Expect(scanner.Visit(visitor)).ToNot(Succeed())
			}).ToNot(Panic())
		},
		Entry("Mbp0", dbn.RType_Mbp0, dbn.Mbp0Msg_Size),
		Entry("Mbp1", dbn.RType_Mbp1, dbn.Mbp1Msg_Size),
		Entry("Mbp10", dbn.RType_Mbp10, dbn.Mbp10Msg_Size),
		Entry("Mbo", dbn.RType_Mbo, dbn.MboMsg_Size),
		Entry("Ohlcv", dbn.RType_Ohlcv1S, dbn.OhlcvMsg_Size),
		Entry("Cmbp1", dbn.RType_Cmbp1, dbn.Cmbp1Msg_Size),
		Entry("Imbalance", dbn.RType_Imbalance, dbn.ImbalanceMsg_Size),
		Entry("Status", dbn.RType_Status, dbn.StatusMsg_Size),
		Entry("Bbo", dbn.RType_Bbo1S, dbn.BboMsg_Size),
		Entry("Statistics", dbn.RType_Statistics, dbn.StatMsgV3_Size),
	)

	It("DecodeStatMsg should reject a short statistics record", func() {
		scanner := dbn.NewDbnScanner(shortRecordStream(dbn.RType_Statistics, dbn.StatMsgV3_Size, 32))
		Expect(scanner.Next()).To(BeTrue())
		_, err := scanner.DecodeStatMsg()
		Expect(err).To(Succeed())

		Expect(scanner.Next()).To(BeTrue())
		Expect(func() {
			record, err := scanner.DecodeStatMsg()
			Expect(err).To(HaveOccurred())
			Expect(record).To(BeNil())
		}).ToNot(Panic())
	})

	It("DecodeStatMsg and DecodeInstrumentDefMsg should return ErrNoMetadata on nil metadata", func() {
		stat, err := dbn.DecodeStatMsg(nil, make([]byte, dbn.StatMsgV3_Size))
		Expect(err).To(MatchError(dbn.ErrNoMetadata))
		Expect(stat).To(BeNil())

		def, err := dbn.DecodeInstrumentDefMsg(nil, make([]byte, dbn.InstrumentDefMsgV3_Size), dbn.InstrumentDefMsgV3_Size)
		Expect(err).To(MatchError(dbn.ErrNoMetadata))
		Expect(def).To(BeNil())
	})
})
