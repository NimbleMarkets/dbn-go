// Copyright (c) 2025 Neomantra Corp

package file

import (
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/NimbleMarkets/dbn-go"
	"github.com/apache/arrow-go/v18/parquet"
	pqfile "github.com/apache/arrow-go/v18/parquet/file"
)

// Expected Parquet column layouts, following Databento's field naming and
// encode order (ts_event first, symbol and ts_recv last, as for the other schemas).
var (
	wantBboColumns = []string{
		"ts_event", "rtype", "publisher_id", "instrument_id",
		"side", "price", "size", "flags", "sequence",
		"bid_px_00", "ask_px_00", "bid_sz_00", "ask_sz_00", "bid_ct_00", "ask_ct_00",
		"symbol", "ts_recv",
	}
	wantCmbp1Columns = []string{
		"ts_event", "rtype", "publisher_id", "instrument_id",
		"action", "side", "price", "size", "flags", "ts_in_delta",
		"bid_px_00", "ask_px_00", "bid_sz_00", "ask_sz_00", "bid_pb_00", "ask_pb_00",
		"symbol", "ts_recv",
	}
	wantStatusColumns = []string{
		"ts_event", "rtype", "publisher_id", "instrument_id",
		"action", "reason", "trading_event", "is_trading", "is_quoting", "is_short_sell_restricted",
		"symbol", "ts_recv",
	}
)

func wantMbp10Columns() []string {
	cols := []string{
		"ts_event", "rtype", "publisher_id", "instrument_id",
		"action", "side", "depth", "price", "size", "flags", "ts_in_delta", "sequence",
	}
	for i := 0; i < 10; i++ {
		for _, f := range []string{"bid_px", "ask_px", "bid_sz", "ask_sz", "bid_ct", "ask_ct"} {
			cols = append(cols, fmt.Sprintf("%s_%02d", f, i))
		}
	}
	return append(cols, "symbol", "ts_recv")
}

func TestWriteDbnFileAsParquet_MissingSchemas(t *testing.T) {
	dataDir := filepath.Join("..", "..", "tests", "data")
	tests := []struct {
		file  string
		cols  []string
		check func(t *testing.T, src string, pq *pqfile.Reader)
	}{
		{"test_data.bbo-1s.dbn", wantBboColumns, checkBboFirstRow},
		{"test_data.bbo-1s.v2.dbn.zst", wantBboColumns, checkBboFirstRow},
		{"test_data.bbo-1s.v3.dbn.zst", wantBboColumns, checkBboFirstRow},
		{"test_data.bbo-1m.dbn", wantBboColumns, checkBboFirstRow},
		{"test_data.bbo-1m.v2.dbn.zst", wantBboColumns, checkBboFirstRow},
		{"test_data.bbo-1m.v3.dbn.zst", wantBboColumns, checkBboFirstRow},
		{"test_data.cmbp-1.v2.dbn.zst", wantCmbp1Columns, checkCmbp1FirstRow},
		{"test_data.cmbp-1.v3.dbn.zst", wantCmbp1Columns, checkCmbp1FirstRow},
		{"test_data.mbp-10.v1.dbn.zst", wantMbp10Columns(), checkMbp10FirstRow},
		{"test_data.mbp-10.v2.dbn.zst", wantMbp10Columns(), checkMbp10FirstRow},
		{"test_data.mbp-10.v3.dbn.zst", wantMbp10Columns(), checkMbp10FirstRow},
		{"test_data.status.v2.dbn.zst", wantStatusColumns, checkStatusFirstRow},
		{"test_data.status.v3.dbn.zst", wantStatusColumns, checkStatusFirstRow},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			src := filepath.Join(dataDir, tt.file)
			dst := filepath.Join(t.TempDir(), "out.parquet")

			if err := WriteDbnFileAsParquet(src, false, dst); err != nil {
				t.Fatalf("WriteDbnFileAsParquet(%s) returned error: %v", tt.file, err)
			}

			pq, err := pqfile.OpenParquetFile(dst, false)
			if err != nil {
				t.Fatalf("OpenParquetFile: %v", err)
			}
			defer pq.Close()

			wantRows := countDBNRecords(t, src, false)
			if wantRows == 0 {
				t.Fatalf("fixture %s has no records", tt.file)
			}
			if got := pq.NumRows(); got != wantRows {
				t.Fatalf("row count: got %d want %d", got, wantRows)
			}

			if got := parquetColumnNames(pq); !reflect.DeepEqual(got, tt.cols) {
				t.Fatalf("columns:\n got  %v\n want %v", got, tt.cols)
			}

			tt.check(t, src, pq)
		})
	}
}

///////////////////////////////////////////////////////////////////////////////

