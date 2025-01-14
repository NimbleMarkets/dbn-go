// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"fmt"
	"slices"
	"time"

	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

const (
	focusCount  = 3
	focusJobs   = 0
	focusFiles  = 1
	focusDetail = 2

	columnJobsReceivedWidth  = 19
	columnJobsDatasetWidth   = 10
	columnJobsStartDateWidth = 10
	columnJobsEndDateWidth   = 10
	columnJobsEncodingWidth  = 8
	columnJobsDeliveryWidth  = 10
	columnJobsSizeWidth      = 8

	columnFilesFilenameIndex = 0
	columnFilesSizeIndex     = 1
	columnFilesHashIndex     = 2
	columnFilesUrlsIndex     = 3
	columnFilesFilenameWidth = 28
	columnFilesSizeWidth     = 8
	columnFilesHashWidth     = 18
	columnFilesUrlsWidth     = 60

	columnDetailsKeyIndex   = 0
	columnDetailsValueIndex = 1
	columnDetailsKeyWidth   = 16
	columnDetailsValueWidth = 32
	tableDetailsWidth       = 1 + columnDetailsKeyWidth + 2 + columnDetailsValueWidth + 1 // border + key + padding + value + padding
)

// Jobs page
type JobsPageModel struct {
	config Config

	jobs          []dbn_hist.BatchJob
	lastJobsError error

	files         []dbn_hist.BatchFileDesc
	lastFileError error

	width  int
	height int

	showExpired bool
	showDetails bool

	selectedJob int
	focusIndex  int
	jobsTable   table.Model
	detailTable table.Model
	filesTable  table.Model
	statusLine  string

	help   help.Model
	keyMap JobsPageKeyMap
}

func NewJobsPage(config Config) JobsPageModel {
	jobsTable := table.New(table.WithColumns([]table.Column{
		{Title: "Received", Width: columnJobsReceivedWidth},
		{Title: "Dataset", Width: columnJobsDatasetWidth},
		{Title: "Start Date", Width: columnJobsStartDateWidth},
		{Title: "End Date", Width: columnJobsEndDateWidth},
		{Title: "Encoding", Width: columnJobsEncodingWidth},
		{Title: "Delivery", Width: columnJobsDeliveryWidth},
		{Title: "Size", Width: columnJobsSizeWidth},
	}), table.WithStyles(nimbleTableStyles),
		table.WithFocused(true))

	detailTable := table.New(table.WithColumns([]table.Column{
		{Title: "Key", Width: columnDetailsKeyWidth},
		{Title: "Value", Width: columnDetailsValueWidth},
	}), table.WithStyles(nimbleTableStyles),
		table.WithFocused(false))

	filesTable := table.New(table.WithColumns([]table.Column{
		{Title: "Filename", Width: columnFilesFilenameWidth},
		{Title: "Size", Width: columnFilesSizeWidth},
		{Title: "Hash", Width: columnFilesHashWidth},
		{Title: "Urls", Width: columnFilesUrlsWidth},
	}), table.WithStyles(nimbleTableStyles),
		table.WithFocused(false))

	m := JobsPageModel{
		config:        config,
		jobs:          nil,
		lastJobsError: nil,
		files:         nil,
		lastFileError: nil,
		width:         20,
		height:        10,
		selectedJob:   -1,
		focusIndex:    0,
		showExpired:   false,
		showDetails:   true,
		jobsTable:     jobsTable,
		detailTable:   detailTable,
		filesTable:    filesTable,
		statusLine:    "",
		help:          help.New(),
		keyMap:        DefaultJobsPageKeyMap(),
	}
	m.updateHeights()
	m.updateWidths()
	return m
}

///////////////////////////////////////////////////////////////////////////////
// JobsPageKeyMap

// JobsPageKeyMap is the all the [key.Binding] for the JobsPageModel
type JobsPageKeyMap struct {
	NextFocus     key.Binding
	Download      key.Binding
	ToggleExpired key.Binding
	ToggleDetails key.Binding
}

// DefaultJobsPageKeyMap returns a default set of key bindings for JobsPageModel
func DefaultJobsPageKeyMap() JobsPageKeyMap {
	return JobsPageKeyMap{
		NextFocus: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "focus->"),
		),
		Download: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "download"),
		),
		ToggleExpired: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "expired"),
		),
		ToggleDetails: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "details"),
		),
	}
}

