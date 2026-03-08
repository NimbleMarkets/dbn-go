// Copyright (c) 2025 Neomantra Corp

package file

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/compress"
	pqfile "github.com/apache/arrow-go/v18/parquet/file"
	pqschema "github.com/apache/arrow-go/v18/parquet/schema"
)

func WriteDbnFileAsParquet(sourceFile string, forceZstdInput bool, destFile string) error {
	// Build the reader, grab the metadata, build a symbol map
	dbnFile, dbnCloser, err := dbn.MakeCompressedReader(sourceFile, forceZstdInput)
	if err != nil {
		return err
	}
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

	errWrite := scanAndWriteParquet(dbnScanner, rgw, dbnSymbolMap)

	// Flush and close the parquet writer
	errClose := rgw.Close()
	errFlush := pw.FlushWithFooter()
	return errors.Join(errWrite, errClose, errFlush)
}

///////////////////////////////////////////////////////////////////////////////

// ParquetSchemaForDbnSchema returns a GroupNode for the given dbnSchema
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
	case dbn.Schema_Statistics:
		return ParquetGroupNode_StatMsg()
	default:
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////

func scanAndWriteParquet(scanner *dbn.DbnScanner, rgw pqfile.BufferedRowGroupWriter, dbnSymbolMap *dbn.TsSymbolMap) error {
	metadata, _ := scanner.Metadata() // we already validated at caller
	switch metadata.Schema {
	case dbn.Schema_Ohlcv1S, dbn.Schema_Ohlcv1M, dbn.Schema_Ohlcv1H, dbn.Schema_Ohlcv1D:
		for scanner.Next() {
			if r, err := dbn.DbnScannerDecode[dbn.OhlcvMsg](scanner); err != nil {
				return err
			} else {
				if err := ParquetWriteRow_OhlcvMsg(rgw, r, dbnSymbolMap); err != nil {
					return err
				}
			}
		}
	case dbn.Schema_Trades:
		for scanner.Next() {
			if r, err := dbn.DbnScannerDecode[dbn.Mbp0Msg](scanner); err != nil {
				return err
			} else {
				if err := ParquetWriteRow_Mbp0Msg(rgw, r, dbnSymbolMap); err != nil {
					return err
				}
			}
		}
	case dbn.Schema_Mbp1, dbn.Schema_Tbbo:
		for scanner.Next() {
			if r, err := dbn.DbnScannerDecode[dbn.Mbp1Msg](scanner); err != nil {
				return err
			} else {
				if err := ParquetWriteRow_Mbp1Msg(rgw, r, dbnSymbolMap); err != nil {
					return err
				}
			}
		}
	case dbn.Schema_Imbalance:
		for scanner.Next() {
			if r, err := dbn.DbnScannerDecode[dbn.ImbalanceMsg](scanner); err != nil {
				return err
			} else {
				if err := ParquetWriteRow_ImbalanceMsg(rgw, r, dbnSymbolMap); err != nil {
					return err
				}
			}
		}
	case dbn.Schema_Statistics:
		for scanner.Next() {
			if r, err := dbn.DbnScannerDecode[dbn.StatMsg](scanner); err != nil {
				return err
			} else {
				if err := ParquetWriteRow_StatMsg(rgw, r, dbnSymbolMap); err != nil {
					return err
				}
			}
		}
	default:
		return fmt.Errorf("no converter for schema %s", metadata.Schema.String())
	}
	if err := scanner.Error(); err != nil && err != io.EOF {
		return err
	}
	return nil
}

func writeInt32Column(rgw pqfile.BufferedRowGroupWriter, idx int, value int32) error {
	cw, err := rgw.Column(idx)
	if err != nil {
		return fmt.Errorf("failed to get column %d: %w", idx, err)
	}
	writer, ok := cw.(*pqfile.Int32ColumnChunkWriter)
	if !ok {
		return fmt.Errorf("column %d has unexpected writer type %T", idx, cw)
	}
	if _, err := writer.WriteBatch([]int32{value}, []int16{1}, nil); err != nil {
		return fmt.Errorf("failed writing int32 column %d: %w", idx, err)
	}
	return nil
}

func writeInt64Column(rgw pqfile.BufferedRowGroupWriter, idx int, value int64) error {
	cw, err := rgw.Column(idx)
	if err != nil {
		return fmt.Errorf("failed to get column %d: %w", idx, err)
	}
	writer, ok := cw.(*pqfile.Int64ColumnChunkWriter)
	if !ok {
		return fmt.Errorf("column %d has unexpected writer type %T", idx, cw)
	}
	if _, err := writer.WriteBatch([]int64{value}, []int16{1}, nil); err != nil {
		return fmt.Errorf("failed writing int64 column %d: %w", idx, err)
	}
	return nil
}

