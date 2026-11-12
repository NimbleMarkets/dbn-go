// Copyright (c) 2025 Neomantra Corp

package tui

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const testDownloadKey = "db-testkeytestkeytestkeytestkey00"

// recordingServer serves body at every path and records the requests it gets.
type recordingServer struct {
	*httptest.Server
	mu       sync.Mutex
	requests []*http.Request
}

func newRecordingServer(t *testing.T, useTLS bool, handler http.HandlerFunc) *recordingServer {
	t.Helper()
	rs := &recordingServer{}
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rs.mu.Lock()
		rs.requests = append(rs.requests, r.Clone(r.Context()))
		rs.mu.Unlock()
		handler(w, r)
	})
	if useTLS {
		rs.Server = httptest.NewTLSServer(wrapped)
	} else {
		rs.Server = httptest.NewServer(wrapped)
	}
	t.Cleanup(rs.Close)
	return rs
}

func (rs *recordingServer) received() []*http.Request {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return append([]*http.Request(nil), rs.requests...)
}

func serveBody(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }
}

// testDownloadManager returns a manager that does not retry and trusts test
// TLS certificates. When trusted servers are given, only their URLs are
// allowed to receive the API key; otherwise the production rule applies.
func testDownloadManager(trusted ...*recordingServer) *DownloadManager {
	noRetries := 0
	dm := &DownloadManager{
		databentoApiKey: testDownloadKey,
		progressCh:      make(chan DownloadProgressMsg, 100_000),
		retryMax:        &noRetries,
		httpClient: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // test servers use self-signed certificates
		}},
	}
	if len(trusted) > 0 {
		dm.urlAllowed = func(u *url.URL) error {
			for _, rs := range trusted {
				if base, _ := url.Parse(rs.URL); u.Scheme == "https" && u.Host == base.Host {
					return nil
				}
			}
			return http.ErrNotSupported
		}
	}
	return dm
}

func sha256Hash(body string) string {
	sum := sha256.Sum256([]byte(body))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// chdirToSubdir makes root/work the current directory, so a path escaping
// the download directory still lands inside the test's temp dir.
func chdirToSubdir(t *testing.T) (root string) {
	t.Helper()
	root = t.TempDir()
	work := filepath.Join(root, "work")
	if err := os.Mkdir(work, 0755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(work)
	return root
}

// filesUnder lists every file under root, relative to root.
func filesUnder(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(root, path)
			files = append(files, rel)
		}
		return nil
	})
	return files
}

func download(dm *DownloadManager, desc DownloadDesc) error {
	return dm.performDownload(DownloadItem{Desc: desc, DestFile: desc.Filename})
}

func TestPerformDownload_WritesVerifiedFile(t *testing.T) {
	const body = "dbn file contents"
	server := newRecordingServer(t, true, serveBody(body))
	root := chdirToSubdir(t)

	err := download(testDownloadManager(server), DownloadDesc{
		Url: server.URL + "/v0/batch/download/job/data.dbn.zst", Filename: "data.dbn.zst",
		FileHash: sha256Hash(body), Size: uint64(len(body)),
	})
	if err != nil {
		t.Fatalf("performDownload: %v", err)
	}
	got, err := os.ReadFile("data.dbn.zst")
	if err != nil || string(got) != body {
		t.Fatalf("downloaded file = %q, %v; want %q", got, err, body)
	}
	if files := filesUnder(t, root); len(files) != 1 {
		t.Errorf("files left behind: %v, want only work/data.dbn.zst", files)
	}
	requests := server.received()
	if len(requests) != 1 || !strings.HasPrefix(requests[0].Header.Get("Authorization"), "Basic ") {
		t.Errorf("expected one authorized request, got %d", len(requests))
	}
}

