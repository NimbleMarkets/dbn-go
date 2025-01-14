// Copyright (c) 2025 Neomantra Corp

package file

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/NimbleMarkets/dbn-go"
)

func WriteDbnFileAsJson(sourceFile string, forceZstdInput bool, writer io.Writer) error {
	dbnFile, dbnCloser, _ := dbn.MakeCompressedReader(sourceFile, forceZstdInput)
	defer dbnCloser.Close()

	dbnScanner := dbn.NewDbnScanner(dbnFile)
	_, err := dbnScanner.Metadata()
	if err != nil {
		return fmt.Errorf("scanner failed to read metadata: %w", err)
	}

	visitor := NewJsonWriterVisitor(writer)
	for dbnScanner.Next() {
		if err := dbnScanner.Visit(visitor); err != nil {
			return fmt.Errorf("json print failed: %w", err)
		}
	}
	if err := dbnScanner.Error(); err != nil && err != io.EOF {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// WriteAsJson writes a value marshalled as JSON to the writer, returning any error.
func WriteAsJson[T any](val *T, writer io.Writer) error {
	jstr, err := json.Marshal(val)
	if err != nil {
		return err
	}
	_, err = writer.Write(jstr)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte{'\n'})
	return err
}

////////////////////////////////////////////////////////////////////////////////

// JsonWriterVisitor is an implementation of all the dbn.Visitor interface.
// It marshals all the records as JSON anout outputs it to its Writer
type JsonWriterVisitor struct {
	writer io.Writer
}

// NewJsonWriterVisitor creates a new JsonWriterVisitor with the given writer.
func NewJsonWriterVisitor(writer io.Writer) *JsonWriterVisitor {
	return &JsonWriterVisitor{writer: writer}
}

func (v *JsonWriterVisitor) OnMbp0(record *dbn.Mbp0Msg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnMbp10(record *dbn.Mbp10Msg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnMbp1(record *dbn.Mbp1Msg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnMbo(record *dbn.MboMsg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnOhlcv(record *dbn.OhlcvMsg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnCbbo(record *dbn.CbboMsg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnImbalance(record *dbn.ImbalanceMsg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnStatMsg(record *dbn.StatMsg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnStatusMsg(record *dbn.StatusMsg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnInstrumentDefMsg(record *dbn.InstrumentDefMsg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnErrorMsg(record *dbn.ErrorMsg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnSystemMsg(record *dbn.SystemMsg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnSymbolMappingMsg(record *dbn.SymbolMappingMsg) error {
	return WriteAsJson(record, v.writer)
}

func (v *JsonWriterVisitor) OnStreamEnd() error {
	return nil
}
