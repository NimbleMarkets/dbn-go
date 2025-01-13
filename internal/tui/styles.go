// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Nimble Color Pallete
	colorDarkPurple  = lipgloss.Color("#3F3080")
	colorLightPurple = lipgloss.Color("#655BA7")
	colorRed         = lipgloss.Color("#E24F36")
	colorGrue        = lipgloss.Color("#4495AA")
	colorGreen       = lipgloss.Color("#AA7D7B1")
	colorYellow      = lipgloss.Color("#FBF4A5")
	colorWhite       = lipgloss.Color("#FFFFFF")
	colorBlack       = lipgloss.Color("#000000")

	nimbleBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), true).
				BorderForeground(colorLightPurple)

	nimbleTableStyles = table.Styles{
		Header:   lipgloss.NewStyle().Bold(true).Foreground(colorRed).Padding(0, 1),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(colorGrue),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	}
)
