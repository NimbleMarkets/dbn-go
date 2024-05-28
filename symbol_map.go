// Copyright (c) 2024 Neomantra Corp

//

package dbn

import (
	"fmt"
	"strconv"
	"time"
)

type tsSymbolKey struct {
	Date uint32 // YMD date
	ID   uint32
}

// TsSymbolMap is a timeseries symbol map. Generally useful for working with historical data
// and is commonly built from a Metadata object.
type TsSymbolMap struct {
	symbolMap map[tsSymbolKey]string
}

func NewTsSymbolMap() *TsSymbolMap {
	return &TsSymbolMap{
		symbolMap: make(map[tsSymbolKey]string),
	}
}

// IsEmpty returns true if there are no mappings.
func (tsm *TsSymbolMap) IsEmpty() bool {
	return len(tsm.symbolMap) == 0
}

// Len returns the number of symbol mappings in the map.
func (tsm *TsSymbolMap) Len() int {
	return len(tsm.symbolMap)
}

// Get returns the symbol mapping for the given date and instrument ID.
// Returns empty string if no mapping exists.
func (tsm *TsSymbolMap) Get(dt time.Time, instrID uint32) string {
	ymd := TimeToYMD(dt)
	key := tsSymbolKey{Date: ymd, ID: instrID}
	symbol, ok := tsm.symbolMap[key]
	if !ok {
		return ""
	}
	return symbol
}

// FillFromMetadata fills the TsSymbolMap with mappings from `metadata`.
func (tsm *TsSymbolMap) FillFromMetadata(metadata *Metadata) error {
	// clear existing mappings
	tsm.symbolMap = make(map[tsSymbolKey]string)

	// handle inverse mappings distinctly
	invMapping, err := metadata.IsInverseMapping()
	if err != nil {
		return err
	}
	if invMapping {
		for _, mapping := range metadata.Mappings {
			// instrID comes from mapping, NOT interval
			instrID, err := strconv.Atoi(mapping.RawSymbol)
			if err != nil {
				return err // really?
			}
			for _, interval := range mapping.Intervals {
				// handle old symbology format
				if interval.Symbol == "" {
					continue
				}
				tsm.Insert(uint32(instrID), interval.StartDate, interval.EndDate, interval.Symbol)
			}
		}
	} else {
		for _, mapping := range metadata.Mappings {
			for _, interval := range mapping.Intervals {
				// handle old symbology format
				if interval.Symbol == "" {
					continue
				}
				// instrID comes from interval, NOT mapping
				instrID, err := strconv.Atoi(interval.Symbol)
				if err != nil {
					return err // really?
				}
				tsm.Insert(uint32(instrID), interval.StartDate, interval.EndDate, mapping.RawSymbol)
			}
		}
	}

	return nil
}

// Insert adds mappings for a date range.
// dates are YYYYMMDD ints
func (tsm *TsSymbolMap) Insert(instrID uint32, startDate uint32, endDate uint32, ticker string) error {
	// convert dates to time.Time
	startTime := YMDToTime(int(startDate), time.UTC)
	endTime := YMDToTime(int(endDate), time.UTC)
	if startTime.After(endTime) {
		return fmt.Errorf("startDate is after endDate")
	}

	// Iterate calendar days from startDate to endDate
	for d := startTime; d.Before(endTime) || d.Equal(endTime); d = d.AddDate(0, 0, 1) {
		ymd := TimeToYMD(d)
		key := tsSymbolKey{Date: ymd, ID: instrID}
		tsm.symbolMap[key] = ticker
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////

// PitSymbolMap is a point-in-time symbol map. Useful for working with live symbology or a
// historical request over a single day or other situations where the symbol
// mappings are known not to change.
// TOOD: handle nuance of int<>string and string<>string mappings based on SType
type PitSymbolMap struct {
	mapping    map[uint32]string
	mappingInv map[string]uint32
}

func NewPitSymbolMap() *PitSymbolMap {
	return &PitSymbolMap{
		mapping:    make(map[uint32]string),
		mappingInv: make(map[string]uint32),
	}
}

// IsEmpty returns `true` if there are no mappings.
func (p *PitSymbolMap) IsEmpty() bool {
	return len(p.mapping) != 0
}

// Returns the number of symbol mappings in the map.
func (p *PitSymbolMap) Len() int {
	return len(p.mapping)
}

// Returns the string mapping of the instrument ID, or empty string if not found.
func (p *PitSymbolMap) Get(instrumentID uint32) string {
	str, ok := p.mapping[instrumentID]
	if !ok {
		return ""
	}
	return str
}

// OnSymbolMappingMsg handles updating the mappings (if required) for a SymbolMappingMsg record.
func (p *PitSymbolMap) OnSymbolMappingMsg(symbolMapping *SymbolMappingMsg) error {
	// Apply from the header's instrumentID to its stype_out
	p.mapping[symbolMapping.Header.InstrumentID] = symbolMapping.StypeOutSymbol
	p.mappingInv[symbolMapping.StypeOutSymbol] = symbolMapping.Header.InstrumentID
	return nil
}

// Fills the PitSymbolMap with mappings from `metadata` for `date`, clearing any original contents
// Returns an error if any.
func (p *PitSymbolMap) FillFromMetadata(metadata *Metadata, timestamp uint64) error {
	// Validate symbol mapping in/out types
	if metadata.StypeIn != SType_InstrumentId && metadata.StypeOut != SType_InstrumentId {
		return ErrWrongStypesForMapping
	}
	// Validate time range
	if timestamp < metadata.Start || timestamp >= metadata.End {
		return ErrDateOutsideQueryRange
	}
	ymd := TimeToYMD(time.Unix(0, int64(timestamp)))

	isInverse, err := metadata.IsInverseMapping()
	if err != nil {
		return err
	}

	p.mapping = make(map[uint32]string, len(metadata.Mappings))
	p.mappingInv = make(map[string]uint32, len(metadata.Mappings))

	for _, mapping := range metadata.Mappings {
		for _, interval := range mapping.Intervals {
			// skip if outside interval
			if ymd < interval.StartDate || ymd >= interval.EndDate {
				continue
			}
			if len(interval.Symbol) == 0 {
				continue
			}

			if isInverse {
				instrID, err := strconv.Atoi(mapping.RawSymbol)
				if err != nil {
					return err
				}
				p.mapping[uint32(instrID)] = interval.Symbol
			} else {
				instrID, err := strconv.Atoi(interval.Symbol)
				if err != nil {
					return err
				}
				p.mapping[uint32(instrID)] = mapping.RawSymbol
			}
		}
	}
	return nil
}