func checkBboFirstRow(t *testing.T, src string, pq *pqfile.Reader) {
	r := firstDBNRecord[dbn.BboMsg](t, src)
	expectInt64(t, pq, "ts_event", int64(r.Header.TsEvent))
	expectInt32(t, pq, "rtype", int32(r.Header.RType))
	expectInt32(t, pq, "instrument_id", int32(r.Header.InstrumentID))
	expectString(t, pq, "side", string([]byte{r.Side}))
	expectFloat64(t, pq, "price", dbn.Fixed9ToFloat64(r.Price))
	expectInt32(t, pq, "size", int32(r.Size))
	expectInt32(t, pq, "sequence", int32(r.Sequence))
	expectFloat64(t, pq, "bid_px_00", dbn.Fixed9ToFloat64(r.Level.BidPx))
	expectFloat64(t, pq, "ask_px_00", dbn.Fixed9ToFloat64(r.Level.AskPx))
	expectInt32(t, pq, "bid_sz_00", int32(r.Level.BidSz))
	expectInt32(t, pq, "ask_ct_00", int32(r.Level.AskCt))
	expectInt64(t, pq, "ts_recv", int64(r.TsRecv))
}

func checkCmbp1FirstRow(t *testing.T, src string, pq *pqfile.Reader) {
	r := firstDBNRecord[dbn.Cmbp1Msg](t, src)
	expectInt64(t, pq, "ts_event", int64(r.Header.TsEvent))
	expectInt32(t, pq, "rtype", int32(r.Header.RType))
	expectInt32(t, pq, "publisher_id", int32(r.Header.PublisherID))
	expectString(t, pq, "action", string([]byte{r.Action}))
	expectString(t, pq, "side", string([]byte{r.Side}))
	expectFloat64(t, pq, "price", dbn.Fixed9ToFloat64(r.Price))
	expectInt32(t, pq, "size", int32(r.Size))
	expectInt32(t, pq, "ts_in_delta", r.TsInDelta)
	expectFloat64(t, pq, "bid_px_00", dbn.Fixed9ToFloat64(r.Level.BidPx))
	expectFloat64(t, pq, "ask_px_00", dbn.Fixed9ToFloat64(r.Level.AskPx))
	expectInt32(t, pq, "ask_sz_00", int32(r.Level.AskSz))
	expectInt32(t, pq, "bid_pb_00", int32(r.Level.BidPb))
	expectInt32(t, pq, "ask_pb_00", int32(r.Level.AskPb))
	expectInt64(t, pq, "ts_recv", int64(r.TsRecv))
}

func checkMbp10FirstRow(t *testing.T, src string, pq *pqfile.Reader) {
	r := firstDBNRecord[dbn.Mbp10Msg](t, src)
	expectInt64(t, pq, "ts_event", int64(r.Header.TsEvent))
	expectInt32(t, pq, "rtype", int32(r.Header.RType))
	expectString(t, pq, "action", string([]byte{r.Action}))
	expectInt32(t, pq, "depth", int32(r.Depth))
	expectFloat64(t, pq, "price", dbn.Fixed9ToFloat64(r.Price))
	expectInt32(t, pq, "size", int32(r.Size))
	expectInt32(t, pq, "sequence", int32(r.Sequence))
	for _, lvl := range []int{0, 3, 9} {
		l := r.Levels[lvl]
		expectFloat64(t, pq, fmt.Sprintf("bid_px_%02d", lvl), dbn.Fixed9ToFloat64(l.BidPx))
		expectFloat64(t, pq, fmt.Sprintf("ask_px_%02d", lvl), dbn.Fixed9ToFloat64(l.AskPx))
		expectInt32(t, pq, fmt.Sprintf("bid_sz_%02d", lvl), int32(l.BidSz))
		expectInt32(t, pq, fmt.Sprintf("ask_sz_%02d", lvl), int32(l.AskSz))
		expectInt32(t, pq, fmt.Sprintf("bid_ct_%02d", lvl), int32(l.BidCt))
		expectInt32(t, pq, fmt.Sprintf("ask_ct_%02d", lvl), int32(l.AskCt))
	}
	// Guard against a fixture whose deeper levels are all zero, which would
	// make the per-level assertions above trivially pass.
	if r.Levels[3].BidPx == 0 && r.Levels[3].AskPx == 0 {
		t.Fatalf("fixture level 3 is empty; pick a different level to assert")
	}
	expectInt64(t, pq, "ts_recv", int64(r.TsRecv))
}

