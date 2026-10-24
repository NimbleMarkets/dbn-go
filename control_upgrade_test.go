// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	dbn "github.com/NimbleMarkets/dbn-go"
)

func TestV1ControlUpgrades(t *testing.T) {
	metadata := &dbn.Metadata{VersionNum: dbn.HeaderVersion1}
	for _, rtype := range []dbn.RType{dbn.RType_Error, dbn.RType_System} {
		header := dbn.RHeader{Length: 20, RType: rtype, PublisherID: 7, InstrumentID: 42, TsEvent: 123456789}
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.LittleEndian, header); err != nil {
			t.Fatal(err)
		}
		buf.Write(make([]byte, 64))
		wire := buf.Bytes()
		copy(wire[16:], "Non-heartbeat message")
		wantHeader := header
		wantHeader.Length = 80 // Upgraded control records are 320 bytes.
		if rtype == dbn.RType_Error {
			r, err := dbn.DecodeErrorMsg(metadata, wire)
			if err != nil {
				t.Fatal(err)
			}
			if r.Header != wantHeader || r.Code != dbn.ErrorCode_Unset || r.IsLast != 255 {
				t.Fatalf("incorrect error upgrade: %+v", r)
			}
			if dbn.TrimNullBytes(r.Error[:]) != "Non-heartbeat message" {
				t.Fatal("error text lost")
			}
		} else {
			r, err := dbn.DecodeSystemMsg(metadata, wire)
			if err != nil {
				t.Fatal(err)
			}
			if r.Header != wantHeader || r.Code != dbn.SystemCode_Unset || r.IsHeartbeat() {
				t.Fatalf("incorrect system upgrade: %+v", r)
			}
			if dbn.TrimNullBytes(r.Message[:]) != "Non-heartbeat message" {
				t.Fatal("system text lost")
			}
			if !bytes.Equal(r.Message[64:], make([]byte, len(r.Message)-64)) {
				t.Fatal("system tail is not zero")
			}
		}
	}
}

func TestControlDecoderShortRecords(t *testing.T) {
	for _, version := range []uint8{1, 2, 3} {
		metadata := &dbn.Metadata{VersionNum: version}
		size := dbn.SystemMsg_Size
		if version == 1 {
			size = 80
		}
		for _, length := range []int{0, 15, size - 1} {
			if _, err := dbn.DecodeErrorMsg(metadata, make([]byte, length)); err == nil {
				t.Errorf("error decoder accepted v%d length %d", version, length)
			}
			if _, err := dbn.DecodeSystemMsg(metadata, make([]byte, length)); err == nil {
				t.Errorf("system decoder accepted v%d length %d", version, length)
			}
		}
	}
}

func TestSystemHeartbeatFallback(t *testing.T) {
	for _, tc := range []struct {
		text string
		code dbn.SystemCode
		want bool
	}{
		{"Heartbeat", dbn.SystemCode_Unset, true},
		{"HEARTBEAT", dbn.SystemCode_Unset, true},
		{"Heartbeat extra", dbn.SystemCode_Unset, false},
		{"Other", dbn.SystemCode_Unset, false},
		{"", dbn.SystemCode_Heartbeat, true},
	} {
		r := dbn.NewSystemMsg()
		r.Code = tc.code
		copy(r.Message[:], tc.text)
		if r.IsHeartbeat() != tc.want {
			t.Errorf("text=%q code=%d: want %t", tc.text, tc.code, tc.want)
		}
	}
}
