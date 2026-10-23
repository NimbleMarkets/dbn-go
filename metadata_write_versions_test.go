// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"testing"

	"github.com/NimbleMarkets/dbn-go"
)

func TestMetadataWriteVersions(t *testing.T) {
	for _, version := range []uint8{dbn.HeaderVersion1, dbn.HeaderVersion2, dbn.HeaderVersion3} {
		// Exercise all eight possible metadata alignment residues.
		for count := 0; count < 8; count++ {
			t.Run(fmt.Sprintf("v%d/symbols_%d", version, count), func(t *testing.T) {
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
				// The body has 100 fixed bytes and five uint32 counts, followed
				// by 22-byte (v1) or 71-byte (v2/v3) symbol strings.
				symbolSize := 71
				if version == dbn.HeaderVersion1 {
					symbolSize = 22
				}
				bodyLength := 120 + count*symbolSize
				expectedLength := bodyLength
				if version == dbn.HeaderVersion3 {
					expectedLength = (bodyLength + 7) &^ 7
				}
				if length != expectedLength {
					t.Fatalf("metadata length: got %d, want %d", length, expectedLength)
				}
				padding := buf.Bytes()[8+bodyLength:]
				if !bytes.Equal(padding, make([]byte, expectedLength-bodyLength)) {
					t.Fatalf("nonzero metadata padding: %v", padding)
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
			})
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

func TestMetadataWriteV3Statistics(t *testing.T) {
	meta := dbn.Metadata{VersionNum: dbn.HeaderVersion3, Dataset: "TEST.DATA",
		Schema: dbn.Schema_Statistics, StypeIn: dbn.SType_RawSymbol,
		StypeOut: dbn.SType_InstrumentId, Symbols: []string{"TEST"}}
	var buf bytes.Buffer
	if err := meta.Write(&buf); err != nil {
		t.Fatal(err)
	}
	record := dbn.StatMsgV3{
		Header:   dbn.RHeader{Length: dbn.StatMsgV3_Size / 4, RType: dbn.RType_Statistics},
		Quantity: int64(math.MaxInt32) + 123,
	}
	if err := binary.Write(&buf, binary.LittleEndian, record); err != nil {
		t.Fatal(err)
	}
	scanner := dbn.NewDbnScanner(bytes.NewReader(buf.Bytes()))
	if !scanner.Next() {
		t.Fatalf("statistics after metadata: %v", scanner.Error())
	}
	got, err := scanner.DecodeStatMsg()
	if err != nil {
		t.Fatal(err)
	}
	if got.Quantity != record.Quantity {
		t.Fatalf("statistics quantity: got %d, want %d", got.Quantity, record.Quantity)
	}
	if scanner.Next() || (scanner.Error() != nil && !errors.Is(scanner.Error(), io.EOF)) {
		t.Fatalf("unexpected trailing data: %v", scanner.Error())
	}
}
