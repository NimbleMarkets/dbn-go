// Copyright (c) 2024 Neomantra Corp

package dbn_test

import (
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
			Expect(unsafe.Sizeof(dbn.Cmbp1Msg{})).To(Equal(uintptr(dbn.Cmbp1Msg_Size)))
			Expect(unsafe.Sizeof(dbn.OhlcvMsg{})).To(Equal(uintptr(dbn.OhlcvMsg_Size)))
			Expect(unsafe.Sizeof(dbn.ImbalanceMsg{})).To(Equal(uintptr(dbn.ImbalanceMsg_Size)))
			Expect(unsafe.Sizeof(dbn.ErrorMsg{})).To(Equal(uintptr(dbn.ErrorMsg_Size)))
			Expect(unsafe.Sizeof(dbn.SystemMsg{})).To(Equal(uintptr(dbn.SystemMsg_Size)))
			Expect(unsafe.Sizeof(dbn.StatMsg{})).To(Equal(uintptr(dbn.StatMsg_Size)))
			Expect(unsafe.Sizeof(dbn.StatusMsg{})).To(Equal(uintptr(dbn.StatusMsg_Size)))
			Expect(unsafe.Sizeof(dbn.BboMsg{})).To(Equal(uintptr(dbn.BboMsg_Size)))
			Expect(unsafe.Sizeof(dbn.InstrumentDefMsg{})).To(Equal(uintptr(dbn.InstrumentDefMsg_Size)))
			Expect(int((&dbn.RHeader{}).RSize())).To(Equal(dbn.RHeader_Size))
			Expect(int((&dbn.Mbp0Msg{}).RSize())).To(Equal(dbn.Mbp0Msg_Size))
			Expect(int((&dbn.Mbp1Msg{}).RSize())).To(Equal(dbn.Mbp1Msg_Size))
			Expect(int((&dbn.Mbp10Msg{}).RSize())).To(Equal(dbn.Mbp10Msg_Size))
			Expect(int((&dbn.Cmbp1Msg{}).RSize())).To(Equal(dbn.Cmbp1Msg_Size))
			Expect(int((&dbn.OhlcvMsg{}).RSize())).To(Equal(dbn.OhlcvMsg_Size))
			Expect(int((&dbn.ImbalanceMsg{}).RSize())).To(Equal(dbn.ImbalanceMsg_Size))
			Expect(int((&dbn.ErrorMsg{}).RSize())).To(Equal(dbn.ErrorMsg_Size))
			Expect(int((&dbn.StatMsg{}).RSize())).To(Equal(dbn.StatMsg_Size))
			Expect(int((&dbn.StatusMsg{}).RSize())).To(Equal(dbn.StatusMsg_Size))
			Expect(int((&dbn.BboMsg{}).RSize())).To(Equal(dbn.BboMsg_Size))
			Expect(int((&dbn.InstrumentDefMsg{}).RSize())).To(Equal(dbn.InstrumentDefMsg_Size))
		})
	})
})
