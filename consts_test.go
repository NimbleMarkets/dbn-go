// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"testing"

	dbn "github.com/NimbleMarkets/dbn-go"
)

func TestSTypeRoundTrip(t *testing.T) {
	tests := []struct {
		stype dbn.SType
		value uint8
		str   string
	}{
		{dbn.SType_FigiTicker, 12, "figi_ticker"},
		{dbn.SType_ListingId, 13, "listing_id"},
		{dbn.SType_IssuerId, 14, "issuer_id"},
		{dbn.SType_SecurityId, 15, "security_id"},
	}
	for _, tt := range tests {
		if uint8(tt.stype) != tt.value {
			t.Errorf("%s: got %d, want %d", tt.str, tt.stype, tt.value)
		}
		if tt.stype.String() != tt.str {
			t.Errorf("SType(%d).String(): got %q, want %q", tt.value, tt.stype.String(), tt.str)
		}
		got, err := dbn.STypeFromString(tt.str)
		if err != nil || got != tt.stype {
			t.Errorf("STypeFromString(%q): got %v, %v, want %v", tt.str, got, err, tt.stype)
		}
	}
}

func TestSystemCodeRoundTrip(t *testing.T) {
	if dbn.SystemCode_UnsubscribeAck != 5 {
		t.Errorf("SystemCode_UnsubscribeAck: got %d, want 5", dbn.SystemCode_UnsubscribeAck)
	}
	code := dbn.SystemCode(dbn.SystemCode_UnsubscribeAck)
	if code.String() != "UNSUBSCRIBE_ACK" {
		t.Errorf("String(): got %q", code.String())
	}
	got, err := dbn.SystemCodeFromString("unsubscribe_ack")
	if err != nil || got != code {
		t.Errorf("SystemCodeFromString: got %v, %v", got, err)
	}
}

func TestErrorCodeRoundTrip(t *testing.T) {
	if dbn.ErrorCode_ReplayDataAgedOut != 8 {
		t.Errorf("ErrorCode_ReplayDataAgedOut: got %d, want 8", dbn.ErrorCode_ReplayDataAgedOut)
	}
	if dbn.ErrorCode_ReplayDataAgedOut.String() != "REPLAY_DATA_AGED_OUT" {
		t.Errorf("String(): got %q", dbn.ErrorCode_ReplayDataAgedOut.String())
	}
	got, err := dbn.ErrorCodeFromString("replay_data_aged_out")
	if err != nil || got != dbn.ErrorCode_ReplayDataAgedOut {
		t.Errorf("ErrorCodeFromString: got %v, %v", got, err)
	}
}

func TestSyncedEnumValues(t *testing.T) {
	if dbn.InstrumentClass_Index != 'I' {
		t.Errorf("InstrumentClass_Index: got %c", dbn.InstrumentClass_Index)
	}
	if dbn.MatchAlgorithm_Allocation != 'A' {
		t.Errorf("MatchAlgorithm_Allocation: got %c", dbn.MatchAlgorithm_Allocation)
	}
	stats := map[dbn.StatType]uint16{
		dbn.StatType_IndicativeClosePrice:        20,
		dbn.StatType_MwcbLevel1:                  21,
		dbn.StatType_MwcbLevel2:                  22,
		dbn.StatType_MwcbLevel3:                  23,
		dbn.StatType_AuctionCollarReferencePrice: 24,
		dbn.StatType_AuctionCollarUpperPrice:     25,
		dbn.StatType_AuctionCollarLowerPrice:     26,
		dbn.StatType_VenueSpecificVolume1:        10001,
		dbn.StatType_VenueSpecificPrice1:         10002,
	}
	for st, want := range stats {
		if uint16(st) != want {
			t.Errorf("StatType: got %d, want %d", st, want)
		}
	}
}
