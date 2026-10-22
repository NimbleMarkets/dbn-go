// Copyright (c) 2024-2025 Neomantra Corp

//

package dbn

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

// symbolInterval is a half-open timestamp range. Intervals for an instrument are
// sorted, non-overlapping, and coalesced when adjacent symbols are equal.
type symbolInterval struct {
	start, end uint64
	symbol     string
}

// TsSymbolMap maps instrument IDs to symbols over time. It can be built from
// metadata or updated from live records with scanner.Visit(tsm).
// The zero value is ready to use. Concurrent mutation requires synchronization.
type TsSymbolMap struct {
	NullVisitor
	symbolMap map[uint32][]symbolInterval
}

var _ Visitor = (*TsSymbolMap)(nil)

func NewTsSymbolMap() *TsSymbolMap { return &TsSymbolMap{} }

// IsEmpty returns true if there are no mappings.
func (tsm *TsSymbolMap) IsEmpty() bool { return len(tsm.symbolMap) == 0 }

// Len returns the number of stored intervals across all instruments, not the
// number of days covered. Adjacent intervals with the same symbol are coalesced.
func (tsm *TsSymbolMap) Len() int {
	n := 0
	for _, intervals := range tsm.symbolMap {
		n += len(intervals)
	}
	return n
}

// Get returns the mapping at UTC midnight of dt's calendar date in its location,
// preserving the date-based API. Use GetForTs for intraday lookups.
// Returns an empty string if no mapping exists or the date is not representable.
func (tsm *TsSymbolMap) Get(dt time.Time, instrID uint32) string {
	ts, err := symbolDateTimestamp(TimeToYMD(dt))
	if err != nil {
		return ""
	}
	return tsm.GetForTs(ts, instrID)
}

// GetForTs returns the symbol at a UNIX timestamp in nanoseconds, or an empty
// string if no mapping exists. UNDEF_TIMESTAMP never matches an interval.
func (tsm *TsSymbolMap) GetForTs(ts uint64, instrID uint32) string {
	intervals := tsm.symbolMap[instrID]
	i := sort.Search(len(intervals), func(i int) bool { return intervals[i].end > ts })
	if i < len(intervals) && intervals[i].start <= ts {
		return intervals[i].symbol
	}
	return ""
}

// FillFromMetadata replaces the map with metadata's mappings. If validation
// fails, the original map is left unchanged.
func (tsm *TsSymbolMap) FillFromMetadata(metadata *Metadata) error {
	invMapping, err := metadata.IsInverseMapping()
	if err != nil {
		return err
	}
	next := NewTsSymbolMap()
	for _, mapping := range metadata.Mappings {
		for _, interval := range mapping.Intervals {
			if interval.Symbol == "" {
				continue
			}
			id, symbol := interval.Symbol, mapping.RawSymbol
			if invMapping {
				id, symbol = mapping.RawSymbol, interval.Symbol
			}
			instrID, err := strconv.ParseUint(id, 10, 32)
			if err != nil {
				return fmt.Errorf("symbol mapping instrument ID %q: %w", id, err)
			}
			if err := next.Insert(uint32(instrID), interval.StartDate, interval.EndDate, symbol); err != nil {
				return fmt.Errorf("symbol mapping for instrument %d: %w", instrID, err)
			}
		}
	}
	tsm.symbolMap = next.symbolMap
	return nil
}

// Insert adds mappings for [startDate, endDate), with an exclusive end date.
// Dates are YYYYMMDD UTC dates within the DBN timestamp range. An empty range
// leaves the map unchanged. Use InsertTs for intraday mappings.
func (tsm *TsSymbolMap) Insert(instrID uint32, startDate uint32, endDate uint32, ticker string) error {
	start, err := symbolDateTimestamp(startDate)
	if err != nil {
		return err
	}
	end, err := symbolDateTimestamp(endDate)
	if err != nil {
		return err
	}
	return tsm.InsertTs(instrID, start, end, ticker)
}