// FullHelp returns bindings to show the full help view.
// Implements bubble's [help.KeyMap] interface.
func (m *JobsPageKeyMap) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{{
		m.NextFocus,
		m.Download,
		m.ToggleExpired,
		m.ToggleDetails,
	}}
	return kb
}

// ShortHelp returns bindings to show in the abbreviated help view. It's part
// of the help.KeyMap interface.
func (m JobsPageKeyMap) ShortHelp() []key.Binding {
	kb := []key.Binding{
		m.NextFocus,
		m.Download,
		m.ToggleExpired,
		m.ToggleDetails,
	}
	return kb
}

//////////////////////////////////////////////////////////////////////////////
// BubbleTea interface

// Init handles the initialization of an JobsPageModel
func (m JobsPageModel) Init() tea.Cmd {
	if len(m.jobs) == 0 {
		return getJobs(m.config.DatabentoApiKey, m.showExpired)
	}
	return nil
}

// Update handles BubbleTea messages for the JobsPageModel
func (m JobsPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateHeights()
		m.updateWidths()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.NextFocus):
			m.focusIndex++
			m.updateFocus()
			return m, nil

		case key.Matches(msg, m.keyMap.Download):
			return m, m.onDownload()

		case key.Matches(msg, m.keyMap.ToggleExpired):
			m.showExpired = !m.showExpired
			return m, getJobs(m.config.DatabentoApiKey, m.showExpired)

		case key.Matches(msg, m.keyMap.ToggleDetails):
			m.showDetails = !m.showDetails
			m.updateWidths()
			m.updateFocus()
			return m, nil
		}

		// pass KeyMsg on to the focused table
		var cmds []tea.Cmd
		var cmd tea.Cmd
		switch m.focusIndex {
		case focusJobs:
			m.jobsTable, cmd = m.jobsTable.Update(msg)
			cmds = append(cmds, cmd)
			cmds = append(cmds, m.onJobSelection())
		case focusFiles:
			m.filesTable, cmd = m.filesTable.Update(msg)
			cmds = append(cmds, cmd)
		case focusDetail:
			m.detailTable, cmd = m.detailTable.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case JobsMsg:
		m.lastJobsError = msg.Error
		m.jobs = msg.Jobs
		slices.SortFunc(m.jobs, func(a dbn_hist.BatchJob, b dbn_hist.BatchJob) int {
			// sort by received time, descending
			if a.TsReceived.After(b.TsReceived) {
				return -1
			} else if a.TsReceived.Before(b.TsReceived) {
				return 1
			}
			return 0
		})

		var rows []table.Row
		for _, job := range m.jobs {
			rows = append(rows, table.Row{
				job.TsReceived.Format(time.DateTime),
				job.Dataset,
				job.Start.Format(time.DateOnly),
				job.End.Format(time.DateOnly),
				job.Encoding.String(),
				job.Delivery.String(),
				humanize.Bytes(job.ActualSize),
			})
		}
		m.jobsTable.SetRows(rows)
		m.jobsTable.SetCursor(0)
		cmd := m.onJobSelection()
		return m, cmd

	case FilesMsg:
		m.lastFileError = msg.Error
		m.files = msg.Files

		var rows []table.Row
		for _, file := range m.files {
			rows = append(rows, table.Row{
				file.Filename,
				humanize.Bytes(file.Size),
				file.Hash,
				file.Urls["https"],
			})
		}
		m.filesTable.SetRows(rows)
		m.filesTable.SetCursor(0)
		return m, nil
	}
	return m, nil
}

