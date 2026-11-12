// Copyright (c) 2024-2025 Neomantra Corp

package file

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NimbleMarkets/dbn-go"
)

func TestSplitFileRejectsInvalidSymbolMappings(t *testing.T) {
	for _, tc := range []struct {
		name      string
		startDate uint32
		id        string
	}{
		{"invalid_date", 20260230, "42"},
		{"negative_id", 20260101, "-1"},
		{"overflowing_id", 20260101, "4294967296"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := uint64(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano())
			metadata := dbn.Metadata{VersionNum: dbn.HeaderVersion3, Dataset: "TEST.DATA",
				Schema: dbn.Schema_Ohlcv1S, Start: start, End: start + uint64(24*time.Hour),
				StypeIn: dbn.SType_RawSymbol, StypeOut: dbn.SType_InstrumentId,
				Mappings: []dbn.SymbolMapping{{RawSymbol: "TEST", Intervals: []dbn.MappingInterval{
					{StartDate: tc.startDate, EndDate: 20260301, Symbol: tc.id},
				}}},
			}
			var buf bytes.Buffer
			if err := metadata.Write(&buf); err != nil {
				t.Fatal(err)
			}
			record := dbn.NewOhlcvMsg(dbn.RType_Ohlcv1S)
			record.Header.InstrumentID, record.Header.TsEvent = 42, start
			if err := binary.Write(&buf, binary.LittleEndian, record); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(t.TempDir(), "input.dbn")
			if err := os.WriteFile(source, buf.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
			dest := t.TempDir()
			if err := SplitFile(source, dest, false, false); err == nil || !strings.Contains(err.Error(), "failed to build symbol map") {
				t.Fatalf("expected symbol-map error, got %v", err)
			}
			entries, err := os.ReadDir(dest)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatal("invalid mappings produced output")
			}
		})
	}
}

// writeSplitSource writes a DBN file with one OHLCV record per (pass, instrument)
// and returns its path. Instrument IDs are 1..len(rawSymbols).
func writeSplitSource(t *testing.T, dataset string, rawSymbols []string, passes int) string {
	t.Helper()
	start := uint64(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano())
	metadata := dbn.Metadata{VersionNum: dbn.HeaderVersion3, Dataset: dataset,
		Schema: dbn.Schema_Ohlcv1S, Start: start, End: start + uint64(24*time.Hour),
		StypeIn: dbn.SType_RawSymbol, StypeOut: dbn.SType_InstrumentId,
	}
	for i, sym := range rawSymbols {
		metadata.Mappings = append(metadata.Mappings, dbn.SymbolMapping{RawSymbol: sym,
			Intervals: []dbn.MappingInterval{{StartDate: 20260101, EndDate: 20260102, Symbol: strconv.Itoa(i + 1)}},
		})
	}
	var buf bytes.Buffer
	if err := metadata.Write(&buf); err != nil {
		t.Fatal(err)
	}
	for pass := 0; pass < passes; pass++ {
		for i := range rawSymbols {
			record := dbn.NewOhlcvMsg(dbn.RType_Ohlcv1S)
			record.Header.InstrumentID = uint32(i + 1)
			record.Header.TsEvent = start + uint64(pass)
			if err := binary.Write(&buf, binary.LittleEndian, record); err != nil {
				t.Fatal(err)
			}
		}
	}
	source := filepath.Join(t.TempDir(), "input.dbn")
	if err := os.WriteFile(source, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return source
}

// regularFiles returns every regular file under root, relative to root.
func regularFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			rel, _ := filepath.Rel(root, path)
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestSplitFileKeepsOutputInsideDestDir(t *testing.T) {
	for _, tc := range []struct {
		name, dataset, symbol string
	}{
		{"dataset_dotdot", "../../ESC", "TEST"},
		{"dataset_is_dotdot", "..", "TEST"},
		{"symbol_dotdot", "TEST.DATA", "../../../ESC"},
		{"symbol_is_dotdot", "TEST.DATA", ".."},
		{"symbol_with_slash", "TEST.DATA", "BRK/B"},
		{"symbol_three_dots", "TEST.DATA", "..."},
		{"symbol_dotdot_space", "TEST.DATA", ".. "},
		{"symbol_trailing_dot", "TEST.DATA", "ABC."},
		{"dataset_three_dots", "...", "TEST"},
		{"dataset_dotdot_space", ".. ", "TEST"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := writeSplitSource(t, tc.dataset, []string{tc.symbol}, 1)
			root := t.TempDir()
			dest := filepath.Join(root, "a", "b", "c", "dest")
			if err := os.MkdirAll(dest, 0755); err != nil {
				t.Fatal(err)
			}
			if err := SplitFile(source, dest, false, false); err != nil {
				t.Fatalf("SplitFile: %v", err)
			}
			// Everything SplitFile created must be under dest: nothing else in root.
			err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				inDest := path == dest || strings.HasPrefix(path, dest+string(filepath.Separator))
				isAncestor := path == root || strings.HasPrefix(dest, path+string(filepath.Separator))
				if !inDest && !isAncestor {
					t.Errorf("SplitFile created %q outside dest %q", path, dest)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			// The record lands in one file, exactly five levels deep:
			// <dataset>/<symbol>/YYYY/MM/DD/<file>
			files := regularFiles(t, dest)
			if len(files) != 1 {
				t.Fatalf("got %d output files %v, want 1", len(files), files)
			}
			if depth := strings.Count(files[0], string(filepath.Separator)); depth != 5 {
				t.Errorf("output %q is %d levels deep, want 5", files[0], depth)
			}
			// Windows drops trailing dots and spaces from each name, so a
			// directory named "..." or ".. " would resolve to a parent there.
			for _, part := range strings.Split(filepath.Dir(files[0]), string(filepath.Separator)) {
				if strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
					t.Errorf("output directory name %q in %q ends with a dot or space", part, files[0])
				}
			}
		})
	}
}
