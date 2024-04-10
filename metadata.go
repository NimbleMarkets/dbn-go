// Copyright (c) 2024 Neomantra Corp

package dbn

import (
	"bytes"
	"encoding/binary"
	"io"
)

const (
	HeaderVersion1            = 1
	HeaderVersion2            = 2
	MetadataV1_SymbolCstrLen  = 22
	MetadataV1_ReservedLen    = 47
	MetadataV2_SymbolCstrLen  = 71
	MetadataV2_ReservedLen    = 53
	Metadata_DatasetCstrLen   = 16
	Metadata_PrefixSize       = 8
	MetadataHeaderV1_Size     = 100 // Size of the fixed-size portion of Metadata v1, without Prefix
	MetadataHeaderV2_Size     = 100 // Size of the fixed-size portion of Metadata v2, without Prefix
	MetadataHeaderV1_SizeFuzz = 12  // Difference between actual layout size and Golang struct
	MetadataHeaderV2_SizeFuzz = 12  // Difference between actual layout size and Golang struct
)

// Normalized Metadata about the data contained in a DBN file or stream. DBN requires the
// Metadata to be included at the start of the encoded data.
type Metadata struct {
	VersionNum       uint8
	Schema           Schema // The data record schema. u16::MAX indicates a potential mix of schemas and record types, which will always be the case for live data.
	Start            uint64 // The start time of query range in UNIX epoch nanoseconds.
	End              uint64 // The end time of query range in UNIX epoch nanoseconds. u64::MAX indicates no end time was provided.
	Limit            uint64 // The maximum number of records to return. 0 indicates no limit.
	StypeIn          SType  // The symbology type of input symbols. u8::MAX indicates a potential mix of types, such as with live data.
	StypeOut         SType  // The symbology type of output symbols.
	TsOut            uint8  // Whether each record has an appended gateway send timestamp.
	SymbolCstrLen    uint16 // The number of bytes in fixed-length string symbols, including a null terminator byte. Version 2 only, symbol strings are always 22 in version 1.
	Dataset          string
	SchemaDefinition []byte // Self-describing schema to be implemented in the future.
	Symbols          []string
	Partial          []string
	NotFound         []string
	Mappings         []SymbolMapping
}

// A raw symbol and its symbol mappings for different time ranges within the query range.
type SymbolMapping struct {
	RawSymbol string            /// The symbol assigned by publisher.
	Intervals []MappingInterval /// The mappings of `native` for different date ranges.
}

// The resolved symbol for a date range.
type MappingInterval struct {
	StartDate uint32 /// The UTC start date of interval (inclusive).
	EndDate   uint32 /// The UTC end date of interval (exclusive).
	Symbol    string /// The resolved symbol for this interval.
}

// IsInverseMapping returns true if either StypeIn or StypeOut is an InstrumentId.
// This means it is capable of bidirectional mapping between symbols and instrument IDs.
func (m *Metadata) IsInverseMapping() bool {
	return m.StypeIn == SType_InstrumentId || m.StypeOut == SType_InstrumentId
}

///////////////////////////////////////////////////////////////////////////////

// The start of every Metadata header, independent of version
type MetadataPrefix struct {
	VersionRaw [4]byte // "DBN" followed by the version of DBN the file is encoded in as a u8.
	Length     uint32  // The length of the remaining metadata header, i.e. excluding MetadataPrefix
}

// Raw DBN Metadata Header V1.
// Every DBN file begins with this header, followed by variable length fields.
// See Metadata for the full nomralized decoded structure.
type MetadataHeaderV1 struct {
	DatasetRaw [Metadata_DatasetCstrLen]byte // The dataset code (string identifier).
	Schema     Schema                        // The data record schema. u16::MAX indicates a potential mix of schemas and record types, which will always be the case for live data.
	Start      uint64                        // The start time of query range in UNIX epoch nanoseconds.
	End        uint64                        // The end time of query range in UNIX epoch nanoseconds. u64::MAX indicates no end time was provided.
	Limit      uint64                        // The maximum number of records to return. 0 indicates no limit.
	ReservedX  [8]byte                       // Reserved padding
	StypeIn    SType                         // The symbology type of input symbols. u8::MAX indicates a potential mix of types, such as with live data.
	StypeOut   SType                         // The symbology type of output symbols.
	TsOut      uint8                         // Whether each record has an appended gateway send timestamp.
	Reserved   [MetadataV1_ReservedLen]byte  // Reserved padding, after is dynamically sized section
}

func (m1 *MetadataHeaderV1) FillFixed_Raw(b []byte) error {
	if len(b) < MetadataHeaderV1_Size {
		return ErrHeaderTooShort
	}
	copy(m1.DatasetRaw[:], b[:Metadata_DatasetCstrLen])
	m1.Schema = Schema(binary.LittleEndian.Uint16(b[Metadata_DatasetCstrLen:18]))
	m1.Start = binary.LittleEndian.Uint64(b[18:26])
	m1.End = binary.LittleEndian.Uint64(b[26:34])
	m1.Limit = binary.LittleEndian.Uint64(b[34:42])
	copy(m1.ReservedX[:], b[42:50])
	m1.StypeIn = SType(b[50])
	m1.StypeOut = SType(b[51])
	m1.TsOut = b[53]
	copy(m1.Reserved[:], b[54:54+MetadataV1_ReservedLen])
	return nil
}

