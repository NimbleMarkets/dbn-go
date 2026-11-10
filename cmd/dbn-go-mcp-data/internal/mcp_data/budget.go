// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"fmt"
	"math"
	"sync"
)

// fetchBudget enforces the spending and size limits on fetch_range.
// A limit of zero or less means unlimited. The zero value has no limits.
type fetchBudget struct {
	maxQueryCost   float64 // Max estimated cost, in dollars, of one fetch
	maxSessionCost float64 // Max total estimated cost, in dollars, of all fetches by this server process
	maxQueryBytes  int64   // Max billable (uncompressed) size, in bytes, of one fetch

	mu          sync.Mutex
	sessionCost float64 // Total estimated cost of the fetches admitted so far
}

// reserve admits a fetch with the given estimated cost and billable size and
// counts its cost against the session, or returns why it is not allowed.
//
// The cost is counted when the fetch is admitted and is not given back if the
// download later fails, because a failed download may still have been billed.
func (b *fetchBudget) reserve(cost float64, sizeBytes int64) error {
	if math.IsNaN(cost) || math.IsInf(cost, 0) || cost < 0 {
		return fmt.Errorf("invalid cost estimate: %v", cost)
	}
	if sizeBytes < 0 {
		return fmt.Errorf("invalid size estimate: %d bytes", sizeBytes)
	}
	if b.maxQueryCost > 0 && cost > b.maxQueryCost {
		return fmt.Errorf("query exceeds budget: estimated cost $%.2f, budget $%.2f", cost, b.maxQueryCost)
	}
	if b.maxQueryBytes > 0 && sizeBytes > b.maxQueryBytes {
		return fmt.Errorf("query exceeds size limit: estimated size %d bytes, limit %d bytes; narrow the date range or symbols, or use a more compact schema",
			sizeBytes, b.maxQueryBytes)
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.maxSessionCost > 0 && b.sessionCost+cost > b.maxSessionCost {
		return fmt.Errorf("query exceeds session budget: estimated cost $%.2f, already spent $%.2f of the $%.2f this server may spend; only the operator can raise it (--max-session-cost)",
			cost, b.sessionCost, b.maxSessionCost)
	}
	b.sessionCost += cost
	return nil
}

// spent returns the total estimated cost of the fetches admitted so far.
func (b *fetchBudget) spent() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sessionCost
}
