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

		// Round-trip: Write -> ReadMetadata must produce an identical Metadata struct.
		var buf bytes.Buffer
		if err := metadata.Write(&buf); err != nil {
			t.Fatalf("metadata.Write failed: %v", err)
		}
		if metadata.VersionNum == dbn.HeaderVersion3 && buf.Len()%8 != 0 {
			t.Fatalf("V3 metadata write size %d is not 8-byte aligned", buf.Len())
		}
		roundTrip, err := dbn.ReadMetadata(&buf)
		if err != nil {
			t.Fatalf("round-trip ReadMetadata failed: %v", err)
		}
		if buf.Len() != 0 {
			t.Fatalf("round-trip left %d unread bytes", buf.Len())
		}
		if !reflect.DeepEqual(metadata, roundTrip) {
			t.Fatalf("round-trip metadata mismatch:\ngot:  %#v\nwant: %#v", roundTrip, metadata)
		}

		// Exercise symbol-map builders on decoded metadata.
		var tsm dbn.TsSymbolMap
		_ = tsm.FillFromMetadata(metadata)
		var psm dbn.PitSymbolMap
		_ = psm.FillFromMetadata(metadata, metadata.Start)
	})
}

func FuzzDbnScanner(f *testing.F) {
	f.Add([]byte{})
	for _, name := range []string{
		"tests/data/test_data.ohlcv-1s.v1.dbn",
		"tests/data/test_data.bbo-1m.dbn",
		"tests/data/test_data.mbo.v3.dbn",
		"tests/data/test_data.definition.dbn",
		"tests/data/test_data.imbalance.dbn",
		"tests/data/test_data.mbp-1.dbn",
		"tests/data/test_data.trades.dbn",
		"tests/data/test_data.statistics.v1.dbn.zst",
		"tests/data/test_data.statistics.v2.dbn.zst",
		"tests/data/test_data.statistics.v3.dbn.zst",
		"tests/data/test_data.definition.v2.dbn.zst",
		"tests/data/test_data.definition.v3.dbn.zst",
		"tests/data/test_data.status.v2.dbn.zst",
	} {
		reader, closer, err := dbn.MakeCompressedReader(name, false)
		if err != nil {
			f.Fatal(err)
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(reader); err != nil {
			_ = closer.Close()
			f.Fatal(err)
		}
		_ = closer.Close()
		raw := buf.Bytes()
		f.Add(raw)
		if len(raw) > 24 {
			f.Add(raw[:len(raw)-5])
		}
	}

	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > maxFuzzInput {
			t.Skip()
		}
		before := bytes.Clone(raw)
		scanner := dbn.NewDbnScanner(bytes.NewReader(raw))
		var tsm dbn.TsSymbolMap
		const maxRecords = 256
		for count := 0; count < maxRecords && scanner.Next(); count++ {
			if scanner.GetLastSize() < dbn.RHeader_Size {
				t.Fatalf("Next() succeeded with lastSize=%d < %d", scanner.GetLastSize(), dbn.RHeader_Size)
			}
			hdr, err := scanner.GetLastHeader()
			if err != nil {
				t.Fatalf("GetLastHeader failed after Next(): %v", err)
			}
			if int(hdr.Length)*4 != scanner.GetLastSize() {
				t.Fatalf("header length %d words != lastSize %d", hdr.Length, scanner.GetLastSize())
			}
			_ = scanner.Visit(&dbn.NullVisitor{})
			_ = scanner.Visit(&tsm)
			_, _ = scanner.DecodeErrorMsg()
			_, _ = scanner.DecodeSystemMsg()
			_, _ = scanner.DecodeSymbolMappingMsg()
			_, _ = scanner.DecodeStatMsg()
			_, _ = scanner.DecodeInstrumentDefMsg()
			_, _ = dbn.DbnScannerDecode[dbn.Mbp0Msg](scanner)
		}
		// When stopped by EOF or a malformed/short record, decoders must not succeed on stale bytes.
		if scanner.GetLastSize() <= dbn.RHeader_Size || scanner.Error() != nil {
			if err := scanner.Visit(&dbn.NullVisitor{}); err == nil {
				t.Fatal("Visit succeeded after failed/short Next()")
			}
			if _, err := scanner.DecodeErrorMsg(); err == nil {
				t.Fatal("DecodeErrorMsg succeeded after failed/short Next()")
			}
			if _, err := scanner.DecodeSystemMsg(); err == nil {
				t.Fatal("DecodeSystemMsg succeeded after failed/short Next()")
			}
			if _, err := scanner.DecodeSymbolMappingMsg(); err == nil {
				t.Fatal("DecodeSymbolMappingMsg succeeded after failed/short Next()")
			}
			if _, err := scanner.DecodeStatMsg(); err == nil {
				t.Fatal("DecodeStatMsg succeeded after failed/short Next()")
			}
			if _, err := scanner.DecodeInstrumentDefMsg(); err == nil {
				t.Fatal("DecodeInstrumentDefMsg succeeded after failed/short Next()")
			}
			if _, err := dbn.DbnScannerDecode[dbn.OhlcvMsg](scanner); err == nil {
				t.Fatal("DbnScannerDecode succeeded after failed/short Next()")
			}
		}
		if !bytes.Equal(raw, before) {
			t.Fatal("DbnScanner mutated input")
		}
	})
}

