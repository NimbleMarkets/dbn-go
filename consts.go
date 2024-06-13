// Copyright (c) 2024 Neomantra Corp
//
// Adapted from DataBento's DBN:
//   https://github.com/databento/dbn/blob/main/rust/dbn/src/enums.rs
//

package dbn

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Side
type Side uint8

const (
	// A sell order or sell aggressor in a trade.
	Side_Ask Side = 'A'
	// A buy order or a buy aggressor in a trade.
	Side_Bid Side = 'B'
	// No side specified by the original source.
	Side_None Side = 'N'
)

// Action
type Action uint8

const (
	// An existing order was modified.
	Action_Modify Action = 'M'
	// A trade executed.
	Action_Trade Action = 'T'
	// An existing order was filled.
	Action_Fill Action = 'F'
	// An order was cancelled.
	Action_Cancel Action = 'C'
	// A new order was added.
	Action_Add Action = 'A'
	// Reset the book; clear all orders for an instrument.
	Action_Clear Action = 'R'
)

// InstrumentClass
type InstrumentClass uint8

const (
	// A bond.
	InstrumentClass_Bond InstrumentClass = 'B'
	// A call option.
	InstrumentClass_Call InstrumentClass = 'C'
	// A future.
	InstrumentClass_Future InstrumentClass = 'F'
	// A stock.
	InstrumentClass_Stock InstrumentClass = 'K'
	// A spread composed of multiple instrument classes
	InstrumentClass_MixedSpread InstrumentClass = 'M'
	// A put option
	InstrumentClass_Put InstrumentClass = 'P'
	// A spread composed of futures.
	InstrumentClass_FutureSpread InstrumentClass = 'S'
	// A spread composed of options.
	InstrumentClass_OptionSpread InstrumentClass = 'T'
	// A foreign exchange spot.
	InstrumentClass_FxSpot InstrumentClass = 'X'
)

func (i InstrumentClass) IsOption() bool {
	return i == InstrumentClass_Call || i == InstrumentClass_Put || i == InstrumentClass_OptionSpread
}

func (i InstrumentClass) IsFuture() bool {
	return i == InstrumentClass_Future || i == InstrumentClass_FutureSpread
}

func (i InstrumentClass) IsSpread() bool {
	return i == InstrumentClass_FutureSpread || i == InstrumentClass_OptionSpread || i == InstrumentClass_MixedSpread
}

// MatchAlgorithm
type MatchAlgorithm uint8

const (
	// First-in-first-out matching.
	MatchAlgorithm_Fifo MatchAlgorithm = 'F'
	// A configurable match algorithm.
	MatchAlgorithm_Configurable MatchAlgorithm = 'K'
	// Trade quantity is allocated to resting orders based on a pro-rata percentage:
	// resting order quantity divided by total quantity.
	MatchAlgorithm_ProRata MatchAlgorithm = 'C'
	// Like [`Self::Fifo`] but with LMM allocations prior to FIFO allocations.
	MatchAlgorithm_FifoLmm MatchAlgorithm = 'T'
	// Like [`Self::ProRata`] but includes a configurable allocation to the first order that improves the market.
	MatchAlgorithm_ThresholdProRata MatchAlgorithm = 'O'
	// Like [`Self::FifoLmm`] but includes a configurable allocation to the first order that improves the market.
	MatchAlgorithm_FifoTopLmm MatchAlgorithm = 'S'
	// Like [`Self::ThresholdProRata`] but includes a special priority to LMMs.
	MatchAlgorithm_ThresholdProRataLmm MatchAlgorithm = 'Q'
	// Special variant used only for Eurodollar futures on CME. See
	// [CME documentation](https://www.cmegroup.com/confluence/display/EPICSANDBOX/Supported+Matching+Algorithms#SupportedMatchingAlgorithms-Pro-RataAllocationforEurodollarFutures).
	MatchAlgorithm_EurodollarFutures MatchAlgorithm = 'Y'
)

// UserDefinedInstrument
type UserDefinedInstrument uint8

const (
	/// The instrument is not user-defined.
	UserDefinedInstrument_No UserDefinedInstrument = 'N'
	/// The instrument is user-defined.
	UserDefinedInstrument_Yes UserDefinedInstrument = 'Y'
)

// SType Symbology type
type SType uint8

