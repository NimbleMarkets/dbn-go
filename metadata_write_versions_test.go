// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/NimbleMarkets/dbn-go"
)

func TestMetadataWriteVersions(t *testing.T) {
	for _, version := range []uint8{dbn.HeaderVersion1, dbn.HeaderVersion2, dbn.HeaderVersion3} {
		// Exercise all eight possible metadata alignment residues.
		for count := 0; count < 8; count++ {
			meta := dbn.Metadata{VersionNum: version, Dataset: "TEST.DATA", Schema: dbn.Schema_Ohlcv1M,
				StypeIn: dbn.SType_RawSymbol, StypeOut: dbn.SType_InstrumentId}
			for i := 0; i < count; i++ {
				meta.Symbols = append(meta.Symbols, "TEST")
			}
			var buf bytes.Buffer
			if err := meta.Write(&buf); err != nil {
				t.Fatal(err)
			}
			if got := buf.Bytes()[:4]; !bytes.Equal(got, []byte{'D', 'B', 'N', version}) {
				t.Fatalf("version %d: magic %v", version, got)
			}
			length := int(binary.LittleEndian.Uint32(buf.Bytes()[4:8]))
			if length+8 != buf.Len() || (version == dbn.HeaderVersion3 && length%8 != 0) {
				t.Fatalf("version %d: header length %d, bytes %d", version, length, buf.Len())
			}
			record := dbn.OhlcvMsg{Header: dbn.RHeader{Length: dbn.OhlcvMsg_Size / 4, RType: dbn.RType_Ohlcv1M}, Volume: 123}
			if err := binary.Write(&buf, binary.LittleEndian, record); err != nil {
				t.Fatal(err)
			}
			scanner := dbn.NewDbnScanner(bytes.NewReader(buf.Bytes()))
			decoded, err := scanner.Metadata()
			if err != nil || decoded.VersionNum != version || len(decoded.Symbols) != count {
				t.Fatalf("version %d metadata: %+v, %v", version, decoded, err)
			}
			if !scanner.Next() {
				t.Fatalf("record after metadata: %v", scanner.Error())
			}
			got, err := dbn.DbnScannerDecode[dbn.OhlcvMsg](scanner)
			if err != nil || got.Volume != record.Volume {
				t.Fatalf("record after metadata: %+v, %v", got, err)
			}
			if scanner.Next() || (scanner.Error() != nil && !errors.Is(scanner.Error(), io.EOF)) {
				t.Fatalf("unexpected trailing data: %v", scanner.Error())
			}
		}
	}
}

func TestMetadataWriteInvalidVersion(t *testing.T) {
	for _, version := range []uint8{0, 4, 255} {
		var buf bytes.Buffer
		err := (&dbn.Metadata{VersionNum: version}).Write(&buf)
		if !errors.Is(err, dbn.ErrInvalidDBNVersion) || buf.Len() != 0 {
			t.Fatalf("version %d: %v, wrote %d bytes", version, err, buf.Len())
		}
	}
}
