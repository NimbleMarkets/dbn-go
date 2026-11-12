// Copyright (c) 2025 Neomantra Corp

package mcp_data

import (
	"strings"

	"github.com/NimbleMarkets/dbn-go"
	"github.com/NimbleMarkets/dbn-go/internal/file"
	pqschema "github.com/apache/arrow-go/v18/parquet/schema"
)

// Each cached {dataset}/{schema} directory is exposed as two DuckDB views:
//
//   - the raw view, named "XNAS.ITCH/trades", reads the Parquet files directly
//     and keeps nanosecond TIMESTAMPTZ_NS timestamps;
//   - the alias view, named xnas_itch_trades, is a plain identifier (no quoting
//     needed) whose timestamp columns are cast to microsecond TIMESTAMPTZ so
//     date_trunc, extract, time_bucket and friends work without a cast.
//
// Both are replaced atomically by ntduckbox.SetAllViews whenever the cache
// changes, so the alias can never outlive its raw view.

// aliasViewName maps a raw view name such as "XNAS.ITCH/ohlcv-1d" to a plain
// lowercase identifier such as "xnas_itch_ohlcv_1d".
func aliasViewName(viewName string) string {
	var b strings.Builder
	b.Grow(len(viewName))
	for _, r := range strings.ToLower(viewName) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

// timestampColumnsForSchema returns the Parquet column names that the writer
// encodes as nanosecond timestamps for the given DBN schema name (for example
// "ts_event", "ts_recv"). Unknown schemas yield nil.
func timestampColumnsForSchema(schemaName string) []string {
	schema, err := dbn.SchemaFromString(schemaName)
	if err != nil {
		return nil
	}
	node := file.ParquetGroupNodeForDbnSchema(schema)
	if node == nil {
		return nil
	}
	tsType := pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos)
	var cols []string
	for i := 0; i < node.NumFields(); i++ {
		prim, ok := node.Field(i).(*pqschema.PrimitiveNode)
		if ok && prim.LogicalType().Equals(tsType) {
			cols = append(cols, prim.Name())
		}
	}
	return cols
}

// aliasViewSQL returns the SELECT body for the alias view over rawView, casting
// each nanosecond timestamp column of schemaName to TIMESTAMPTZ.
func aliasViewSQL(rawView, schemaName string) string {
	from := ` FROM "` + strings.ReplaceAll(rawView, `"`, `""`) + `"`
	tsCols := timestampColumnsForSchema(schemaName)
	if len(tsCols) == 0 {
		return "SELECT *" + from
	}
	casts := make([]string, len(tsCols))
	for i, c := range tsCols {
		casts[i] = `"` + c + `"::TIMESTAMPTZ AS "` + c + `"`
	}
	return "SELECT * REPLACE (" + strings.Join(casts, ", ") + ")" + from
}