const (
	/// Symbology using a unique numeric ID.
	SType_InstrumentId SType = 0
	/// Symbology using the original symbols provided by the publisher.
	SType_RawSymbol SType = 1
	/// Deprecated: A set of Databento-specific symbologies for referring to groups of symbols.
	SType_Smart SType = 2
	/// A Databento-specific symbology where one symbol may point to different
	/// instruments at different points of time, e.g. to always refer to the front month
	/// future.
	SType_Continuous SType = 3
	/// A Databento-specific symbology for referring to a group of symbols by one
	/// "parent" symbol, e.g. ES.FUT to refer to all ES futures.
	SType_Parent SType = 4
	/// Symbology for US equities using NASDAQ Integrated suffix conventions.
	SType_Nasdaq SType = 5
	/// Symbology for US equities using CMS suffix conventions.
	SType_Cms SType = 6
)

// Returns the string representation of the SType, or empty string if unknown.
func (s SType) String() string {
	switch s {
	case SType_InstrumentId:
		return "instrument_id"
	case SType_RawSymbol:
		return "raw_symbol"
	case SType_Smart:
		return "smart"
	case SType_Continuous:
		return "continuous"
	case SType_Parent:
		return "parent"
	case SType_Nasdaq:
		return "nasdaq"
	case SType_Cms:
		return "cms"
	default:
		return ""
	}
}

// STypeFromString converts a string to an SType.
// Returns an error if the string is unknown.
// Possible string values: instrument_id, id, instr, raw_symbol, raw, smart, continuous, parent, nasdaq, cms
func STypeFromString(str string) (SType, error) {
	str = strings.ToLower(str)
	switch str {
	case "instrument_id", "id", "instr":
		return SType_InstrumentId, nil
	case "raw_symbol", "raw":
		return SType_RawSymbol, nil
	case "smart":
		return SType_Smart, nil
	case "continuous":
		return SType_Continuous, nil
	case "parent":
		return SType_Parent, nil
	case "nasdaq":
		return SType_Nasdaq, nil
	case "cms":
		return SType_Cms, nil
	default:
		return SType_InstrumentId, fmt.Errorf("unknown stype: %s", str)
	}
}

func (s SType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *SType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	js, err := STypeFromString(str)
	if err != nil {
		return err
	}
	*s = js
	return nil
}

// Type implements pflag.Value.Type.  Returns "dbn.SType".
func (*SType) Type() string {
	return "dbn.SType"
}

// Set implements the flag.Value interface.
func (s *SType) Set(value string) error {
	stype, err := STypeFromString(value)
	if err == nil {
		*s = stype
	}
	return err
}

///////////////////////////////////////////////////////////////////////////////

type RType uint8

const (
	// Sentinel values for different DBN record types.
	// comments from: https://github.com/databento/dbn/blob/main/rust/dbn/src/enums.rs
	RType_Mbp0            RType = 0x00 // Denotes a market-by-price record with a book depth of 0 (used for the Trades schema)
	RType_Mbp1            RType = 0x01 // Denotes a market-by-price record with a book depth of 1 (also used for the Tbbo schema)
	RType_Mbp10           RType = 0x0A // Denotes a market-by-price record with a book depth of 10.
	RType_OhlcvDeprecated RType = 0x11 // Deprecated in 0.4.0. Denotes an open, high, low, close, and volume record at an unspecified cadence.
	RType_Ohlcv1S         RType = 0x20 // Denotes an open, high, low, close, and volume record at a 1-second cadence.
	RType_Ohlcv1M         RType = 0x21 // Denotes an open, high, low, close, and volume record at a 1-minute cadence.
	RType_Ohlcv1H         RType = 0x22 // Denotes an open, high, low, close, and volume record at an hourly cadence.
	RType_Ohlcv1D         RType = 0x23 // Denotes an open, high, low, close, and volume record at a daily cadence based on the UTC date.
	RType_OhlcvEod        RType = 0x24 // Denotes an open, high, low, close, and volume record at a daily cadence based on the end of the trading session.
	RType_Status          RType = 0x12 // Denotes an exchange status record.
	RType_InstrumentDef   RType = 0x13 // Denotes an instrument definition record.
	RType_Imbalance       RType = 0x14 // Denotes an order imbalance record.
	RType_Error           RType = 0x15 // Denotes an error from gateway.
	RType_SymbolMapping   RType = 0x16 // Denotes a symbol mapping record.
	RType_System          RType = 0x17 // Denotes a non-error message from the gateway. Also used for heartbeats.
	RType_Statistics      RType = 0x18 // Denotes a statistics record from the publisher (not calculated by Databento).
	RType_Mbo             RType = 0xA0 // Denotes a market by order record.
	RType_Cbbo            RType = 0xB1 /// Denotes a consolidated best bid and offer record.
	RType_Cbbo1S          RType = 0xC0 /// Denotes a consolidated best bid and offer record subsampled on a one-second interval.
	RType_Cbbo1M          RType = 0xC1 /// Denotes a consolidated best bid and offer record subsampled on a one-minute interval.
	RType_Tcbbo           RType = 0xC2 /// Denotes a consolidated best bid and offer trade record containing the consolidated BBO before the trade.
	RType_Bbo1S           RType = 0xC3 /// Denotes a best bid and offer record subsampled on a one-second interval.
	RType_Bbo1M           RType = 0xC4 /// Denotes a best bid and offer record subsampled on a one-minute interval.
	RType_Unknown         RType = 0xFF // Golang-only: Unknown or invalid record type
)

