// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	dbn "github.com/NimbleMarkets/dbn-go"
)

func TestJSONSymbolMappingTimestampEncodings(t *testing.T) {
	// Above 2^53 and not exactly representable as float64. Also exercise the
	// undefined timestamp sentinel, which exceeds the signed int64 range.
	const start uint64 = 1609113600000000001
	for _, quoted := range []bool{false, true} {
		t.Run(fmt.Sprintf("quoted=%t", quoted), func(t *testing.T) {
			format := `{"hd":{"rtype":22},"start_ts":%d,"end_ts":%d}`
			if quoted {
				format = `{"hd":{"rtype":22},"start_ts":"%d","end_ts":"%d"}`
			}
			input := fmt.Sprintf(format, start, uint64(math.MaxUint64))
			v1, err := dbn.ReadJsonToSlice[dbn.SymbolMappingMsgV1](strings.NewReader(input))
			if err != nil || len(v1) != 1 {
				t.Fatalf("V1: records=%d err=%v", len(v1), err)
			}
			if v1[0].StartTs != start || v1[0].EndTs != math.MaxUint64 {
				t.Errorf("V1: timestamps = %d, %d", v1[0].StartTs, v1[0].EndTs)
			}
			v2, err := dbn.ReadJsonToSlice[dbn.SymbolMappingMsgV2](strings.NewReader(input))
			if err != nil || len(v2) != 1 {
				t.Fatalf("V2: records=%d err=%v", len(v2), err)
			}
			if v2[0].StartTs != start || v2[0].EndTs != math.MaxUint64 {
				t.Errorf("V2: timestamps = %d, %d", v2[0].StartTs, v2[0].EndTs)
			}
		})
	}
}