// The API key must only be sent over HTTPS to Databento, whatever URL the
// file listing names.
func TestPerformDownload_NeverSendsKeyToUntrustedURL(t *testing.T) {
	plain := newRecordingServer(t, false, serveBody("x"))
	secure := newRecordingServer(t, true, serveBody("x"))
	chdirToSubdir(t)

	for _, rawURL := range []string{
		plain.URL + "/file",  // not https
		secure.URL + "/file", // https, but not a Databento host
		"http://hist.databento.com/file",
		"https://hist.databento.com.evil.example/file",
		"https://evil.example/hist.databento.com/file",
		"https://databento.com@evil.example/file",
		"https://key:secret@hist.databento.com/file",
		"https://hist.databento.com:8443/file",
		"ftp://hist.databento.com/file",
		"file:///etc/passwd",
		"://bad",
	} {
		err := download(testDownloadManager(), DownloadDesc{Url: rawURL, Filename: "data.dbn.zst"})
		if err == nil {
			t.Errorf("download from %q succeeded, want an error", rawURL)
		}
	}
	if n := len(plain.received()) + len(secure.received()); n != 0 {
		t.Errorf("%d requests reached servers that are not Databento", n)
	}
}

func TestDownloadURLAllowed_AcceptsDatabentoHTTPS(t *testing.T) {
	for _, rawURL := range []string{
		"https://hist.databento.com/v0/batch/download/JOB/file.dbn.zst",
		"https://HIST.Databento.com/v0/batch/download/JOB/file.dbn.zst",
		"https://hist.databento.com:443/file",
		"https://databento.com/file",
	} {
		u, err := url.Parse(rawURL)
		if err != nil {
			t.Fatal(err)
		}
		if err := databentoDownloadURLAllowed(u); err != nil {
			t.Errorf("databentoDownloadURLAllowed(%q) = %v, want nil", rawURL, err)
		}
	}
}

// A redirect must not carry the API key to a host that may not receive it,
// and must not downgrade to plain HTTP.
func TestPerformDownload_RedirectDoesNotLeakKey(t *testing.T) {
	const body = "redirected contents"
	other := newRecordingServer(t, true, serveBody(body))
	plain := newRecordingServer(t, false, serveBody(body))
	chdirToSubdir(t)

	toOther := newRecordingServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/cdn/file", http.StatusFound)
	})
	if err := download(testDownloadManager(toOther), DownloadDesc{Url: toOther.URL + "/file", Filename: "a.dbn.zst",
		FileHash: sha256Hash(body), Size: uint64(len(body))}); err != nil {
		t.Fatalf("download via https redirect: %v", err)
	}
	requests := other.received()
	if len(requests) != 1 {
		t.Fatalf("redirect target got %d requests, want 1", len(requests))
	}
	if auth := requests[0].Header.Get("Authorization"); auth != "" {
		t.Errorf("API key was forwarded to the redirect target: Authorization %q", auth)
	}

	toPlain := newRecordingServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL+"/file", http.StatusFound)
	})
	if err := download(testDownloadManager(toPlain), DownloadDesc{Url: toPlain.URL + "/file", Filename: "b.dbn.zst",
		FileHash: sha256Hash(body), Size: uint64(len(body))}); err == nil {
		t.Error("download via redirect to plain http succeeded, want an error")
	}
	if n := len(plain.received()); n != 0 {
		t.Errorf("plain-http redirect target got %d requests, want 0", n)
	}
}

// The file name comes from the server; it must name a file in the download
// directory and nothing else.
func TestPerformDownload_RejectsFilenameOutsideDownloadDir(t *testing.T) {
	const body = "x"
	server := newRecordingServer(t, true, serveBody(body))
	root := chdirToSubdir(t)
	abs := filepath.Join(root, "abs.dbn.zst")

	for _, name := range []string{"../escape.dbn.zst", abs, "sub/file.dbn.zst", `sub\file.dbn.zst`, "..", ".", "", "C:file.dbn.zst", "a\x00b", "a\nb",
		"...", ".. ", "file.dbn.zst.", "file.dbn.zst ", " "} {
		err := download(testDownloadManager(server), DownloadDesc{
			Url: server.URL + "/file", Filename: name, FileHash: sha256Hash(body), Size: 1,
		})
		if err == nil {
			t.Errorf("download to %q succeeded, want an error", name)
		}
	}
	if files := filesUnder(t, root); len(files) != 0 {
		t.Errorf("files were created: %v", files)
	}
	if n := len(server.received()); n != 0 {
		t.Errorf("%d requests were sent for rejected file names, want 0", n)
	}
}