// Returns the string representation of the RType, or empty string if unknown.
func (s RType) String() string {
	switch s {
	case RType_Mbp0:
		return "mbp-0"
	case RType_Mbp1:
		return "mbp-1"
	case RType_Mbp10:
		return "mbp-10"
	case RType_OhlcvDeprecated:
		return "ohlcv-deprecated"
	case RType_Ohlcv1S:
		return "ohlcv-1s"
	case RType_Ohlcv1M:
		return "ohlcv-1m"
	case RType_Ohlcv1H:
		return "ohlcv-1h"
	case RType_Ohlcv1D:
		return "ohlcv-1d"
	case RType_OhlcvEod:
		return "ohlcv-eod"
	case RType_Status:
		return "status"
	case RType_InstrumentDef:
		return "instrument-def"
	case RType_Imbalance:
		return "imbalance"
	case RType_Error:
		return "error"
	case RType_SymbolMapping:
		return "symbol-mapping"
	case RType_System:
		return "system"
	case RType_Statistics:
		return "statistics"
	case RType_Mbo:
		return "mbo"
	case RType_Cbbo:
		return "cbbo"
	case RType_Cbbo1S:
		return "cbbo-1s"
	case RType_Cbbo1M:
		return "cbbo-1m"
	case RType_Tcbbo:
		return "tcbbo"
	case RType_Bbo1S:
		return "bbo-1s"
	case RType_Bbo1M:
		return "bbo-1m"
	case RType_Unknown:
		return "unknown"
	}
	return ""
}

///////////////////////////////////////////////////////////////////////////////

type Schema uint16

const (
	/// The data record schema. u16::MAX indicates a potential mix of schemas and record types, which will always be the case for live data.
	Schema_Mixed Schema = 0xFFFF
	/// Market by order.
	Schema_Mbo Schema = 0
	/// Market by price with a book depth of 1.
	Schema_Mbp1 Schema = 1
	/// Market by price with a book depth of 10.
	Schema_Mbp10 Schema = 2
	/// All trade events with the best bid and offer (BBO) immediately **before** the effect of the trade.
	Schema_Tbbo Schema = 3
	/// All trade events.
	Schema_Trades Schema = 4
	/// Open, high, low, close, and volume at a one-second interval.
	Schema_Ohlcv1S Schema = 5
	/// Open, high, low, close, and volume at a one-minute interval.
	Schema_Ohlcv1M Schema = 6
	/// Open, high, low, close, and volume at an hourly interval.
	Schema_Ohlcv1H Schema = 7
	/// Open, high, low, close, and volume at a daily interval based on the UTC date.
	Schema_Ohlcv1D Schema = 8
	/// Instrument definitions.
	Schema_Definition Schema = 9
	/// Additional data disseminated by publishers.
	Schema_Statistics Schema = 10
	/// Trading status events.
	Schema_Status Schema = 11
	/// Auction imbalance events.
	Schema_Imbalance Schema = 12
	/// Open, high, low, close, and volume at a daily cadence based on the end of the trading session.
	Schema_OhlcvEod Schema = 13
	/// Consolidated best bid and offer.
	Schema_Cbbo Schema = 14
	/// Consolidated best bid and offer subsampled at one-second intervals, in addition to trades.
	Schema_Cbbo1S Schema = 15
	/// Consolidated best bid and offer subsampled at one-minute intervals, in addition to trades.
	Schema_Cbbo1M Schema = 16
	/// All trade events with the consolidated best bid and offer (CBBO) immediately **before** the effect of the trade.6
	Schema_Tcbbo Schema = 17
	/// Best bid and offer subsampled at one-second intervals, in addition to trades.
	Schema_Bbo1S Schema = 18
	/// Best bid and offer subsampled at one-minute intervals, in addition to trades.
	Schema_Bbo1M Schema = 19
)

