// Copyright (c) 2024 Neomantra Corp

package dbn

import (
	"bytes"
	"time"
)

// / The denominator of fixed prices in DBN.
const FIXED_PRICE_SCALE float64 = 1000000000.0

func Fixed9ToFloat64(fixed int64) float64 {
	return float64(fixed) / FIXED_PRICE_SCALE
}

// TrimNullBytes removes trailing nulls from a byte slice and returns a string.
func TrimNullBytes(b []byte) string {
	return string(bytes.TrimRight(b, "\x00"))
}

// TimestampToSecNanos converts a DBN timestamp to seconds and nanoseconds.
func TimestampToSecNanos(dbnTimestamp uint64) (int64, int64) {
	secs := int64(dbnTimestamp / 1e9)
	nano := int64(dbnTimestamp) - int64(secs*1e9)
	return secs, nano
}

// TimestampToTime converts a DBN timestamp to time.Time
func TimestampToTime(dbnTimestamp uint64) time.Time {
	secs := int64(dbnTimestamp / 1e9)
	nano := int64(dbnTimestamp) - int64(secs*1e9)
	return time.Unix(secs, nano)
}

// TimeToYMD returns the YYYYMMDD for the time.Time in that Time's location.
// A zero time returns a 0 value.
// From  https://github.com/neomantra/ymdflag/blob/main/ymdflag.go#L49
func TimeToYMD(t time.Time) uint32 {
	if t.IsZero() {
		return 0
	} else {
		return uint32(10000*t.Year() + 100*int(t.Month()) + t.Day())
	}
}