// InsertTs adds a mapping for [startTs, endTs), expressed as UNIX nanoseconds.
// The new mapping replaces overlapping portions of existing mappings, preserving
// their non-overlapping portions. Equal bounds are a no-op; reversed bounds
// return an error. UNDEF_TIMESTAMP may be used as an open-ended end bound.
func (tsm *TsSymbolMap) InsertTs(instrID uint32, startTs, endTs uint64, ticker string) error {
	if startTs > endTs {
		return fmt.Errorf("mapping start timestamp is after end timestamp")
	}
	if startTs == endTs {
		return nil
	}
	if tsm.symbolMap == nil {
		tsm.symbolMap = make(map[uint32][]symbolInterval)
	}
	intervals := tsm.symbolMap[instrID]
	added := symbolInterval{startTs, endTs, ticker}
	// Chronological updates are common. Append without copying older intervals.
	if len(intervals) == 0 || intervals[len(intervals)-1].end <= startTs {
		tsm.symbolMap[instrID] = appendSymbolInterval(intervals, added)
		return nil
	}
	// Split overwritten intervals at the new bounds and coalesce equal neighbors.
	result := make([]symbolInterval, 0, len(intervals)+2)
	inserted := false
	for _, old := range intervals {
		if old.end <= startTs {
			result = appendSymbolInterval(result, old)
			continue
		}
		if old.start >= endTs {
			if !inserted {
				result = appendSymbolInterval(result, added)
				inserted = true
			}
			result = appendSymbolInterval(result, old)
			continue
		}
		if old.start < startTs {
			result = appendSymbolInterval(result, symbolInterval{old.start, startTs, old.symbol})
		}
		if !inserted {
			result = appendSymbolInterval(result, added)
			inserted = true
		}
		if old.end > endTs {
			result = appendSymbolInterval(result, symbolInterval{endTs, old.end, old.symbol})
		}
	}
	if !inserted {
		result = appendSymbolInterval(result, added)
	}
	tsm.symbolMap[instrID] = result
	return nil
}

func appendSymbolInterval(intervals []symbolInterval, next symbolInterval) []symbolInterval {
	if len(intervals) > 0 {
		last := &intervals[len(intervals)-1]
		if last.end == next.start && last.symbol == next.symbol {
			last.end = next.end
			return intervals
		}
	}
	return append(intervals, next)
}

func symbolDateTimestamp(date uint32) (uint64, error) {
	dt := YMDToTime(int(date), time.UTC)
	seconds := dt.Unix()
	if date == 0 || TimeToYMD(dt) != date || seconds < 0 || uint64(seconds) > (UNDEF_TIMESTAMP-1)/uint64(time.Second) {
		return 0, fmt.Errorf("invalid symbol mapping date %d", date)
	}
	return uint64(seconds) * uint64(time.Second), nil
}

// OnSymbolMappingMsg updates the map from a live mapping record. An undefined
// start applies from the UNIX epoch; an undefined end applies indefinitely.
// The record header's instrument ID maps to StypeOutSymbol, matching PitSymbolMap.
// Other record types are ignored through the embedded NullVisitor.
func (tsm *TsSymbolMap) OnSymbolMappingMsg(record *SymbolMappingMsg) error {
	start := record.StartTs
	if start == UNDEF_TIMESTAMP {
		start = 0
	}
	return tsm.InsertTs(record.Header.InstrumentID, start, record.EndTs, record.StypeOutSymbol)
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
	return len(p.mapping) == 0
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
	ymd := TimeToYMD(time.Unix(0, int64(timestamp)).UTC())

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

			id, symbol := interval.Symbol, mapping.RawSymbol
			if isInverse {
				id, symbol = mapping.RawSymbol, interval.Symbol
			}
			instrID, err := strconv.ParseUint(id, 10, 32)
			if err != nil {
				return fmt.Errorf("symbol mapping instrument ID %q: %w", id, err)
			}
			p.mapping[uint32(instrID)] = symbol
		}
	}
	return nil
}