// Returns the string representation of the Schema, or empty string if unknown.
func (s Schema) String() string {
	switch s {
	case Schema_Mixed:
		return "mixed"
	case Schema_Mbo:
		return "mbo"
	case Schema_Mbp1:
		return "mbp-1"
	case Schema_Mbp10:
		return "mbp-10"
	case Schema_Tbbo:
		return "tbbo"
	case Schema_Trades:
		return "trades"
	case Schema_Ohlcv1S:
		return "ohlcv-1s"
	case Schema_Ohlcv1M:
		return "ohlcv-1m"
	case Schema_Ohlcv1H:
		return "ohlcv-1h"
	case Schema_Ohlcv1D:
		return "ohlcv-1d"
	case Schema_Definition:
		return "definition"
	case Schema_Statistics:
		return "statistics"
	case Schema_Status:
		return "status"
	case Schema_Imbalance:
		return "imbalance"
	case Schema_OhlcvEod:
		return "ohlcv-eod"
	case Schema_Cbbo:
		return "cbbo"
	case Schema_Cbbo1S:
		return "cbbo-1s"
	case Schema_Cbbo1M:
		return "cbbo-1m"
	case Schema_Tcbbo:
		return "tcbbo"
	case Schema_Bbo1S:
		return "bbo-1s"
	case Schema_Bbo1M:
		return "bbo-1m"
	default:
		return ""
	}
}

// SchemaFromString converts a string to a Schema.
// Returns an error if the string is unknown.
func SchemaFromString(str string) (Schema, error) {
	str = strings.ToLower(str)
	switch str {
	case "mixed":
		return Schema_Mixed, nil
	case "mbo":
		return Schema_Mbo, nil
	case "mbp-1":
		return Schema_Mbp1, nil
	case "mbp-10":
		return Schema_Mbp10, nil
	case "tbbo":
		return Schema_Tbbo, nil
	case "trades":
		return Schema_Trades, nil
	case "ohlcv-1s":
		return Schema_Ohlcv1S, nil
	case "ohlcv-1m":
		return Schema_Ohlcv1M, nil
	case "ohlcv-1h":
		return Schema_Ohlcv1H, nil
	case "ohlcv-1d":
		return Schema_Ohlcv1D, nil
	case "definition":
		return Schema_Definition, nil
	case "statistics":
		return Schema_Statistics, nil
	case "status":
		return Schema_Status, nil
	case "imbalance":
		return Schema_Imbalance, nil
	case "ohlcv-eod":
		return Schema_OhlcvEod, nil
	case "cbbo":
		return Schema_Cbbo, nil
	case "cbbo-1s":
		return Schema_Cbbo1S, nil
	case "cbbo-1m":
		return Schema_Cbbo1M, nil
	case "tcbbo":
		return Schema_Tcbbo, nil
	case "bbo-1s":
		return Schema_Bbo1S, nil
	case "bbo-1m":
		return Schema_Bbo1M, nil
	default:
		return Schema_Mixed, fmt.Errorf("unknown schema: %s", str)
	}
}

func (s Schema) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Schema) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	js, err := SchemaFromString(str)
	if err != nil {
		return err
	}
	*s = js
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// / Encoding A data encoding format.
type Encoding uint8

const (
	/// Databento Binary Encoding.
	Encoding_Dbn Encoding = 0
	/// Comma-separated values.
	Encoding_Csv Encoding = 1
	/// JavaScript object notation.
	Encoding_Json Encoding = 2
)

// Returns the string representation of the Encoding, or empty string if unknown.
func (e Encoding) String() string {
	switch e {
	case Encoding_Dbn:
		return "dbn"
	case Encoding_Csv:
		return "csv"
	case Encoding_Json:
		return "json"
	default:
		return ""
	}
}

