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
	"slices"
	"sync"
	"time"

	dbn_hist "github.com/NimbleMarkets/dbn-go/hist"
	"github.com/hashicorp/go-retryablehttp"
)

type DownloadState string

const (
	DownloadNone     DownloadState = ""
	DownloadQueued   DownloadState = "queued"
	DownloadActive   DownloadState = "active"
	DownloadComplete DownloadState = "complete"
	DownloadFailed   DownloadState = "failed"
)

///////////////////////////////////////////////////////////////////////////////

// DownloadDesc decribes a specific download
type DownloadDesc struct {
	JobID    string
	Url      string
	Filename string
	FileHash string
	Size     uint64
}

type DownloadItem struct {
	Desc          DownloadDesc
	DestFile      string
	CurrentSize   float64
	State         DownloadState
	ContextCancel context.CancelFunc
}

type DownloadCompleteMsg struct {
	Desc  DownloadDesc
	State DownloadState
}

type DownloadProgressMsg struct {
	Desc        DownloadDesc
	CurrentSize uint64
	State       DownloadState
	Error       error
}

///////////////////////////////////////////////////////////////////////////////

type DownloadManager struct {
	// config
	databentoApiKey    string
	maxActiveDownloads int

	// concurrency
	progressCh chan DownloadProgressMsg // results are sent here

	queueTicker *time.Ticker
	queueExitCh chan int

	// state
	queueMtx        sync.Mutex
	queuedDownloads []DownloadItem
	activeDownloads []DownloadItem
	pastDownloads   []DownloadItem

	progressMtx     sync.Mutex
	progressBacklog []DownloadProgressMsg
}

func NewDownloadManager(databentoApiKey string, maxActiveDownloads int) *DownloadManager {
	dm := &DownloadManager{
		databentoApiKey:    databentoApiKey,
		maxActiveDownloads: maxActiveDownloads,
		progressCh:         make(chan DownloadProgressMsg, 500),
		queueExitCh:        make(chan int, 10),
		queueTicker:        time.NewTicker(100 * time.Millisecond),
	}
	go dm.queueHandler()
	return dm
}

func (dm *DownloadManager) ProgressChannel() chan DownloadProgressMsg {
	return dm.progressCh
}

// Counts returns the number of queued, active, and past downloads
func (dm *DownloadManager) Counts() (queued, active, past int) {
	dm.queueMtx.Lock()
	defer dm.queueMtx.Unlock()
	return len(dm.queuedDownloads), len(dm.activeDownloads), len(dm.pastDownloads)
}

// QueueDownload queues a download for the specified job and file.
func (dm *DownloadManager) QueueDownload(jobID string, file dbn_hist.BatchFileDesc) bool {
	// extract https url
	httpsUrl := file.Urls["https"]
	if httpsUrl == "" {
		return false // we only support https
	}

	// check if already queued or processed
	desc := DownloadDesc{
		JobID:    jobID,
		Url:      httpsUrl,
		Filename: file.Filename,
		FileHash: file.Hash,
		Size:     file.Size,
	}
	added := dm.enqueueDownload(desc)
	if added {
		dm.sendProgress(&DownloadProgressMsg{
			Desc:        desc,
			CurrentSize: 0,
			State:       DownloadQueued,
		})
	}
	return added
}

func (dm *DownloadManager) CancelDownload(jobID string, file dbn_hist.BatchFileDesc) bool {
	// extract https url
	httpsUrl := file.Urls["https"]
	if httpsUrl == "" {
		return false // we only support https
	}
	// TODO
	return false
}

// Close closes the channels and exits the queue
func (dm *DownloadManager) Close() {
	dm.queueTicker.Stop()
	dm.queueExitCh <- 0
	// TODO cancel all active contexts
}

// enqueueDownload adds a download item to the queue.  Returns true if added or false if duplicate.
func (dm *DownloadManager) enqueueDownload(desc DownloadDesc) bool {
	dm.queueMtx.Lock()
	defer dm.queueMtx.Unlock()

	// check if already queued or completed (this TUI session)
	for _, download := range dm.queuedDownloads {
		if download.Desc == desc {
			return false
		}
	}
	for _, download := range dm.activeDownloads {
		if download.Desc == desc {
			return false
		}
	}
	for _, download := range dm.pastDownloads {
		if download.Desc == desc {
			return false
		}
	}

	downloadItem := DownloadItem{
		Desc:          desc,
		DestFile:      desc.Filename, // TODO: dest path, right now is cwd
		CurrentSize:   0,
		State:         DownloadQueued,
		ContextCancel: nil,
	}

	dm.queuedDownloads = append(dm.queuedDownloads, downloadItem)
	return true
}

