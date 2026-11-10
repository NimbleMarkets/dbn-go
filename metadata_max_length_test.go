// Copyright (c) 2025 Neomantra Corp

package dbn_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/NimbleMarkets/dbn-go"
)

// metadataPrefix returns the 8-byte DBN prelude declaring `length` bytes of metadata.
func metadataPrefix(length uint32) []byte {
	return binary.LittleEndian.AppendUint32([]byte{'D', 'B', 'N', dbn.HeaderVersion3}, length)
}

func validMetadataBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	metadata := dbn.Metadata{VersionNum: dbn.HeaderVersion3, Dataset: "TEST.DATA", Schema: dbn.Schema_Ohlcv1S,
		StypeIn: dbn.SType_RawSymbol, StypeOut: dbn.SType_InstrumentId, Symbols: []string{"AAPL", "MSFT"}}
	if err := metadata.Write(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func setMaxMetadataLength(t *testing.T, length uint32) {
	t.Helper()
	saved := dbn.MaxMetadataLength
	dbn.MaxMetadataLength = length
	t.Cleanup(func() { dbn.MaxMetadataLength = saved })
}

func TestReadMetadataRejectsLengthAboveDefaultLimit(t *testing.T) {
	// The declared length is rejected from the prelude alone, before any
	// metadata bytes are read or buffered.
	for _, length := range []uint32{dbn.MaxMetadataLength + 1, 1<<32 - 1} {
		_, err := dbn.ReadMetadata(bytes.NewReader(metadataPrefix(length)))
		if !errors.Is(err, dbn.ErrMetadataTooLarge) {
			t.Errorf("length %d: got error %v, want ErrMetadataTooLarge", length, err)
		}
	}
}

func TestReadMetadataHonorsConfiguredLimit(t *testing.T) {
	data := validMetadataBytes(t)
	length := uint32(len(data) - 8)

	setMaxMetadataLength(t, length-1)
	if _, err := dbn.ReadMetadata(bytes.NewReader(data)); !errors.Is(err, dbn.ErrMetadataTooLarge) {
		t.Errorf("limit below length: got error %v, want ErrMetadataTooLarge", err)
	}

	dbn.MaxMetadataLength = length
	if _, err := dbn.ReadMetadata(bytes.NewReader(data)); err != nil {
		t.Errorf("limit equal to length: got error %v, want nil", err)
	}
}

func TestReadMetadataZeroLimitMeansUnlimited(t *testing.T) {
	setMaxMetadataLength(t, 0)
	if _, err := dbn.ReadMetadata(bytes.NewReader(validMetadataBytes(t))); err != nil {
		t.Errorf("got error %v, want nil", err)
	}
}

func TestDbnScannerReportsMetadataTooLarge(t *testing.T) {
	scanner := dbn.NewDbnScanner(bytes.NewReader(metadataPrefix(1<<32 - 1)))
	if _, err := scanner.Metadata(); !errors.Is(err, dbn.ErrMetadataTooLarge) {
		t.Errorf("got error %v, want ErrMetadataTooLarge", err)
	}
}