// EncodingFromString converts a string to an Encoding.
// Returns an error if the string is unknown.
func EncodingFromString(str string) (Encoding, error) {
	str = strings.ToLower(str)
	switch str {
	case "dbn":
		return Encoding_Dbn, nil
	case "csv":
		return Encoding_Csv, nil
	case "json":
		return Encoding_Json, nil
	default:
		return Encoding_Dbn, fmt.Errorf("unknown encoding: %s", str)
	}
}

func (e Encoding) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *Encoding) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	js, err := EncodingFromString(str)
	if err != nil {
		return err
	}
	*e = js
	return nil
}

// Type implements pflag.Value.Type.  Returns "dbn.Encoding".
func (*Encoding) Type() string {
	return "dbn.Encoding"
}

// Set implements the flag.Value interface.
func (e *Encoding) Set(value string) error {
	enc, err := EncodingFromString(value)
	if err == nil {
		*e = enc
	}
	return err
}

///////////////////////////////////////////////////////////////////////////////

// Compression is the compression format or none if uncompressed.
type Compression uint8

const (
	/// Uncompressed.
	Compress_None Compression = 0
	/// Zstandard compressed.
	Compress_ZStd Compression = 1
)

// Returns the string representation of the Compression ('zstd' or 'none'), or empty string if unknown.
func (c Compression) String() string {
	switch c {
	case Compress_None:
		return "none"
	case Compress_ZStd:
		return "zstd"
	default:
		return ""
	}
}

// CompressionFromString converts a string to a Compression.
// Returns an error if the string is unknown.
func CompressionFromString(str string) (Compression, error) {
	str = strings.ToLower(str)
	switch str {
	case "none":
		return Compress_None, nil
	case "zstd":
		return Compress_ZStd, nil
	default:
		return Compress_None, fmt.Errorf("unknown encoding: %s", str)
	}
}

func (c Compression) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *Compression) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	js, err := CompressionFromString(str)
	if err != nil {
		return err
	}
	*c = js
	return nil
}

// Type implements pflag.Value.Type.  Returns "dbn.Compression".
func (*Compression) Type() string {
	return "dbn.Compression"
}

// Set implements the flag.Value interface.
func (c *Compression) Set(value string) error {
	comp, err := CompressionFromString(value)
	if err == nil {
		*c = comp
	}
	return err
}

///////////////////////////////////////////////////////////////////////////////

// / Constants for the bit flag record fields.
const (
	/// Indicates it's the last message in the packet from the venue for a given
	/// `instrument_id`.
	RFlag_LAST uint8 = 1 << 7
	/// Indicates a top-of-book message, not an individual order.
	RFlag_TOB uint8 = 1 << 6
	/// Indicates the message was sourced from a replay, such as a snapshot server.
	RFlag_SNAPSHOT uint8 = 1 << 5
	/// Indicates an aggregated price level message, not an individual order.
	RFlag_MBP uint8 = 1 << 4
	/// Indicates the `ts_recv` value is inaccurate due to clock issues or packet
	/// reordering.
	RFlag_BAD_TS_RECV uint8 = 1 << 3
	/// Indicates an unrecoverable gap was detected in the channel.
	RFlag_MAYBE_BAD_BOOK uint8 = 1 << 2
)

// / The type of [`InstrumentDefMsg`](crate::record::InstrumentDefMsg) update.
type SecurityUpdateAction uint8

const (
	/// A new instrument definition.
	Add SecurityUpdateAction = 'A'
	/// A modified instrument definition of an existing one.
	Modify SecurityUpdateAction = 'M'
	/// Removal of an instrument definition.
	Delete SecurityUpdateAction = 'D'
	// Deprecated: Still present in legacy files."
	Invalid SecurityUpdateAction = '~'
)

// / The type of statistic contained in a [`StatMsg`](crate::record::StatMsg).
type StatType uint8

