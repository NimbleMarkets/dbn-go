// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestFetchBudget_RejectsQueryOverPerQueryCost(t *testing.T) {
	b := &fetchBudget{maxQueryCost: 1}
	if err := b.reserve(1.01, 0); err == nil || !strings.Contains(err.Error(), "query exceeds budget") {
		t.Errorf("reserve(1.01) = %v, want a per-query budget error", err)
	}
	if got := b.spent(); got != 0 {
		t.Errorf("a rejected query counted $%.2f against the session", got)
	}
	if err := b.reserve(1, 0); err != nil {
		t.Errorf("reserve(1.00) = %v, want nil", err)
	}
}

func TestFetchBudget_RejectsQueryThatWouldExceedSessionCost(t *testing.T) {
	// Each query is under the per-query limit; repeating it must not be unlimited.
	b := &fetchBudget{maxQueryCost: 1, maxSessionCost: 2.5}
	for i := 0; i < 2; i++ {
		if err := b.reserve(1, 0); err != nil {
			t.Fatalf("query %d: reserve(1.00) = %v, want nil", i+1, err)
		}
	}
	err := b.reserve(1, 0)
	if err == nil || !strings.Contains(err.Error(), "session budget") {
		t.Fatalf("third query: reserve(1.00) = %v, want a session budget error", err)
	}
	if got := b.spent(); got != 2 {
		t.Errorf("spent = $%.2f, want $2.00", got)
	}
	// A query that still fits is admitted.
	if err := b.reserve(0.5, 0); err != nil {
		t.Errorf("reserve(0.50) = %v, want nil", err)
	}
}

func TestFetchBudget_RejectsQueryOverSizeLimit(t *testing.T) {
	// Flat-rate plans report zero cost, so size is the only bound there.
	b := &fetchBudget{maxQueryCost: 1, maxSessionCost: 10, maxQueryBytes: 1000}
	if err := b.reserve(0, 1001); err == nil || !strings.Contains(err.Error(), "size") {
		t.Errorf("reserve(size 1001) = %v, want a size error", err)
	}
	if err := b.reserve(0, 1000); err != nil {
		t.Errorf("reserve(size 1000) = %v, want nil", err)
	}
}

func TestFetchBudget_RejectsInvalidEstimates(t *testing.T) {
	b := &fetchBudget{maxQueryCost: 1, maxSessionCost: 10, maxQueryBytes: 1000}
	for _, cost := range []float64{math.NaN(), -1, math.Inf(1), math.Inf(-1)} {
		if err := b.reserve(cost, 0); err == nil {
			t.Errorf("reserve(cost %v) = nil, want an error", cost)
		}
	}
	if err := b.reserve(0, -1); err == nil {
		t.Error("reserve(size -1) = nil, want an error")
	}
	if got := b.spent(); got != 0 {
		t.Errorf("spent = %v, want 0", got)
	}
}

func TestFetchBudget_ZeroLimitsAreUnlimited(t *testing.T) {
	var b fetchBudget
	for i := 0; i < 3; i++ {
		if err := b.reserve(1e6, 1<<50); err != nil {
			t.Fatalf("reserve = %v, want nil", err)
		}
	}
	if got := b.spent(); got != 3e6 {
		t.Errorf("spent = %v, want 3e6", got)
	}
}

func TestFetchBudget_ConcurrentQueriesCannotOverspend(t *testing.T) {
	b := &fetchBudget{maxQueryCost: 1, maxSessionCost: 10}
	var admitted atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b.reserve(1, 0) == nil {
				admitted.Add(1)
			}
		}()
	}
	wg.Wait()
	if got := admitted.Load(); got != 10 {
		t.Errorf("admitted %d $1 queries against a $10 session budget, want 10", got)
	}
}
