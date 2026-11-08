// Copyright (c) 2025 Neomantra Corp

package file

import (
	"bytes"
	"testing"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	pqfile "github.com/apache/arrow-go/v18/parquet/file"
	pqschema "github.com/apache/arrow-go/v18/parquet/schema"
)

// Databento's UNDEF_PRICE and UNDEF_TIMESTAMP sentinels mean "no value". The
// Parquet columns are Optional, so they must be written as NULL rather than as
// the finite values i64::MAX/1e9 or -1ns (1969-12-31), which silently poison
// aggregates. This matches Databento's Python to_parquet, which emits NaN/NaT.

const (
	undefTestInstrumentID uint32 = 15144
	undefTestSymbol              = "SPY"
)

var (
	// 2026-09-15 13:31:00 UTC, the end of a bbo-1m interval.
	undefTestTsRecv = uint64(time.Date(2026, 9, 15, 13, 31, 0, 0, time.UTC).UnixNano())
	// A ts_event in a day the symbol map does NOT cover, so a lookup keyed on
	// ts_event cannot accidentally resolve to the right symbol.
	undefTestTsEventOtherDay = uint64(time.Date(2026, 9, 10, 13, 30, 59, 0, time.UTC).UnixNano())
)

// undefTestSymbolMap maps the test instrument to SPY on 2026-09-15 only.
func undefTestSymbolMap(t *testing.T) *dbn.TsSymbolMap {
	t.Helper()
	m := dbn.NewTsSymbolMap()
	if err := m.Insert(undefTestInstrumentID, 20260915, 20260916, undefTestSymbol); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	return m
}

// writeSingleRowParquet writes one row with writeRow and returns a reader over the result.
func writeSingleRowParquet(t *testing.T, node *pqschema.GroupNode, writeRow func(pqfile.BufferedRowGroupWriter) error) *pqfile.Reader {
	t.Helper()
	var buf bytes.Buffer
	pw := pqfile.NewParquetWriter(&buf, node)
	rgw, err := pw.AppendBufferedRowGroupChecked()
	if err != nil {
		t.Fatalf("AppendBufferedRowGroupChecked: %v", err)
	}
	if err := writeRow(rgw); err != nil {
		t.Fatalf("write row: %v", err)
	}
	if err := rgw.Close(); err != nil {
		t.Fatalf("row group Close: %v", err)
	}
	if err := pw.Close(); err != nil {
		t.Fatalf("writer Close: %v", err)
	}
	pq, err := pqfile.NewParquetReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("NewParquetReader: %v", err)
	}
	t.Cleanup(func() { pq.Close() })
	return pq
}

// expectNull asserts the first value of an int64 or float64 column is NULL.
func expectNull(t *testing.T, pq *pqfile.Reader, name string) {
	t.Helper()
	defs := make([]int16, 1)
	var total int64
	var n int
	var err error
	switch cr := parquetColumnReader(t, pq, name).(type) {
	case *pqfile.Int64ColumnChunkReader:
		total, n, err = cr.ReadBatch(1, make([]int64, 1), defs, nil)
	case *pqfile.Float64ColumnChunkReader:
		total, n, err = cr.ReadBatch(1, make([]float64, 1), defs, nil)
	default:
		t.Fatalf("column %q has unsupported reader type %T", name, cr)
	}
	if err != nil || total != 1 {
		t.Fatalf("read %q: total=%d err=%v", name, total, err)
	}
	if n != 0 || defs[0] != 0 {
		t.Errorf("%s: expected NULL, got %d value(s) with def level %d", name, n, defs[0])
	}
}