// Raw DBN Metadata Header V2.
// Every DBN file begins with this header, followed by variable length fields.
// See Metadata for the full nomralized decoded structure.
type MetadataHeaderV2 struct {
	DatasetRaw    [Metadata_DatasetCstrLen]byte // The dataset code (string identifier).
	Schema        Schema                        // The data record schema. u16::MAX indicates a potential mix of schemas and record types, which will always be the case for live data.
	Start         uint64                        // The start time of query range in UNIX epoch nanoseconds.
	End           uint64                        // The end time of query range in UNIX epoch nanoseconds. u64::MAX indicates no end time was provided.
	Limit         uint64                        // The maximum number of records to return. 0 indicates no limit.
	StypeIn       SType                         // The symbology type of input symbols. u8::MAX indicates a potential mix of types, such as with live data.
	StypeOut      SType                         // The symbology type of output symbols.
	TsOut         uint8                         // Whether each record has an appended gateway send timestamp.
	SymbolCstrLen uint16                        // The number of bytes in fixed-length string symbols, including a null terminator byte. Version 2 only, symbol strings are always 22 in version 1.
	Reserved      [MetadataV2_ReservedLen]byte  // Reserved padding, after is dynamically sized section
}

func (m2 *MetadataHeaderV2) FillFixed_Raw(b []byte) error {
	if len(b) < MetadataHeaderV2_Size {
		return ErrHeaderTooShort
	}
	copy(m2.DatasetRaw[:], b[:Metadata_DatasetCstrLen])
	m2.Schema = Schema(binary.LittleEndian.Uint16(b[Metadata_DatasetCstrLen:18]))
	m2.Start = binary.LittleEndian.Uint64(b[18:26])
	m2.End = binary.LittleEndian.Uint64(b[26:34])
	m2.Limit = binary.LittleEndian.Uint64(b[34:42])
	m2.StypeIn = SType(b[42])
	m2.StypeOut = SType(b[43])
	m2.TsOut = b[44]
	m2.SymbolCstrLen = binary.LittleEndian.Uint16(b[45:47])
	copy(m2.Reserved[:], b[47:47+MetadataV2_ReservedLen])
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// ReadMetadata reads the Metadata from a DBN stream over an io.Reader.
func ReadMetadata(r io.Reader) (*Metadata, error) {
	// Read the version and length
	var mp MetadataPrefix
	if err := binary.Read(r, binary.LittleEndian, &mp); err != nil {
		return nil, err
	}
	// Verify DBN header and extract version
	if (mp.VersionRaw[0] != 'D') || (mp.VersionRaw[1] != 'B') || (mp.VersionRaw[2] != 'N') {
		return nil, ErrInvalidDBNFile
	}

	// Extract the remaining bytes of the metadata header
	b := make([]byte, mp.Length)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}

	// Dispatch to the version's decoder
	switch versionNum := mp.VersionRaw[3]; versionNum {
	case HeaderVersion1:
		return makeMetadataV1(b, mp)
	case HeaderVersion2:
		return readMetadataV2(b, mp)
	default:
		return nil, ErrInvalidDBNVersion
	}
}

///////////////////////////////////////////////////////////////////////////////

func makeMetadataV1(b []byte, mp MetadataPrefix) (*Metadata, error) {
	// Read the MetadataHeader which is a fixed length
	var mhv1 MetadataHeaderV1
	if err := mhv1.FillFixed_Raw(b); err != nil {
		return nil, err
	}

	// Fill normalized Metadata struct
	m := Metadata{
		VersionNum:    mp.VersionRaw[3],
		Dataset:       TrimNullBytes(mhv1.DatasetRaw[:]),
		Schema:        mhv1.Schema,
		Start:         mhv1.Start,
		End:           mhv1.End,
		Limit:         mhv1.Limit,
		StypeIn:       mhv1.StypeIn,
		StypeOut:      mhv1.StypeOut,
		TsOut:         mhv1.TsOut,
		SymbolCstrLen: MetadataV1_SymbolCstrLen,
	}

	// Make a bytes.Reader to handle the rest of the buffer
	r := bytes.NewReader(b[MetadataHeaderV1_Size:])

	// Decode the the SchemaDefinition
	var schemaDefLen uint32
	if err := binary.Read(r, binary.LittleEndian, &schemaDefLen); err != nil {
		return nil, err
	}
	var schemaDefBytes = make([]byte, schemaDefLen*MetadataV1_SymbolCstrLen)
	if err := binary.Read(r, binary.LittleEndian, &schemaDefBytes); err != nil {
		return nil, err
	}
	m.SchemaDefinition = schemaDefBytes

	// Decode the Symbols, Partials, NotFounds
	if err := decodeToStringArray(r, MetadataV1_SymbolCstrLen, &m.Symbols); err != nil {
		return nil, err
	}
	if err := decodeToStringArray(r, MetadataV1_SymbolCstrLen, &m.Partial); err != nil {
		return nil, err
	}
	if err := decodeToStringArray(r, MetadataV1_SymbolCstrLen, &m.NotFound); err != nil {
		return nil, err
	}

	// Decode the Mapping
	if err := decodeToSymbolMapping(r, MetadataV1_SymbolCstrLen, &m.Mappings); err != nil {
		return nil, err
	}

	return &m, nil
}

