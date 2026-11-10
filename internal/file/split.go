// Copyright (c) 2025 Neomantra Corp

package file

import (
	"container/list"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	"github.com/klauspost/compress/zstd"
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
	if err := dbnSymbolMap.FillFromMetadata(sourceMetadata); err != nil {
		return fmt.Errorf("failed to build symbol map for '%s': %w", sourceFilename, err)
	}

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

	// Iterate over all the records, keeping a bounded set of output files open.
	outputs := newSplitOutputs(maxOpenSplitOutputs)
	defer outputs.closeAll()

	for dbnScanner.Next() {
		// Get the RHeader
		rheader, err := dbnScanner.GetLastHeader()
		if err != nil {
			return fmt.Errorf("failed to read RHeader: %w", err)
		}
		recordTime := time.Unix(0, int64(rheader.TsEvent)).UTC()
		recordYMD := ymdflag.TimeToYMD(recordTime)
		recordYMDStr := strconv.Itoa(recordYMD)
		fileKey := fmt.Sprintf("%s-%d-%s", rheader.RType, rheader.InstrumentID, recordYMDStr)

		// Get the open writer, reopen a file we closed earlier, or create a new one
		writer := outputs.get(fileKey)
		if writer == nil {
			if fullDestPath, started := outputs.paths[fileKey]; started {
				if writer, err = outputs.open(fileKey, fullDestPath, true); err != nil {
					fmt.Fprintf(os.Stderr, "failed to reopen dest file '%s': %s\n", fullDestPath, err.Error())
					return err
				}
			}
		}
		if writer == nil {
			// TODO: check denylist
			// Create the dest file
			// <dest>/<dataset>/<symbol>/YYYY/MM/DD/<symbol>-YYYYYMMDD.<schema>.dbn.zst
			// The dataset and symbol come from the source file, so they are
			// encoded to stay single path components inside destDir.
			dbnSymbol := dbnSymbolMap.GetForTs(rheader.TsEvent, rheader.InstrumentID)
			safeSymbol := SafePathComponent(dbnSymbol)
			datePath := recordTime.Format(ymdPathFormat)
			destPath := filepath.Join(destDir, SafePathComponent(sourceMetadata.Dataset), safeSymbol, datePath)
			destFile := fmt.Sprintf("%s.%d.%s.dbn.zst", safeSymbol, recordYMD, sourceMetadata.Schema.String())
			fullDestPath := filepath.Join(destPath, destFile)
			if rel, err := filepath.Rel(destDir, fullDestPath); err != nil || !filepath.IsLocal(rel) {
				return fmt.Errorf("refusing to write '%s' outside of '%s'", fullDestPath, destDir)
			}

			err := os.MkdirAll(destPath, os.ModePerm)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create dest path '%s': %s\n", destPath, err.Error())
				return err

			}

			writer, err = outputs.open(fileKey, fullDestPath, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create dest file '%s': %s\n", fullDestPath, err.Error())
				return err
				// TODO: add to deny list
				// we don't want to keep creating failed files
			}

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
	if closeErr := outputs.closeAll(); err == nil {
		err = closeErr
	}
	return err
}

// maxOpenSplitOutputs bounds how many output files SplitFile holds open at
// once. The source file decides how many outputs there are, so without a bound
// a small input could exhaust file descriptors and memory.
const maxOpenSplitOutputs = 256

// splitOutput is one open zstd-compressed output file.
type splitOutput struct {
	key  string
	file *os.File
	zw   *zstd.Encoder
}

func (o *splitOutput) close() error {
	err := o.zw.Close()
	if closeErr := o.file.Close(); err == nil {
		err = closeErr
	}
	return err
}

// splitOutputs holds the most recently used output files open, up to max.
// A file that is closed to make room is reopened in append mode when its key
// comes up again; the extra zstd frame this starts is valid in a .zst file.
type splitOutputs struct {
	max   int
	byKey map[string]*list.Element // key -> element of lru holding *splitOutput
	lru   *list.List               // front is most recently used
	paths map[string]string        // key -> path, for every file started so far
}

func newSplitOutputs(max int) *splitOutputs {
	return &splitOutputs{
		max:   max,
		byKey: make(map[string]*list.Element),
		lru:   list.New(),
		paths: make(map[string]string),
	}
}

// get returns the open writer for key, or nil if it is not open.
func (s *splitOutputs) get(key string) io.Writer {
	elem, ok := s.byKey[key]
	if !ok {
		return nil
	}
	s.lru.MoveToFront(elem)
	return elem.Value.(*splitOutput).zw
}

// open creates the file at path (or appends to it), closing the least recently
// used output first if the limit is reached.
func (s *splitOutputs) open(key string, path string, appendTo bool) (io.Writer, error) {
	for s.lru.Len() >= s.max {
		oldest := s.lru.Back()
		output := s.lru.Remove(oldest).(*splitOutput)
		delete(s.byKey, output.key)
		if err := output.close(); err != nil {
			return nil, fmt.Errorf("failed to close '%s': %w", s.paths[output.key], err)
		}
	}

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if appendTo {
		flags = os.O_WRONLY | os.O_APPEND
	}
	file, err := os.OpenFile(path, flags, 0666)
	if err != nil {
		return nil, err
	}
	zw, err := zstd.NewWriter(file)
	if err != nil {
		file.Close()
		return nil, err
	}
	s.byKey[key] = s.lru.PushFront(&splitOutput{key: key, file: file, zw: zw})
	s.paths[key] = path
	return zw, nil
}

// closeAll closes every open output and returns the first error.
func (s *splitOutputs) closeAll() error {
	var firstErr error
	for elem := s.lru.Front(); elem != nil; elem = elem.Next() {
		if err := elem.Value.(*splitOutput).close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	s.lru.Init()
	clear(s.byKey)
	return firstErr
}
