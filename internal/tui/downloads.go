// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/hashicorp/go-retryablehttp"
)

type DownloadItemState string

const (
	DownloadNone     DownloadItemState = ""
	DownloadQueued   DownloadItemState = "queued"
	DownloadActive   DownloadItemState = "active"
	DownloadComplete DownloadItemState = "complete"
	DownloadFailed   DownloadItemState = "failed"

	MaxActiveDownloads = 2
)

type DownloadItem struct {
	JobID         string
	Url           string
	Filename      string
	FileHash      string
	DestFile      string
	State         DownloadItemState
	ExpectedSize  uint64
	Progress      float64
	ContextCancel context.CancelFunc
}

///////////////////////////////////////////////////////////////////////////////

// Downloads page
type DownloadsPageModel struct {
	config Config

	downloads   []DownloadItem
	queuedCount int
	activeCount int

	progressCh chan DownloadProgressMsg

	width  int
	height int

	downloadsTable table.Model
	lastError      error
	help           help.Model
	keyMap         DownloadsPageKeyMap
}

const (
	downloadsJobColumn      = 0
	downloadsFileColumn     = 1
	downloadsSizeColumn     = 2
	downloadsStateColumn    = 3
	downloadsProgressColumn = 4

	downloadsSizeColumnSize  = 12
	downloadsStateColumnSize = 10
)

func NewDownloadsPage(config Config) DownloadsPageModel {
	downloadsTable := table.New(table.WithColumns([]table.Column{
		{Title: "Job", Width: 24},
		{Title: "File", Width: 35},
		{Title: "Size", Width: downloadsSizeColumnSize},
		{Title: "State", Width: downloadsStateColumnSize},
		{Title: "Progress", Width: 10},
	}), table.WithStyles(nimbleTableStyles),
		table.WithFocused(true))

	m := DownloadsPageModel{
		config:         config,
		downloads:      nil,
		queuedCount:    0,
		activeCount:    0,
		progressCh:     make(chan DownloadProgressMsg, 100),
		width:          20,
		height:         10,
		downloadsTable: downloadsTable,
		help:           help.New(),
		keyMap:         DefaultDownloadsPageKeyMap(),
	}
	return m
}

///////////////////////////////////////////////////////////////////////////////
// DownloadsPageKeyMap

// DownloadsPageKeyMap is the all the [key.Binding] for the DownloadsPageModel
type DownloadsPageKeyMap struct {
	// Cancel key.Binding
}

// DefaultDownloadsPageKeyMap returns a default set of key bindings for DownloadsPageModel
func DefaultDownloadsPageKeyMap() DownloadsPageKeyMap {
	return DownloadsPageKeyMap{
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
		// m.Cancel,
	}}
	return kb
}

