// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"fmt"
	"strconv"

	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Publishers page
type PublishersPageModel struct {
	config     Config
	publishers []dbn_hist.PublisherDetail
	lastError  error

	table  table.Model
	width  int
	height int
}

func NewPublishersPage(config Config) PublishersPageModel {
	table := table.New(table.WithColumns([]table.Column{
		{Title: "ID", Width: 4},
		{Title: "Dataset", Width: 16},
		{Title: "Venue", Width: 8},
		{Title: "Description", Width: 80},
	}), table.WithStyles(nimbleTableStyles),
		table.WithFocused(true))

	return PublishersPageModel{
		config:     config,
		publishers: nil,
		lastError:  nil,
		table:      table,
		width:      20,
		height:     10,
	}
}

//////////////////////////////////////////////////////////////////////////////
// BubbleTea interface

// Init handles the initialization of an PublishersPageModel
func (m PublishersPageModel) Init() tea.Cmd {
	if len(m.publishers) == 0 {
		return getPublishers(m.config.DatabentoApiKey)
	}
	return nil
}

// Update handles BubbleTea messages for the Publishers Page
func (m PublishersPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width - 2)
		m.table.SetHeight(msg.Height - 4)

	case PublishersMsg:
		m.lastError = msg.Error
		m.publishers = msg.Publishers

		var rows []table.Row
		for _, pub := range m.publishers {
			rows = append(rows, table.Row{
				strconv.Itoa(int(pub.PublisherID)),
				pub.Dataset,
				pub.Venue,
				pub.Description,
			})
		}
		m.table.SetRows(rows)
	default:
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View renders the PublishersPageModel's view.
func (m PublishersPageModel) View() string {
	var pane string
	if m.lastError == nil {
		pane = m.table.View()
	} else {
		pane = lipgloss.NewStyle().Width(m.table.Width()).Render(
			fmt.Sprintf("Error: %s", m.lastError.Error()))
	}

	return nimbleBorderStyle.Render(pane)
}

//////////////////////////////////////////////////////////////////////////////

type PublishersMsg struct {
	Publishers []dbn_hist.PublisherDetail
	Error      error
}

func getPublishers(databentoApiKey string) tea.Cmd {
	return func() tea.Msg {
		publishers, err := dbn_hist.ListPublishers(databentoApiKey)
		return PublishersMsg{Publishers: publishers, Error: err}
	}
}
