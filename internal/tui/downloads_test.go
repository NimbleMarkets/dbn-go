// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"strings"
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
