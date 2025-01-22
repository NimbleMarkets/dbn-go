// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"fmt"

	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Datasets page
type DatasetsPageModel struct {
	config Config

	datasets        []string
	datasetStrlen   int
	schemas         []string
	selectedDataset int
	selectedSchema  int
	lastError       error

	width        int
	height       int
	datasetTable table.Model
	schemasTable table.Model
}

func NewDatasetsPage(config Config) DatasetsPageModel {
	datasetTable := table.New(table.WithColumns([]table.Column{
		{Title: "Datasets", Width: 16},
	}), table.WithStyles(nimbleTableStyles),
		table.WithFocused(true))

	schemasTableStyle := nimbleTableStyles
	schemasTableStyle.Selected = lipgloss.NewStyle()
	schemasTable := table.New(table.WithColumns([]table.Column{
		{Title: "Schemas", Width: 16},
		{Title: "Live $/GB", Width: 16},
		{Title: "Hist $/GB", Width: 16},
	}), table.WithStyles(schemasTableStyle),
		table.WithFocused(false))

	m := DatasetsPageModel{
		config:          config,
		datasets:        nil,
		datasetStrlen:   0,
		schemas:         nil,
		selectedDataset: -1,
		selectedSchema:  -1,
		lastError:       nil,
		datasetTable:    datasetTable,
		schemasTable:    schemasTable,
		width:           20,
		height:          10,
	}
	m.updateSizes()
	return m
}

//////////////////////////////////////////////////////////////////////////////
// BubbleTea interface

// Init handles the initialization of an DatasetsPageModel
func (m DatasetsPageModel) Init() tea.Cmd {
	if len(m.datasets) == 0 {
		return getDatasets(m.config.DatabentoApiKey)
	}
	return nil
}

// Update handles BubbleTea messages for the DatasetsPageModel
func (m DatasetsPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()

	case DatasetsMsg:
		m.lastError = msg.Error
		m.datasets = msg.Datasets

		var rows []table.Row
		var datasetStrlen int
		for _, dataset := range m.datasets {
			rows = append(rows, table.Row{dataset})
			datasetStrlen = maxInt(datasetStrlen, len(dataset))
		}
		m.datasetTable.SetRows(rows)
		m.datasetStrlen = datasetStrlen

		m.schemas = nil
		m.schemasTable.SetRows(nil)
		m.updateSizes()
		cmd := m.onDatasetSelection()
		return m, cmd

	case SchemasMsg:
		m.lastError = msg.Error
		m.schemas = msg.Schemas
		var rows []table.Row
		for _, schema := range m.schemas {
			rows = append(rows, table.Row{
				schema,
				fmt.Sprintf("%8.2f", msg.LivePrices[schema]),
				fmt.Sprintf("%8.2f", msg.HistPrices[schema])})
		}
		m.schemasTable.SetRows(rows)
		m.updateSizes()

	default:
		var cmd1, cmd2 tea.Cmd
		m.datasetTable, cmd1 = m.datasetTable.Update(msg)
		m.schemasTable, cmd2 = m.schemasTable.Update(msg)
		m.updateSizes()
		cmd3 := m.onDatasetSelection()
		return m, tea.Batch(cmd1, cmd2, cmd3)
	}
	return m, nil
}

func (m *DatasetsPageModel) onDatasetSelection() tea.Cmd {
	cursor := m.datasetTable.Cursor()
	if cursor < 0 || cursor >= len(m.datasets) || cursor == m.selectedDataset {
		return nil
	}
	m.selectedDataset = cursor

	dataset := m.datasets[m.selectedDataset]
	cmd := getSchemas(m.config.DatabentoApiKey, dataset)
	return cmd
}

// View renders the DatasetsPageModel's view.
func (m DatasetsPageModel) View() string {
	if m.lastError != nil {
		return fmt.Sprintf("Error: %s", m.lastError.Error())
	}

	datasetPane := nimbleBorderStyle.Render(m.datasetTable.View())
	schemaPane := nimbleBorderStyle.Render(m.schemasTable.View())

	return lipgloss.JoinHorizontal(lipgloss.Top,
		datasetPane,
		schemaPane,
	)
}

//////////////////////////////////////////////////////////////////////////////

// updateSizes update the heights/widths of widgets
func (m *DatasetsPageModel) updateSizes() {
	availHeight := m.height - 2 - 2 // 2xAppHeaderFooter 2xPaneBorder
	m.datasetTable.SetHeight(availHeight)
	m.schemasTable.SetHeight(availHeight)

	availWidth := m.width - 2 // 2xPaneBorder
	datasetWidth := maxInt(0, minInt(availWidth, m.datasetStrlen+3))
	m.datasetTable.SetWidth(datasetWidth)
	const datasetColumnIndex = 0
	m.datasetTable.Columns()[datasetColumnIndex].Width = datasetWidth - 1

	availWidth -= m.datasetTable.Width() + 3 // 2xTableBorder
	m.schemasTable.SetWidth(maxInt(0, availWidth))
}

//////////////////////////////////////////////////////////////////////////////

type DatasetsMsg struct {
	Datasets []string
	Error    error
}

type SchemasMsg struct {
	Dataset    string
	Schemas    []string
	HistPrices map[string]float64
	LivePrices map[string]float64
	Error      error
}

func getDatasets(databentoApiKey string) tea.Cmd {
	return func() tea.Msg {
		datasets, err := dbn_hist.ListDatasets(databentoApiKey, dbn_hist.DateRange{})
		return DatasetsMsg{Datasets: datasets, Error: err}
	}
}

func getSchemas(databentoApiKey string, dataset string) tea.Cmd {
	return func() tea.Msg {
		msg := SchemasMsg{
			Dataset: dataset,
		}

		unitPricesVec, err := dbn_hist.ListUnitPrices(databentoApiKey, dataset)
		if err != nil {
			msg.Error = err
			return msg
		}

		schemasMap := make(map[string]bool)
		msg.HistPrices = make(map[string]float64)
		msg.LivePrices = make(map[string]float64)
		for _, unitPricesForMode := range unitPricesVec {
			var destMap *map[string]float64
			switch unitPricesForMode.Mode {
			case dbn_hist.FeedMode_Historical:
				destMap = &msg.HistPrices
			case dbn_hist.FeedMode_Live:
				destMap = &msg.LivePrices
			default:
				continue
			}
			for schema, cost := range unitPricesForMode.UnitPrices {
				schemasMap[schema] = true
				(*destMap)[schema] = cost
			}
		}
		for schema := range schemasMap {
			msg.Schemas = append(msg.Schemas, schema)
		}
		return msg
	}
}