const (
	/// The price of the first trade of an instrument. `price` will be set.
	StatType_OpeningPrice StatType = 1
	/// The probable price of the first trade of an instrument published during pre-
	/// open. Both `price` and `quantity` will be set.
	StatType_IndicativeOpeningPrice StatType = 2
	/// The settlement price of an instrument. `price` will be set and `flags` indicate
	/// whether the price is final or preliminary and actual or theoretical. `ts_ref`
	/// will indicate the trading date of the settlement price.
	StatType_SettlementPrice StatType = 3
	/// The lowest trade price of an instrument during the trading session. `price` will
	/// be set.
	StatType_TradingSessionLowPrice StatType = 4
	/// The highest trade price of an instrument during the trading session. `price` will
	/// be set.
	StatType_TradingSessionHighPrice StatType = 5
	/// The number of contracts cleared for an instrument on the previous trading date.
	/// `quantity` will be set. `ts_ref` will indicate the trading date of the volume.
	StatType_ClearedVolume StatType = 6
	/// The lowest offer price for an instrument during the trading session. `price`
	/// will be set.
	StatType_LowestOffer StatType = 7
	/// The highest bid price for an instrument during the trading session. `price`
	/// will be set.
	StatType_HighestBid StatType = 8
	/// The current number of outstanding contracts of an instrument. `quantity` will
	/// be set. `ts_ref` will indicate the trading date for which the open interest was
	/// calculated.
	StatType_OpenInterest StatType = 9
	/// The volume-weighted average price (VWAP) for a fixing period. `price` will be
	/// set.
	StatType_FixingPrice StatType = 10
	/// The last trade price during a trading session. `price` will be set.
	StatType_ClosePrice StatType = 11
	/// The change in price from the close price of the previous trading session to the
	/// most recent trading session. `price` will be set.
	StatType_NetChange StatType = 12
	/// The volume-weighted average price (VWAP) during the trading session.
	/// `price` will be set to the VWAP while `quantity` will be the traded
	/// volume.
	StatType_Vwap StatType = 13
)

// / The type of [`StatMsg`](crate::record::StatMsg) update.
type StatUpdateAction uint8

const (
	/// A new statistic.
	StatUpdateAction_New StatUpdateAction = 1
	/// A removal of a statistic.
	StatUpdateAction_Delete StatUpdateAction = 2
)

// / The primary enum for the type of [`StatusMsg`](crate::record::StatusMsg) update.
type StatusAction uint8

const (
	/// No change.
	StatusAction_None StatusAction = 0
	/// The instrument is in a pre-open period.
	StatusAction_PreOpen StatusAction = 1
	/// The instrument is in a pre-cross period.
	StatusAction_PreCross StatusAction = 2
	/// The instrument is quoting but not trading.
	StatusAction_Quoting StatusAction = 3
	/// The instrument is in a cross/auction.
	StatusAction_Cross StatusAction = 4
	/// The instrument is being opened through a trading rotation.
	StatusAction_Rotation StatusAction = 5
	/// A new price indication is available for the instrument.
	StatusAction_NewPriceIndication StatusAction = 6
	/// The instrument is trading.
	StatusAction_Trading StatusAction = 7
	/// Trading in the instrument has been halted.
	StatusAction_Halt StatusAction = 8
	/// Trading in the instrument has been paused.
	StatusAction_Pause StatusAction = 9
	/// Trading in the instrument has been suspended.
	StatusAction_Suspend StatusAction = 10
	/// The instrument is in a pre-close period.
	StatusAction_PreClose StatusAction = 11
	/// Trading in the instrument has closed.
	StatusAction_Close StatusAction = 12
	/// The instrument is in a post-close period.
	StatusAction_PostClose StatusAction = 13
	/// A change in short-selling restrictions.
	StatusAction_SsrChange StatusAction = 14
	/// The instrument is not available for trading, either trading has closed or been
	/// halted.
	StatusAction_NotAvailableForTrading StatusAction = 15
)

// / The secondary enum for a [`StatusMsg`](crate::record::StatusMsg) update, explains
// / the cause of a halt or other change in `action`.
type StatusReason uint8

