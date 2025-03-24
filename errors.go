// Copyright (c) 2024 Neomantra Corp

package dbn

import "fmt"

var (
	ErrInvalidDBNVersion     = fmt.Errorf("invalid DBN version")
	ErrInvalidDBNFile        = fmt.Errorf("invalid DBN file")
	ErrHeaderTooShort        = fmt.Errorf("header shorter than expected")
	ErrHeaderTooLong         = fmt.Errorf("header longer than expected")
	ErrUnexpectedCStrLength  = fmt.Errorf("unexpected cstr length")
	ErrNoRecord              = fmt.Errorf("no record scanned")
	ErrMalformedRecord       = fmt.Errorf("malformed record")
	ErrUnknownRType          = fmt.Errorf("unknown rtype")
	ErrDateOutsideQueryRange = fmt.Errorf("date outside the query range")
	ErrWrongStypesForMapping = fmt.Errorf("wrong stypes for mapping")
	ErrNoMetadata            = fmt.Errorf("no metadata")
)

func unexpectedBytesError(got int, want int) error {
	return fmt.Errorf("expected %d bytes, got %d", want, got)
}

func unexpectedRTypeError(got RType, want RType) error {
	return fmt.Errorf("expected RType %d , got %d", want, got)
}
