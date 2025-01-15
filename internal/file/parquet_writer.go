// Copyright (c) 2025 Neomantra Corp

package file

import (
	"fmt"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/compress"
	pqfile "github.com/apache/arrow-go/v18/parquet/file"
	pqschema "github.com/apache/arrow-go/v18/parquet/schema"
)

func WriteDbnFileAsParquet(sourceFile string, forceZstdInput bool, destFile string) error {
	// Build the reader, grab the metadata, build a symbol map
	dbnFile, dbnCloser, err := dbn.MakeCompressedReader(sourceFile, forceZstdInput)
	defer dbnCloser.Close()

	dbnScanner := dbn.NewDbnScanner(dbnFile)
	metadata, err := dbnScanner.Metadata()
	if err != nil {
		return fmt.Errorf("failed to read metadata %w", err)
	}

	dbnSymbolMap := dbn.NewTsSymbolMap()
	err = dbnSymbolMap.FillFromMetadata(metadata)
	if err != nil {
		return fmt.Errorf("failed to fill symbol map: %w", err)
	}

	// Prepare for writing
	outfile, outfileCloser, err := dbn.MakeCompressedWriter(destFile, false)
	if err != nil {
		return fmt.Errorf("failed to create writer %w", err)
	}
	defer outfileCloser()

	pwProperties := parquet.NewWriterProperties(
		parquet.WithVersion(parquet.V2_LATEST),
		parquet.WithCompression(compress.Codecs.Snappy))

	// Grab the appropriate Parquet schema
	pqGroupNode := ParquetGroupNodeForDbnSchema(metadata.Schema)
	if pqGroupNode == nil {
		return fmt.Errorf("no converter for schema %s", metadata.Schema.String())
	}

	pw := pqfile.NewParquetWriter(outfile, pqGroupNode, pqfile.WithWriterProps(pwProperties))
	defer pw.Close()

	rgw := pw.AppendBufferedRowGroup()

	err = scanAndWriteParquet(dbnScanner, rgw, dbnSymbolMap)

	// Flush and close the parquet writer
	rgw.Close()
	err = pw.FlushWithFooter()
	if err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// ParquetSchemaForDbnSchema returns a GroupNode
func ParquetGroupNodeForDbnSchema(dbnSchema dbn.Schema) *pqschema.GroupNode {
	switch dbnSchema {
	case dbn.Schema_Ohlcv1S, dbn.Schema_Ohlcv1M, dbn.Schema_Ohlcv1H, dbn.Schema_Ohlcv1D:
		return ParquetGroupNode_OhlcvMsg()
	case dbn.Schema_Trades:
		return ParquetGroupNode_Mbp0Msg()
	case dbn.Schema_Mbp1, dbn.Schema_Tbbo:
		return ParquetGroupNode_Mbp1Msg()
	case dbn.Schema_Imbalance:
		return ParquetGroupNode_ImbalanceMsg()
	default:
		return nil
	}
}

func scanAndWriteParquet(scanner *dbn.DbnScanner, rgw pqfile.BufferedRowGroupWriter, dbnSymbolMap *dbn.TsSymbolMap) error {
	metadata, _ := scanner.Metadata() // we already validated at caller
	switch metadata.Schema {
	case dbn.Schema_Ohlcv1S, dbn.Schema_Ohlcv1M, dbn.Schema_Ohlcv1H, dbn.Schema_Ohlcv1D:
		for scanner.Next() {
			if r, err := dbn.DbnScannerDecode[dbn.OhlcvMsg](scanner); err != nil {
				return err
			} else {
				ParquetWriteRow_OhlcvMsg(rgw, r, dbnSymbolMap)
			}
		}
	case dbn.Schema_Trades:
		for scanner.Next() {
			if r, err := dbn.DbnScannerDecode[dbn.Mbp0Msg](scanner); err != nil {
				return err
			} else {
				ParquetWriteRow_Mbp0Msg(rgw, r, dbnSymbolMap)
			}
		}
	case dbn.Schema_Mbp1, dbn.Schema_Tbbo:
		for scanner.Next() {
			if r, err := dbn.DbnScannerDecode[dbn.Mbp1Msg](scanner); err != nil {
				return err
			} else {
				ParquetWriteRow_Mbp1Msg(rgw, r, dbnSymbolMap)
			}
		}
	case dbn.Schema_Imbalance:
		for scanner.Next() {
			if r, err := dbn.DbnScannerDecode[dbn.ImbalanceMsg](scanner); err != nil {
				return err
			} else {
				ParquetWriteRow_ImbalanceMsg(rgw, r, dbnSymbolMap)
			}
		}
	default:
		return fmt.Errorf("no converter for schema %s", metadata.Schema.String())
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// ParquetGroupNode_OhlcvMsg returns the Parquet Schema's Group Node for OhlcvMsg.
//
// optional int32 field_id=-1 rtype (Int(bitWidth=8, isSigned=false));
// optional int32 field_id=-1 publisher_id (Int(bitWidth=16, isSigned=false));
// optional int32 field_id=-1 instrument_id (Int(bitWidth=32, isSigned=false));
// optional double field_id=-1 open;
// optional double field_id=-1 high;
// optional double field_id=-1 low;
// optional double field_id=-1 close;
// optional int64 field_id=-1 volume (Int(bitWidth=64, isSigned=false));
// optional binary field_id=-1 symbol (String);
// optional int64 field_id=-1 ts_event (Timestamp(isAdjustedToUTC=true, timeUnit=nanoseconds, is_from_converted_type=false, force_set_converted_type=false));
func ParquetGroupNode_OhlcvMsg() *pqschema.GroupNode {
	return pqschema.MustGroup(pqschema.NewGroupNode("schema", parquet.Repetitions.Required, pqschema.FieldList{
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("rtype", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("publisher_id", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(16, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("instrument_id", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.NewFloat64Node("open", parquet.Repetitions.Optional, -1),
		pqschema.NewFloat64Node("high", parquet.Repetitions.Optional, -1),
		pqschema.NewFloat64Node("low", parquet.Repetitions.Optional, -1),
		pqschema.NewFloat64Node("close", parquet.Repetitions.Optional, -1),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("volume", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(64, false), parquet.Types.Int64, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("symbol", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ts_event", parquet.Repetitions.Optional, pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos), parquet.Types.Int64, 0, -1)),
	}, -1))
}

func ParquetWriteRow_OhlcvMsg(rgw pqfile.BufferedRowGroupWriter, record *dbn.OhlcvMsg, dbnSymbolMap *dbn.TsSymbolMap) error {
	// TODO: handle errors
	cw, _ := rgw.Column(0)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.RType)}, []int16{1}, nil)
	cw, _ = rgw.Column(1)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.PublisherID)}, []int16{1}, nil)
	cw, _ = rgw.Column(2)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.InstrumentID)}, []int16{1}, nil)
	cw, _ = rgw.Column(3)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.Open)}, []int16{1}, nil)
	cw, _ = rgw.Column(4)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.High)}, []int16{1}, nil)
	cw, _ = rgw.Column(5)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.Low)}, []int16{1}, nil)
	cw, _ = rgw.Column(6)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.Close)}, []int16{1}, nil)
	cw, _ = rgw.Column(7)
	cw.(*pqfile.Int64ColumnChunkWriter).WriteBatch([]int64{int64(record.Volume)}, []int16{1}, nil)
	cw, _ = rgw.Column(8)
	recordTime := time.Unix(0, int64(record.Header.TsEvent)).UTC()
	dbnSymbol := dbnSymbolMap.Get(recordTime, record.Header.InstrumentID)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{parquet.ByteArray(dbnSymbol)}, []int16{1}, nil)
	cw, _ = rgw.Column(9)
	cw.(*pqfile.Int64ColumnChunkWriter).WriteBatch([]int64{int64(record.Header.TsEvent)}, []int16{1}, nil)
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// ParquetGroupNode_Mbp0Msg returns the Parquet Schema's Group Node for OhlcvMsg.
//
//	optional int64 field_id=-1 ts_event (Timestamp(isAdjustedToUTC=true, timeUnit=nanoseconds, is_from_converted_type=false, force_set_converted_type=false));
//	optional int32 field_id=-1 rtype (Int(bitWidth=8, isSigned=false));
//	optional int32 field_id=-1 publisher_id (Int(bitWidth=16, isSigned=false));
//	optional int32 field_id=-1 instrument_id (Int(bitWidth=32, isSigned=false));
//	optional binary field_id=-1 action (String);
//	optional binary field_id=-1 side (String);
//	optional int32 field_id=-1 depth (Int(bitWidth=8, isSigned=false));
//	optional double field_id=-1 price;
//	optional int32 field_id=-1 size (Int(bitWidth=32, isSigned=false));
//	optional int32 field_id=-1 flags (Int(bitWidth=8, isSigned=false));
//	optional int32 field_id=-1 ts_in_delta;
//	optional int32 field_id=-1 sequence (Int(bitWidth=32, isSigned=false));
//	optional binary field_id=-1 symbol (String);
//	optional int64 field_id=-1 ts_recv (Timestamp(isAdjustedToUTC=true, timeUnit=nanoseconds, is_from_converted_type=false, force_set_converted_type=false));
func ParquetGroupNode_Mbp0Msg() *pqschema.GroupNode {
	return pqschema.MustGroup(pqschema.NewGroupNode("schema", parquet.Repetitions.Required, pqschema.FieldList{
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ts_event", parquet.Repetitions.Optional, pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos), parquet.Types.Int64, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("rtype", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("publisher_id", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(16, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("instrument_id", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("action", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("side", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("depth", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.NewFloat64Node("price", parquet.Repetitions.Optional, -1),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("size", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("flags", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.NewInt32Node("ts_in_delta", parquet.Repetitions.Optional, -1),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("sequence", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("symbol", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ts_recv", parquet.Repetitions.Optional, pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos), parquet.Types.Int64, 0, -1)),
	}, -1))
}

