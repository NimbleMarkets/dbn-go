// Copyright (c) 2024 Neomantra Corp

//

package dbn

import (
	"strconv"
	"time"
)

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
	// TODO
	return ErrMalformedRecord
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
	is_inverse := metadata.IsInverseMapping()
	ymd := TimeToYMD(time.Unix(0, int64(timestamp)))

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
			atoi, err := strconv.Atoi(interval.Symbol)
			if err != nil {
				return ErrMalformedRecord
			}
			p.mapping[uint32(atoi)] = mapping.RawSymbol
			if is_inverse {
				p.mappingInv[mapping.RawSymbol] = uint32(atoi)
			}
		}
	}
	return nil
}

/*

/// Handles updating the mappings (if required) for a generic record.
///
/// # Errors
/// This function returns an error when `record` contains a [`SymbolMappingMsg`] but
/// it contains invalid UTF-8.
pub fn on_record(&mut self, record: RecordRef) -> crate::Result<()> {
	if matches!(record.rtype(), Ok(RType::SymbolMapping)) {
		// >= to allow WithTsOut
		if record.record_size() >= std::mem::size_of::<SymbolMappingMsg>() {
			// Safety: checked rtype and length
			self.on_symbol_mapping(unsafe { record.get_unchecked::<SymbolMappingMsg>() })
		} else {
			// Use get here to get still perform length checks
			self.on_symbol_mapping(record.get::<compat::SymbolMappingMsgV1>().unwrap())
		}
	} else {
		Ok(())
	}
}

/// Handles updating the mappings for a symbol mapping record.
///
/// # Errors
/// This function returns an error when `symbol_mapping` contains invalid UTF-8.
pub fn on_symbol_mapping<S: compat::SymbolMappingRec>(
	&mut self,
	symbol_mapping: &S,
) -> crate::Result<()> {
	let stype_out_symbol = symbol_mapping.stype_out_symbol()?;
	self.0.insert(
		symbol_mapping.header().instrument_id,
		stype_out_symbol.to_owned(),
	);
	Ok(())
}

/// Returns a reference to the mapping for the given instrument ID.
pub fn get(&self, instrument_id: u32) -> Option<&String> {
	self.0.get(&instrument_id)
}

/// Returns a reference to the inner map.
pub fn inner(&self) -> &HashMap<u32, String> {
	&self.0
}

/// Returns a mutable reference to the inner map.
pub fn inner_mut(&mut self) -> &mut HashMap<u32, String> {
	&mut self.0
}
}*/
