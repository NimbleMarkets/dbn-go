// Copyright (c) 2025 Neomantra Corp

package file

import "testing"

func TestSafePathComponent(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		// Ordinary names are unchanged, so existing output paths stay stable.
		{"AAPL", "AAPL"},
		{"ES.c.0", "ES.c.0"},
		{"XNAS.ITCH", "XNAS.ITCH"},
		{"ESH4-ESM4", "ESH4-ESM4"},
		{"AAPL  240119C00150000", "AAPL  240119C00150000"},
		{"AAPL,MSFT+2more", "AAPL,MSFT+2more"},
		{"", ""},
		// Separators, traversal names and reserved characters are encoded.
		{"BRK/B", "BRK%2FB"},
		{`BRK\B`, "BRK%5CB"},
		{"..", "%2E%2E"},
		{".", "%2E"},
		{"../../x", "..%2F..%2Fx"},
		{"a%b", "a%25b"},
		{"a\x00b", "a%00b"},
		{"a\nb", "a%0Ab"},
		{"UD:1V", "UD%3A1V"},
		{`a*?"<>|b`, "a%2A%3F%22%3C%3E%7Cb"},
	} {
		if got := SafePathComponent(tc.in); got != tc.want {
			t.Errorf("SafePathComponent(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
