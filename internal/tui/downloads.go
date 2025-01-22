// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"fmt"
	"strings"

	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/dustin/go-humanize"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type QueueDownloadMsg struct {
	JobID string
	Files []dbn_hist.BatchFileDesc
}

///////////////////////////////////////////////////////////////////////////////

// Downloads page
type DownloadsPageModel struct {
	config Config

	downloadManager *DownloadManager

	width  int
	height int

	downloadsTable table.Model

	progressBar progress.Model // We use this in the table via ViewAs

	lastError error
	help      help.Model
	keyMap    DownloadsPageKeyMap
}

const (
	columnDownloadsJobKey      = "Job"
	columnDownloadsFileKey     = "File"
	columnDownloadsSizeKey     = "Size"
	columnDownloadsStateKey    = "State"
	columnDownloadsProgressKey = "Progress"

	columnDownloadsJobIndex      = 0
	columnDownloadsFileIndex     = 1
	columnDownloadsSizeIndex     = 2
	columnDownloadsStateIndex    = 3
	columnDownloadsProgressIndex = 4

	columnDownloadsJobWidth      = 24
	columnDownloadsFileWidth     = 35
	columnDownloadsSizeWidth     = 12
	columnDownloadsStateWidth    = 10
	columnDownloadsProgressWidth = 20

	columnDownloadsJobMinWidth      = 10
	columnDownloadsFileMinWidth     = 10
	columnDownloadsSizeMinWidth     = 12
	columnDownloadsStateMinWidth    = 10
	columnDownloadsProgressMinWidth = 12

	countsReportFormat = "Finished %d of %d (%d active)"
	countsReportWidth  = 40 // "%d done of %d (%d active)"
)

func NewDownloadsPage(config Config) DownloadsPageModel {
	downloadsTable := table.New(table.WithColumns([]table.Column{
		{Title: "Job", Width: columnDownloadsJobWidth},
		{Title: "File", Width: columnDownloadsFileWidth},
		{Title: "Size", Width: columnDownloadsSizeWidth},
		{Title: "State", Width: columnDownloadsStateWidth},
		{Title: "Progress", Width: columnDownloadsProgressWidth},
	}), table.WithStyles(nimbleTableStyles),
		table.WithFocused(true))

	progressBar := progress.New(
		progress.WithGradient(rgbaNimbleLightPurple, rgbaNimbleGrue),
		progress.WithWidth(20))

	m := DownloadsPageModel{
		config:          config,
		downloadManager: NewDownloadManager(config.DatabentoApiKey, config.MaxActiveDownloads),
		width:           20,
		height:          10,
		downloadsTable:  downloadsTable,
		progressBar:     progressBar,
		help:            help.New(),
		keyMap:          DefaultDownloadsPageKeyMap(),
	}
	return m
}

///////////////////////////////////////////////////////////////////////////////
// DownloadsPageKeyMap

// DownloadsPageKeyMap is the all the [key.Binding] for the DownloadsPageModel
type DownloadsPageKeyMap struct {
	CursorUp   key.Binding
	CursorDown key.Binding
	// Cancel key.Binding
}

// DefaultDownloadsPageKeyMap returns a default set of key bindings for DownloadsPageModel
func DefaultDownloadsPageKeyMap() DownloadsPageKeyMap {
	return DownloadsPageKeyMap{
		CursorUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "cursor up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "cursor down"),
		),
		// Cancel: key.NewBinding(
		// 	key.WithKeys("x"),
		// 	key.WithHelp("x", "cancel"),
		// ),
	}
}

// FullHelp returns bindings to show the full help view.
// Implements bubble's [help.KeyMap] interface.
func (m *DownloadsPageKeyMap) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{{
		m.CursorUp,
		m.CursorDown,
	}}
	return kb
}

// ShortHelp returns bindings to show in the abbreviated help view. It's part
// of the help.KeyMap interface.
func (m DownloadsPageKeyMap) ShortHelp() []key.Binding {
	kb := []key.Binding{
		m.CursorUp,
		m.CursorDown,
	}
	return kb
}

//////////////////////////////////////////////////////////////////////////////
// BubbleTea interface

// Init handles the initialization of an DownloadsPageModel
func (m DownloadsPageModel) Init() tea.Cmd {
	return m.listenForProgress() // start the progress listener
}