// ShortHelp returns bindings to show in the abbreviated help view. It's part
// of the help.KeyMap interface.
func (m DownloadsPageKeyMap) ShortHelp() []key.Binding {
	kb := []key.Binding{
		// m.Cancel,
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
		cmd := m.onQueueDownload(msg)
		return m, cmd

	case PerformDownloadMsg:
		cmd := m.onPerformDownload()
		return m, cmd

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
	viewStr += m.help.View(&m.keyMap)
	return viewStr
}

///////////////////////////////////////////////////////////////////////////////

// updateSizes update the heights/widths of widgets
func (m *DownloadsPageModel) updateSizes() {
	availHeight := m.height - 2 - 2 - 2 // 2xAppHeaderFooter 2xPaneBorder
	m.downloadsTable.SetHeight(availHeight)
	m.downloadsTable.SetWidth(m.width - 2)

	availWidth := m.width - 2 // 2xPaneBorder
	if m.lastError != nil {
		availWidth -= lipgloss.Width(fmt.Sprintf("Error: %s ", m.lastError))
	}
	m.help.Width = availWidth
}

// listenForProgress is a command that waits for the responses on the channel
func (m *DownloadsPageModel) listenForProgress() tea.Cmd {
	return func() tea.Msg {
		return DownloadProgressMsg(<-m.progressCh)
	}
}

// checkDownloadQueue is a command that waits for the download queue and begins downloads
func (m *DownloadsPageModel) checkDownloadQueue() tea.Cmd {
	return func() tea.Msg {
		// count the number of active downloads
		if m.queuedCount > 0 && m.activeCount < MaxActiveDownloads {
			// start the first queued download
			return PerformDownloadMsg{}
		}
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// HTTP Download Queue

type QueueDownloadMsg struct {
	JobID string
	Files []dbn_hist.BatchFileDesc
}

type PerformDownloadMsg struct{}

type DownloadProgressMsg struct {
	JobID        string
	File         string
	Url          string
	ExpectedSize uint64
	CurrentSize  uint64
	Error        error
}

// DownloadProgressWriter is an io.Writer that reports download progress via a channel
// An instance is fed to TeeWriter to track the HTTP download progress
type DownloadProgressWriter struct {
	JobID        string
	File         string
	Url          string
	ExpectedSize uint64
	CurrentSize  uint64
	ProgressCh   chan DownloadProgressMsg
}

// Write implements the io.Writer interface, tracking and reporting bytes read on the channel
func (w *DownloadProgressWriter) Write(p []byte) (int, error) {
	if p == nil {
		return 0, nil
	}
	n := len(p)
	w.CurrentSize += uint64(n)
	w.ProgressCh <- DownloadProgressMsg{
		JobID:        w.JobID,
		File:         w.File,
		Url:          w.Url,
		ExpectedSize: w.ExpectedSize,
		CurrentSize:  w.CurrentSize,
		Error:        nil,
	}
	return n, nil
}

// onQueueDownload handles a QueueDownloadMsg, enqueueing all the download items
func (m *DownloadsPageModel) onQueueDownload(msg QueueDownloadMsg) tea.Cmd {
	var cmds []tea.Cmd
	for _, file := range msg.Files {
		// extract https url
		httpsUrl := file.Urls["https"]
		if httpsUrl == "" {
			continue
		}
		// check if already queued
		for _, download := range m.downloads {
			if download.JobID == msg.JobID && download.Url == httpsUrl {
				continue
			}
		}
		// queue download
		downloadItem := DownloadItem{
			JobID:         msg.JobID,
			Filename:      file.Filename,
			FileHash:      file.Hash,
			DestFile:      file.Filename, // TODO: dest path, right now is cwd
			Url:           httpsUrl,
			ExpectedSize:  file.Size,
			State:         DownloadQueued,
			ContextCancel: nil,
		}
		m.downloads = append(m.downloads, downloadItem)
		m.queuedCount++

		m.downloadsTable.SetRows(append(m.downloadsTable.Rows(), table.Row{
			downloadItem.JobID,
			downloadItem.Filename,
			lipgloss.NewStyle().Width(downloadsSizeColumnSize).Align(lipgloss.Right).
				Render(humanize.Comma(int64(downloadItem.ExpectedSize))),
			lipgloss.NewStyle().Width(downloadsStateColumnSize).Align(lipgloss.Center).
				Render(string(downloadItem.State)),
			"0.0%",
		}))
		cmds = append(cmds, m.checkDownloadQueue())
	}
	return tea.Batch(cmds...)
}

// onPerformDownload handles a PerformDownloadMsg, performing the download
func (m *DownloadsPageModel) onPerformDownload() tea.Cmd {
	// find the first queeud download and activate it
	for i, download := range m.downloads {
		if download.State == DownloadQueued {
			m.activeCount++
			m.queuedCount--
			m.downloads[i].State = DownloadActive
			return m.startPerformDownload(m.downloads[i])
		}
	}
	return nil
}

// onDownloadProgress handles a DownloadProgressMsg, updating the state of DownloadItems
func (m *DownloadsPageModel) onDownloadProgress(msg DownloadProgressMsg) tea.Cmd {
	// Find the download
	cmds := []tea.Cmd{m.listenForProgress()}
	for i, download := range m.downloads {
		if download.JobID == msg.JobID || download.Url == msg.Url {
			thisDownload := &m.downloads[i]
			// Update the download
			if msg.Error != nil {
				thisDownload.State = DownloadFailed
				m.activeCount--
				m.lastError = msg.Error
				cmds = append(cmds, m.checkDownloadQueue()) // start the next download
			} else {
				thisDownload.Progress = minFloat(1.0, float64(msg.CurrentSize)/float64(msg.ExpectedSize))
				if thisDownload.Progress >= 1.0 {
					thisDownload.Progress = 1.0
					thisDownload.State = DownloadComplete
					m.activeCount--
					cmds = append(cmds, m.checkDownloadQueue()) // start the next download
				}
			}

			// Update the row in the table
			for _, row := range m.downloadsTable.Rows() {
				if row[downloadsJobColumn] == msg.JobID && row[downloadsFileColumn] == msg.File {
					row[downloadsStateColumn] = string(thisDownload.State)
					row[downloadsProgressColumn] = fmt.Sprintf("%.2f%%", thisDownload.Progress*100)
					m.downloadsTable.UpdateViewport()
					break
				}
			}
			break
		}
	}
	return tea.Batch(cmds...)
}

///////////////////////////////////////////////////////////////////////////////

// startPerformDownload is a tea.Msg wrapper for perfomDownload
func (m *DownloadsPageModel) startPerformDownload(item DownloadItem) tea.Cmd {
	return func() tea.Msg {
		return m.performDownload(item)
	}
}

// performDownload downlaods the specified file and reports progress
// Adapted from: https://github.com/joncrlsn/go-examples/blob/master/http-download-with-progress.go#L15
func (m *DownloadsPageModel) performDownload(item DownloadItem) tea.Msg {
	// Create a DownloadProgressMsg to propogate errors
	progressMsg := DownloadProgressMsg{
		JobID:        item.JobID,
		File:         item.Filename,
		Url:          item.Url,
		ExpectedSize: item.ExpectedSize,
	}

	// Create the request
	apiUrl, err := url.Parse(item.Url)
	if err != nil {
		progressMsg.Error = err
		m.progressCh <- progressMsg
		return nil
	}

	ctx, _ := context.WithCancel(context.Background()) // TODO: use cancelFunc
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", apiUrl.String(), nil)
	if err != nil {
		progressMsg.Error = err
		m.progressCh <- progressMsg
		return nil
	}

	auth := base64.StdEncoding.EncodeToString([]byte(m.config.DatabentoApiKey + ":"))
	req.Header.Add("Authorization", "Basic "+auth)

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	tmpFile, err := os.Create(item.DestFile + ".tmp")
	if err != nil {
		progressMsg.Error = err
		m.progressCh <- progressMsg
		return nil
	}
	defer tmpFile.Close()

	// Get the data
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10
	retryClient.Logger = log.New(io.Discard, "", log.LstdFlags)
	resp, err := retryClient.Do(req)
	if err != nil {
		progressMsg.Error = err
		m.progressCh <- progressMsg
		return nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		progressMsg.Error = fmt.Errorf("key not authorized")
		m.progressCh <- progressMsg
		return nil
	} else if resp.StatusCode == http.StatusTooManyRequests {
		progressMsg.Error = fmt.Errorf("%s", resp.Status)
		m.progressCh <- progressMsg
		return nil
	} else if resp.StatusCode != http.StatusOK {
		progressMsg.Error = fmt.Errorf("%s", resp.Status)
		m.progressCh <- progressMsg
		return nil
	}
	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	progressWriter := &DownloadProgressWriter{
		JobID:        item.JobID,
		File:         item.Filename,
		Url:          item.Url,
		ExpectedSize: item.ExpectedSize,
		CurrentSize:  0,
		ProgressCh:   m.progressCh,
	}
	bytesCopied, err := io.Copy(tmpFile, io.TeeReader(resp.Body, progressWriter))
	if err != nil {
		progressMsg.Error = err
		m.progressCh <- progressMsg
		return nil
	}

	// Close the file without defer so it can happen before Rename()
	tmpFile.Close()
	if err = os.Rename(item.DestFile+".tmp", item.DestFile); err != nil {
		progressMsg.Error = err
		m.progressCh <- progressMsg
		return nil
	}

	progressMsg.CurrentSize = uint64(bytesCopied)
	m.progressCh <- progressMsg

	return nil
}
