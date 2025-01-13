// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

//////////////////////////////////////////////////////////////////////////////

func niceTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func niceBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func niceInt[I int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64](i I) string {
	return fmt.Sprintf("%d", i)
}

func maxInt[I int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64](a, b I) I {
	if a > b {
		return a
	}
	return b
}

func minInt[I int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64](a, b I) I {
	if a < b {
		return a
	}
	return b
}

func minFloat[F float32 | float64](a, b F) F {
	if a < b {
		return a
	}
	return b
}

// cmdize is a utility function to convert a given value into a `tea.Cmd`
func teaCmdize[T any](t T) tea.Cmd {
	return func() tea.Msg {
		return t
	}
}