// completeItem
func (dm *DownloadManager) completeDownload(completedDesc DownloadDesc, state DownloadState) bool {
	dm.queueMtx.Lock()
	defer dm.queueMtx.Unlock()

	// find the active download
	for i, download := range dm.activeDownloads {
		if download.Desc == completedDesc {
			// found, change its state and complete it
			download.State = state
			dm.activeDownloads = slices.Delete(dm.activeDownloads, i, i+1)
			dm.pastDownloads = append(dm.pastDownloads, download)
			return true
		}
	}
	return false
}

func (dm *DownloadManager) sendProgress(progressMsg *DownloadProgressMsg) {
	dm.progressMtx.Lock()
	defer dm.progressMtx.Unlock()

	if progressMsg != nil {
		dm.progressBacklog = append(dm.progressBacklog, *progressMsg)
	}

	// first try to send the backlog
	for i, msg := range dm.progressBacklog {
		if success := TrySendChannel(msg, dm.progressCh); !success {
			// cut the successful messages
			dm.progressBacklog = dm.progressBacklog[i:]
			return
		}
	}
	dm.progressBacklog = nil // all sent
}

///////////////////////////////////////////////////////////////////////////////

// queueHandler is the main loop for the downloads manager
// It is the only goroutine that should access the downloads slice
func (dm *DownloadManager) queueHandler() {
	for {
		select {
		case <-dm.queueExitCh:
			return // all done!
		case <-dm.queueTicker.C:
			dm.sendProgress(nil)
			for dm.checkQueue() { // loop till we can't queue anymore
			}
		}
	}
}

// checkQueue checks the queue, activating a download if possible.  Returns true if activated.
func (dm *DownloadManager) checkQueue() bool {
	dm.queueMtx.Lock()

	// check the queue and if we have active slots available
	if len(dm.queuedDownloads) == 0 || len(dm.activeDownloads) >= dm.maxActiveDownloads {
		dm.queueMtx.Unlock()
		return false
	}

	// pop item from queue and put into active state
	var item DownloadItem
	item, dm.queuedDownloads = dm.queuedDownloads[0], dm.queuedDownloads[1:]
	item.State = DownloadActive
	dm.activeDownloads = append(dm.activeDownloads, item)
	dm.queueMtx.Unlock()

	go func() {
		progressMsg := DownloadProgressMsg{Desc: item.Desc}
		if err := dm.performDownload(item); err != nil {
			progressMsg.State = DownloadFailed
			progressMsg.Error = err
		} else {
			progressMsg.State = DownloadComplete
			progressMsg.CurrentSize = item.Desc.Size
		}
		dm.sendProgress(&progressMsg)
		dm.completeDownload(item.Desc, progressMsg.State)
	}()
	return true
}

///////////////////////////////////////////////////////////////////////////////

// performDownload downloads the specified file and reports progress on the channel
// Adapted from: https://github.com/joncrlsn/go-examples/blob/master/http-download-with-progress.go#L15
func (dm *DownloadManager) performDownload(item DownloadItem) error {
	// Create the request
	apiUrl, err := url.Parse(item.Desc.Url)
	if err != nil {
		return err
	}

	//ctx, _ := context.WithCancel(context.Background()) // TODO: use cancelFunc
	ctx := context.Background()
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", apiUrl.String(), nil)
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(dm.databentoApiKey + ":"))
	req.Header.Add("Authorization", "Basic "+auth)

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	tmpFile, err := os.Create(item.DestFile + ".tmp")
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	// Get the data
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10
	retryClient.Logger = log.New(io.Discard, "", log.LstdFlags)
	resp, err := retryClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("key not authorized")
	} else if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("%s", resp.Status)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}
	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	progressWriter := &DownloadProgressWriter{
		Desc:        item.Desc,
		CurrentSize: 0,
		ProgressCh:  dm.progressCh,
	}
	_, err = io.Copy(tmpFile, io.TeeReader(resp.Body, progressWriter))
	if err != nil {
		return err
	}

	// Close the file without defer so it can happen before Rename()
	tmpFile.Close()
	if err = os.Rename(item.DestFile+".tmp", item.DestFile); err != nil {
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// HTTP Download Queue

// DownloadProgressWriter is an io.Writer that reports download progress via a channel
// An instance is fed to TeeWriter to track the HTTP download progress
type DownloadProgressWriter struct {
	Desc        DownloadDesc
	CurrentSize uint64
	ProgressCh  chan DownloadProgressMsg
}

// Write implements the io.Writer interface, tracking and reporting bytes read on the channel
func (w *DownloadProgressWriter) Write(p []byte) (int, error) {
	if p == nil {
		return 0, nil
	}
	n := len(p)
	w.CurrentSize += uint64(n)
	w.ProgressCh <- DownloadProgressMsg{
		Desc:        w.Desc,
		CurrentSize: w.CurrentSize,
		State:       DownloadActive,
		Error:       nil,
	}
	return n, nil
}
