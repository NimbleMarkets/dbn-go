// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"testing"

	dbn "github.com/NimbleMarkets/dbn-go"
)

// TestReadMetadataBodyEndsAtFieldBoundary checks metadata whose declared
// length stops exactly where a count field should start. The decoder runs out
// of bytes on a field boundary, which must surface as truncation rather than
// io.EOF, because ReadDBNToSlice and file splitting treat io.EOF as success.
func TestReadMetadataBodyEndsAtFieldBoundary(t *testing.T) {
	for _, version := range []uint8{dbn.HeaderVersion1, dbn.HeaderVersion2, dbn.HeaderVersion3} {
		fixedSize := dbn.MetadataHeaderV2_Size
		if version == dbn.HeaderVersion1 {
			fixedSize = dbn.MetadataHeaderV1_Size
		}
		md := &dbn.Metadata{
			VersionNum: version,
			Schema:     dbn.Schema_Trades,
			Start:      1,
			End:        dbn.UNDEF_TIMESTAMP,
			StypeIn:    dbn.SType_RawSymbol,
			StypeOut:   dbn.SType_InstrumentId,
			Dataset:    "GLBX.MDP3",
		}
		buf := &bytes.Buffer{}
		if err := md.Write(buf); err != nil {
			t.Fatalf("v%d: Write: %v", version, err)
		}
		full := buf.Bytes()

		// Cut the body after the fixed header and after each of the five
		// length/count fields that follow it, except the last.
		for field := 0; field < 5; field++ {
			name := fmt.Sprintf("v%d/before field %d", version, field)
			bodyLen := fixedSize + 4*field
			b := append([]byte(nil), full[:8+bodyLen]...)
			binary.LittleEndian.PutUint32(b[4:8], uint32(bodyLen))

			_, err := dbn.ReadMetadata(bytes.NewReader(b))
			if !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Errorf("%s: got err %v, want io.ErrUnexpectedEOF", name, err)
			}
			records, metadata, err := dbn.ReadDBNToSlice[dbn.MboMsg](bytes.NewReader(b))
			if err == nil {
				t.Errorf("%s: ReadDBNToSlice got success (records=%d, metadata=%v), want an error", name, len(records), metadata)
			}
		}
	}
}
