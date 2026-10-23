// Copyright (c) 2024-2025 Neomantra Corp

package dbn

import "fmt"

var (
	ErrInvalidDBNVersion         = fmt.Errorf("invalid DBN version")
	ErrInvalidDBNFile            = fmt.Errorf("invalid DBN file")
	ErrHeaderTooShort            = fmt.Errorf("header shorter than expected")
	ErrHeaderTooLong             = fmt.Errorf("header longer than expected")
	ErrUnexpectedCStrLength      = fmt.Errorf("unexpected cstr length")
	ErrNoRecord                  = fmt.Errorf("no record scanned")
	ErrMalformedRecord           = fmt.Errorf("malformed record")
	ErrUnknownRType              = fmt.Errorf("unknown rtype")
	ErrDateOutsideQueryRange     = fmt.Errorf("date outside the query range")
	ErrWrongStypesForMapping     = fmt.Errorf("wrong stypes for mapping")
	ErrNoMetadata                = fmt.Errorf("no metadata")
	ErrInvalidMetadataCount      = fmt.Errorf("invalid metadata count")
	ErrInvalidSymbolMappingRange = fmt.Errorf("invalid symbol mapping range")
	ErrInvalidSymbolMappingDate  = fmt.Errorf("invalid symbol mapping date")
)

func unexpectedBytesError(got int, want int) error {
	return fmt.Errorf("expected %d bytes, got %d", want, got)
}

func unexpectedRTypeError(got RType, want RType) error {
	return fmt.Errorf("expected RType %d , got %d", want, got)
}

func unexpectedCStrLenError(got uint16) error {
	return fmt.Errorf("expected CStrLen of %d bytes", got)
}

// checkedMetadataCount verifies that `count` items of at least `minItemSize` bytes
// can fit in the `remaining` bytes of the metadata, so that an overstated count
// from a malformed header is rejected before it sizes an allocation or a loop.
func checkedMetadataCount(what string, count uint32, minItemSize int, remaining int) error {
	capacity := 0
	if minItemSize > 0 {
		capacity = remaining / minItemSize
	}
	if uint64(count) > uint64(capacity) {
		return fmt.Errorf("%w: %s declares %d but only %d fit in the remaining %d bytes",
			ErrInvalidMetadataCount, what, count, capacity, remaining)
	}
	return nil
}
