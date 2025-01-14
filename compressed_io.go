// Copyright (c) 2025 Neomantra Corp
// Reader/Writer Compression helpers
//
// Adapted from Neomantra's Gist, but simplified to only support zstd.:
//
// https://gist.github.com/neomantra/691a6028cdf2ac3fc6ec97d00e8ea802
//

package dbn

import (
	"io"
	"os"
	"strings"

	"github.com/klauspost/compress/zstd"
)

///////////////////////////////////////////////////////////////////////////////

// Returns an io.Writer for the given filename, or os.Stdout if filename is "-".  Also returns a closing function to defer and any error.
// If the filename ends in ".zst" or ".zstd", or if useZstd is true, the writer will zstd-compress the output.
func MakeCompressedWriter(filename string, useZstd bool) (io.Writer, func(), error) {
	var writer io.Writer
	var closer io.Closer
	fileCloser := func() {
		if closer != nil {
			closer.Close()
		}
	}
	if filename != "-" {
		if file, err := os.Create(filename); err == nil {
			writer, closer = file, file
		} else {
			return nil, nil, err
		}
	} else {
		writer, closer = os.Stdout, nil
	}

	if useZstd || strings.HasSuffix(filename, ".zst") || strings.HasSuffix(filename, ".zstd") {
		zstdWriter, err := zstd.NewWriter(writer)
		if err != nil {
			fileCloser()
			return nil, nil, err
		}
		zstdCloser := func() {
			zstdWriter.Close()
			fileCloser()
		}
		return zstdWriter, zstdCloser, nil
	} else {
		return writer, fileCloser, nil
	}
}

///////////////////////////////////////////////////////////////////////////////

// Returns a io.Reader for the given filename, or os.Stdout if filename is "-". Also returns a closing function to defer.
// If the filename ends in ".zst" or ".zstd", or if useZstd is true, the reader will zstd-decompress the input.
func MakeCompressedReader(filename string, useZstd bool) (io.Reader, io.Closer, error) {
	var reader io.Reader
	var closer io.Closer

	if filename != "-" {
		if file, err := os.Open(filename); err == nil {
			reader, closer = file, file
		} else {
			return nil, nil, err
		}
	} else {
		reader, closer = os.Stdin, nil
	}

	var err error
	if useZstd || strings.HasSuffix(filename, ".zst") || strings.HasSuffix(filename, ".zstd") {
		reader, err = zstd.NewReader(reader)
	}

	if err != nil {
		// clean up file
		if closer != nil {
			closer.Close()
		}
		return nil, nil, err
	}
	return reader, closer, nil
}
