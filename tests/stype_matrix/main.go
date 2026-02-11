// Probes the Databento symbology.resolve API to determine which stype_in and stype_out
// values are supported for each dataset. Outputs a markdown file with two tables.
//
// Only uses the free symbology.resolve endpoint (no billing).
//
// Usage: DATABENTO_API_KEY=db-... go run ./tests/stype_matrix > tests/stype_matrix.md
package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	dbn "github.com/NimbleMarkets/dbn-go"
	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
)

var datasets = []string{
	"ARCX.PILLAR", "BATS.PITCH", "BATY.PITCH", "DBEQ.BASIC",
	"EDGA.PITCH", "EDGX.PITCH", "EPRL.DOM", "EQUS.MINI",
	"EQUS.SUMMARY", "GLBX.MDP3", "IEXG.TOPS", "IFEU.IMPACT",
	"IFLL.IMPACT", "IFUS.IMPACT", "MEMX.MEMOIR", "NDEX.IMPACT",
	"OPRA.PILLAR", "XASE.PILLAR", "XBOS.ITCH", "XCHI.PILLAR",
	"XCIS.TRADESBBO", "XEEE.EOBI", "XEUR.EOBI", "XNAS.BASIC",
	"XNAS.ITCH", "XNYS.PILLAR", "XPSX.ITCH",
}

var stypes = []dbn.SType{
	dbn.SType_RawSymbol,
	dbn.SType_InstrumentId,
	dbn.SType_Smart,
	dbn.SType_Continuous,
	dbn.SType_Parent,
	dbn.SType_NasdaqSymbol,
	dbn.SType_CmsSymbol,
	dbn.SType_Isin,
	dbn.SType_UsCode,
	dbn.SType_BbgCompId,
	dbn.SType_BbgCompTicker,
	dbn.SType_Figi,
	dbn.SType_FigiTicker,
}

// Test symbols per stype_in that are plausibly formatted for each symbology.
// We don't need them to resolve; we just need a non-422 response to confirm the combo is accepted.
var testSymbolForStypeIn = map[dbn.SType]string{
	dbn.SType_RawSymbol:    "ZZZTESTZZ",
	dbn.SType_InstrumentId: "999999999",
	dbn.SType_Smart:        "ZZZTESTZZ",
	dbn.SType_Continuous:   "ZZ.c.0",
	dbn.SType_Parent:       "ZZ",
	dbn.SType_NasdaqSymbol: "ZZZTESTZZ",
	dbn.SType_CmsSymbol:    "ZZZTESTZZ",
	dbn.SType_Isin:         "US0000000000",
	dbn.SType_UsCode:       "000000000",
	dbn.SType_BbgCompId:    "000000000",
	dbn.SType_BbgCompTicker: "ZZZTESTZZ",
	dbn.SType_Figi:         "BBG000000000",
	dbn.SType_FigiTicker:   "ZZZTESTZZ",
}

type probeResult struct {
	dataset string
	stype   dbn.SType
	ok      bool // true = accepted (200), false = rejected (422)
	err     string
}

func probe(apiKey string, dataset string, stypeIn dbn.SType, stypeOut dbn.SType, symbol string) probeResult {
	params := dbn_hist.ResolveParams{
		Dataset: dataset,
		Symbols: []string{symbol},
		StypeIn: stypeIn,
		StypeOut: stypeOut,
		DateRange: dbn_hist.DateRange{
			Start: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
		},
	}
	_, err := dbn_hist.SymbologyResolve(apiKey, params)
	if err != nil {
		errStr := err.Error()
		// 422 with "symbology_invalid_request" means the combo is unsupported
		if strings.Contains(errStr, "422") && strings.Contains(errStr, "symbology_invalid_request") {
			return probeResult{dataset: dataset, stype: stypeIn, ok: false, err: "unsupported"}
		}
		// Any other error (e.g. 400 bad request for bad symbol format)
		return probeResult{dataset: dataset, stype: stypeIn, ok: false, err: errStr}
	}
	return probeResult{dataset: dataset, stype: stypeIn, ok: true}
}

