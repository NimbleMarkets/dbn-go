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

func clampInt[I int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64](v, low, high I) I {
	return minInt(maxInt(v, low), high)
}

//////////////////////////////////////////////////////////////////////////////

func minFloat[F float32 | float64](a, b F) F {
	if a < b {
		return a
	}
	return b
}

func maxFloat[F float32 | float64](a, b F) F {
	if a > b {
		return a
	}
	return b
}

func clampFloat[F float32 | float64](v, low, high F) F {
	return minFloat(maxFloat(v, low), high)
}

//////////////////////////////////////////////////////////////////////////////

// cmdize is a utility function to convert a given value into a `tea.Cmd`
func teaCmdize[T any](t T) tea.Cmd {
	return func() tea.Msg {
		return t
	}
}

//////////////////////////////////////////////////////////////////////////////

// TrySendChannel attempts to send a msg to a channel, returning true if successful.
func TrySendChannel[T any](msg T, ch chan T) bool {
	// non-blocking send
	select {
	case ch <- msg:
		return true
	default:
		return false
	}
}
