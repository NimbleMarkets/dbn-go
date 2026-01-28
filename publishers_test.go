// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"encoding/json"
	"testing"

	"github.com/NimbleMarkets/dbn-go"
)

///////////////////////////////////////////////////////////////////////////////
// Venue Tests

func TestVenue_Values(t *testing.T) {
	tests := []struct {
		venue dbn.Venue
		want  uint16
		str   string
	}{
		{dbn.Venue_Glbx, 1, "GLBX"},
		{dbn.Venue_Xnas, 2, "XNAS"},
		{dbn.Venue_Xnys, 9, "XNYS"},
		{dbn.Venue_Equs, 47, "EQUS"},
		{dbn.Venue_Ifus, 48, "IFUS"},
		{dbn.Venue_Xcbf, 52, "XCBF"},
		{dbn.Venue_Ocea, 53, "OCEA"},
	}

	for _, tt := range tests {
		if uint16(tt.venue) != tt.want {
			t.Errorf("Venue %d: got %d, want %d", tt.venue, uint16(tt.venue), tt.want)
		}
		if got := tt.venue.String(); got != tt.str {
			t.Errorf("Venue.String() %d: got %q, want %q", tt.venue, got, tt.str)
		}
	}
}

func TestVenueFromString(t *testing.T) {
	tests := []struct {
		input string
		want  dbn.Venue
	}{
		{"GLBX", dbn.Venue_Glbx},
		{"glbx", dbn.Venue_Glbx},
		{"XNAS", dbn.Venue_Xnas},
		{"xnas", dbn.Venue_Xnas},
		{"EQUS", dbn.Venue_Equs},
		{"IFUS", dbn.Venue_Ifus},
		{"XCBF", dbn.Venue_Xcbf},
		{"OCEA", dbn.Venue_Ocea},
	}

	for _, tt := range tests {
		got, err := dbn.VenueFromString(tt.input)
		if err != nil {
			t.Errorf("VenueFromString(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("VenueFromString(%q): got %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestVenueFromString_Invalid(t *testing.T) {
	_, err := dbn.VenueFromString("INVALID")
	if err == nil {
		t.Error("VenueFromString(\"INVALID\"): expected error, got nil")
	}
}

func TestVenue_JSON(t *testing.T) {
	venue := dbn.Venue_Xnas
	data, err := json.Marshal(venue)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(data) != `"XNAS"` {
		t.Errorf("json.Marshal: got %s, want \"XNAS\"", string(data))
	}

	var decoded dbn.Venue
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if decoded != venue {
		t.Errorf("json.Unmarshal: got %v, want %v", decoded, venue)
	}
}

///////////////////////////////////////////////////////////////////////////////
// Dataset Tests

func TestDataset_Values(t *testing.T) {
	tests := []struct {
		dataset dbn.Dataset
		want    uint16
		str     string
	}{
		{dbn.Dataset_GlbxMdp3, 1, "GLBX.MDP3"},
		{dbn.Dataset_XnasItch, 2, "XNAS.ITCH"},
		{dbn.Dataset_EqusMini, 35, "EQUS.MINI"},
		{dbn.Dataset_IfusImpact, 36, "IFUS.IMPACT"},
		{dbn.Dataset_XcbfPitch, 40, "XCBF.PITCH"},
		{dbn.Dataset_OceaMemoir, 41, "OCEA.MEMOIR"},
	}

	for _, tt := range tests {
		if uint16(tt.dataset) != tt.want {
			t.Errorf("Dataset %d: got %d, want %d", tt.dataset, uint16(tt.dataset), tt.want)
		}
		if got := tt.dataset.String(); got != tt.str {
			t.Errorf("Dataset.String() %d: got %q, want %q", tt.dataset, got, tt.str)
		}
	}
}

func TestDatasetFromString(t *testing.T) {
	tests := []struct {
		input string
		want  dbn.Dataset
	}{
		{"GLBX.MDP3", dbn.Dataset_GlbxMdp3},
		{"glbx.mdp3", dbn.Dataset_GlbxMdp3},
		{"XNAS.ITCH", dbn.Dataset_XnasItch},
		{"EQUS.MINI", dbn.Dataset_EqusMini},
		{"IFUS.IMPACT", dbn.Dataset_IfusImpact},
		{"XCBF.PITCH", dbn.Dataset_XcbfPitch},
		{"OCEA.MEMOIR", dbn.Dataset_OceaMemoir},
	}

	for _, tt := range tests {
		got, err := dbn.DatasetFromString(tt.input)
		if err != nil {
			t.Errorf("DatasetFromString(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("DatasetFromString(%q): got %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestDataset_Publishers(t *testing.T) {
	// Test single publisher dataset
	pubs := dbn.Dataset_GlbxMdp3.Publishers()
	if len(pubs) != 1 || pubs[0] != dbn.Publisher_GlbxMdp3Glbx {
		t.Errorf("Dataset_GlbxMdp3.Publishers(): got %v, want [Publisher_GlbxMdp3Glbx]", pubs)
	}

	// Test multi-publisher dataset
	pubs = dbn.Dataset_OpraPillar.Publishers()
	if len(pubs) != 19 {
		t.Errorf("Dataset_OpraPillar.Publishers(): got %d publishers, want 19", len(pubs))
	}

	// Test deprecated dataset
	pubs = dbn.Dataset_FinnNls.Publishers()
	if len(pubs) != 0 {
		t.Errorf("Dataset_FinnNls.Publishers(): got %d publishers, want 0 (deprecated)", len(pubs))
	}
}

func TestDataset_JSON(t *testing.T) {
	dataset := dbn.Dataset_EqusMini
	data, err := json.Marshal(dataset)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(data) != `"EQUS.MINI"` {
		t.Errorf("json.Marshal: got %s, want \"EQUS.MINI\"", string(data))
	}

	var decoded dbn.Dataset
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if decoded != dataset {
		t.Errorf("json.Unmarshal: got %v, want %v", decoded, dataset)
	}
}

///////////////////////////////////////////////////////////////////////////////
// Publisher Tests

func TestPublisher_Values(t *testing.T) {
	tests := []struct {
		publisher dbn.Publisher
		want      uint16
		str       string
	}{
		{dbn.Publisher_GlbxMdp3Glbx, 1, "GLBX.MDP3.GLBX"},
		{dbn.Publisher_XnasItchXnas, 2, "XNAS.ITCH.XNAS"},
		{dbn.Publisher_EqusMiniEqus, 95, "EQUS.MINI.EQUS"},
		{dbn.Publisher_IfusImpactIfus, 97, "IFUS.IMPACT.IFUS"},
		{dbn.Publisher_OceaMemoirOcea, 107, "OCEA.MEMOIR.OCEA"},
	}

	for _, tt := range tests {
		if uint16(tt.publisher) != tt.want {
			t.Errorf("Publisher %d: got %d, want %d", tt.publisher, uint16(tt.publisher), tt.want)
		}
		if got := tt.publisher.String(); got != tt.str {
			t.Errorf("Publisher.String() %d: got %q, want %q", tt.publisher, got, tt.str)
		}
	}
}

func TestPublisher_Venue(t *testing.T) {
	tests := []struct {
		publisher dbn.Publisher
		want      dbn.Venue
	}{
		{dbn.Publisher_GlbxMdp3Glbx, dbn.Venue_Glbx},
		{dbn.Publisher_XnasItchXnas, dbn.Venue_Xnas},
		{dbn.Publisher_EqusMiniEqus, dbn.Venue_Equs},
		{dbn.Publisher_IfusImpactIfus, dbn.Venue_Ifus},
		{dbn.Publisher_IfusImpactXoff, dbn.Venue_Xoff},
		{dbn.Publisher_OceaMemoirOcea, dbn.Venue_Ocea},
	}

	for _, tt := range tests {
		if got := tt.publisher.Venue(); got != tt.want {
			t.Errorf("Publisher(%d).Venue(): got %v, want %v", tt.publisher, got, tt.want)
		}
	}
}

func TestPublisher_Dataset(t *testing.T) {
	tests := []struct {
		publisher dbn.Publisher
		want      dbn.Dataset
	}{
		{dbn.Publisher_GlbxMdp3Glbx, dbn.Dataset_GlbxMdp3},
		{dbn.Publisher_XnasItchXnas, dbn.Dataset_XnasItch},
		{dbn.Publisher_EqusMiniEqus, dbn.Dataset_EqusMini},
		{dbn.Publisher_IfusImpactIfus, dbn.Dataset_IfusImpact},
		{dbn.Publisher_XcbfPitchXcbf, dbn.Dataset_XcbfPitch},
		{dbn.Publisher_OceaMemoirOcea, dbn.Dataset_OceaMemoir},
	}

	for _, tt := range tests {
		if got := tt.publisher.Dataset(); got != tt.want {
			t.Errorf("Publisher(%d).Dataset(): got %v, want %v", tt.publisher, got, tt.want)
		}
	}
}

func TestPublisherFromString(t *testing.T) {
	tests := []struct {
		input string
		want  dbn.Publisher
	}{
		{"GLBX.MDP3.GLBX", dbn.Publisher_GlbxMdp3Glbx},
		{"glbx.mdp3.glbx", dbn.Publisher_GlbxMdp3Glbx},
		{"XNAS.ITCH.XNAS", dbn.Publisher_XnasItchXnas},
		{"EQUS.MINI.EQUS", dbn.Publisher_EqusMiniEqus},
		{"IFUS.IMPACT.IFUS", dbn.Publisher_IfusImpactIfus},
		{"XCBF.PITCH.XCBF", dbn.Publisher_XcbfPitchXcbf},
		{"OCEA.MEMOIR.OCEA", dbn.Publisher_OceaMemoirOcea},
	}

	for _, tt := range tests {
		got, err := dbn.PublisherFromString(tt.input)
		if err != nil {
			t.Errorf("PublisherFromString(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("PublisherFromString(%q): got %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestPublisherFromString_Invalid(t *testing.T) {
	_, err := dbn.PublisherFromString("INVALID.PUB")
	if err == nil {
		t.Error("PublisherFromString(\"INVALID.PUB\"): expected error, got nil")
	}
}

func TestPublisherFromDatasetVenue(t *testing.T) {
	tests := []struct {
		dataset dbn.Dataset
		venue   dbn.Venue
		want    dbn.Publisher
	}{
		{dbn.Dataset_GlbxMdp3, dbn.Venue_Glbx, dbn.Publisher_GlbxMdp3Glbx},
		{dbn.Dataset_XnasItch, dbn.Venue_Xnas, dbn.Publisher_XnasItchXnas},
		{dbn.Dataset_EqusMini, dbn.Venue_Equs, dbn.Publisher_EqusMiniEqus},
		{dbn.Dataset_IfusImpact, dbn.Venue_Ifus, dbn.Publisher_IfusImpactIfus},
		{dbn.Dataset_IfusImpact, dbn.Venue_Xoff, dbn.Publisher_IfusImpactXoff},
		{dbn.Dataset_OceaMemoir, dbn.Venue_Ocea, dbn.Publisher_OceaMemoirOcea},
	}

	for _, tt := range tests {
		got, err := dbn.PublisherFromDatasetVenue(tt.dataset, tt.venue)
		if err != nil {
			t.Errorf("PublisherFromDatasetVenue(%s, %s): unexpected error: %v",
				tt.dataset.String(), tt.venue.String(), err)
			continue
		}
		if got != tt.want {
			t.Errorf("PublisherFromDatasetVenue(%s, %s): got %v, want %v",
				tt.dataset.String(), tt.venue.String(), got, tt.want)
		}
	}
}

func TestPublisherFromDatasetVenue_Invalid(t *testing.T) {
	// Invalid combination
	_, err := dbn.PublisherFromDatasetVenue(dbn.Dataset_GlbxMdp3, dbn.Venue_Xnas)
	if err == nil {
		t.Error("PublisherFromDatasetVenue(GLBX.MDP3, XNAS): expected error, got nil")
	}
}

func TestPublisher_JSON(t *testing.T) {
	publisher := dbn.Publisher_OceaMemoirOcea
	data, err := json.Marshal(publisher)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(data) != `"OCEA.MEMOIR.OCEA"` {
		t.Errorf("json.Marshal: got %s, want \"OCEA.MEMOIR.OCEA\"", string(data))
	}

	var decoded dbn.Publisher
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if decoded != publisher {
		t.Errorf("json.Unmarshal: got %v, want %v", decoded, publisher)
	}
}

///////////////////////////////////////////////////////////////////////////////
// Count Constants

func TestCounts(t *testing.T) {
	if dbn.VENUE_COUNT != 53 {
		t.Errorf("VENUE_COUNT: got %d, want 53", dbn.VENUE_COUNT)
	}
	if dbn.DATASET_COUNT != 41 {
		t.Errorf("DATASET_COUNT: got %d, want 41", dbn.DATASET_COUNT)
	}
	if dbn.PUBLISHER_COUNT != 107 {
		t.Errorf("PUBLISHER_COUNT: got %d, want 107", dbn.PUBLISHER_COUNT)
	}
}

///////////////////////////////////////////////////////////////////////////////
// Round-trip Tests

func TestVenue_RoundTrip(t *testing.T) {
	// Test that all valid venues can be converted to string and back
	venues := []dbn.Venue{
		dbn.Venue_Glbx, dbn.Venue_Xnas, dbn.Venue_Xnys,
		dbn.Venue_Equs, dbn.Venue_Ifus, dbn.Venue_Xcbf, dbn.Venue_Ocea,
	}

	for _, v := range venues {
		str := v.String()
		got, err := dbn.VenueFromString(str)
		if err != nil {
			t.Errorf("Venue %d round-trip: error: %v", v, err)
			continue
		}
		if got != v {
			t.Errorf("Venue %d round-trip: got %d, want %d", v, got, v)
		}
	}
}

func TestDataset_RoundTrip(t *testing.T) {
	datasets := []dbn.Dataset{
		dbn.Dataset_GlbxMdp3, dbn.Dataset_XnasItch, dbn.Dataset_EqusMini,
		dbn.Dataset_IfusImpact, dbn.Dataset_XcbfPitch, dbn.Dataset_OceaMemoir,
	}

	for _, d := range datasets {
		str := d.String()
		got, err := dbn.DatasetFromString(str)
		if err != nil {
			t.Errorf("Dataset %d round-trip: error: %v", d, err)
			continue
		}
		if got != d {
			t.Errorf("Dataset %d round-trip: got %d, want %d", d, got, d)
		}
	}
}

func TestPublisher_RoundTrip(t *testing.T) {
	publishers := []dbn.Publisher{
		dbn.Publisher_GlbxMdp3Glbx, dbn.Publisher_XnasItchXnas,
		dbn.Publisher_EqusMiniEqus, dbn.Publisher_IfusImpactIfus,
		dbn.Publisher_XcbfPitchXcbf, dbn.Publisher_OceaMemoirOcea,
	}

	for _, p := range publishers {
		str := p.String()
		got, err := dbn.PublisherFromString(str)
		if err != nil {
			t.Errorf("Publisher %d round-trip: error: %v", p, err)
			continue
		}
		if got != p {
			t.Errorf("Publisher %d round-trip: got %d, want %d", p, got, p)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// Consistency Tests

func TestPublisher_Consistency(t *testing.T) {
	// Test that Publisher.Venue() and Publisher.Dataset() are consistent
	// with PublisherFromDatasetVenue
	publishers := []dbn.Publisher{
		dbn.Publisher_GlbxMdp3Glbx,
		dbn.Publisher_XnasItchXnas,
		dbn.Publisher_EqusMiniEqus,
		dbn.Publisher_IfusImpactIfus,
		dbn.Publisher_IfusImpactXoff,
		dbn.Publisher_OceaMemoirOcea,
	}

	for _, p := range publishers {
		dataset := p.Dataset()
		venue := p.Venue()

		got, err := dbn.PublisherFromDatasetVenue(dataset, venue)
		if err != nil {
			t.Errorf("Publisher %d consistency: PublisherFromDatasetVenue error: %v", p, err)
			continue
		}
		if got != p {
			t.Errorf("Publisher %d consistency: got %d, want %d", p, got, p)
		}
	}
}

func TestDataset_Publishers_Consistency(t *testing.T) {
	// Test that all publishers returned by Dataset.Publishers() have the correct Dataset
	datasets := []dbn.Dataset{
		dbn.Dataset_GlbxMdp3,
		dbn.Dataset_XnasItch,
		dbn.Dataset_EqusMini,
		dbn.Dataset_IfusImpact,
		dbn.Dataset_OceaMemoir,
	}

	for _, d := range datasets {
		for _, p := range d.Publishers() {
			if p.Dataset() != d {
				t.Errorf("Dataset %s: Publisher %s has wrong Dataset %s",
					d.String(), p.String(), p.Dataset().String())
			}
		}
	}
}
