// Copyright (c) 2025 Neomantra Corp

package file

import "strings"

// SafePathComponent encodes s so it can be used as a single file or directory
// name. Symbols, datasets and similar values come from API callers, DBN files
// and the network, so they must never be joined into a path as-is.
//
// Path separators, control characters, '%' and the characters Windows
// reserves in file names are replaced by %XX (uppercase hex). The names "."
// and ".." are encoded entirely. Everything else, including spaces and
// non-ASCII text, is left alone, so ordinary names map to themselves and the
// encoding is reversible.
func SafePathComponent(s string) string {
	if s == "." || s == ".." {
		return strings.Repeat("%2E", len(s))
	}
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c == 0x7f || strings.IndexByte(`/\%:*?"<>|`, c) >= 0 {
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0f])
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}
