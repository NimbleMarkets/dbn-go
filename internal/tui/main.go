// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Config struct {
	DatabentoApiKey    string
	MaxActiveDownloads int // note that default would be 0, which is no downloads
}

func Run(config Config) error {
	model := NewAppModel(config)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

//////////////////////////////////////////////////////////////////////////////

type AppModel struct {
	config Config

	datasets     []string
	datasetIndex int

	pages       []tea.Model
	pageNames   []string
	currentPage int

	width            int
	height           int
	help             help.Model
	keyMap           AppKeyMap
	headerStyle      lipgloss.Style
	footerStyle      lipgloss.Style
	inactiveTabStyle lipgloss.Style
	activeTabStyle   lipgloss.Style
}

func NewAppModel(config Config) AppModel {
	m := AppModel{
		config:       config,
		datasets:     nil,
		datasetIndex: -1,
		currentPage:  0,
		pageNames:    []string{"1-Jobs", "2-Downloads", "3-Datasets", "4-Publishers"},
		pages: []tea.Model{
			NewJobsPage(config),
			NewDownloadsPage(config),
			NewDatasetsPage(config),
			NewPublishersPage(config),
		},
		width:  20,
		height: 10,
		help:   help.New(),
		keyMap: DefaultAppKeyMap(),
		headerStyle: lipgloss.NewStyle().
			Foreground(colorYellow).
			Background(colorDarkPurple),
		footerStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorYellow)).
			Background(lipgloss.Color(colorDarkPurple)),
		inactiveTabStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorYellow)).
			Background(lipgloss.Color(colorDarkPurple)),
		activeTabStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorYellow)).
			Background(lipgloss.Color(colorGrue)),
	}
	return m
}

///////////////////////////////////////////////////////////////////////////////
// AppKeyMap

// AppKeyMap is the all the [key.Binding] for the AppModel
type AppKeyMap struct {
	Quit            key.Binding
	FocusJobs       key.Binding
	FocusDownloads  key.Binding
	FocusDatasets   key.Binding
	FocusPublishers key.Binding
}

// DefaultAppKeyMap returns a default set of key bindings for AppModel
func DefaultAppKeyMap() AppKeyMap {
	return AppKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("esc", "quit"),
		),
		FocusJobs: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "jobs"),
		),
		FocusDownloads: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "downloads"),
		),
		FocusDatasets: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "datasets"),
		),
		FocusPublishers: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "publishers"),
		),
	}
}

// FullHelp returns bindings to show the full help view.
// Implements bubble's [help.KeyMap] interface.
func (m *AppKeyMap) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{{
		m.Quit,
		m.FocusJobs,
		m.FocusDownloads,
		m.FocusDatasets,
		m.FocusPublishers,
	}}
	return kb
}

// ShortHelp returns bindings to show in the abbreviated help view. It's part
// of the help.KeyMap interface.
func (m AppKeyMap) ShortHelp() []key.Binding {
	kb := []key.Binding{
		m.Quit,
		m.FocusJobs,
		m.FocusDownloads,
		m.FocusDatasets,
		m.FocusPublishers,
	}
	return kb
}

//////////////////////////////////////////////////////////////////////////////
// BubbleTea interface

// Init handles the initialization of an Session
func (m AppModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, page := range m.pages {
		cmds = append(cmds, page.Init())
	}
	return tea.Batch(cmds...)
}

// Update handles BubbleTea messages for the Session
// This is for starting/stopping/updating generation.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.FocusJobs):
			m.currentPage = 0
		case key.Matches(msg, m.keyMap.FocusDownloads):
			m.currentPage = 1
		case key.Matches(msg, m.keyMap.FocusDatasets):
			m.currentPage = 2
		case key.Matches(msg, m.keyMap.FocusPublishers):
			m.currentPage = 3
		}

		// only active page gets key evens
		pageModel, cmd := m.pages[m.currentPage].Update(msg)
		m.pages[m.currentPage] = pageModel
		return m, cmd
	}

	// propogate message to all pages
	var cmds []tea.Cmd
	for i := 0; i < len(m.pages); i++ {
		pageModel, cmd := m.pages[i].Update(msg)
		m.pages[i] = pageModel
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// View renders the ModelChooser's view.
func (m AppModel) View() string {
	viewStr := m.headerView() + "\n"
	if m.currentPage < 0 || m.currentPage >= len(m.pages) {
		viewStr += "Error: bad page\n"
	} else {
		viewStr += m.pages[m.currentPage].View() + "\n"
	}
	viewStr += m.footerView()
	return viewStr
}

///////////////////////////////////////////////////////////////////////////////

func (m *AppModel) headerView() string {
	header := m.headerStyle.Render(" dbn-go-tui   ")
	for i, name := range m.pageNames {
		if i == m.currentPage {
			name = "[ " + name + " ]"
			header += m.activeTabStyle.Render(name)
		} else {
			name = "| " + name + " |"
			header += m.inactiveTabStyle.Render(name)
		}
		header += m.headerStyle.Render(" ")
	}

	const bigHeart = "\u2764"
	headerSuffix := m.headerStyle.Render(bigHeart + "nm ")
	restOfLine := maxInt(0, m.width-lipgloss.Width(header)-lipgloss.Width(headerSuffix))
	header += m.headerStyle.Render(strings.Repeat(" ", restOfLine))
	header += headerSuffix
	return header
}

func (m *AppModel) footerView() string {
	return m.help.View(&m.keyMap)
}
