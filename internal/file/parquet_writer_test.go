// Copyright (c) 2025 Neomantra Corp

package file

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/NimbleMarkets/dbn-go"
	pqfile "github.com/apache/arrow-go/v18/parquet/file"
)

func TestWriteDbnFileAsParquet_ValidInput(t *testing.T) {
	src := filepath.Join("..", "..", "tests", "data", "test_data.trades.dbn")
	dst := filepath.Join(t.TempDir(), "out.parquet")

	if err := WriteDbnFileAsParquet(src, false, dst); err != nil {
		t.Fatalf("WriteDbnFileAsParquet(valid) returned error: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("expected output parquet file to exist: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("expected output parquet file to be non-empty")
	}
}

func TestWriteDbnFileAsParquet_TruncatedInputReturnsError(t *testing.T) {
	orig := filepath.Join("..", "..", "tests", "data", "test_data.trades.dbn")
	data, err := os.ReadFile(orig)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	if len(data) < 2 {
		t.Fatalf("fixture too small to truncate")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "truncated.dbn")
	dst := filepath.Join(dir, "out.parquet")
	if err := os.WriteFile(src, data[:len(data)-1], 0644); err != nil {
		t.Fatalf("failed to write truncated fixture: %v", err)
	}

	if err := WriteDbnFileAsParquet(src, false, dst); err == nil {
		t.Fatalf("expected error for truncated input, got nil")
	}
}

func TestWriteDbnFileAsParquet_WritesAllRowsForAffectedSchemas(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{
			name: "mbp1",
			src:  filepath.Join("..", "..", "tests", "data", "test_data.mbp-1.v2.dbn.zst"),
		},
		{
			name: "tbbo",
			src:  filepath.Join("..", "..", "tests", "data", "test_data.tbbo.v2.dbn.zst"),
		},
		{
			name: "imbalance",
			src:  filepath.Join("..", "..", "tests", "data", "test_data.imbalance.v2.dbn.zst"),
		},
		{
			name: "statistics",
			src:  filepath.Join("..", "..", "tests", "data", "test_data.statistics.v2.dbn.zst"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantRows := countDBNRecords(t, tt.src, false)
			dst := filepath.Join(t.TempDir(), "out.parquet")

			if err := WriteDbnFileAsParquet(tt.src, false, dst); err != nil {
				t.Fatalf("WriteDbnFileAsParquet(%s) returned error: %v", tt.name, err)
			}

			gotRows := countParquetRows(t, dst)
			if gotRows != wantRows {
				t.Fatalf("parquet row count mismatch for %s: got %d want %d", tt.name, gotRows, wantRows)
			}
		})
	}
}

func countDBNRecords(t *testing.T, filename string, forceZstd bool) int64 {
	t.Helper()

	reader, closer, err := dbn.MakeCompressedReader(filename, forceZstd)
	if err != nil {
		t.Fatalf("MakeCompressedReader(%q): %v", filename, err)
	}
	defer closer.Close()

	scanner := dbn.NewDbnScanner(reader)
	if _, err := scanner.Metadata(); err != nil {
		t.Fatalf("Metadata(%q): %v", filename, err)
	}

	var rows int64
	for scanner.Next() {
		rows++
	}

	if err := scanner.Error(); err != nil && err != io.EOF {
		t.Fatalf("scanner error for %q: %v", filename, err)
	}

	return rows
}

func countParquetRows(t *testing.T, filename string) int64 {
	t.Helper()

	reader, err := pqfile.OpenParquetFile(filename, false)
	if err != nil {
		t.Fatalf("OpenParquetFile(%q): %v", filename, err)
	}
	defer reader.Close()

	return reader.NumRows()
}
