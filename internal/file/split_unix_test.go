// Copyright (c) 2025 Neomantra Corp

//go:build !windows

package file

import (
	"fmt"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/NimbleMarkets/dbn-go"
)

// A source with more instruments than the process may hold open files must
// still split, and instruments revisited after their writer was closed must
// keep all their records.
func TestSplitFileBoundsOpenFiles(t *testing.T) {
	const fdLimit, numInstruments, passes = 320, 400, 2

	var saved syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &saved); err != nil {
		t.Skipf("getrlimit: %v", err)
	}
	limited := saved
	limited.Cur = fdLimit
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limited); err != nil {
		t.Skipf("setrlimit: %v", err)
	}
	t.Cleanup(func() { syscall.Setrlimit(syscall.RLIMIT_NOFILE, &saved) })

	symbols := make([]string, numInstruments)
	for i := range symbols {
		symbols[i] = fmt.Sprintf("SYM%03d", i)
	}
	source := writeSplitSource(t, "TEST.DATA", symbols, passes)
	dest := t.TempDir()
	if err := SplitFile(source, dest, false, false); err != nil {
		t.Fatalf("SplitFile: %v", err)
	}

	files := regularFiles(t, dest)
	if len(files) != numInstruments {
		t.Fatalf("got %d output files, want %d", len(files), numInstruments)
	}
	for _, rel := range files {
		reader, closer, err := dbn.MakeCompressedReader(filepath.Join(dest, rel), false)
		if err != nil {
			t.Fatal(err)
		}
		scanner := dbn.NewDbnScanner(reader)
		count := 0
		for scanner.Next() {
			count++
		}
		closer.Close()
		if count != passes {
			t.Fatalf("%s has %d records, want %d (scanner error: %v)", rel, count, passes, scanner.Error())
		}
	}
}