func writeFloat64Column(rgw pqfile.BufferedRowGroupWriter, idx int, value float64) error {
	cw, err := rgw.Column(idx)
	if err != nil {
		return fmt.Errorf("failed to get column %d: %w", idx, err)
	}
	writer, ok := cw.(*pqfile.Float64ColumnChunkWriter)
	if !ok {
		return fmt.Errorf("column %d has unexpected writer type %T", idx, cw)
	}
	if _, err := writer.WriteBatch([]float64{value}, []int16{1}, nil); err != nil {
		return fmt.Errorf("failed writing float64 column %d: %w", idx, err)
	}
	return nil
}

func writeByteArrayColumn(rgw pqfile.BufferedRowGroupWriter, idx int, value parquet.ByteArray) error {
	cw, err := rgw.Column(idx)
	if err != nil {
		return fmt.Errorf("failed to get column %d: %w", idx, err)
	}
	writer, ok := cw.(*pqfile.ByteArrayColumnChunkWriter)
	if !ok {
		return fmt.Errorf("column %d has unexpected writer type %T", idx, cw)
	}
	if _, err := writer.WriteBatch([]parquet.ByteArray{value}, []int16{1}, nil); err != nil {
		return fmt.Errorf("failed writing byte-array column %d: %w", idx, err)
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
	if err := writeInt32Column(rgw, 0, int32(record.Header.RType)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 1, int32(record.Header.PublisherID)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 2, int32(record.Header.InstrumentID)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 3, dbn.Fixed9ToFloat64(record.Open)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 4, dbn.Fixed9ToFloat64(record.High)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 5, dbn.Fixed9ToFloat64(record.Low)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 6, dbn.Fixed9ToFloat64(record.Close)); err != nil {
		return err
	}
	if err := writeInt64Column(rgw, 7, int64(record.Volume)); err != nil {
		return err
	}
	recordTime := time.Unix(0, int64(record.Header.TsEvent)).UTC()
	dbnSymbol := dbnSymbolMap.Get(recordTime, record.Header.InstrumentID)
	if err := writeByteArrayColumn(rgw, 8, parquet.ByteArray(dbnSymbol)); err != nil {
		return err
	}
	if err := writeInt64Column(rgw, 9, int64(record.Header.TsEvent)); err != nil {
		return err
	}
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
	if err := writeInt64Column(rgw, 0, int64(record.Header.TsEvent)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 1, int32(record.Header.RType)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 2, int32(record.Header.PublisherID)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 3, int32(record.Header.InstrumentID)); err != nil {
		return err
	}
	if err := writeByteArrayColumn(rgw, 4, parquet.ByteArray{record.Action}); err != nil {
		return err
	}
	if err := writeByteArrayColumn(rgw, 5, parquet.ByteArray{record.Side}); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 6, int32(record.Depth)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 7, dbn.Fixed9ToFloat64(record.Price)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 8, int32(record.Size)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 9, int32(record.Flags)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 10, int32(record.TsInDelta)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 11, int32(record.Sequence)); err != nil {
		return err
	}
	recordTime := time.Unix(0, int64(record.Header.TsEvent)).UTC()
	dbnSymbol := dbnSymbolMap.Get(recordTime, record.Header.InstrumentID)
	if err := writeByteArrayColumn(rgw, 12, parquet.ByteArray(dbnSymbol)); err != nil {
		return err
	}
	if err := writeInt64Column(rgw, 13, int64(record.TsRecv)); err != nil {
		return err
	}
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
	if err := writeInt64Column(rgw, 0, int64(record.Header.TsEvent)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 1, int32(record.Header.RType)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 2, int32(record.Header.PublisherID)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 3, int32(record.Header.InstrumentID)); err != nil {
		return err
	}
	if err := writeByteArrayColumn(rgw, 4, parquet.ByteArray{record.Action}); err != nil {
		return err
	}
	if err := writeByteArrayColumn(rgw, 5, parquet.ByteArray{record.Side}); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 6, int32(record.Depth)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 7, dbn.Fixed9ToFloat64(record.Price)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 8, int32(record.Size)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 9, int32(record.Flags)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 10, int32(record.TsInDelta)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 11, int32(record.Sequence)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 12, dbn.Fixed9ToFloat64(record.Level.BidPx)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 13, dbn.Fixed9ToFloat64(record.Level.AskPx)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 14, int32(record.Level.BidSz)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 15, int32(record.Level.AskSz)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 16, int32(record.Level.BidCt)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 17, int32(record.Level.AskCt)); err != nil {
		return err
	}
	recordTime := time.Unix(0, int64(record.Header.TsEvent)).UTC()
	dbnSymbol := dbnSymbolMap.Get(recordTime, record.Header.InstrumentID)
	if err := writeByteArrayColumn(rgw, 18, parquet.ByteArray(dbnSymbol)); err != nil {
		return err
	}
	if err := writeInt64Column(rgw, 19, int64(record.TsRecv)); err != nil {
		return err
	}
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
	if err := writeInt64Column(rgw, 0, int64(record.Header.TsEvent)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 1, int32(record.Header.RType)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 2, int32(record.Header.PublisherID)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 3, int32(record.Header.InstrumentID)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 4, dbn.Fixed9ToFloat64(record.RefPrice)); err != nil {
		return err
	}
	if err := writeInt64Column(rgw, 5, int64(record.AuctionTime)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 6, dbn.Fixed9ToFloat64(record.ContBookClrPrice)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 7, dbn.Fixed9ToFloat64(record.AuctInterestClrPrice)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 8, dbn.Fixed9ToFloat64(record.SsrFillingPrice)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 9, dbn.Fixed9ToFloat64(record.IndMatchPrice)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 10, dbn.Fixed9ToFloat64(record.UpperCollar)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 11, dbn.Fixed9ToFloat64(record.LowerCollar)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 12, int32(record.PairedQty)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 13, int32(record.TotalImbalanceQty)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 14, int32(record.MarketImbalanceQty)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 15, int32(record.UnpairedQty)); err != nil {
		return err
	}
	if err := writeByteArrayColumn(rgw, 16, parquet.ByteArray{record.AuctionType}); err != nil {
		return err
	}
	if err := writeByteArrayColumn(rgw, 17, parquet.ByteArray{record.Side}); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 18, int32(record.AuctionStatus)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 19, int32(record.FreezeStatus)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 20, int32(record.NumExtensions)); err != nil {
		return err
	}
	if err := writeByteArrayColumn(rgw, 21, parquet.ByteArray{record.UnpairedSide}); err != nil {
		return err
	}
	if err := writeByteArrayColumn(rgw, 22, parquet.ByteArray{record.SignificantImbalance}); err != nil {
		return err
	}
	recordTime := time.Unix(0, int64(record.Header.TsEvent)).UTC()
	dbnSymbol := dbnSymbolMap.Get(recordTime, record.Header.InstrumentID)
	if err := writeByteArrayColumn(rgw, 23, parquet.ByteArray(dbnSymbol)); err != nil {
		return err
	}
	if err := writeInt64Column(rgw, 24, int64(record.TsRecv)); err != nil {
		return err
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////

// ParquetGroupNode_StatMsg returns the Parquet Schema's Group Node for StatMsg.
//
// optional int64 field_id=-1 ts_event (Timestamp(isAdjustedToUTC=true, timeUnit=nanoseconds, is_from_converted_type=false, force_set_converted_type=false));
// optional int32 field_id=-1 rtype (Int(bitWidth=8, isSigned=false));
// optional int32 field_id=-1 publisher_id (Int(bitWidth=16, isSigned=false));
// optional int32 field_id=-1 instrument_id (Int(bitWidth=32, isSigned=false));
// optional double field_id=-1 price;
// optional int64 field_id=-1 quantity (Int(bitWidth=32, isSigned=true));
// optional int32 field_id=-1 sequence (Int(bitWidth=32, isSigned=false));
// optional int32 field_id=-1 stat_type (Int(bitWidth=16, isSigned=false));
// optional int32 field_id=-1 channel_id (Int(bitWidth=16, isSigned=false));
// optional int32 field_id=-1 update_action (Int(bitWidth=8, isSigned=false));
// optional int32 field_id=-1 stat_flags (Int(bitWidth=8, isSigned=false));
// optional binary field_id=-1 symbol (String);
// optional int32 field_id=-1 ts_in_delta;
// optional int64 field_id=-1 ts_recv (Timestamp(isAdjustedToUTC=true, timeUnit=nanoseconds, is_from_converted_type=false, force_set_converted_type=false));
// optional int64 field_id=-1 ts_ref (Timestamp(isAdjustedToUTC=true, timeUnit=nanoseconds, is_from_converted_type=false, force_set_converted_type=false));
func ParquetGroupNode_StatMsg() *pqschema.GroupNode {
	return pqschema.MustGroup(pqschema.NewGroupNode("schema", parquet.Repetitions.Required, pqschema.FieldList{
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ts_event", parquet.Repetitions.Optional, pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos), parquet.Types.Int64, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("rtype", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("publisher_id", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(16, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("instrument_id", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, false), parquet.Types.Int32, 0, -1)),
		pqschema.NewFloat64Node("price", parquet.Repetitions.Optional, -1),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("quantity", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(32, true), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("sequence", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(16, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("stat_type", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(16, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("channel_id", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(16, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("update_action", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("stat_flags", parquet.Repetitions.Optional, pqschema.NewIntLogicalType(8, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("symbol", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.NewInt32Node("ts_in_delta", parquet.Repetitions.Optional, -1),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ts_recv", parquet.Repetitions.Optional, pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos), parquet.Types.Int64, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("ts_ref", parquet.Repetitions.Optional, pqschema.NewTimestampLogicalType(true, pqschema.TimeUnitNanos), parquet.Types.Int64, 0, -1)),
	}, -1))
}

func ParquetWriteRow_StatMsg(rgw pqfile.BufferedRowGroupWriter, record *dbn.StatMsg, dbnSymbolMap *dbn.TsSymbolMap) error {
	if err := writeInt64Column(rgw, 0, int64(record.Header.TsEvent)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 1, int32(record.Header.RType)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 2, int32(record.Header.PublisherID)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 3, int32(record.Header.InstrumentID)); err != nil {
		return err
	}
	if err := writeFloat64Column(rgw, 4, dbn.Fixed9ToFloat64(record.Price)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 5, int32(record.Quantity)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 6, int32(record.Sequence)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 7, int32(record.StatType)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 8, int32(record.ChannelID)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 9, int32(record.UpdateAction)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 10, int32(record.StatFlags)); err != nil {
		return err
	}
	recordTime := time.Unix(0, int64(record.Header.TsEvent)).UTC()
	dbnSymbol := dbnSymbolMap.Get(recordTime, record.Header.InstrumentID)
	if err := writeByteArrayColumn(rgw, 11, parquet.ByteArray(dbnSymbol)); err != nil {
		return err
	}
	if err := writeInt32Column(rgw, 12, int32(record.TsInDelta)); err != nil {
		return err
	}
	if err := writeInt64Column(rgw, 13, int64(record.TsRecv)); err != nil {
		return err
	}
	if err := writeInt64Column(rgw, 14, int64(record.TsRef)); err != nil {
		return err
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////

func WritePublishersAsParquet(publishers []dbn_hist.PublisherDetail, forceZstdInput bool, destFile string) error {
	// Prepare file for writing
	outfile, outfileCloser, err := dbn.MakeCompressedWriter(destFile, false)
	if err != nil {
		return fmt.Errorf("failed to create writer %w", err)
	}
	defer outfileCloser()

	// Prepare parquet writer
	pwProperties := parquet.NewWriterProperties(
		parquet.WithVersion(parquet.V2_LATEST),
		parquet.WithCompression(compress.Codecs.Snappy))
	pqGroupNode := ParquetGroupNode_Publisher()
	pw := pqfile.NewParquetWriter(outfile, pqGroupNode, pqfile.WithWriterProps(pwProperties))
	defer pw.Close()

	// Write the publishers
	rgw := pw.AppendBufferedRowGroup()
	var errWrite error
	for _, publisher := range publishers {
		if err := ParquetWriteRow_Publisher(rgw, &publisher); err != nil {
			errWrite = fmt.Errorf("failed to write publisher row %w", err)
			break
		}
	}

	// Flush and close the parquet writer
	errClose := rgw.Close()
	errFlush := pw.FlushWithFooter()
	return errors.Join(errWrite, errClose, errFlush)
}

// ParquetGroupNode_Publisher returns the Parquet Schema's Group Node for DBN Publishers.
//
// int32 field_id=-1 publisher_id (Int(bitWidth=16, isSigned=false));
// byte_array field_id=-1 dataset (String);
// byte_array field_id=-1 venue (String);
// optional byte_array field_id=-1 description (String);
func ParquetGroupNode_Publisher() *pqschema.GroupNode {
	return pqschema.MustGroup(pqschema.NewGroupNode("publisher", parquet.Repetitions.Required, pqschema.FieldList{
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeLogical("publisher_id", parquet.Repetitions.Required, pqschema.NewIntLogicalType(16, false), parquet.Types.Int32, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("dataset", parquet.Repetitions.Required, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("venue", parquet.Repetitions.Required, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
		pqschema.MustPrimitive(pqschema.NewPrimitiveNodeConverted("description", parquet.Repetitions.Optional, parquet.Types.ByteArray, pqschema.ConvertedTypes.UTF8, 0, 0, 0, -1)),
	}, -1))
}

// ParquetWriteRow_Publisher writes a single publisherDetail to the given RowGroupWriter.
func ParquetWriteRow_Publisher(rgw pqfile.BufferedRowGroupWriter, publisher *dbn_hist.PublisherDetail) error {
	if err := writeInt32Column(rgw, 0, int32(publisher.PublisherID)); err != nil {
		return err
	}
	if err := writeByteArrayColumn(rgw, 1, parquet.ByteArray(publisher.Dataset)); err != nil {
		return err
	}
	if err := writeByteArrayColumn(rgw, 2, parquet.ByteArray(publisher.Venue)); err != nil {
		return err
	}
	if err := writeByteArrayColumn(rgw, 3, parquet.ByteArray(publisher.Description)); err != nil {
		return err
	}
	return nil
}
