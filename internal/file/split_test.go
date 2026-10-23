// Copyright (c) 2024-2025 Neomantra Corp

package file

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NimbleMarkets/dbn-go"
)

func TestSplitFileRejectsInvalidSymbolMappings(t *testing.T) {
	for _, tc := range []struct {
		name      string
		startDate uint32
		id        string
	}{
		{"invalid_date", 20260230, "42"},
		{"negative_id", 20260101, "-1"},
		{"overflowing_id", 20260101, "4294967296"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := uint64(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano())
			metadata := dbn.Metadata{VersionNum: dbn.HeaderVersion3, Dataset: "TEST.DATA",
				Schema: dbn.Schema_Ohlcv1S, Start: start, End: start + uint64(24*time.Hour),
				StypeIn: dbn.SType_RawSymbol, StypeOut: dbn.SType_InstrumentId,
				Mappings: []dbn.SymbolMapping{{RawSymbol: "TEST", Intervals: []dbn.MappingInterval{
					{StartDate: tc.startDate, EndDate: 20260301, Symbol: tc.id},
				}}},
			}
			var buf bytes.Buffer
			if err := metadata.Write(&buf); err != nil {
				t.Fatal(err)
			}
			record := dbn.NewOhlcvMsg(dbn.RType_Ohlcv1S)
			record.Header.InstrumentID, record.Header.TsEvent = 42, start
			if err := binary.Write(&buf, binary.LittleEndian, record); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(t.TempDir(), "input.dbn")
			if err := os.WriteFile(source, buf.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
			dest := t.TempDir()
			if err := SplitFile(source, dest, false, false); err == nil || !strings.Contains(err.Error(), "failed to build symbol map") {
				t.Fatalf("expected symbol-map error, got %v", err)
			}
			entries, err := os.ReadDir(dest)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatal("invalid mappings produced output")
			}
		})
	}
}
