// Copyright (c) 2025 Neomantra Corp

package tui

import (
	stickersTable "github.com/76creates/stickers/table"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var (

	// Nimble Color Pallete
	rgbaNimbleDarkPurple  = "#3F3080"
	rgbaNimbleLightPurple = "#655BA7"
	rgbaNimbleRed         = "#E24F36"
	rgbaNimbleGrue        = "#4495AA"
	rgbaNimbleGreen       = "#AA7D7B1"
	rgbaNimbleYellow      = "#FBF4A5"
	rgbaNimbleWhite       = "#FFFFFF"
	rgbaNimbleBlack       = "#000000"

	colorDarkPurple  = lipgloss.Color(rgbaNimbleDarkPurple)
	colorLightPurple = lipgloss.Color(rgbaNimbleLightPurple)
	colorRed         = lipgloss.Color(rgbaNimbleRed)
	colorGrue        = lipgloss.Color(rgbaNimbleGrue)
	colorGreen       = lipgloss.Color(rgbaNimbleGreen)
	colorYellow      = lipgloss.Color(rgbaNimbleYellow)
	colorWhite       = lipgloss.Color(rgbaNimbleWhite)
	colorBlack       = lipgloss.Color(rgbaNimbleBlack)

	nimbleBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), true).
				BorderForeground(colorLightPurple)

	nimbleTableStyles = table.Styles{
		Header:   lipgloss.NewStyle().Bold(true).Foreground(colorRed).Padding(0, 1),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(colorGrue),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	}

	nimbleStickerTableStyles = map[stickersTable.StyleKey]lipgloss.Style{
		stickersTable.StyleKeyHeader: lipgloss.NewStyle().
			Foreground(colorRed),
		stickersTable.StyleKeyFooter: lipgloss.NewStyle().
			Align(lipgloss.Right).
			Height(1).
			Foreground(colorRed).
			Background(colorLightPurple),
		stickersTable.StyleKeyRows: lipgloss.NewStyle().
			Foreground(colorWhite),
		stickersTable.StyleKeyRowsSubsequent: lipgloss.NewStyle().
			Foreground(colorWhite),
		stickersTable.StyleKeyRowsCursor: lipgloss.NewStyle().
			Foreground(colorGrue),
		stickersTable.StyleKeyCellCursor: lipgloss.NewStyle().
			Foreground(colorGrue),
	}
)