const (
	/// No reason is given.
	StatusReason_None StatusAction = 0
	/// The change in status occurred as scheduled.
	StatusReason_Scheduled StatusAction = 1
	/// The instrument stopped due to a market surveillance intervention.
	StatusReason_SurveillanceIntervention StatusAction = 2
	/// The status changed due to activity in the market.
	StatusReason_MarketEvent StatusAction = 3
	/// The derivative instrument began trading.
	StatusReason_InstrumentActivation StatusAction = 4
	/// The derivative instrument expired.
	StatusReason_InstrumentExpiration StatusAction = 5
	/// Recovery in progress.
	StatusReason_RecoveryInProcess StatusAction = 6
	/// The status change was caused by a regulatory action.
	StatusReason_Regulatory StatusAction = 10
	/// The status change was caused by an administrative action.
	StatusReason_Administrative StatusAction = 11
	/// The status change was caused by the issuer not being compliance with regulatory
	/// requirements.
	StatusReason_NonCompliance StatusAction = 12
	/// Trading halted because the issuer's filings are not current.
	StatusReason_FilingsNotCurrent StatusAction = 13
	/// Trading halted due to an SEC trading suspension.
	StatusReason_SecTradingSuspension StatusAction = 14
	/// The status changed because a new issue is available.
	StatusReason_NewIssue StatusAction = 15
	/// The status changed because an issue is available.
	StatusReason_IssueAvailable StatusAction = 16
	/// The status changed because the issue was reviewed.
	StatusReason_IssuesReviewed StatusAction = 17
	/// The status changed because the filing requirements were satisfied.
	StatusReason_FilingReqsSatisfied StatusAction = 18
	/// Relevant news is pending.
	StatusReason_NewsPending StatusAction = 30
	/// Relevant news was released.
	StatusReason_NewsReleased StatusAction = 31
	/// The news has been fully disseminated and times are available for the resumption
	/// in quoting and trading.
	StatusReason_NewsAndResumptionTimes StatusAction = 32
	/// The relevants news was not forthcoming.
	StatusReason_NewsNotForthcoming StatusAction = 33
	/// Halted for order imbalance.
	StatusReason_OrderImbalance StatusAction = 40
	/// The instrument hit limit up or limit down.
	StatusReason_LuldPause StatusAction = 50
	/// An operational issue occurred with the venue.
	StatusReason_Operational StatusAction = 60
	/// The status changed until the exchange receives additional information.
	StatusReason_AdditionalInformationRequested StatusAction = 70
	/// Trading halted due to merger becoming effective.
	StatusReason_MergerEffective StatusAction = 80
	/// Trading is halted in an ETF due to conditions with the component securities.
	StatusReason_Etf StatusAction = 90
	/// Trading is halted for a corporate action.
	StatusReason_CorporateAction StatusAction = 100
	/// Trading is halted because the instrument is a new offering.
	StatusReason_NewSecurityOffering StatusAction = 110
	/// Halted due to the market-wide circuit breaker level 1.
	StatusReason_MarketWideHaltLevel1 StatusAction = 120
	/// Halted due to the market-wide circuit breaker level 2.
	StatusReason_MarketWideHaltLevel2 StatusAction = 121
	/// Halted due to the market-wide circuit breaker level 3.
	StatusReason_MarketWideHaltLevel3 StatusAction = 122
	/// Halted due to the carryover of a market-wide circuit breaker from the previous
	/// trading day.
	StatusReason_MarketWideHaltCarryover StatusAction = 123
	/// Resumption due to the end of the a market-wide circuit breaker halt.
	StatusReason_MarketWideHaltResumption StatusAction = 124
	/// Halted because quotation is not available.
	StatusReason_QuotationNotAvailable StatusAction = 130
)

// / Further information about a status update.
type TradingEvent uint8

const (
	/// No additional information given.
	TradingEvent_None TradingEvent = 0
	/// Order entry and modification are not allowed.
	TradingEvent_NoCancel TradingEvent = 1
	/// A change of trading session occurred. Daily statistics are reset.
	TradingEvent_ChangeTradingSession TradingEvent = 2
	/// Implied matching is available.
	TradingEvent_ImpliedMatchingOn TradingEvent = 3
	/// Implied matching is not available.
	TradingEvent_ImpliedMatchingOff TradingEvent = 4
)

// / An enum for representing unknown, true, or false values. Equivalent to
// / `Option<bool>` but with a human-readable repr.
type TriState uint8

const (
	/// The value is not applicable or not known.
	TriState_NotAvailable TradingEvent = '~'
	/// False
	TriState_No TradingEvent = 'N'
	/// True
	TriState_Yes TradingEvent = 'Y'
)

// / How to handle decoding DBN data from a prior version.
type VersionUpgradePolicy uint8

const (
	/// Decode data from previous versions as-is.
	VersionUpgradePolicy_AsIs VersionUpgradePolicy = 0
	/// Decode data from previous versions converting it to the latest version. This
	/// breaks zero-copy decoding for structs that need updating, but makes usage
	/// simpler.
	VersionUpgradePolicy_Upgrade VersionUpgradePolicy = 1
)
