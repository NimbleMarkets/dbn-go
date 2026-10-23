// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"bytes"
	"testing"

	dbn "github.com/NimbleMarkets/dbn-go"
)

func TestMetadataV1FixedHeaderOffsets(t *testing.T) {
	raw := make([]byte, 100)
	raw[52] = 1 // ts_out follows stype_in and stype_out at offsets 50 and 51.
	for i := 53; i < len(raw); i++ {
		raw[i] = byte(i)
	}
	var header dbn.MetadataHeaderV1
	if err := header.FillFixed_Raw(raw); err != nil {
		t.Fatal(err)
	}
	if header.TsOut != 1 || !bytes.Equal(header.Reserved[:], raw[53:100]) {
		t.Fatalf("incorrect ts_out or padding: %+v", header)
	}
	var encoded bytes.Buffer
	md := dbn.Metadata{VersionNum: dbn.HeaderVersion1, TsOut: 1}
	if err := md.Write(&encoded); err != nil {
		t.Fatal(err)
	}
	got, err := dbn.ReadMetadata(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got.TsOut != 1 {
		t.Fatal("lost v1 ts_out flag")
	}
}