///////////////////////////////////////////////////////////////////////////////

func readMetadataV2(b []byte, mp MetadataPrefix) (*Metadata, error) {
	// Read the MetadataHeader which is a fixed length
	var mhv2 MetadataHeaderV2
	if err := mhv2.FillFixed_Raw(b); err != nil {
		return nil, err
	}

	if mhv2.SymbolCstrLen != MetadataV2_SymbolCstrLen {
		return nil, ErrUnexpectedCStrLength
	}

	// Fill normalized Metadata struct
	m := Metadata{
		VersionNum:    mp.VersionRaw[3],
		Dataset:       TrimNullBytes(mhv2.DatasetRaw[:]),
		Schema:        mhv2.Schema,
		Start:         mhv2.Start,
		End:           mhv2.End,
		Limit:         mhv2.Limit,
		StypeIn:       mhv2.StypeIn,
		StypeOut:      mhv2.StypeOut,
		TsOut:         mhv2.TsOut,
		SymbolCstrLen: mhv2.SymbolCstrLen,
	}

	// Make a bytes.Reader to handle the rest of the buffer
	r := bytes.NewReader(b[MetadataHeaderV2_Size:])

	// Decode the the SchemaDefinition
	var schemaDefLen uint32
	if err := binary.Read(r, binary.LittleEndian, &schemaDefLen); err != nil {
		return nil, err
	}
	var schemaDefBytes = make([]byte, schemaDefLen)
	if err := binary.Read(r, binary.LittleEndian, &schemaDefBytes); err != nil {
		return nil, err
	}
	m.SchemaDefinition = schemaDefBytes

	// Decode the Symbols, Partials, NotFounds
	if err := decodeToStringArray(r, mhv2.SymbolCstrLen, &m.Symbols); err != nil {
		return nil, err
	}
	if err := decodeToStringArray(r, mhv2.SymbolCstrLen, &m.Partial); err != nil {
		return nil, err
	}
	if err := decodeToStringArray(r, mhv2.SymbolCstrLen, &m.NotFound); err != nil {
		return nil, err
	}

	// Decode the Mapping
	if err := decodeToSymbolMapping(r, mhv2.SymbolCstrLen, &m.Mappings); err != nil {
		return nil, err
	}

	return &m, nil
}

///////////////////////////////////////////////////////////////////////////////

// Decode a fixed-witdth string arrays from the Reader.
// Returns number of bytes read and any error.
func decodeToStringArray(r io.Reader, cstrLength uint16, strArray *[]string) error {
	var arrayLen uint32
	if err := binary.Read(r, binary.LittleEndian, &arrayLen); err != nil {
		return err
	}

	strBytes := make([]byte, cstrLength) // we reuse this
	for i := uint32(0); i < arrayLen; i++ {
		if err := binary.Read(r, binary.LittleEndian, &strBytes); err != nil {
			return err
		}
		str := TrimNullBytes(strBytes)
		*strArray = append(*strArray, str)
	}
	return nil
}

// Decode "mappings" file data to Golang SymbolMapping structure
// Returns number of bytes read and any error.
func decodeToSymbolMapping(r io.Reader, cstrLength uint16, mappings *[]SymbolMapping) error {
	var mappingLen uint32
	if err := binary.Read(r, binary.LittleEndian, &mappingLen); err != nil {
		return err
	}

	strBytes := make([]byte, cstrLength) // we reuse this
	for i := uint32(0); i < mappingLen; i++ {
		var mapping SymbolMapping
		// raw symbol
		if err := binary.Read(r, binary.LittleEndian, &strBytes); err != nil {
			return err
		}
		mapping.RawSymbol = TrimNullBytes(strBytes)
		// intervals
		var intervalLen uint32
		if err := binary.Read(r, binary.LittleEndian, &intervalLen); err != nil {
			return err
		}
		for j := uint32(0); j < intervalLen; j++ {
			var interval MappingInterval
			if err := binary.Read(r, binary.LittleEndian, &interval.StartDate); err != nil {
				return err
			}
			if err := binary.Read(r, binary.LittleEndian, &interval.EndDate); err != nil {
				return err
			}
			if err := binary.Read(r, binary.LittleEndian, &strBytes); err != nil {
				return err
			}
			interval.Symbol = TrimNullBytes(strBytes)
			mapping.Intervals = append(mapping.Intervals, interval)
		}
		// append to dest array
		*mappings = append(*mappings, mapping)
	}
	return nil
}