var fuzzJsonDecoders = []func(*dbn.JsonScanner) error{
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.Mbp0Msg](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.MboMsg](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.Mbp1Msg](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.Mbp10Msg](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.Cmbp1Msg](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.BboMsg](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.OhlcvMsg](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.ImbalanceMsg](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.StatusMsg](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.StatMsgV2](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.StatMsgV3](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.InstrumentDefMsgV2](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.InstrumentDefMsgV3](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.ErrorMsgV1](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.ErrorMsgV2](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.SystemMsg](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.SymbolMappingMsgV1](s); return err },
	func(s *dbn.JsonScanner) error { _, err := dbn.JsonScannerDecode[dbn.SymbolMappingMsgV2](s); return err },
}

func FuzzJsonScanner(f *testing.F) {
	for i, seed := range []string{
		"",
		"{}\n",
		"{\"hd\":null}\n",
		"{\"hd\":{\"rtype\":1},\"levels\":[]}\n",
		"{\"hd\":{\"rtype\":1},\"levels\":[null]}\n",
		"{\"ts_recv\":\"1704186000404085841\",\"hd\":{\"ts_event\":\"1704186000403918695\",\"rtype\":0,\"publisher_id\":2,\"instrument_id\":15144},\"action\":\"T\",\"side\":\"B\",\"depth\":0,\"price\":\"476370000000\",\"size\":40,\"flags\":130,\"ts_in_delta\":167146,\"sequence\":277449}\n",
		"{\"hd\":{\"ts_event\":\"1702987922000000000\",\"rtype\":32,\"publisher_id\":40,\"instrument_id\":15144},\"open\":\"472600000000\",\"high\":\"472600000000\",\"low\":\"472600000000\",\"close\":\"472600000000\",\"volume\":\"300\"}\n",
		"{\"ts_recv\":\"1711027500000942123\",\"hd\":{\"ts_event\":\"1711027500000776211\",\"rtype\":20,\"publisher_id\":2,\"instrument_id\":17598},\"ref_price\":\"0\",\"auction_time\":\"0\",\"cont_book_clr_price\":\"0\",\"auct_interest_clr_price\":\"0\",\"ssr_filling_price\":\"0\",\"ind_match_price\":\"0\",\"upper_collar\":\"0\",\"lower_collar\":\"0\",\"paired_qty\":0,\"total_imbalance_qty\":0,\"market_imbalance_qty\":0,\"unpaired_qty\":0,\"auction_type\":\"O\",\"side\":\"N\",\"auction_status\":0,\"freeze_status\":0,\"num_extensions\":0,\"unpaired_side\":\"N\",\"significant_imbalance\":\"~\"}\n",
		"{\"ts_recv\":\"1609160400000000000\",\"hd\":{\"ts_event\":\"1609160400000000000\",\"rtype\":24,\"publisher_id\":1,\"instrument_id\":5482},\"ts_ref\":\"0\",\"price\":\"100\",\"quantity\":\"-5\",\"sequence\":1,\"ts_in_delta\":0,\"stat_type\":1,\"channel_id\":0,\"update_action\":1,\"stat_flags\":0}\n",
		upstreamMboV3Json,
		upstreamMbp1V3Json,
		upstreamMbp10V3Json,
		upstreamCmbp1V3Json,
		upstreamStatusV3Json,
		upstreamLiveControlStreamV3Json,
		upstreamLiveControlStreamV1Json,
		upstreamBbo1sJson,
		upstreamBbo1mJson,
		upstreamDefinitionV3Record1Json,
	} {
		f.Add(uint8(i), []byte(seed))
	}

	f.Fuzz(func(t *testing.T, kind uint8, raw []byte) {
		if len(raw) > maxFuzzInput {
			t.Skip()
		}
		before := bytes.Clone(raw)
		scanner := dbn.NewJsonScanner(bytes.NewReader(raw))
		decode := fuzzJsonDecoders[int(kind)%len(fuzzJsonDecoders)]
		const maxLines = 128
		for line := 0; line < maxLines && scanner.Next(); line++ {
			if scanner.GetLastSize() != len(scanner.GetLastRecord()) {
				t.Fatalf("GetLastSize %d != len(GetLastRecord) %d", scanner.GetLastSize(), len(scanner.GetLastRecord()))
			}
			_ = scanner.Visit(&dbn.NullVisitor{})
			_ = decode(scanner)
		}
		_ = scanner.Error()
		if !bytes.Equal(raw, before) {
			t.Fatal("JsonScanner mutated input")
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