// Update handles BubbleTea messages for the DownloadsPageModel
func (m DownloadsPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()
		return m, nil

	case tea.KeyMsg:
		// pass KeyMsg on to the focused table
		var cmd tea.Cmd
		m.downloadsTable, cmd = m.downloadsTable.Update(msg)
		return m, cmd

	case QueueDownloadMsg:
		for _, file := range msg.Files {
			m.downloadManager.QueueDownload(msg.JobID, file)
		}
		return m, nil

	case DownloadProgressMsg:
		cmd := m.onDownloadProgress(msg)
		return m, cmd
	}
	return m, nil
}

// View renders the DownloadsPageModel's view.
func (m DownloadsPageModel) View() string {
	viewStr := nimbleBorderStyle.Render(m.downloadsTable.View()) + "\n"
	if m.lastError != nil {
		viewStr += fmt.Sprintf("Error: %s ", m.lastError)
	}

	helpStr := m.help.View(&m.keyMap)
	countsWidth := m.width - lipgloss.Width(helpStr)
	queued, active, past := m.downloadManager.Counts()
	countsStr := lipgloss.NewStyle().Width(countsWidth).Align(lipgloss.Right).
		Render(fmt.Sprintf(countsReportFormat, past, (past + queued + active), active))

	viewStr += lipgloss.JoinHorizontal(lipgloss.Top, helpStr, countsStr)
	return viewStr
}

///////////////////////////////////////////////////////////////////////////////

// updateSizes update the heights/widths of widgets
func (m *DownloadsPageModel) updateSizes() {
	availHeight := m.height - 2 - 2 - 2 // 2xAppHeaderFooter 2xPaneBorder
	m.downloadsTable.SetHeight(availHeight)

	availWidth := m.width - 2 // 2xPaneBorder
	m.downloadsTable.SetWidth(availWidth)

	if m.lastError != nil {
		availWidth -= lipgloss.Width(fmt.Sprintf("Error: %s ", m.lastError))
	}
	m.help.Width = availWidth
}

// renderProgressBar renders a string progress bar with a given [0, 1] times width
func renderProgressBar(width int, progress float64) string {
	progress = clampFloat(progress, 0.0, 1.0)
	fullLen := maxInt(0, width-6) // 5 = len(" 100% ")
	barLen := int(float64(fullLen) * progress)
	return fmt.Sprintf("%s%s % 3.0f%%",
		strings.Repeat("█", barLen),
		strings.Repeat(".", fullLen-barLen),
		progress*100)
}

///////////////////////////////////////////////////////////////////////////////

// listenForProgress is a command that waits for the responses on the channel
func (m *DownloadsPageModel) listenForProgress() tea.Cmd {
	return func() tea.Msg {
		return DownloadProgressMsg(<-m.downloadManager.ProgressChannel())
	}
}

// onDownloadProgress handles a DownloadProgressMsg, updating the tables for the DownloadItems
func (m *DownloadsPageModel) onDownloadProgress(msg DownloadProgressMsg) tea.Cmd {
	// Render progress or report an error
	var progressStr string
	if msg.Error != nil {
		progressStr = fmt.Sprintf("%s", msg.Error.Error())
		m.lastError = msg.Error
	} else {
		progress := clampFloat(float64(msg.CurrentSize)/float64(msg.Desc.Size), 0, 1)
		progressStr = renderProgressBar(columnDownloadsProgressWidth, progress)
	}

	// Update the download's row
	found := false
	downloadsRows := m.downloadsTable.Rows()
	for i := 0; i < len(downloadsRows); i++ {
		row := downloadsRows[i]
		if row[columnDownloadsJobIndex] == msg.Desc.JobID && row[columnDownloadsFileIndex] == msg.Desc.Filename {
			row[columnDownloadsStateIndex] = lipgloss.NewStyle().Width(columnDownloadsStateWidth).Align(lipgloss.Center).
				Render(string(msg.State))
			row[columnDownloadsProgressIndex] = progressStr
			downloadsRows[i] = row
			m.downloadsTable.UpdateViewport()
			found = true
			break
		}
	}
	if !found {
		m.downloadsTable.SetRows(append(m.downloadsTable.Rows(), table.Row{
			msg.Desc.JobID,
			msg.Desc.Filename,
			lipgloss.NewStyle().Width(columnDownloadsSizeWidth).Align(lipgloss.Right).
				Render(humanize.Comma(int64(msg.Desc.Size))),
			lipgloss.NewStyle().Width(columnDownloadsStateWidth).Align(lipgloss.Center).
				Render(string(msg.State)),
			// lipgloss.NewStyle().Width(columnDownloadsStateWidth).Render(
			progressStr,
		}))
	}

	return m.listenForProgress()
}