func TestParquetWriteRow_BboMsg_UndefSentinelsAreNull(t *testing.T) {
	record := &dbn.BboMsg{
		Header: dbn.RHeader{
			RType:        dbn.RType_Bbo1M,
			PublisherID:  95,
			InstrumentID: undefTestInstrumentID,
			TsEvent:      dbn.UNDEF_TIMESTAMP, // no trade in the interval
		},
		Price:  dbn.UNDEF_PRICE, // no last trade
		Size:   0,
		Side:   'N',
		Flags:  128,
		TsRecv: undefTestTsRecv,
		Level: dbn.BidAskPair{
			BidPx: dbn.UNDEF_PRICE, // one-sided book
			AskPx: 331_070_000_000,
			AskSz: 40,
		},
	}
	symbols := undefTestSymbolMap(t)
	pq := writeSingleRowParquet(t, ParquetGroupNode_BboMsg(), func(rgw pqfile.BufferedRowGroupWriter) error {
		return ParquetWriteRow_BboMsg(rgw, record, symbols)
	})

	expectNull(t, pq, "ts_event")
	expectNull(t, pq, "price")
	expectNull(t, pq, "bid_px_00")
	expectFloat64(t, pq, "ask_px_00", 331.07)
	expectInt32(t, pq, "ask_sz_00", 40)
	expectInt64(t, pq, "ts_recv", int64(undefTestTsRecv))
	// The symbol must resolve from ts_recv (the record's index timestamp),
	// not from the undefined ts_event.
	expectString(t, pq, "symbol", undefTestSymbol)
}

func TestParquetWriteRow_Mbp0Msg_SymbolResolvesFromTsRecv(t *testing.T) {
	record := &dbn.Mbp0Msg{
		Header: dbn.RHeader{
			RType:        dbn.RType_Mbp0,
			InstrumentID: undefTestInstrumentID,
			TsEvent:      undefTestTsEventOtherDay,
		},
		Action: 'T',
		Side:   'B',
		Price:  766_190_000_000,
		Size:   100,
		TsRecv: undefTestTsRecv,
	}
	symbols := undefTestSymbolMap(t)
	pq := writeSingleRowParquet(t, ParquetGroupNode_Mbp0Msg(), func(rgw pqfile.BufferedRowGroupWriter) error {
		return ParquetWriteRow_Mbp0Msg(rgw, record, symbols)
	})

	expectInt64(t, pq, "ts_event", int64(undefTestTsEventOtherDay))
	expectFloat64(t, pq, "price", 766.19)
	expectString(t, pq, "symbol", undefTestSymbol)
}

func TestParquetWriteRow_OhlcvMsg_UndefPricesAreNull(t *testing.T) {
	record := &dbn.OhlcvMsg{
		Header: dbn.RHeader{
			RType:        dbn.RType_Ohlcv1M,
			InstrumentID: undefTestInstrumentID,
			TsEvent:      undefTestTsRecv, // OHLCV has no ts_recv; ts_event is its index
		},
		Open:   dbn.UNDEF_PRICE,
		High:   dbn.UNDEF_PRICE,
		Low:    dbn.UNDEF_PRICE,
		Close:  dbn.UNDEF_PRICE,
		Volume: 0,
	}
	symbols := undefTestSymbolMap(t)
	pq := writeSingleRowParquet(t, ParquetGroupNode_OhlcvMsg(), func(rgw pqfile.BufferedRowGroupWriter) error {
		return ParquetWriteRow_OhlcvMsg(rgw, record, symbols)
	})

	for _, col := range []string{"open", "high", "low", "close"} {
		expectNull(t, pq, col)
	}
	expectInt64(t, pq, "volume", 0)
	expectInt64(t, pq, "ts_event", int64(undefTestTsRecv))
	expectString(t, pq, "symbol", undefTestSymbol)
}

func TestParquetWriteRow_StatMsg_UndefTsRefAndPriceAreNull(t *testing.T) {
	record := &dbn.StatMsg{
		Header: dbn.RHeader{
			RType:        dbn.RType_Statistics,
			InstrumentID: undefTestInstrumentID,
			TsEvent:      undefTestTsEventOtherDay,
		},
		TsRecv:   undefTestTsRecv,
		TsRef:    dbn.UNDEF_TIMESTAMP,
		Price:    dbn.UNDEF_PRICE,
		Quantity: 12345,
	}
	symbols := undefTestSymbolMap(t)
	pq := writeSingleRowParquet(t, ParquetGroupNode_StatMsg(), func(rgw pqfile.BufferedRowGroupWriter) error {
		return ParquetWriteRow_StatMsg(rgw, record, symbols)
	})

	expectNull(t, pq, "ts_ref")
	expectNull(t, pq, "price")
	expectInt64(t, pq, "quantity", 12345)
	expectInt64(t, pq, "ts_recv", int64(undefTestTsRecv))
	expectString(t, pq, "symbol", undefTestSymbol)
}

