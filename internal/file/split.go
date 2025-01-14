// Copyright (c) 2025 Neomantra Corp

package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	"github.com/neomantra/ymdflag"
)

const ymdPathFormat = "2006" + string(filepath.Separator) + "01" + string(filepath.Separator) + "02"

// SplitFile splits a source file into "<feed>/<instrument_id>/Y/M/D/feed-YMD.type.dbn.zst"`
func SplitFile(sourceFilename string, destDir string, forceZstdInput bool, verbose bool) error {
	// Open file, possibly using decompression wrapper
	sourceReader, sourceCloser, err := dbn.MakeCompressedReader(sourceFilename, forceZstdInput)
	if err != nil {
		return fmt.Errorf("failed to open '%s' for reading: %w", sourceFilename, err)
	}
	defer sourceCloser.Close()

	// Start scanning the DBN file, first extracting its metadata
	dbnScanner := dbn.NewDbnScanner(sourceReader)
	sourceMetadata, err := dbnScanner.Metadata()
	if err != nil {
		return fmt.Errorf("failed to read metadata for '%s': %w", sourceFilename, err)
	}
	dbnSymbolMap := dbn.NewTsSymbolMap()
	dbnSymbolMap.FillFromMetadata(sourceMetadata)

	singleMetadata := dbn.Metadata{
		VersionNum:       dbn.HeaderVersion2,
		Schema:           sourceMetadata.Schema,
		Start:            sourceMetadata.Start,
		End:              sourceMetadata.End,
		Limit:            sourceMetadata.Limit,
		StypeIn:          dbn.SType_Parent,
		StypeOut:         dbn.SType_InstrumentId,
		TsOut:            sourceMetadata.TsOut,
		SymbolCstrLen:    sourceMetadata.SymbolCstrLen,
		Dataset:          sourceMetadata.Dataset,
		SchemaDefinition: nil,
		Symbols:          nil,
		Partial:          nil,
		NotFound:         nil,
		Mappings:         nil,
	}

	// Iterate over all the records
	writerMap := make(map[string]io.Writer)
	closerMap := make(map[string]func())
	defer func() {
		// Close all the files
		for _, closer := range closerMap {
			closer()
		}
	}()

	for dbnScanner.Next() {
		// Get the RHeader
		rheader, err := dbnScanner.GetLastHeader()
		if err != nil {
			return fmt.Errorf("failed to read rheader: %w", err)
		}
		recordTime := time.Unix(0, int64(rheader.TsEvent)).UTC()
		recordYMD := ymdflag.TimeToYMD(recordTime)
		recordYMDStr := strconv.Itoa(recordYMD)
		fileKey := fmt.Sprintf("%s-%d-%s", rheader.RType, rheader.InstrumentID, recordYMDStr)

		// Get the file handle (or create)
		writer, ok := writerMap[fileKey]
		if !ok {
			// TODO: check denylist
			// Create the dest file
			// <dest>/<dataset>/<symbol>/YYYY/MM/DD/<symbol>-YYYYYMMDD.<schema>.dbn.zst
			dbnSymbol := dbnSymbolMap.Get(recordTime, rheader.InstrumentID)
			datePath := recordTime.Format(ymdPathFormat)
			destPath := filepath.Join(destDir, sourceMetadata.Dataset, dbnSymbol, datePath)
			err := os.MkdirAll(destPath, os.ModePerm)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create dest path '%s': %s\n", destPath, err.Error())
				return err

			}

			destFile := fmt.Sprintf("%s.%d.%s.dbn.zst", dbnSymbol, recordYMD, sourceMetadata.Schema.String())
			fullDestPath := filepath.Join(destPath, destFile)

			destWriter, destCloser, err := dbn.MakeCompressedWriter(fullDestPath, true)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create dest file '%s': %s\n", fullDestPath, err.Error())
				return err
				// TODO: add to deny list
				// we don't want to keep creating failed files
			}
			writerMap[fileKey] = destWriter
			closerMap[fileKey] = destCloser
			writer = destWriter

			if verbose {
				fmt.Fprintf(os.Stderr, "writing to '%s'\n", fullDestPath)
			}

			singleMetadata.Symbols = []string{dbnSymbol}
			singleMetadata.Mappings = []dbn.SymbolMapping{
				{
					RawSymbol: dbnSymbol,
					Intervals: []dbn.MappingInterval{
						{
							StartDate: uint32(recordYMD),
							EndDate:   uint32(ymdflag.TimeToYMD(recordTime.AddDate(0, 0, 1))),
							Symbol:    strconv.Itoa(int(rheader.InstrumentID)),
						},
					},
				},
			}
			if err := singleMetadata.Write(writer); err != nil {
				fmt.Fprintf(os.Stderr, "failed to write file header '%s': %s\n", fullDestPath, err.Error())
				return err
			}
		}

		_, err = writer.Write(dbnScanner.GetLastRecord()[:dbnScanner.GetLastSize()])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to write record: %s\n", err.Error())
			return err
		}
	}

	// Check for any errors
	err = dbnScanner.Error()
	if err == io.EOF {
		// EOF is not propagated as an error
		err = nil
	}
	return err
}