func ParquetWriteRow_Mbp0Msg(rgw pqfile.BufferedRowGroupWriter, record *dbn.Mbp0Msg, dbnSymbolMap *dbn.TsSymbolMap) error {
	// TODO: handle errors
	cw, _ := rgw.Column(0)
	cw.(*pqfile.Int64ColumnChunkWriter).WriteBatch([]int64{int64(record.Header.TsEvent)}, []int16{1}, nil)
	cw, _ = rgw.Column(1)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.RType)}, []int16{1}, nil)
	cw, _ = rgw.Column(2)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.PublisherID)}, []int16{1}, nil)
	cw, _ = rgw.Column(3)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.InstrumentID)}, []int16{1}, nil)
	cw, _ = rgw.Column(4)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{{record.Action}}, []int16{1}, nil)
	cw, _ = rgw.Column(5)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{{record.Side}}, []int16{1}, nil)
	cw, _ = rgw.Column(6)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Depth)}, []int16{1}, nil)
	cw, _ = rgw.Column(7)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.Price)}, []int16{1}, nil)
	cw, _ = rgw.Column(8)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Size)}, []int16{1}, nil)
	cw, _ = rgw.Column(9)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Flags)}, []int16{1}, nil)
	cw, _ = rgw.Column(10)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.TsInDelta)}, []int16{1}, nil)
	cw, _ = rgw.Column(11)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Sequence)}, []int16{1}, nil)
	cw, _ = rgw.Column(12)
	recordTime := time.Unix(0, int64(record.Header.TsEvent)).UTC()
	dbnSymbol := dbnSymbolMap.Get(recordTime, record.Header.InstrumentID)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{parquet.ByteArray(dbnSymbol)}, []int16{1}, nil)
	cw, _ = rgw.Column(13)
	cw.(*pqfile.Int64ColumnChunkWriter).WriteBatch([]int64{int64(record.TsRecv)}, []int16{1}, nil)
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// ParquetGroupNode_Mbp1Msg returns the Parquet Schema's Group Node for OhlcvMsg.
//
// optional int64 field_id=-1 ts_event (Timestamp(isAdjustedToUTC=true, timeUnit=nanoseconds, is_from_converted_type=false, force_set_converted_type=false));
// optional int32 field_id=-1 rtype (Int(bitWidth=8, isSigned=false));
// optional int32 field_id=-1 publisher_id (Int(bitWidth=16, isSigned=false));
// optional int32 field_id=-1 instrument_id (Int(bitWidth=32, isSigned=false));
// optional binary field_id=-1 action (String);
// optional binary field_id=-1 side (String);
// optional int32 field_id=-1 depth (Int(bitWidth=8, isSigned=false));
// optional double field_id=-1 price;
// optional int32 field_id=-1 size (Int(bitWidth=32, isSigned=false));
// optional int32 field_id=-1 flags (Int(bitWidth=8, isSigned=false));
// optional int32 field_id=-1 ts_in_delta;
// optional int32 field_id=-1 sequence (Int(bitWidth=32, isSigned=false));
// optional double field_id=-1 bid_px_00;
// optional double field_id=-1 ask_px_00;
// optional int32 field_id=-1 bid_sz_00 (Int(bitWidth=32, isSigned=false));
// optional int32 field_id=-1 ask_sz_00 (Int(bitWidth=32, isSigned=false));
// optional int32 field_id=-1 bid_ct_00 (Int(bitWidth=32, isSigned=false));
// optional int32 field_id=-1 ask_ct_00 (Int(bitWidth=32, isSigned=false));
// optional binary field_id=-1 symbol (String);
// optional int64 field_id=-1 ts_recv (Timestamp(isAdjustedToUTC=true, timeUnit=nanoseconds, is_from_converted_type=false, force_set_converted_type=false));
func ParquetGroupNode_Mbp1Msg() *pqschema.GroupNode {
	return pqschema.MustGroup(pqschema.NewGroupNode("schema", parquet.Repetitions.Required, pqschema.FieldList{
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ts_event", parquet.Repetitions.Optional, pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos), parquet.Types.Int64, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("rtype", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("publisher_id", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(16, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("instrument_id", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("action", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("side", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("depth", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.NewFloat64Node("price", parquet.Repetitions.Optional, -1),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("size", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("flags", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.NewInt32Node("ts_in_delta", parquet.Repetitions.Optional, -1),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("sequence", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.NewFloat64Node("bid_px_00", parquet.Repetitions.Optional, -1),
		pqschema.NewFloat64Node("ask_px_00", parquet.Repetitions.Optional, -1),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("bid_sz_00", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ask_sz_00", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("bid_ct_00", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ask_ct_00", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("symbol", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ts_recv", parquet.Repetitions.Optional, pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos), parquet.Types.Int64, 0, -1)),
	}, -1))
}

func ParquetWriteRow_Mbp1Msg(rgw pqfile.BufferedRowGroupWriter, record *dbn.Mbp1Msg, dbnSymbolMap *dbn.TsSymbolMap) error {
	// TODO: handle errors
	cw, _ := rgw.Column(0)
	cw.(*pqfile.Int64ColumnChunkWriter).WriteBatch([]int64{int64(record.Header.TsEvent)}, []int16{1}, nil)
	cw, _ = rgw.Column(1)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.RType)}, []int16{1}, nil)
	cw, _ = rgw.Column(2)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.PublisherID)}, []int16{1}, nil)
	cw, _ = rgw.Column(3)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.InstrumentID)}, []int16{1}, nil)
	cw, _ = rgw.Column(4)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{{record.Action}}, []int16{1}, nil)
	cw, _ = rgw.Column(5)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{{record.Side}}, []int16{1}, nil)
	cw, _ = rgw.Column(6)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Depth)}, []int16{1}, nil)
	cw, _ = rgw.Column(7)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.Price)}, []int16{1}, nil)
	cw, _ = rgw.Column(8)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Size)}, []int16{1}, nil)
	cw, _ = rgw.Column(9)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Flags)}, []int16{1}, nil)
	cw, _ = rgw.Column(10)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.TsInDelta)}, []int16{1}, nil)
	cw, _ = rgw.Column(11)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Sequence)}, []int16{1}, nil)
	cw, _ = rgw.Column(12)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.Level.BidPx)}, []int16{1}, nil)
	cw, _ = rgw.Column(13)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.Level.AskPx)}, []int16{1}, nil)
	cw, _ = rgw.Column(14)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Level.BidSz)}, []int16{1}, nil)
	cw, _ = rgw.Column(15)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Level.AskSz)}, []int16{1}, nil)
	cw, _ = rgw.Column(16)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Level.BidCt)}, []int16{1}, nil)
	cw, _ = rgw.Column(17)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Level.AskCt)}, []int16{1}, nil)
	cw, _ = rgw.Column(18)
	recordTime := time.Unix(0, int64(record.Header.TsEvent)).UTC()
	dbnSymbol := dbnSymbolMap.Get(recordTime, record.Header.InstrumentID)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{parquet.ByteArray(dbnSymbol)}, []int16{1}, nil)
	cw, _ = rgw.Column(19)
	cw.(*pqfile.Int64ColumnChunkWriter).WriteBatch([]int64{int64(record.TsRecv)}, []int16{1}, nil)
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// ParquetGroupNode_ImbalanceMsg returns the Parquet Schema's Group Node for OhlcvMsg.
//
//	optional int64 field_id=-1 ts_event (Timestamp(isAdjustedToUTC=true, timeUnit=nanoseconds, is_from_converted_type=false, force_set_converted_type=false));
//	optional int32 field_id=-1 rtype (Int(bitWidth=8, isSigned=false));
//	optional int32 field_id=-1 publisher_id (Int(bitWidth=16, isSigned=false));
//	optional int32 field_id=-1 instrument_id (Int(bitWidth=32, isSigned=false));
//	optional double field_id=-1 ref_price;
//	optional int64 field_id=-1 auction_time (Int(bitWidth=64, isSigned=false));
//	optional double field_id=-1 cont_book_clr_price;
//	optional double field_id=-1 auct_interest_clr_price;
//	optional double field_id=-1 ssr_filling_price;
//	optional double field_id=-1 ind_match_price;
//	optional double field_id=-1 upper_collar;
//	optional double field_id=-1 lower_collar;
//	optional int32 field_id=-1 paired_qty (Int(bitWidth=32, isSigned=false));
//	optional int32 field_id=-1 total_imbalance_qty (Int(bitWidth=32, isSigned=false));
//	optional int32 field_id=-1 market_imbalance_qty (Int(bitWidth=32, isSigned=false));
//	optional int32 field_id=-1 unpaired_qty (Int(bitWidth=32, isSigned=false));
//	optional binary field_id=-1 auction_type (String);
//	optional binary field_id=-1 side (String);
//	optional int32 field_id=-1 auction_status (Int(bitWidth=8, isSigned=false));
//	optional int32 field_id=-1 freeze_status (Int(bitWidth=8, isSigned=false));
//	optional int32 field_id=-1 num_extensions (Int(bitWidth=8, isSigned=false));
//	optional binary field_id=-1 unpaired_side (String);
//	optional binary field_id=-1 significant_imbalance (String);
//	optional binary field_id=-1 symbol (String);
//	optional int64 field_id=-1 ts_recv (Timestamp(isAdjustedToUTC=true, timeUnit=nanoseconds, is_from_converted_type=false, force_set_converted_type=false));
func ParquetGroupNode_ImbalanceMsg() *pqschema.GroupNode {
	return pqschema.MustGroup(pqschema.NewGroupNode("schema", parquet.Repetitions.Required, pqschema.FieldList{
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ts_event", parquet.Repetitions.Optional, pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos), parquet.Types.Int64, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("rtype", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("publisher_id", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(16, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("instrument_id", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.NewFloat64Node("ref_price", parquet.Repetitions.Optional, -1),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("auction_time", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(64, false), parquet.Types.Int64, 0, -1)),
		pqschema.NewFloat64Node("cont_book_clr_price", parquet.Repetitions.Optional, -1),
		pqschema.NewFloat64Node("auct_interest_clr_price", parquet.Repetitions.Optional, -1),
		pqschema.NewFloat64Node("ssr_filling_price", parquet.Repetitions.Optional, -1),
		pqschema.NewFloat64Node("ind_match_price", parquet.Repetitions.Optional, -1),
		pqschema.NewFloat64Node("upper_collar", parquet.Repetitions.Optional, -1),
		pqschema.NewFloat64Node("lower_collar", parquet.Repetitions.Optional, -1),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("paired_qty", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("total_imbalance_qty", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("market_imbalance_qty", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("unpaired_qty", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("auction_type", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("side", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("auction_status", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("freeze_status", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("num_extensions", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("unpaired_side", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("significant_imbalance", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("symbol", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ts_recv", parquet.Repetitions.Optional, pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos), parquet.Types.Int64, 0, -1)),
	}, -1))
}

func ParquetWriteRow_ImbalanceMsg(rgw pqfile.BufferedRowGroupWriter, record *dbn.ImbalanceMsg, dbnSymbolMap *dbn.TsSymbolMap) error {
	// TODO: handle errors
	cw, _ := rgw.Column(0)
	cw.(*pqfile.Int64ColumnChunkWriter).WriteBatch([]int64{int64(record.Header.TsEvent)}, []int16{1}, nil)
	cw, _ = rgw.Column(1)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.RType)}, []int16{1}, nil)
	cw, _ = rgw.Column(2)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.PublisherID)}, []int16{1}, nil)
	cw, _ = rgw.Column(3)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.Header.InstrumentID)}, []int16{1}, nil)
	cw, _ = rgw.Column(4)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.RefPrice)}, []int16{1}, nil)
	cw, _ = rgw.Column(5)
	cw.(*pqfile.Int64ColumnChunkWriter).WriteBatch([]int64{int64(record.AuctionTime)}, []int16{1}, nil)
	cw, _ = rgw.Column(6)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.ContBookClrPrice)}, []int16{1}, nil)
	cw, _ = rgw.Column(7)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.AuctInterestClrPrice)}, []int16{1}, nil)
	cw, _ = rgw.Column(8)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.SsrFillingPrice)}, []int16{1}, nil)
	cw, _ = rgw.Column(9)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.IndMatchPrice)}, []int16{1}, nil)
	cw, _ = rgw.Column(10)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.UpperCollar)}, []int16{1}, nil)
	cw, _ = rgw.Column(11)
	cw.(*pqfile.Float64ColumnChunkWriter).WriteBatch([]float64{dbn.Fixed9ToFloat64(record.LowerCollar)}, []int16{1}, nil)
	cw, _ = rgw.Column(12)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.PairedQty)}, []int16{1}, nil)
	cw, _ = rgw.Column(13)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.TotalImbalanceQty)}, []int16{1}, nil)
	cw, _ = rgw.Column(14)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.MarketImbalanceQty)}, []int16{1}, nil)
	cw, _ = rgw.Column(15)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.UnpairedQty)}, []int16{1}, nil)
	cw, _ = rgw.Column(16)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{{record.AuctionType}}, []int16{1}, nil)
	cw, _ = rgw.Column(17)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{{record.Side}}, []int16{1}, nil)
	cw, _ = rgw.Column(18)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.AuctionStatus)}, []int16{1}, nil)
	cw, _ = rgw.Column(19)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.FreezeStatus)}, []int16{1}, nil)
	cw, _ = rgw.Column(20)
	cw.(*pqfile.Int32ColumnChunkWriter).WriteBatch([]int32{int32(record.NumExtensions)}, []int16{1}, nil)
	cw, _ = rgw.Column(21)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{{record.UnpairedSide}}, []int16{1}, nil)
	cw, _ = rgw.Column(22)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{{record.SignificantImbalance}}, []int16{1}, nil)
	cw, _ = rgw.Column(23)
	recordTime := time.Unix(0, int64(record.Header.TsEvent)).UTC()
	dbnSymbol := dbnSymbolMap.Get(recordTime, record.Header.InstrumentID)
	cw.(*pqfile.ByteArrayColumnChunkWriter).WriteBatch([]parquet.ByteArray{parquet.ByteArray(dbnSymbol)}, []int16{1}, nil)
	cw, _ = rgw.Column(24)
	cw.(*pqfile.Int64ColumnChunkWriter).WriteBatch([]int64{int64(record.TsRecv)}, []int16{1}, nil)
	return nil
}