func TestParquetWriteRow_ImbalanceMsg_UndefAuctionTimeIsNull(t *testing.T) {
	record := &dbn.ImbalanceMsg{
		Header: dbn.RHeader{
			RType:        dbn.RType_Imbalance,
			InstrumentID: undefTestInstrumentID,
			TsEvent:      undefTestTsEventOtherDay,
		},
		TsRecv:               undefTestTsRecv,
		RefPrice:             229_430_000_000,
		AuctionTime:          dbn.UNDEF_TIMESTAMP,
		ContBookClrPrice:     dbn.UNDEF_PRICE,
		AuctInterestClrPrice: dbn.UNDEF_PRICE,
		SsrFillingPrice:      dbn.UNDEF_PRICE,
		IndMatchPrice:        dbn.UNDEF_PRICE,
		UpperCollar:          dbn.UNDEF_PRICE,
		LowerCollar:          dbn.UNDEF_PRICE,
	}
	symbols := undefTestSymbolMap(t)
	pq := writeSingleRowParquet(t, ParquetGroupNode_ImbalanceMsg(), func(rgw pqfile.BufferedRowGroupWriter) error {
		return ParquetWriteRow_ImbalanceMsg(rgw, record, symbols)
	})

	// auction_time is a timestamp like ts_event/ts_recv, not a bare integer.
	if lt := pq.MetaData().Schema.Column(pq.MetaData().Schema.ColumnIndexByName("auction_time")).LogicalType(); !lt.Equals(pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos)) {
		t.Errorf("auction_time logical type: got %s, want Timestamp(ns, UTC)", lt)
	}
	expectNull(t, pq, "auction_time")
	for _, col := range []string{"cont_book_clr_price", "auct_interest_clr_price", "ssr_filling_price", "ind_match_price", "upper_collar", "lower_collar"} {
		expectNull(t, pq, col)
	}
	expectFloat64(t, pq, "ref_price", 229.43)
	expectInt64(t, pq, "ts_recv", int64(undefTestTsRecv))
	expectString(t, pq, "symbol", undefTestSymbol)
}

// Guard: a defined value adjacent to a NULL must still be written with def level 1.
func TestParquetWriteRow_BboMsg_DefinedValuesKeepDefLevel(t *testing.T) {
	record := &dbn.BboMsg{
		Header: dbn.RHeader{RType: dbn.RType_Bbo1M, InstrumentID: undefTestInstrumentID, TsEvent: undefTestTsRecv - 1},
		Price:  331_130_000_000,
		TsRecv: undefTestTsRecv,
		Level:  dbn.BidAskPair{BidPx: 330_910_000_000, AskPx: 331_070_000_000},
	}
	symbols := undefTestSymbolMap(t)
	pq := writeSingleRowParquet(t, ParquetGroupNode_BboMsg(), func(rgw pqfile.BufferedRowGroupWriter) error {
		return ParquetWriteRow_BboMsg(rgw, record, symbols)
	})
	cr, ok := parquetColumnReader(t, pq, "price").(*pqfile.Float64ColumnChunkReader)
	if !ok {
		t.Fatalf("price column is not float64")
	}
	vals, defs := make([]float64, 1), make([]int16, 1)
	if _, n, err := cr.ReadBatch(1, vals, defs, nil); err != nil || n != 1 || defs[0] != 1 {
		t.Fatalf("price: n=%d def=%d err=%v", n, defs[0], err)
	}
	if vals[0] != 331.13 {
		t.Errorf("price: got %v want 331.13", vals[0])
	}
	expectInt64(t, pq, "ts_event", int64(undefTestTsRecv-1))
}
