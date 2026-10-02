// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"testing"

	dbn "github.com/NimbleMarkets/dbn-go"
)

// Offsets within an encoded V3 header with empty lists:
// prefix (8) + fixed header (100) + schema definition length + symbols + partial + not_found + mappings counts.
const (
	countsOffset = 8 + dbn.MetadataHeaderV2_Size
)

func encodeTestMetadata(t *testing.T, mappings []dbn.SymbolMapping) []byte {
	t.Helper()
	md := &dbn.Metadata{
		VersionNum:    dbn.HeaderVersion3,
		Schema:        dbn.Schema_Trades,
		Start:         1,
		End:           dbn.UNDEF_TIMESTAMP,
		StypeIn:       dbn.SType_RawSymbol,
		StypeOut:      dbn.SType_InstrumentId,
		SymbolCstrLen: dbn.MetadataV3_SymbolCstrLen,
		Dataset:       "GLBX.MDP3",
		Symbols:       []string{},
		Partial:       []string{},
		NotFound:      []string{},
		Mappings:      mappings,
	}
	buf := &bytes.Buffer{}
	if err := md.Write(buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	return buf.Bytes()
}

func TestReadMetadataRejectsOverstatedCounts(t *testing.T) {
	names := []string{"schema definition length", "symbols", "partial", "not_found", "mappings"}
	for i, name := range names {
		b := encodeTestMetadata(t, []dbn.SymbolMapping{})
		if want := countsOffset + 5*4; len(b) != want {
			t.Fatalf("unexpected encoded length: got %d, want %d", len(b), want)
		}
		binary.LittleEndian.PutUint32(b[countsOffset+4*i:], math.MaxUint32)
		_, err := dbn.ReadMetadata(bytes.NewReader(b))
		if !errors.Is(err, dbn.ErrInvalidMetadataCount) {
			t.Errorf("%s: got err %v, want ErrInvalidMetadataCount", name, err)
		}
	}
}

func TestReadMetadataRejectsOverstatedIntervalCount(t *testing.T) {
	b := encodeTestMetadata(t, []dbn.SymbolMapping{{
		RawSymbol: "ESZ4",
		Intervals: []dbn.MappingInterval{{StartDate: 20240101, EndDate: 20240102, Symbol: "1"}},
	}})
	// ... mappings count, then raw symbol, then the interval count
	at := countsOffset + 5*4 + dbn.MetadataV3_SymbolCstrLen
	if got := binary.LittleEndian.Uint32(b[at:]); got != 1 {
		t.Fatalf("expected interval count 1 at offset %d, got %d", at, got)
	}
	binary.LittleEndian.PutUint32(b[at:], math.MaxUint32)
	_, err := dbn.ReadMetadata(bytes.NewReader(b))
	if !errors.Is(err, dbn.ErrInvalidMetadataCount) {
		t.Errorf("got err %v, want ErrInvalidMetadataCount", err)
	}
}

func TestReadMetadataValidMappingsStillDecode(t *testing.T) {
	want := []dbn.SymbolMapping{{
		RawSymbol: "ESZ4",
		Intervals: []dbn.MappingInterval{{StartDate: 20240101, EndDate: 20240102, Symbol: "123"}},
	}}
	md, err := dbn.ReadMetadata(bytes.NewReader(encodeTestMetadata(t, want)))
	if err != nil {
		t.Fatalf("ReadMetadata: %v", err)
	}
	if len(md.Mappings) != 1 || md.Mappings[0].RawSymbol != "ESZ4" || len(md.Mappings[0].Intervals) != 1 ||
		md.Mappings[0].Intervals[0] != want[0].Intervals[0] {
		t.Errorf("mappings did not round-trip: %+v", md.Mappings)
	}
}

func TestReadMetadataTruncatedOverstatedLength(t *testing.T) {
	b := encodeTestMetadata(t, []dbn.SymbolMapping{})
	// Claim a huge metadata length but provide far fewer bytes
	binary.LittleEndian.PutUint32(b[4:8], math.MaxUint32)
	_, err := dbn.ReadMetadata(bytes.NewReader(b))
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("got err %v, want io.ErrUnexpectedEOF", err)
	}
}