func TestPerformDownload_RejectsWrongContent(t *testing.T) {
	const body = "the real contents"
	for _, tc := range []struct {
		name string
		desc DownloadDesc
	}{
		{"hash_mismatch", DownloadDesc{FileHash: sha256Hash("something else"), Size: uint64(len(body))}},
		{"larger_than_declared", DownloadDesc{FileHash: sha256Hash(body), Size: uint64(len(body)) - 1}},
		{"smaller_than_declared", DownloadDesc{FileHash: sha256Hash(body), Size: uint64(len(body)) + 1}},
		{"unsupported_hash", DownloadDesc{FileHash: "md5:d41d8cd98f00b204e9800998ecf8427e", Size: uint64(len(body))}},
		{"missing_hash", DownloadDesc{Size: uint64(len(body))}},
		{"zero_size_listed", DownloadDesc{FileHash: sha256Hash(body), Size: 0}},
		{"zero_size_and_missing_hash", DownloadDesc{}},
		{"size_beyond_int64", DownloadDesc{FileHash: sha256Hash(body), Size: 1 << 63}},
		{"size_max_uint64", DownloadDesc{FileHash: sha256Hash(body), Size: 1<<64 - 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newRecordingServer(t, true, serveBody(body))
			root := chdirToSubdir(t)
			tc.desc.Url, tc.desc.Filename = server.URL+"/file", "data.dbn.zst"
			if err := download(testDownloadManager(server), tc.desc); err == nil {
				t.Error("download succeeded, want an error")
			}
			if files := filesUnder(t, root); len(files) != 0 {
				t.Errorf("files left behind: %v", files)
			}
		})
	}
}

// A failed download must not leave its temporary file behind.
func TestPerformDownload_RemovesTempFileOnHTTPError(t *testing.T) {
	server := newRecordingServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	})
	root := chdirToSubdir(t)
	if err := download(testDownloadManager(server), DownloadDesc{Url: server.URL + "/file", Filename: "data.dbn.zst",
		FileHash: sha256Hash("x"), Size: 1}); err == nil {
		t.Error("download succeeded, want an error")
	}
	if files := filesUnder(t, root); len(files) != 0 {
		t.Errorf("files left behind: %v", files)
	}
}

// A body served from a redirect target is held to the listed size and hash
// like any other.
func TestPerformDownload_VerifiesRedirectedBody(t *testing.T) {
	const listed = "the listed contents"
	other := newRecordingServer(t, true, serveBody("something else entirely, and longer than what was listed"))
	redirector := newRecordingServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/cdn/file", http.StatusFound)
	})
	root := chdirToSubdir(t)
	err := download(testDownloadManager(redirector), DownloadDesc{Url: redirector.URL + "/file", Filename: "a.dbn.zst",
		FileHash: sha256Hash(listed), Size: uint64(len(listed))})
	if err == nil {
		t.Error("download succeeded, want an error")
	}
	if files := filesUnder(t, root); len(files) != 0 {
		t.Errorf("files left behind: %v", files)
	}
}

// An empty file is still a valid download when the listing says so.
func TestPerformDownload_AcceptsListedEmptyFile(t *testing.T) {
	server := newRecordingServer(t, true, serveBody(""))
	chdirToSubdir(t)
	err := download(testDownloadManager(server), DownloadDesc{Url: server.URL + "/file", Filename: "empty.json",
		FileHash: sha256Hash(""), Size: 0})
	if err != nil {
		t.Fatalf("performDownload: %v", err)
	}
	if info, err := os.Stat("empty.json"); err != nil || info.Size() != 0 {
		t.Errorf("empty.json: %v, %v; want an empty file", info, err)
	}
}
