// Copyright (c) 2025 Neomantra Corp

package file

import (
	"os"
	"path/filepath"
	"testing"
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
