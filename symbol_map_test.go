// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"testing"
	"time"

	dbn "github.com/NimbleMarkets/dbn-go"
)

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
