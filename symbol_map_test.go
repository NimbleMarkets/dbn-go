// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	dbn "github.com/NimbleMarkets/dbn-go"
)

func TestPitSymbolMapIsEmpty(t *testing.T) {
	var zero dbn.PitSymbolMap
	if !zero.IsEmpty() {
		t.Fatal("zero-value map should be empty")
	}
	m := dbn.NewPitSymbolMap()
	if !m.IsEmpty() {
		t.Fatal("new map should be empty")
	}
	record := dbn.NewSymbolMappingMsgV2()
	record.Header.InstrumentID, record.StypeOutSymbol = 42, "TEST"
	if err := m.OnSymbolMappingMsg(&record); err != nil {
		t.Fatal(err)
	}
	if m.IsEmpty() || m.Len() != 1 {
		t.Fatal("populated map should not be empty")
	}
}

func TestPitSymbolMapMetadataInstrumentIDs(t *testing.T) {
	start := uint64(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano())
	for _, inverse := range []bool{false, true} {
		for _, tc := range []struct {
			id    string
			want  uint32
			valid bool
		}{
			{"-1", 0, false}, {"4294967296", 0, false}, {"not-an-id", 0, false},
			{"0", 0, true}, {"4294967295", 4294967295, true},
		} {
			t.Run(fmt.Sprintf("inverse_%v/id_%s", inverse, tc.id), func(t *testing.T) {
				md := &dbn.Metadata{Start: start, End: start + uint64(24*time.Hour),
					StypeIn: dbn.SType_RawSymbol, StypeOut: dbn.SType_InstrumentId,
					Mappings: []dbn.SymbolMapping{{RawSymbol: "TEST", Intervals: []dbn.MappingInterval{
						{StartDate: 20260101, EndDate: 20260102, Symbol: tc.id},
					}}},
				}
				if inverse {
					md.StypeIn, md.StypeOut = md.StypeOut, md.StypeIn
					md.Mappings[0].RawSymbol = tc.id
					md.Mappings[0].Intervals[0].Symbol = "TEST"
				}
				m := dbn.NewPitSymbolMap()
				err := m.FillFromMetadata(md, start)
				if !tc.valid {
					var numErr *strconv.NumError
					if !errors.As(err, &numErr) {
						t.Fatalf("expected ID parse error, got %v", err)
					}
					if !m.IsEmpty() {
						t.Fatal("invalid ID was inserted")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if got := m.Get(tc.want); got != "TEST" {
					t.Fatalf("instrument %d: got %q, want TEST", tc.want, got)
				}
			})
		}
	}
}

func TestTsSymbolMapExclusiveEnd(t *testing.T) {
	m := dbn.NewTsSymbolMap()
	if err := m.Insert(1, 20240228, 20240301, "OLD"); err != nil {
		t.Fatal(err)
	}
	for _, date := range []int{20240228, 20240229} {
		if got := m.Get(dbn.YMDToTime(date, time.UTC), 1); got != "OLD" {
			t.Fatalf("%d: got %q", date, got)
		}
	}
	if got := m.Get(dbn.YMDToTime(20240301, time.UTC), 1); got != "" {
		t.Fatalf("exclusive end: got %q", got)
	}
	// An earlier interval must not overwrite a mapping starting at its end.
	if err := m.Insert(1, 20240301, 20240302, "NEW"); err != nil {
		t.Fatal(err)
	}
	if err := m.Insert(1, 20240228, 20240301, "OLD"); err != nil {
		t.Fatal(err)
	}
	if got := m.Get(dbn.YMDToTime(20240301, time.UTC), 1); got != "NEW" {
		t.Fatalf("boundary: got %q", got)
	}
	before := m.Len()
	if err := m.Insert(1, 20240301, 20240301, "EMPTY"); err != nil {
		t.Fatal(err)
	}
	if m.Len() != before || m.Get(dbn.YMDToTime(20240301, time.UTC), 1) != "NEW" {
		t.Fatal("empty range changed map")
	}
	if err := m.Insert(1, 20240302, 20240301, "INVALID"); err == nil {
		t.Fatal("expected reversed range error")
	}
}

// Compare interval updates against a simple point-by-point reference, including
// out-of-order insertion, overlap, gaps, multiple instruments, and empty ranges.
func TestTsSymbolMapIntervalsAgainstReference(t *testing.T) {
	var m dbn.TsSymbolMap
	var want [3][128]string
	rng := rand.New(rand.NewSource(42))
	for step := 0; step < 500; step++ {
		id := rng.Intn(len(want))
		start := rng.Intn(128)
		end := start + rng.Intn(129-start)
		symbol := []string{"A", "B", "C"}[rng.Intn(3)]
		if err := m.InsertTs(uint32(id), uint64(start), uint64(end), symbol); err != nil {
			t.Fatal(err)
		}
		for ts := start; ts < end; ts++ {
			want[id][ts] = symbol
		}
		runs := 0
		for iid, points := range want {
			previous := ""
			for ts, expected := range points {
				if got := m.GetForTs(uint64(ts), uint32(iid)); got != expected {
					t.Fatalf("step %d, id %d, ts %d: got %q want %q", step, iid, ts, got, expected)
				}
				if expected != "" && expected != previous {
					runs++
				}
				previous = expected
			}
			if m.GetForTs(128, uint32(iid)) != "" {
				t.Fatal("matched exclusive end")
			}
		}
		if m.Len() != runs {
			t.Fatalf("step %d: %d intervals, want %d coalesced runs", step, m.Len(), runs)
		}
	}
	before := m.Len()
	if err := m.InsertTs(0, 50, 40, "INVALID"); err == nil {
		t.Fatal("expected invalid range error")
	}
	if m.Len() != before {
		t.Fatal("invalid range changed map")
	}
}

func TestTsSymbolMapDateCompatibilityAndLongRange(t *testing.T) {
	m := dbn.NewTsSymbolMap()
	if !m.IsEmpty() {
		t.Fatal("new map is not empty")
	}
	if err := m.Insert(1, 20000101, 20500101, "LONG"); err != nil {
		t.Fatal(err)
	}
	if m.Len() != 1 {
		t.Fatalf("50-year range stored as %d intervals", m.Len())
	}
	if m.IsEmpty() {
		t.Fatal("populated map is empty")
	}
	// The legacy wrapper uses the calendar date in the supplied location.
	dt := time.Date(2049, 12, 31, 23, 0, 0, 0, time.FixedZone("west", -5*3600))
	if got := m.Get(dt, 1); got != "LONG" {
		t.Fatalf("date wrapper: %q", got)
	}
	if got := m.GetForTs(uint64(dt.UnixNano()), 1); got != "" {
		t.Fatalf("timestamp lookup outside range: %q", got)
	}
	for _, date := range []uint32{0, 19691231, 20240230, 30000101} {
		if err := m.Insert(1, date, date, "INVALID"); err == nil {
			t.Errorf("accepted invalid date %d", date)
		}
	}
	// Dates past the signed int64 nanosecond range must not wrap.
	if err := m.Insert(2, 23000101, 23000102, "FUTURE"); err != nil {
		t.Fatal(err)
	}
	future := uint64(time.Date(2300, 1, 1, 0, 0, 0, 0, time.UTC).Unix()) * uint64(time.Second)
	if got := m.GetForTs(future, 2); got != "FUTURE" {
		t.Fatalf("future timestamp: %q", got)
	}
}

func TestTsSymbolMapMetadata(t *testing.T) {
	for _, inverse := range []bool{false, true} {
		md := &dbn.Metadata{StypeIn: dbn.SType_RawSymbol, StypeOut: dbn.SType_InstrumentId,
			Mappings: []dbn.SymbolMapping{{RawSymbol: "TEST", Intervals: []dbn.MappingInterval{
				{StartDate: 20260101, EndDate: 20260103, Symbol: "123"},
				{StartDate: 20260103, EndDate: 20260105, Symbol: "123"},
				{StartDate: 20260105, EndDate: 20260106, Symbol: ""},
			}}},
		}
		if inverse {
			md.StypeIn, md.StypeOut = md.StypeOut, md.StypeIn
			md.Mappings[0].RawSymbol = "123"
			md.Mappings[0].Intervals[0].Symbol = "TEST"
			md.Mappings[0].Intervals[1].Symbol = "TEST"
		}
		m := dbn.NewTsSymbolMap()
		if err := m.FillFromMetadata(md); err != nil {
			t.Fatal(err)
		}
		if m.Len() != 1 {
			t.Fatalf("inverse=%v: expected coalesced interval", inverse)
		}
		if got := m.Get(dbn.YMDToTime(20260104, time.UTC), 123); got != "TEST" {
			t.Fatalf("inverse=%v: %q", inverse, got)
		}
		if m.Get(dbn.YMDToTime(20260105, time.UTC), 123) != "" {
			t.Fatal("end date included")
		}
		md.Mappings[0].Intervals[0].EndDate = 20251231
		if err := m.FillFromMetadata(md); err == nil {
			t.Fatal("ignored invalid metadata range")
		}
		if m.Get(dbn.YMDToTime(20260104, time.UTC), 123) != "TEST" {
			t.Fatal("failed load changed original map")
		}
	}
	for _, id := range []string{"-1", "4294967296", "not-an-id"} {
		for _, inverse := range []bool{false, true} {
			md := &dbn.Metadata{StypeIn: dbn.SType_RawSymbol, StypeOut: dbn.SType_InstrumentId,
				Mappings: []dbn.SymbolMapping{{RawSymbol: "TEST", Intervals: []dbn.MappingInterval{{StartDate: 20260101, EndDate: 20260102, Symbol: id}}}},
			}
			if inverse {
				md.StypeIn, md.StypeOut = md.StypeOut, md.StypeIn
				md.Mappings[0].RawSymbol = id
				md.Mappings[0].Intervals[0].Symbol = "TEST"
			}
			if err := dbn.NewTsSymbolMap().FillFromMetadata(md); err == nil {
				t.Errorf("accepted instrument ID %q (inverse=%v)", id, inverse)
			}
		}
	}
}

func TestTsSymbolMapLiveUpdates(t *testing.T) {
	m := dbn.NewTsSymbolMap()
	remap := uint64(time.Date(2026, 9, 16, 13, 30, 0, 0, time.UTC).UnixNano())
	record := dbn.NewSymbolMappingMsgV2()
	record.Header.InstrumentID = 42
	record.StypeOutSymbol = "OLD"
	if err := m.OnSymbolMappingMsg(&record); err != nil {
		t.Fatal(err)
	}
	record.StartTs, record.StypeOutSymbol = remap, "NEW"
	if err := m.OnSymbolMappingMsg(&record); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		ts     uint64
		symbol string
	}{
		{0, "OLD"}, {remap - 1, "OLD"}, {remap, "NEW"}, {dbn.UNDEF_TIMESTAMP - 1, "NEW"}, {dbn.UNDEF_TIMESTAMP, ""},
	} {
		if got := m.GetForTs(tt.ts, 42); got != tt.symbol {
			t.Errorf("ts %d: got %q want %q", tt.ts, got, tt.symbol)
		}
	}
	if m.Len() != 2 {
		t.Fatalf("expected two intervals, got %d", m.Len())
	}
	record.StartTs, record.EndTs = remap, remap-1
	if err := m.OnSymbolMappingMsg(&record); err == nil {
		t.Fatal("accepted reversed live interval")
	}
	// Use the scanner's ordinary dispatch, including an unrelated record.
	scanner := dbn.NewJsonScanner(strings.NewReader(`{"hd":{"rtype":22,"instrument_id":7},"stype_out_symbol":"LIVE","start_ts":10,"end_ts":20}
{"hd":{"rtype":23},"msg":"heartbeat","code":0}
`))
	for scanner.Next() {
		if err := scanner.Visit(m); err != nil {
			t.Fatal(err)
		}
	}
	if err := scanner.Error(); err != nil {
		t.Fatal(err)
	}
	if m.GetForTs(10, 7) != "LIVE" || m.GetForTs(20, 7) != "" {
		t.Fatal("scanner live mapping failed")
	}
}