func main() {
	apiKey := os.Getenv("DATABENTO_API_KEY")
	if apiKey == "" {
		keyBytes, err := os.ReadFile(os.ExpandEnv("$HOME/secrets/dbn.key"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "DATABENTO_API_KEY not set and ~/secrets/dbn.key not readable")
			os.Exit(1)
		}
		apiKey = strings.TrimSpace(string(keyBytes))
	}

	sem := make(chan struct{}, 8) // concurrency limit
	var mu sync.Mutex

	// Table 1: stype_in support (stype_out fixed to instrument_id)
	stypeInResults := make(map[string]map[string]probeResult)
	var wg sync.WaitGroup

	fmt.Fprintf(os.Stderr, "Probing stype_in support (%d calls)...\n", len(datasets)*len(stypes))
	for _, ds := range datasets {
		for _, st := range stypes {
			wg.Add(1)
			go func(ds string, st dbn.SType) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				symbol := testSymbolForStypeIn[st]
				r := probe(apiKey, ds, st, dbn.SType_InstrumentId, symbol)

				mu.Lock()
				if stypeInResults[ds] == nil {
					stypeInResults[ds] = make(map[string]probeResult)
				}
				stypeInResults[ds][st.String()] = r
				mu.Unlock()
			}(ds, st)
		}
	}
	wg.Wait()
	fmt.Fprintf(os.Stderr, "Done.\n")

	// Table 2: stype_out support (stype_in fixed to raw_symbol)
	stypeOutResults := make(map[string]map[string]probeResult)

	fmt.Fprintf(os.Stderr, "Probing stype_out support (%d calls)...\n", len(datasets)*len(stypes))
	for _, ds := range datasets {
		for _, st := range stypes {
			wg.Add(1)
			go func(ds string, st dbn.SType) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				r := probe(apiKey, ds, dbn.SType_RawSymbol, st, "ZZZTESTZZ")

				mu.Lock()
				if stypeOutResults[ds] == nil {
					stypeOutResults[ds] = make(map[string]probeResult)
				}
				stypeOutResults[ds][st.String()] = r
				mu.Unlock()
			}(ds, st)
		}
	}
	wg.Wait()
	fmt.Fprintf(os.Stderr, "Done.\n")

	// Output markdown
	stypeNames := make([]string, len(stypes))
	for i, st := range stypes {
		stypeNames[i] = st.String()
	}

	fmt.Println("# Databento Symbology Support Matrix")
	fmt.Println()
	fmt.Printf("Generated: %s\n", time.Now().UTC().Format("2006-01-02"))
	fmt.Println()
	fmt.Println("Probed using `symbology.resolve` with dummy symbols. A checkmark means the API")
	fmt.Println("accepted the stype for that dataset (HTTP 200). An `x` means the API rejected")
	fmt.Println("the combination as unsupported (HTTP 422 `symbology_invalid_request`).")
	fmt.Println()

	// Table 1
	fmt.Println("## stype_in support")
	fmt.Println()
	fmt.Println("Each cell shows whether `stype_in=<column>` is accepted when `stype_out=instrument_id`.")
	fmt.Println()
	printTable(datasets, stypeNames, stypeInResults)
	fmt.Println()

	// Table 2
	fmt.Println("## stype_out support")
	fmt.Println()
	fmt.Println("Each cell shows whether `stype_out=<column>` is accepted when `stype_in=raw_symbol`.")
	fmt.Println()
	printTable(datasets, stypeNames, stypeOutResults)
}

func printTable(datasets []string, stypeNames []string, results map[string]map[string]probeResult) {
	// Header
	fmt.Printf("| Dataset |")
	for _, name := range stypeNames {
		fmt.Printf(" %s |", name)
	}
	fmt.Println()

	// Separator
	fmt.Printf("| --- |")
	for range stypeNames {
		fmt.Printf(" :---: |")
	}
	fmt.Println()

	// Rows
	for _, ds := range datasets {
		fmt.Printf("| %s |", ds)
		for _, name := range stypeNames {
			r, exists := results[ds][name]
			if !exists {
				fmt.Printf(" ? |")
			} else if r.ok {
				fmt.Printf(" Y |")
			} else {
				fmt.Printf(" - |")
			}
		}
		fmt.Println()
	}
}
