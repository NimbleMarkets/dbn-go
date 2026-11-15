// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestDownloadsPage_CancelKeyCancelsSelectedDownload(t *testing.T) {
	server := newRecordingServer(t, true, serveBody("x"))
	page := NewDownloadsPage(Config{})
	page.downloadManager.Close()
	dm := cancelTestManager(server)
	page.downloadManager = dm

	update := func(msg tea.Msg) {
		t.Helper()
		model, _ := page.Update(msg)
		page = model.(DownloadsPageModel)
	}

	if !dm.QueueDownload("JOB", batchFile(server, "data.dbn.zst", 1)) {
		t.Fatal("download was not queued")
	}
	update(waitForProgress(t, dm, func(msg DownloadProgressMsg) bool { return msg.State == DownloadQueued }))

	if help := page.help.View(&page.keyMap); !strings.Contains(help, "cancel") {
		t.Errorf("help %q does not mention cancel", help)
	}

	update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	msg := waitForProgress(t, dm, isTerminal)
	if msg.State != DownloadCancelled {
		t.Fatalf("state after pressing x = %q, want %q", msg.State, DownloadCancelled)
	}
	update(msg)
	if row := page.downloadsTable.SelectedRow(); !strings.Contains(row[columnDownloadsStateIndex], string(DownloadCancelled)) {
		t.Errorf("state cell = %q, want %q", row[columnDownloadsStateIndex], DownloadCancelled)
	}
}

func TestDownloadsPage_RequeueFailedDownloadClearsErrorBanner(t *testing.T) {
	var attempts atomic.Int32
	server := newRecordingServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			// First attempt returns a wrong body so SHA-256 verification fails.
			w.Write([]byte("w"))
			return
		}
		w.Write([]byte("x"))
	})
	chdirToSubdir(t)
	page := NewDownloadsPage(Config{})
	page.downloadManager.Close()
	dm := cancelTestManager(server)
	page.downloadManager = dm

	update := func(msg tea.Msg) {
		t.Helper()
		model, _ := page.Update(msg)
		page = model.(DownloadsPageModel)
	}

	file := batchFile(server, "data.dbn.zst", 1)
	if !dm.QueueDownload("JOB", file) || !dm.checkQueue() {
		t.Fatal("failed download was not started")
	}
	update(waitForProgress(t, dm, func(msg DownloadProgressMsg) bool { return msg.State == DownloadQueued }))
	failedMsg := waitForProgress(t, dm, isTerminal)
	if failedMsg.State != DownloadFailed || failedMsg.Error == nil {
		t.Fatalf("state = %q, err = %v; want %q with error", failedMsg.State, failedMsg.Error, DownloadFailed)
	}
	update(failedMsg)
	if !strings.Contains(page.View().Content, "Error:") {
		t.Fatalf("View() = %q, want Error banner after failure", page.View().Content)
	}

	// Re-queueing the same file removes its failed pastDownloads entry, clears
	// the error banner, and completes cleanly.
	if !dm.QueueDownload("JOB", file) {
		t.Fatal("re-queueing failed download returned false")
	}
	if queued, active, past := dm.Counts(); queued != 1 || active != 0 || past != 0 {
		t.Errorf("Counts() after re-queue = %d, %d, %d; want 1, 0, 0", queued, active, past)
	}
	update(waitForProgress(t, dm, func(msg DownloadProgressMsg) bool { return msg.State == DownloadQueued }))
	if strings.Contains(page.View().Content, "Error:") {
		t.Errorf("View() still shows Error banner after re-queue: %q", page.View().Content)
	}
	if !dm.checkQueue() {
		t.Fatal("re-queued download was not started")
	}
	doneMsg := waitForProgress(t, dm, isTerminal)
	if doneMsg.State != DownloadComplete || doneMsg.Error != nil {
		t.Fatalf("state = %q, err = %v; want %q", doneMsg.State, doneMsg.Error, DownloadComplete)
	}
	update(doneMsg)
	if strings.Contains(page.View().Content, "Error:") {
		t.Errorf("View() still shows Error banner after completion: %q", page.View().Content)
	}
	if row := page.downloadsTable.SelectedRow(); !strings.Contains(row[columnDownloadsStateIndex], string(DownloadComplete)) {
		t.Errorf("state cell = %q, want %q", row[columnDownloadsStateIndex], DownloadComplete)
	}
	if queued, active, past := dm.Counts(); queued != 0 || active != 0 || past != 1 {
		t.Errorf("Counts() after completion = %d, %d, %d; want 0, 0, 1", queued, active, past)
	}
}
