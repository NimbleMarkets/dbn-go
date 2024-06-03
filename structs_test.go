// Copyright (c) 2024 Neomantra Corp

package dbn_test

import (
	"os"
	"unsafe"

	dbn "github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Struct", func() {
	Context("correctness", func() {
		It("struct consts should be correct", func() {
			Expect(unsafe.Sizeof(dbn.RHeader{})).To(Equal(uintptr(dbn.RHeader_Size)))
			Expect(unsafe.Sizeof(dbn.BidAskPair{})).To(Equal(uintptr(dbn.BidAskPair_Size)))
			Expect(unsafe.Sizeof(dbn.Mbp0Msg{})).To(Equal(uintptr(dbn.Mbp0Msg_Size)))
			Expect(unsafe.Sizeof(dbn.Mbp1Msg{})).To(Equal(uintptr(dbn.Mbp1Msg_Size)))
			Expect(unsafe.Sizeof(dbn.Mbp10Msg{})).To(Equal(uintptr(dbn.Mbp10Msg_Size)))
			Expect(unsafe.Sizeof(dbn.OhlcvMsg{})).To(Equal(uintptr(dbn.OhlcvMsg_Size)))
			Expect(unsafe.Sizeof(dbn.ImbalanceMsg{})).To(Equal(uintptr(dbn.ImbalanceMsg_Size)))
			Expect(unsafe.Sizeof(dbn.ErrorMsg{})).To(Equal(uintptr(dbn.ErrorMsg_Size)))
			Expect(unsafe.Sizeof(dbn.SystemMsg{})).To(Equal(uintptr(dbn.SystemMsg_Size)))
			Expect(unsafe.Sizeof(dbn.StatMsg{})).To(Equal(uintptr(dbn.StatMsg_Size)))
			Expect(int((&dbn.RHeader{}).RSize())).To(Equal(dbn.RHeader_Size))
			Expect(int((&dbn.Mbp0Msg{}).RSize())).To(Equal(dbn.Mbp0Msg_Size))
			Expect(int((&dbn.Mbp1Msg{}).RSize())).To(Equal(dbn.Mbp1Msg_Size))
			Expect(int((&dbn.Mbp10Msg{}).RSize())).To(Equal(dbn.Mbp10Msg_Size))
			Expect(int((&dbn.OhlcvMsg{}).RSize())).To(Equal(dbn.OhlcvMsg_Size))
			Expect(int((&dbn.ImbalanceMsg{}).RSize())).To(Equal(dbn.ImbalanceMsg_Size))
			Expect(int((&dbn.ErrorMsg{}).RSize())).To(Equal(dbn.ErrorMsg_Size))
			Expect(int((&dbn.StatMsg{}).RSize())).To(Equal(dbn.StatMsg_Size))
		})
	})

	Context("v1 file testing", func() {
		It("should read v1 olhc-1s correctly", func() {
			file, err := os.Open("./tests/data/test_data.ohlcv-1s.v1.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			// dbn -J ./tests/data/test_data.ohlcv-1s.v1.dbn
			// {"hd":{"ts_event":"1609160400000000000","rtype":32,"publisher_id":1,"instrument_id":5482},"open":"372025000000000","high":"372050000000000","low":"372025000000000","close":"372050000000000","volume":"57"}
			// {"hd":{"ts_event":"1609160401000000000","rtype":32,"publisher_id":1,"instrument_id":5482},"open":"372050000000000","high":"372050000000000","low":"372050000000000","close":"372050000000000","volume":"13"}

			records, metadata, err := dbn.ReadDBNToSlice[dbn.OhlcvMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609160400000000000)))
			Expect(r0h.RType).To(Equal(dbn.RType(32)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r0.Open).To(Equal(int64(372025000000000)))
			Expect(r0.High).To(Equal(int64(372050000000000)))
			Expect(r0.Low).To(Equal(int64(372025000000000)))
			Expect(r0.Close).To(Equal(int64(372050000000000)))
			Expect(r0.Volume).To(Equal(uint64(57)))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609160401000000000)))
			Expect(r1h.RType).To(Equal(dbn.RType(32)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r1.Open).To(Equal(int64(372050000000000)))
			Expect(r1.High).To(Equal(int64(372050000000000)))
			Expect(r1.Low).To(Equal(int64(372050000000000)))
			Expect(r1.Close).To(Equal(int64(372050000000000)))
			Expect(r1.Volume).To(Equal(uint64(13)))
		})
	})

	Context("v2 files", func() {
		It("should read a v2 ohlc-1s correctly", func() {
			file, err := os.Open("./tests/data/test_data.ohlcv-1s.dbn")
			Expect(err).To(BeNil())
			defer file.Close()

			records, metadata, err := dbn.ReadDBNToSlice[dbn.OhlcvMsg](file)
			Expect(err).To(BeNil())
			Expect(metadata).ToNot(BeNil())
			Expect(len(records)).To(Equal(2))

			// dbn -J ./tests/data/test_data.ohlcv-1s.dbn
			// {"hd":{"ts_event":"1609160400000000000","rtype":32,"publisher_id":1,"instrument_id":5482},"open":"372025000000000","high":"372050000000000","low":"372025000000000","close":"372050000000000","volume":"57"}
			// {"hd":{"ts_event":"1609160401000000000","rtype":32,"publisher_id":1,"instrument_id":5482},"open":"372050000000000","high":"372050000000000","low":"372050000000000","close":"372050000000000","volume":"13"}

			r0, r0h := records[0], records[0].Header
			Expect(r0h.TsEvent).To(Equal(uint64(1609160400000000000)))
			Expect(r0h.RType).To(Equal(dbn.RType(32)))
			Expect(r0h.PublisherID).To(Equal(uint16(1)))
			Expect(r0h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r0.Open).To(Equal(int64(372025000000000)))
			Expect(r0.High).To(Equal(int64(372050000000000)))
			Expect(r0.Low).To(Equal(int64(372025000000000)))
			Expect(r0.Close).To(Equal(int64(372050000000000)))
			Expect(r0.Volume).To(Equal(uint64(57)))

			r1, r1h := records[1], records[1].Header
			Expect(r1h.TsEvent).To(Equal(uint64(1609160401000000000)))
			Expect(r1h.RType).To(Equal(dbn.RType(32)))
			Expect(r1h.PublisherID).To(Equal(uint16(1)))
			Expect(r1h.InstrumentID).To(Equal(uint32(5482)))
			Expect(r1.Open).To(Equal(int64(372050000000000)))
			Expect(r1.High).To(Equal(int64(372050000000000)))
			Expect(r1.Low).To(Equal(int64(372050000000000)))
			Expect(r1.Close).To(Equal(int64(372050000000000)))
			Expect(r1.Volume).To(Equal(uint64(13)))
		})
	})
})
