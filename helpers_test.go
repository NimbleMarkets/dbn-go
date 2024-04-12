// Copyright (c) 2024 Neomantra Corp

package dbn_test

import (
	"time"

	"github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helpers", func() {
	Context("conversion", func() {
		It("converts fixed9 to float correctly", func() {
			Expect(dbn.Fixed9ToFloat64(1234567890123456789)).To(Equal(float64(1234567890.123456789)))
		})
		It("converts timestamp to sec, nanos correctly", func() {
			sec, nanos := dbn.TimestampToSecNanos(1234567890123456789)
			Expect(sec).To(Equal(int64(1234567890)))
			Expect(nanos).To(Equal(int64(123456789)))
		})
		It("converts Times to Time correctly", func() {
			Expect(dbn.TimestampToTime(0).UTC()).To(Equal(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)))
			Expect(dbn.TimestampToTime(1234567890123456789).UTC()).To(Equal(time.Date(2009, 02, 13, 23, 31, 30, 123456789, time.UTC)))
		})
		It("converts Times to YMD correctly", func() {
			Expect(dbn.TimeToYMD(time.Time{})).To(Equal(uint32(0)))
			Expect(dbn.TimeToYMD(time.Date(2024, 04, 12, 0, 0, 0, 0, time.UTC))).To(Equal(uint32(20240412)))
		})
	})
	Context("modification", func() {
		It("trims null bytes correctly", func() {
			Expect(dbn.TrimNullBytes([]byte("hello\x00\x00\x00\x00"))).To(Equal("hello"))
		})
		It("does not malform regular strings", func() {
			Expect(dbn.TrimNullBytes([]byte("hello"))).To(Equal("hello"))
		})
	})
})