// View renders the JobsPageModel's view.
func (m JobsPageModel) View() string {
	var jobsPane, filesPane, detailPane string

	if m.lastJobsError == nil {
		jobsPane = m.jobsTable.View()
	} else {
		jobsPane = lipgloss.NewStyle().Width(m.jobsTable.Width()).Render(
			fmt.Sprintf(" %s", m.lastJobsError.Error()))
	}

	if m.showDetails {
		detailPane = m.detailTable.View()
	}

	if m.lastFileError == nil {
		filesPane = m.filesTable.View()
	} else {
		filesPane = lipgloss.NewStyle().Width(m.filesTable.Width()).Render(
			fmt.Sprintf(" %s", m.lastFileError.Error()))
	}

	switch m.focusIndex {
	case focusJobs:
		jobsPane = nimbleBorderStyle.BorderStyle(lipgloss.ThickBorder()).Render(jobsPane)
		filesPane = nimbleBorderStyle.Render(filesPane)
		detailPane = nimbleBorderStyle.Render(detailPane)
	case focusFiles:
		jobsPane = nimbleBorderStyle.Render(jobsPane)
		filesPane = nimbleBorderStyle.BorderStyle(lipgloss.ThickBorder()).Render(filesPane)
		detailPane = nimbleBorderStyle.Render(detailPane)
	case focusDetail:
		jobsPane = nimbleBorderStyle.Render(jobsPane)
		filesPane = nimbleBorderStyle.Render(filesPane)
		detailPane = nimbleBorderStyle.BorderStyle(lipgloss.ThickBorder()).Render(detailPane)
	}

	var viewStr string
	leftPane := lipgloss.JoinVertical(lipgloss.Left,
		jobsPane,
		filesPane)
	viewStr += leftPane
	if m.showDetails {
		viewStr = lipgloss.JoinHorizontal(lipgloss.Top,
			viewStr, detailPane)
	}

	helpView := m.help.View(&m.keyMap)
	viewStr += "\n" + helpView
	viewStr += lipgloss.NewStyle().Width(m.width - lipgloss.Width(helpView)).Align(lipgloss.Right).
		Render(" " + m.statusLine)
	return viewStr
}

///////////////////////////////////////////////////////////////////////////////

func (m *JobsPageModel) updateFocus() {
	// don't focus details if it is not shown!
	if m.focusIndex == focusDetail && !m.showDetails {
		m.focusIndex++
	}
	if m.focusIndex >= focusCount {
		m.focusIndex = m.focusIndex % focusCount
	}
	switch m.focusIndex {
	case focusJobs:
		m.jobsTable.Focus()
		m.filesTable.Blur()
		m.detailTable.Blur()
	case focusFiles:
		m.jobsTable.Blur()
		m.filesTable.Focus()
		m.detailTable.Blur()
	case focusDetail:
		m.jobsTable.Blur()
		m.filesTable.Blur()
		m.detailTable.Focus()
	}
}

// updateHeights update the heights of widgets
func (m *JobsPageModel) updateHeights() {
	availHeight := m.height - 2 // app header+status bars

	helpView := m.help.View(&m.keyMap)
	availHeight -= lipgloss.Height(helpView)

	m.detailTable.SetHeight(availHeight - 2)

	filesHeight := maxInt(1, availHeight/2)
	m.filesTable.SetHeight(filesHeight)
	availHeight -= filesHeight + 2

	jobsHeight := maxInt(1, availHeight-2)
	m.jobsTable.SetHeight(jobsHeight)
}

// updateWidths update the heights of widgets
func (m *JobsPageModel) updateWidths() {
	availbleWidth := m.width - 2 // -2 for details borders
	m.detailTable.SetWidth(tableDetailsWidth)
	if m.showDetails {
		availbleWidth = maxInt(1, availbleWidth-tableDetailsWidth) - 2 // -2 for details border
	}

	m.help.Width = m.width
	m.jobsTable.SetWidth(availbleWidth)

	m.filesTable.SetWidth(availbleWidth)
	const leftOfURLWidth = columnFilesFilenameWidth + columnFilesSizeWidth + columnFilesHashWidth + 8 // +9 for left borders/padding
	m.filesTable.Columns()[columnFilesUrlsIndex].Width = availbleWidth - leftOfURLWidth
	m.filesTable.UpdateViewport()
}

func (m *JobsPageModel) onJobSelection() tea.Cmd {
	cursor := m.jobsTable.Cursor()
	if cursor < 0 || cursor >= len(m.jobs) || cursor == m.selectedJob {
		return nil
	}
	m.selectedJob = cursor

	job := m.jobs[m.selectedJob]
	detailRows := jobDetailRows(job)
	m.detailTable.SetRows(detailRows)

	cmd := getFiles(m.config.DatabentoApiKey, job.Id)
	return cmd
}

