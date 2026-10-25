// Copyright (c) 2025 Neomantra Corp

//lint:file-ignore U1000 keep full Nimble color palette

package tui

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
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
)