func checkStatusFirstRow(t *testing.T, src string, pq *pqfile.Reader) {
	r := firstDBNRecord[dbn.StatusMsg](t, src)
	expectInt64(t, pq, "ts_event", int64(r.Header.TsEvent))
	expectInt32(t, pq, "rtype", int32(r.Header.RType))
	expectInt32(t, pq, "instrument_id", int32(r.Header.InstrumentID))
	expectInt32(t, pq, "action", int32(r.Action))
	expectInt32(t, pq, "reason", int32(r.Reason))
	expectInt32(t, pq, "trading_event", int32(r.TradingEvent))
	expectString(t, pq, "is_trading", string([]byte{r.IsTrading}))
	expectString(t, pq, "is_quoting", string([]byte{r.IsQuoting}))
	expectString(t, pq, "is_short_sell_restricted", string([]byte{r.IsShortSellRestricted}))
	expectInt64(t, pq, "ts_recv", int64(r.TsRecv))
}

///////////////////////////////////////////////////////////////////////////////

func firstDBNRecord[R dbn.Record, RP dbn.RecordPtr[R]](t *testing.T, filename string) *R {
	t.Helper()
	reader, closer, err := dbn.MakeCompressedReader(filename, false)
	if err != nil {
		t.Fatalf("MakeCompressedReader(%q): %v", filename, err)
	}
	defer closer.Close()

	scanner := dbn.NewDbnScanner(reader)
	if _, err := scanner.Metadata(); err != nil {
		t.Fatalf("Metadata(%q): %v", filename, err)
	}
	if !scanner.Next() {
		if err := scanner.Error(); err != nil && err != io.EOF {
			t.Fatalf("scanner error: %v", err)
		}
		t.Fatalf("no records in %q", filename)
	}
	r, err := dbn.DbnScannerDecode[R, RP](scanner)
	if err != nil {
		t.Fatalf("decode first record of %q: %v", filename, err)
	}
	return r
}

func parquetColumnNames(pq *pqfile.Reader) []string {
	schema := pq.MetaData().Schema
	names := make([]string, schema.NumColumns())
	for i := range names {
		names[i] = schema.Column(i).Name()
	}
	return names
}

func parquetColumnReader(t *testing.T, pq *pqfile.Reader, name string) pqfile.ColumnChunkReader {
	t.Helper()
	idx := pq.MetaData().Schema.ColumnIndexByName(name)
	if idx < 0 {
		t.Fatalf("parquet column %q not found", name)
	}
	cr, err := pq.RowGroup(0).Column(idx)
	if err != nil {
		t.Fatalf("column %q: %v", name, err)
	}
	return cr
}

func expectInt32(t *testing.T, pq *pqfile.Reader, name string, want int32) {
	t.Helper()
	cr, ok := parquetColumnReader(t, pq, name).(*pqfile.Int32ColumnChunkReader)
	if !ok {
		t.Fatalf("column %q is not int32", name)
	}
	vals, defs := make([]int32, 1), make([]int16, 1)
	if _, n, err := cr.ReadBatch(1, vals, defs, nil); err != nil || n != 1 {
		t.Fatalf("read %q: n=%d err=%v", name, n, err)
	}
	if vals[0] != want {
		t.Errorf("%s: got %d want %d", name, vals[0], want)
	}
}

func expectInt64(t *testing.T, pq *pqfile.Reader, name string, want int64) {
	t.Helper()
	cr, ok := parquetColumnReader(t, pq, name).(*pqfile.Int64ColumnChunkReader)
	if !ok {
		t.Fatalf("column %q is not int64", name)
	}
	vals, defs := make([]int64, 1), make([]int16, 1)
	if _, n, err := cr.ReadBatch(1, vals, defs, nil); err != nil || n != 1 {
		t.Fatalf("read %q: n=%d err=%v", name, n, err)
	}
	if vals[0] != want {
		t.Errorf("%s: got %d want %d", name, vals[0], want)
	}
}

func expectFloat64(t *testing.T, pq *pqfile.Reader, name string, want float64) {
	t.Helper()
	cr, ok := parquetColumnReader(t, pq, name).(*pqfile.Float64ColumnChunkReader)
	if !ok {
		t.Fatalf("column %q is not float64", name)
	}
	vals, defs := make([]float64, 1), make([]int16, 1)
	if _, n, err := cr.ReadBatch(1, vals, defs, nil); err != nil || n != 1 {
		t.Fatalf("read %q: n=%d err=%v", name, n, err)
	}
	if vals[0] != want {
		t.Errorf("%s: got %v want %v", name, vals[0], want)
	}
}

func expectString(t *testing.T, pq *pqfile.Reader, name string, want string) {
	t.Helper()
	cr, ok := parquetColumnReader(t, pq, name).(*pqfile.ByteArrayColumnChunkReader)
	if !ok {
		t.Fatalf("column %q is not byte-array", name)
	}
	vals, defs := make([]parquet.ByteArray, 1), make([]int16, 1)
	if _, n, err := cr.ReadBatch(1, vals, defs, nil); err != nil || n != 1 {
		t.Fatalf("read %q: n=%d err=%v", name, n, err)
	}
	if string(vals[0]) != want {
		t.Errorf("%s: got %q want %q", name, string(vals[0]), want)
	}
}