func (m *JobsPageModel) onDownload() tea.Cmd {
	if m.selectedJob == -1 || len(m.files) == 0 {
		return nil // no selection or files
	}
	job := m.jobs[m.selectedJob]

	// depends on focus
	var files []dbn_hist.BatchFileDesc
	switch m.focusIndex {
	case focusDetail:
		fallthrough
	case focusJobs:
		// focus is on job, download all files
		// TODO: do you want to download all files?
		files = m.files

	case focusFiles:
		// focus is on file, download this file
		cursor := m.filesTable.Cursor()
		if cursor < 0 || cursor >= len(m.files) {
			return nil
		}
		files = append(files, m.files[cursor])
	}
	m.statusLine = fmt.Sprintf("Queued %d files...", len(files))
	return teaCmdize(QueueDownloadMsg{JobID: job.Id, Files: files})
}

////////////////////////////////////////////////////////////////////////////////

func jobDetailRows(j dbn_hist.BatchJob) []table.Row {
	var rows []table.Row
	rows = append(rows, table.Row{"ID", j.Id})
	if j.UserID != nil {
		rows = append(rows, table.Row{"User ID", *j.UserID})
	}
	if j.BillID != nil {
		rows = append(rows, table.Row{"Bill ID", *j.BillID})
	}
	if j.CostUSD != nil {
		rows = append(rows, table.Row{"Cost USD",
			fmt.Sprintf("%0.2f", *j.CostUSD)})
	}
	rows = append(rows,
		table.Row{"Dataset", j.Dataset},
		table.Row{"Symbols", j.Symbols},
		table.Row{"StypeIn", j.StypeIn.String()},
		table.Row{"StypeOut", j.StypeOut.String()},
		table.Row{"Schema", j.Schema.String()},
		table.Row{"Start", niceTime(j.Start)},
		table.Row{"End", niceTime(j.End)},
		table.Row{"Limit", niceInt(j.Limit)},
		table.Row{"Encoding", j.Encoding.String()},
		table.Row{"Compression", j.Compression.String()},
		table.Row{"Pretty Px", niceBool(j.PrettyPx)},
		table.Row{"Pretty Ts", niceBool(j.PrettyTs)},
		table.Row{"Map Symbols", niceBool(j.MapSymbols)},
		table.Row{"Split Symbols", niceBool(j.SplitSymbols)},
		table.Row{"Split Duration", j.SplitDuration},
		table.Row{"Split Size", niceInt(j.SplitSize)},
		table.Row{"Packaging", j.Packaging.String()},
		table.Row{"Delivery", j.Delivery.String()},
		table.Row{"Record Count", niceInt(j.RecordCount)},
		table.Row{"Billed Size", niceInt(j.BilledSize)},
		table.Row{"Actual Size", niceInt(j.ActualSize)},
		table.Row{"Package Size", niceInt(j.PackageSize)},
		table.Row{"State", j.State.String()},
		table.Row{"Ts Received", niceTime(j.TsReceived)},
		table.Row{"Ts Queued", niceTime(j.TsQueued)},
		table.Row{"Ts Process Start", niceTime(j.TsProcessStart)},
		table.Row{"Ts Process Done", niceTime(j.TsProcessDone)},
		table.Row{"Ts Expiration", niceTime(j.TsExpiration)},
	)
	return rows
}

////////////////////////////////////////////////////////////////////////////////

type JobsMsg struct {
	Jobs  []dbn_hist.BatchJob
	Error error
}

type FilesMsg struct {
	Files []dbn_hist.BatchFileDesc
	Error error
}

func getJobs(databentoApiKey string, showExpired bool) tea.Cmd {
	return func() tea.Msg {
		stateFilter := "received,queued,processing,done"
		if showExpired {
			stateFilter += ",expired"
		}
		since := time.Now().Add(time.Hour * 24 * 365 * -2)
		jobs, err := dbn_hist.ListJobs(databentoApiKey, stateFilter, since)
		return JobsMsg{Jobs: jobs, Error: err}
	}
}

func getFiles(databentoApiKey string, jobID string) tea.Cmd {
	return func() tea.Msg {
		files, err := dbn_hist.ListFiles(databentoApiKey, jobID)
		return FilesMsg{Files: files, Error: err}
	}
}
