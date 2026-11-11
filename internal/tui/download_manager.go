// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
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

	// Overridable for tests; nil or zero means the production default.
	httpClient *http.Client           // HTTP client for downloads
	urlAllowed func(u *url.URL) error // Decides whether the API key may be sent to u
	retryMax   *int                   // Max retries per download

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

	desc := DownloadDesc{
		JobID:    jobID,
		Url:      httpsUrl,
		Filename: file.Filename,
		FileHash: file.Hash,
		Size:     file.Size,
	}

	dm.queueMtx.Lock()
	defer dm.queueMtx.Unlock()

	// Try to find and cancel active download
	for i, download := range dm.activeDownloads {
		if download.Desc == desc {
			if download.ContextCancel != nil {
				download.ContextCancel()
			}
			// Move from active to past with failed state
			download.State = DownloadFailed
			dm.activeDownloads = slices.Delete(dm.activeDownloads, i, i+1)
			dm.pastDownloads = append(dm.pastDownloads, download)
			return true
		}
	}

	// Try to remove from queue if still queued
	for i, download := range dm.queuedDownloads {
		if download.Desc == desc {
			dm.queuedDownloads = slices.Delete(dm.queuedDownloads, i, i+1)
			return true
		}
	}

	return false
}

// Close closes the channels and exits the queue
func (dm *DownloadManager) Close() {
	dm.queueTicker.Stop()
	dm.queueExitCh <- 0
	// Cancel all active download contexts
	dm.queueMtx.Lock()
	for _, download := range dm.activeDownloads {
		if download.ContextCancel != nil {
			download.ContextCancel()
		}
	}
	dm.queueMtx.Unlock()
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
//
// The URL, file name, size and hash all come from the batch file listing, so
// none of them is trusted: the API key is only sent to a Databento HTTPS URL,
// the file name must be a plain name in the download directory, and the file
// is only kept if its size and SHA-256 hash match the listing.
func (dm *DownloadManager) performDownload(item DownloadItem) error {
	// Check everything from the listing before touching the disk or network
	apiUrl, err := url.Parse(item.Desc.Url)
	if err != nil {
		return fmt.Errorf("invalid download URL: %w", err)
	}
	urlAllowed := dm.urlAllowed
	if urlAllowed == nil {
		urlAllowed = databentoDownloadURLAllowed
	}
	if err := urlAllowed(apiUrl); err != nil {
		return fmt.Errorf("refusing download URL: %w", err)
	}
	if err := checkDownloadFilename(item.Desc.Filename); err != nil {
		return err
	}
	if item.DestFile != item.Desc.Filename {
		return fmt.Errorf("unexpected destination %q for file %q", item.DestFile, item.Desc.Filename)
	}
	wantHash, err := parseSHA256Hash(item.Desc.FileHash)
	if err != nil {
		return err
	}
	wantSize := item.Desc.Size
	if wantSize >= math.MaxInt64 {
		return fmt.Errorf("refusing download with listed size %d", wantSize)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Store the cancel function in the item so it can be cancelled later
	dm.queueMtx.Lock()
	for i := range dm.activeDownloads {
		if dm.activeDownloads[i].Desc == item.Desc {
			dm.activeDownloads[i].ContextCancel = cancel
			break
		}
	}
	dm.queueMtx.Unlock()
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", apiUrl.String(), nil)
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(dm.databentoApiKey + ":"))
	req.Header.Add("Authorization", "Basic "+auth)

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	tmpPath := item.DestFile + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	keepTmp := false
	defer func() {
		tmpFile.Close()
		if !keepTmp {
			os.Remove(tmpPath)
		}
	}()

	// Get the data
	httpClient := &http.Client{}
	if dm.httpClient != nil {
		*httpClient = *dm.httpClient
	}
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		if req.URL.Scheme != "https" {
			return fmt.Errorf("refusing redirect to non-https URL %q", req.URL.Redacted())
		}
		if urlAllowed(req.URL) != nil {
			req.Header.Del("Authorization") // the key stays with Databento
		}
		return nil
	}
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = httpClient
	retryClient.RetryMax = 10
	if dm.retryMax != nil {
		retryClient.RetryMax = *dm.retryMax
	}
	retryClient.Logger = log.New(io.Discard, "", log.LstdFlags)
	resp, err := retryClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("key not authorized")
	} else if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("%s", resp.Status)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}

	// Create our progress reporter and pass it to be used alongside our writer
	progressWriter := &DownloadProgressWriter{
		Desc:        item.Desc,
		CurrentSize: 0,
		ProgressCh:  dm.progressCh,
	}
	// Never write more than the listing declared (one extra byte detects an
	// overrun). This holds for every listed size, including zero.
	body := io.LimitReader(resp.Body, int64(wantSize)+1)
	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(tmpFile, hasher), io.TeeReader(body, progressWriter))
	if err != nil {
		return err
	}
	if uint64(written) != wantSize {
		if uint64(written) > wantSize {
			return fmt.Errorf("download is larger than the %d bytes listed", wantSize)
		}
		return fmt.Errorf("download is %d bytes, but %d were listed", written, wantSize)
	}
	if !bytes.Equal(hasher.Sum(nil), wantHash) {
		return fmt.Errorf("download does not match the listed SHA-256 hash")
	}

	// Close the file without defer so it can happen before Rename()
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmpPath, item.DestFile); err != nil {
		return err
	}
	keepTmp = true

	return nil
}

// databentoDownloadURLAllowed reports whether the API key may be sent to u:
// only over HTTPS, on the default port, to databento.com or a subdomain.
func databentoDownloadURLAllowed(u *url.URL) error {
	if u.Scheme != "https" {
		return fmt.Errorf("%q is not an https URL", u.Redacted())
	}
	if u.User != nil {
		return errors.New("URL must not contain credentials")
	}
	if port := u.Port(); port != "" && port != "443" {
		return fmt.Errorf("unexpected port %q", port)
	}
	host := strings.ToLower(u.Hostname())
	if host != "databento.com" && !strings.HasSuffix(host, ".databento.com") {
		return fmt.Errorf("%q is not a Databento host", host)
	}
	return nil
}

// checkDownloadFilename returns an error unless name is a plain file name:
// no directory part, drive or stream separator, or control characters, and
// no trailing dot or space. Windows drops trailing dots and spaces, so "..."
// or ".. " would name a parent directory there; "." and ".." end in a dot too.
func checkDownloadFilename(name string) error {
	invalid := name == "" || strings.ContainsAny(name, `/\:`) || strings.TrimRight(name, ". ") != name
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			invalid = true
		}
	}
	if invalid {
		return fmt.Errorf("refusing unsafe file name %q", name)
	}
	return nil
}

// parseSHA256Hash decodes a listing hash of the form "sha256:<hex>".
// A hash is required: without one the content cannot be verified, and it may
// have been served by a host other than the one the listing named.
func parseSHA256Hash(hash string) ([]byte, error) {
	if hash == "" {
		return nil, errors.New("refusing download with no listed hash")
	}
	hexDigest, ok := strings.CutPrefix(hash, "sha256:")
	if !ok {
		return nil, fmt.Errorf("unsupported file hash %q", hash)
	}
	digest, err := hex.DecodeString(hexDigest)
	if err != nil || len(digest) != sha256.Size {
		return nil, fmt.Errorf("malformed file hash %q", hash)
	}
	return digest, nil
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
