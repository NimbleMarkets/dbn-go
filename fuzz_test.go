// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"reflect"
	"testing"

	dbn "github.com/NimbleMarkets/dbn-go"
)

// Keep generated inputs small enough for local fuzzing and CI. Header-declared
// lengths and counts are deliberately not capped or repaired before decoding.
const maxFuzzInput = 64 * 1024

type fuzzRecord interface {
	Fill_Raw([]byte) error
	RSize() uint16
}

var fuzzRecordFactories = []func() fuzzRecord{
	func() fuzzRecord { return &dbn.RHeader{} },
	func() fuzzRecord { return &dbn.Mbp0Msg{} },
	func() fuzzRecord { return &dbn.MboMsg{} },
	func() fuzzRecord { return &dbn.Mbp1Msg{} },
	func() fuzzRecord { return &dbn.Mbp10Msg{} },
	func() fuzzRecord { return &dbn.Cmbp1Msg{} },
	func() fuzzRecord { return &dbn.BboMsg{} },
	func() fuzzRecord { return &dbn.OhlcvMsg{} },
	func() fuzzRecord { return &dbn.ImbalanceMsg{} },
	func() fuzzRecord { return &dbn.StatusMsg{} },
	func() fuzzRecord { return &dbn.StatMsgV2{} },
	func() fuzzRecord { return &dbn.StatMsgV3{} },
	func() fuzzRecord { return &dbn.InstrumentDefMsgV2{} },
	func() fuzzRecord { return &dbn.InstrumentDefMsgV3{} },
	func() fuzzRecord { return &dbn.ErrorMsgV1{} },
	func() fuzzRecord { return &dbn.ErrorMsgV2{} },
	func() fuzzRecord { return &dbn.SystemMsg{} },
	func() fuzzRecord { return &dbn.SymbolMappingMsgV1{} },
	func() fuzzRecord { return &dbn.SymbolMappingMsgV2{} },
}

func FuzzRecordDecode(f *testing.F) {
	for i, factory := range fuzzRecordFactories {
		size := int(factory().RSize())
		raw := make([]byte, size+8) // includes a possible ts_out suffix
		for j := range raw {
			raw[j] = byte(j*37 + 11)
		}
		raw[0] = byte(size / 4)
		for _, n := range []int{0, 1, 15, size - 1, size, size + 8} {
			f.Add(uint8(i), raw[:n])
		}
	}
	f.Fuzz(func(t *testing.T, kind uint8, raw []byte) {
		if len(raw) > 1024 {
			t.Skip()
		}
		factory := fuzzRecordFactories[int(kind)%len(fuzzRecordFactories)]
		record := factory()
		before := bytes.Clone(raw)
		err := record.Fill_Raw(raw) // Panics are failures, never recovered or hidden.
		if !bytes.Equal(raw, before) {
			t.Fatal("decoder mutated input")
		}
		size := int(record.RSize())
		if len(raw) < size {
			if err == nil {
				t.Fatalf("%T accepted %d bytes; needs %d", record, len(raw), size)
			}
			return
		}
		if err != nil {
			return
		}
		// Bytes after the fixed record, including ts_out, must not change its fields.
		exact := factory()
		if err := exact.Fill_Raw(raw[:size]); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(record, exact) {
			t.Fatalf("%T decoded trailing bytes as record fields", record)
		}
	})
}

func FuzzReadMetadata(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("DBN\x03\xff\xff\xff\xff"))
	// Real upstream fixtures cover the supported metadata versions and mappings.
	for _, name := range []string{
		"tests/data/test_data.ohlcv-1s.v1.dbn",
		"tests/data/test_data.bbo-1m.dbn",
		"tests/data/test_data.mbo.v3.dbn",
	} {
		raw, err := os.ReadFile(name)
		if err != nil {
			f.Fatal(err)
		}
		if _, err := dbn.ReadMetadata(bytes.NewReader(raw)); err != nil {
			f.Fatalf("seed %s: %v", name, err)
		}
		f.Add(raw)
		f.Add(raw[:7])
		f.Add(raw[:len(raw)/2])
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > maxFuzzInput {
			t.Skip()
		}
		before := bytes.Clone(raw)
		reader := bytes.NewReader(raw)
		metadata, err := dbn.ReadMetadata(reader)
		if !bytes.Equal(raw, before) {
			t.Fatal("metadata decoder mutated input")
		}
		if err != nil {
			return
		}
		if metadata == nil {
			t.Fatal("successful decode returned nil metadata")
		}
		if len(raw) < 8 {
			t.Fatal("accepted incomplete prefix")
		}
		if metadata.VersionNum < 1 || metadata.VersionNum > 3 {
			t.Fatalf("accepted version %d", metadata.VersionNum)
		}
		want := uint64(8) + uint64(binary.LittleEndian.Uint32(raw[4:8]))
		if got := uint64(len(raw) - reader.Len()); got != want {
			t.Fatalf("consumed %d bytes, declared %d", got, want)
		}
	})
}

// A small per-timestamp oracle checks arbitrary overlapping insertions, gaps,
// and coalescing without duplicating the interval implementation.
func FuzzSymbolMapIntervals(f *testing.F) {
	f.Add([]byte{0, 0, 32, 1, 0, 8, 24, 2, 0, 4, 28, 1})
	f.Add([]byte{1, 63, 64, 3, 1, 0, 63, 3, 2, 10, 10, 1, 2, 20, 10, 2})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 1024 {
			t.Skip()
		}
		var m dbn.TsSymbolMap
		var want [4][64]string
		for i := 0; i+4 <= len(operations); i += 4 {
			id, start, end := operations[i]%4, operations[i+1]%65, operations[i+2]%65
			symbol := string(rune('A' + operations[i+3]%4))
			err := m.InsertTs(uint32(id), uint64(start), uint64(end), symbol)
			if start > end {
				if err == nil {
					t.Fatal("accepted reversed range")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				for ts := start; ts < end; ts++ {
					want[id][ts] = symbol
				}
			}
			runs := 0
			for instrument, points := range want {
				previous := ""
				for ts, symbol := range points {
					if got := m.GetForTs(uint64(ts), uint32(instrument)); got != symbol {
						t.Fatalf("instrument %d timestamp %d: %q != %q", instrument, ts, got, symbol)
					}
					if symbol != "" && symbol != previous {
						runs++
					}
					previous = symbol
				}
				if m.GetForTs(64, uint32(instrument)) != "" {
					t.Fatal("matched exclusive end")
				}
			}
			if m.Len() != runs {
				t.Fatalf("%d intervals, want %d", m.Len(), runs)
			}
			if m.IsEmpty() != (runs == 0) {
				t.Fatal("incorrect empty status")
			}
		}
	})
}
