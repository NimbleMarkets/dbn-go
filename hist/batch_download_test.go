//go:build !integration

// Copyright (c) 2026 Neomantra Corp

package dbn_hist

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

const testDownloadKey = "db-test-download-key"

// fakeBatchServer serves batch.list_files and the listed file URLs. Every
// request the hist client makes is routed to it, whatever the URL's host, so
// the descriptors can carry real Databento URLs.
type fakeBatchServer struct {
	*httptest.Server
	mu       sync.Mutex
	files    []BatchFileDesc
	bodies   map[string]string // URL path -> body
	requests []*http.Request
}

func (s *fakeBatchServer) received() []*http.Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.requests)
}

// fileRequests counts requests for file contents (not the listing).
func (s *fakeBatchServer) fileRequests() int {
	n := 0
	for _, r := range s.received() {
		if r.URL.Path != "/v0/batch.list_files" {
			n++
		}
	}
	return n
}

func sha256Hash(body string) string {
	sum := sha256.Sum256([]byte(body))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// fileDesc lists body at a Databento download URL with its size and hash.
func fileDesc(name, body string) BatchFileDesc {
	return BatchFileDesc{
		Filename: name,
		Size:     uint64(len(body)),
		Hash:     sha256Hash(body),
		Urls: map[string]string{
			"https": "https://api.databento.com/v0/job.download/JOB-1/" + name,
			"ftp":   "ftp://ftp.databento.com/JOB-1/" + name,
		},
	}
}

// newFakeBatchServer starts a server for files with the given bodies (keyed by
// URL path) and points the package HTTP client at it for the test.
func newFakeBatchServer(t *testing.T, files []BatchFileDesc, bodies map[string]string) *fakeBatchServer {
	t.Helper()
	s := &fakeBatchServer{files: files, bodies: bodies}
	s.Server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.requests = append(s.requests, r)
		s.mu.Unlock()
		if r.URL.Path == "/v0/batch.list_files" {
			_ = json.NewEncoder(w).Encode(s.files)
			return
		}
		body, ok := s.bodies[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(s.Close)

	addr := s.Listener.Addr().String()
	oldClient, oldRetry, oldWait := httpClient, downloadRetryMax, downloadRetryWait
	// Trust the test server's certificate, which is issued for example.com,
	// whatever host name the request uses.
	transport := s.Client().Transport.(*http.Transport).Clone()
	transport.TLSClientConfig.ServerName = "example.com"
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, addr)
	}
	httpClient = &http.Client{Transport: transport}
	downloadRetryMax, downloadRetryWait = 0, 0
	t.Cleanup(func() { httpClient, downloadRetryMax, downloadRetryWait = oldClient, oldRetry, oldWait })
	return s
}

// filesUnder lists every file under root, relative to root and slash-separated.
func filesUnder(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(root, path)
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	slices.Sort(files)
	return files
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func wantBasicAuth(t *testing.T, r *http.Request) {
	t.Helper()
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte(testDownloadKey+":"))
	if got := r.Header.Get("Authorization"); got != want {
		t.Errorf("%s: Authorization = %q, want %q", r.URL.Path, got, want)
	}
}

func TestDownloadJob_DownloadsAllFiles(t *testing.T) {
	a, b := "first file", "second file contents"
	files := []BatchFileDesc{fileDesc("a.dbn.zst", a), fileDesc("metadata.json", b)}
	srv := newFakeBatchServer(t, files, map[string]string{
		"/v0/job.download/JOB-1/a.dbn.zst":     a,
		"/v0/job.download/JOB-1/metadata.json": b,
	})
	out := t.TempDir()

	paths, err := DownloadJob(testDownloadKey, DownloadJobParams{JobID: "JOB-1", OutputDir: out})
	if err != nil {
		t.Fatalf("DownloadJob: %v", err)
	}
	wantPaths := []string{filepath.Join(out, "JOB-1", "a.dbn.zst"), filepath.Join(out, "JOB-1", "metadata.json")}
	if !slices.Equal(paths, wantPaths) {
		t.Errorf("paths = %q, want %q", paths, wantPaths)
	}
	if got := readFile(t, wantPaths[0]); got != a {
		t.Errorf("a.dbn.zst = %q, want %q", got, a)
	}
	if got := readFile(t, wantPaths[1]); got != b {
		t.Errorf("metadata.json = %q, want %q", got, b)
	}
	if got, want := filesUnder(t, out), []string{"JOB-1/a.dbn.zst", "JOB-1/metadata.json"}; !slices.Equal(got, want) {
		t.Errorf("files = %q, want %q (no temp files left)", got, want)
	}

	reqs := srv.received()
	if len(reqs) != 3 {
		t.Fatalf("got %d requests, want 3", len(reqs))
	}
	if reqs[0].URL.Path != "/v0/batch.list_files" || reqs[0].URL.Query().Get("job_id") != "JOB-1" {
		t.Errorf("first request = %s, want batch.list_files?job_id=JOB-1", reqs[0].URL)
	}
	for _, r := range reqs {
		wantBasicAuth(t, r)
	}
}

func TestDownloadJob_DownloadsSingleFile(t *testing.T) {
	a, b := "first file", "second file"
	files := []BatchFileDesc{fileDesc("a.dbn.zst", a), fileDesc("b.dbn.zst", b)}
	srv := newFakeBatchServer(t, files, map[string]string{
		"/v0/job.download/JOB-1/a.dbn.zst": a,
		"/v0/job.download/JOB-1/b.dbn.zst": b,
	})
	out := t.TempDir()

	paths, err := DownloadJob(testDownloadKey, DownloadJobParams{JobID: "JOB-1", OutputDir: out, FilenameToDownload: "b.dbn.zst"})
	if err != nil {
		t.Fatalf("DownloadJob: %v", err)
	}
	want := filepath.Join(out, "JOB-1", "b.dbn.zst")
	if !slices.Equal(paths, []string{want}) {
		t.Errorf("paths = %q, want [%q]", paths, want)
	}
	if got := readFile(t, want); got != b {
		t.Errorf("b.dbn.zst = %q, want %q", got, b)
	}
	if got := filesUnder(t, out); !slices.Equal(got, []string{"JOB-1/b.dbn.zst"}) {
		t.Errorf("files = %q, want only JOB-1/b.dbn.zst", got)
	}
	if n := srv.fileRequests(); n != 1 {
		t.Errorf("file requests = %d, want 1", n)
	}
}

func TestDownloadJob_UnknownFilenameErrors(t *testing.T) {
	files := []BatchFileDesc{fileDesc("a.dbn.zst", "x")}
	srv := newFakeBatchServer(t, files, map[string]string{"/v0/job.download/JOB-1/a.dbn.zst": "x"})
	out := t.TempDir()

	_, err := DownloadJob(testDownloadKey, DownloadJobParams{JobID: "JOB-1", OutputDir: out, FilenameToDownload: "nope.dbn.zst"})
	if err == nil || !strings.Contains(err.Error(), "nope.dbn.zst") {
		t.Fatalf("err = %v, want an error naming the missing file", err)
	}
	if n := srv.fileRequests(); n != 0 {
		t.Errorf("file requests = %d, want 0", n)
	}
	if got := filesUnder(t, out); len(got) != 0 {
		t.Errorf("files = %q, want none", got)
	}
}

func TestDownloadJob_RejectsUnsafeJobID(t *testing.T) {
	srv := newFakeBatchServer(t, nil, nil)
	for _, jobID := range []string{"", "..", "../JOB", "a/b", `a\b`, "C:x"} {
		if _, err := DownloadJob(testDownloadKey, DownloadJobParams{JobID: jobID, OutputDir: t.TempDir()}); err == nil {
			t.Errorf("job ID %q: expected error", jobID)
		}
	}
	if n := len(srv.received()); n != 0 {
		t.Errorf("requests = %d, want 0", n)
	}
}

func TestDownloadJob_ExpiredJob(t *testing.T) {
	srv := newFakeBatchServer(t, nil, nil)
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"expired"}`, http.StatusGone)
	})
	_, err := DownloadJob(testDownloadKey, DownloadJobParams{JobID: "JOB-1", OutputDir: t.TempDir()})
	var expired JobExpiredError
	if !errors.As(err, &expired) {
		t.Fatalf("err = %v, want JobExpiredError", err)
	}
}

func TestDownloadJobFile_HashMismatchLeavesNoFile(t *testing.T) {
	desc := fileDesc("a.dbn.zst", "expected content")
	newFakeBatchServer(t, nil, map[string]string{"/v0/job.download/JOB-1/a.dbn.zst": "tampered content"}) // same length
	dir := t.TempDir()

	_, err := DownloadJobFile(testDownloadKey, desc, dir)
	if err == nil || !strings.Contains(err.Error(), "hash") {
		t.Fatalf("err = %v, want hash mismatch", err)
	}
	if got := filesUnder(t, dir); len(got) != 0 {
		t.Errorf("files = %q, want none (no final or temp file)", got)
	}
}

func TestDownloadJobFile_HashMismatchKeepsExistingFile(t *testing.T) {
	desc := fileDesc("a.dbn.zst", "expected content")
	newFakeBatchServer(t, nil, map[string]string{"/v0/job.download/JOB-1/a.dbn.zst": "tampered content"})
	dir := t.TempDir()
	dest := filepath.Join(dir, "a.dbn.zst")
	if err := os.WriteFile(dest, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := DownloadJobFile(testDownloadKey, desc, dir); err == nil {
		t.Fatal("expected hash mismatch error")
	}
	if got := readFile(t, dest); got != "old" {
		t.Errorf("existing file = %q, want it untouched", got)
	}
	if got := filesUnder(t, dir); !slices.Equal(got, []string{"a.dbn.zst"}) {
		t.Errorf("files = %q, want only a.dbn.zst", got)
	}
}

func TestDownloadJobFile_SizeMismatch(t *testing.T) {
	for name, body := range map[string]string{"short": "abc", "long": "abcdefgh-and-more"} {
		t.Run(name, func(t *testing.T) {
			desc := fileDesc("a.dbn.zst", "abcdefgh")
			desc.Hash = "" // only the size can catch it
			newFakeBatchServer(t, nil, map[string]string{"/v0/job.download/JOB-1/a.dbn.zst": body})
			dir := t.TempDir()
			if _, err := DownloadJobFile(testDownloadKey, desc, dir); err == nil {
				t.Fatal("expected size error")
			}
			if got := filesUnder(t, dir); len(got) != 0 {
				t.Errorf("files = %q, want none", got)
			}
		})
	}
}

func TestDownloadJobFile_WithoutHashChecksSizeOnly(t *testing.T) {
	desc := fileDesc("a.dbn.zst", "content")
	desc.Hash = ""
	newFakeBatchServer(t, nil, map[string]string{"/v0/job.download/JOB-1/a.dbn.zst": "content"})
	dir := t.TempDir()
	path, err := DownloadJobFile(testDownloadKey, desc, dir)
	if err != nil {
		t.Fatalf("DownloadJobFile: %v", err)
	}
	if got := readFile(t, path); got != "content" {
		t.Errorf("content = %q", got)
	}
}

func TestDownloadJobFile_RejectsMalformedHash(t *testing.T) {
	srv := newFakeBatchServer(t, nil, map[string]string{"/v0/job.download/JOB-1/a.dbn.zst": "content"})
	for _, hash := range []string{"md5:abcd", "sha256:zz", "sha256:abcd", "deadbeef"} {
		desc := fileDesc("a.dbn.zst", "content")
		desc.Hash = hash
		if _, err := DownloadJobFile(testDownloadKey, desc, t.TempDir()); err == nil {
			t.Errorf("hash %q: expected error", hash)
		}
	}
	if n := srv.fileRequests(); n != 0 {
		t.Errorf("file requests = %d, want 0", n)
	}
}

func TestDownloadJobFile_RejectsPathTraversal(t *testing.T) {
	srv := newFakeBatchServer(t, nil, nil)
	root := t.TempDir()
	dir := filepath.Join(root, "out")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", ".", "..", "...", ".. ", "../escape", "a/b", `..\escape`, `a\b`, "C:escape", "/etc/passwd", "bad\x00name", "bad\nname"} {
		desc := fileDesc("x", "content")
		desc.Filename = name
		if _, err := DownloadJobFile(testDownloadKey, desc, dir); err == nil {
			t.Errorf("filename %q: expected error", name)
		}
	}
	if got := filesUnder(t, root); len(got) != 0 {
		t.Errorf("files = %q, want none", got)
	}
	if n := len(srv.received()); n != 0 {
		t.Errorf("requests = %d, want 0", n)
	}
}

func TestDownloadJob_RejectsTraversalFromListing(t *testing.T) {
	files := []BatchFileDesc{fileDesc("../escape.dbn", "x")}
	srv := newFakeBatchServer(t, files, map[string]string{"/v0/job.download/JOB-1/../escape.dbn": "x"})
	root := t.TempDir()
	out := filepath.Join(root, "out")

	if _, err := DownloadJob(testDownloadKey, DownloadJobParams{JobID: "JOB-1", OutputDir: out}); err == nil {
		t.Fatal("expected error for unsafe listed file name")
	}
	if got := filesUnder(t, root); len(got) != 0 {
		t.Errorf("files = %q, want none", got)
	}
	if n := srv.fileRequests(); n != 0 {
		t.Errorf("file requests = %d, want 0", n)
	}
}

func TestDownloadJobFile_SkipsCompleteExistingFile(t *testing.T) {
	desc := fileDesc("a.dbn.zst", "complete content")
	srv := newFakeBatchServer(t, nil, map[string]string{"/v0/job.download/JOB-1/a.dbn.zst": "complete content"})
	dir := t.TempDir()
	dest := filepath.Join(dir, "a.dbn.zst")
	if err := os.WriteFile(dest, []byte("complete content"), 0o644); err != nil {
		t.Fatal(err)
	}

	path, err := DownloadJobFile(testDownloadKey, desc, dir)
	if err != nil {
		t.Fatalf("DownloadJobFile: %v", err)
	}
	if path != dest {
		t.Errorf("path = %q, want %q", path, dest)
	}
	if n := srv.fileRequests(); n != 0 {
		t.Errorf("file requests = %d, want 0 for a complete file", n)
	}
}

func TestDownloadJobFile_ReplacesIncompleteOrWrongFile(t *testing.T) {
	for name, existing := range map[string]string{
		"partial":   "complete",
		"same size": "COMPLETE content",
		"longer":    "complete content and then some",
	} {
		t.Run(name, func(t *testing.T) {
			desc := fileDesc("a.dbn.zst", "complete content")
			srv := newFakeBatchServer(t, nil, map[string]string{"/v0/job.download/JOB-1/a.dbn.zst": "complete content"})
			dir := t.TempDir()
			dest := filepath.Join(dir, "a.dbn.zst")
			if err := os.WriteFile(dest, []byte(existing), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := DownloadJobFile(testDownloadKey, desc, dir); err != nil {
				t.Fatalf("DownloadJobFile: %v", err)
			}
			if got := readFile(t, dest); got != "complete content" {
				t.Errorf("content = %q, want the downloaded file", got)
			}
			if n := srv.fileRequests(); n != 1 {
				t.Errorf("file requests = %d, want 1", n)
			}
		})
	}
}

func TestDownloadJobFile_NeverSendsKeyToUntrustedURL(t *testing.T) {
	srv := newFakeBatchServer(t, nil, map[string]string{"/f": "content"})
	for _, u := range []string{
		"http://api.databento.com/f",
		"https://evil.example.com/f",
		"https://databento.com.evil.example/f",
		"https://user:pw@api.databento.com/f",
		"https://api.databento.com:8443/f",
	} {
		desc := fileDesc("a.dbn.zst", "content")
		desc.Urls = map[string]string{"https": u}
		if _, err := DownloadJobFile(testDownloadKey, desc, t.TempDir()); err == nil {
			t.Errorf("URL %q: expected error", u)
		}
	}
	desc := fileDesc("a.dbn.zst", "content")
	desc.Urls = map[string]string{"ftp": "ftp://ftp.databento.com/a"}
	if _, err := DownloadJobFile(testDownloadKey, desc, t.TempDir()); err == nil {
		t.Error("no https URL: expected error")
	}
	if n := len(srv.received()); n != 0 {
		t.Errorf("requests = %d, want 0", n)
	}
}

func TestDownloadJobFile_RedirectDoesNotLeakKey(t *testing.T) {
	desc := fileDesc("a.dbn.zst", "content")
	srv := newFakeBatchServer(t, nil, map[string]string{"/elsewhere": "content"})
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.mu.Lock()
		srv.requests = append(srv.requests, r)
		srv.mu.Unlock()
		if r.Host == "api.databento.com" {
			http.Redirect(w, r, "https://cdn.example.com/elsewhere", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("content"))
	})

	if _, err := DownloadJobFile(testDownloadKey, desc, t.TempDir()); err != nil {
		t.Fatalf("DownloadJobFile: %v", err)
	}
	reqs := srv.received()
	if len(reqs) != 2 {
		t.Fatalf("requests = %d, want 2", len(reqs))
	}
	wantBasicAuth(t, reqs[0])
	if got := reqs[1].Header.Get("Authorization"); got != "" {
		t.Errorf("redirected request to %s carried Authorization %q", reqs[1].Host, got)
	}
}

func TestDownloadJobFile_AcceptsListedEmptyFile(t *testing.T) {
	desc := fileDesc("empty.csv", "")
	newFakeBatchServer(t, nil, map[string]string{"/v0/job.download/JOB-1/empty.csv": ""})
	dir := t.TempDir()
	path, err := DownloadJobFile(testDownloadKey, desc, dir)
	if err != nil {
		t.Fatalf("DownloadJobFile: %v", err)
	}
	if got := readFile(t, path); got != "" {
		t.Errorf("content = %q, want empty", got)
	}
}

func TestDownloadJobFile_RetriesServerErrors(t *testing.T) {
	desc := fileDesc("a.dbn.zst", "content")
	srv := newFakeBatchServer(t, nil, nil)
	failures := 2
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.mu.Lock()
		srv.requests = append(srv.requests, r)
		fail := len(srv.requests) <= failures
		srv.mu.Unlock()
		if fail {
			// A partial body before the error must not end up in the file
			w.Header().Set("Content-Length", "7")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("garbage"))
			return
		}
		_, _ = w.Write([]byte("content"))
	})
	downloadRetryMax = failures

	dir := t.TempDir()
	path, err := DownloadJobFile(testDownloadKey, desc, dir)
	if err != nil {
		t.Fatalf("DownloadJobFile: %v", err)
	}
	if got := readFile(t, path); got != "content" {
		t.Errorf("content = %q", got)
	}
	if n := len(srv.received()); n != failures+1 {
		t.Errorf("requests = %d, want %d", n, failures+1)
	}
	if got := filesUnder(t, dir); !slices.Equal(got, []string{"a.dbn.zst"}) {
		t.Errorf("files = %q, want only a.dbn.zst", got)
	}
}

func TestDownloadJobFile_DoesNotRetryClientErrors(t *testing.T) {
	desc := fileDesc("a.dbn.zst", "content")
	srv := newFakeBatchServer(t, nil, nil) // every file is 404
	downloadRetryMax = 3

	dir := t.TempDir()
	if _, err := DownloadJobFile(testDownloadKey, desc, dir); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("err = %v, want HTTP 404", err)
	}
	if n := len(srv.received()); n != 1 {
		t.Errorf("requests = %d, want 1", n)
	}
	if got := filesUnder(t, dir); len(got) != 0 {
		t.Errorf("files = %q, want none", got)
	}
}

func TestDownloadJobFile_RefusesNonHTTPSRedirect(t *testing.T) {
	desc := fileDesc("a.dbn.zst", "content")
	srv := newFakeBatchServer(t, nil, nil)
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.mu.Lock()
		srv.requests = append(srv.requests, r)
		srv.mu.Unlock()
		http.Redirect(w, r, "http://api.databento.com/elsewhere", http.StatusFound)
	})
	downloadRetryMax = 3

	dir := t.TempDir()
	_, err := DownloadJobFile(testDownloadKey, desc, dir)
	if !errors.Is(err, errRefusedRedirect) {
		t.Fatalf("err = %v, want errRefusedRedirect", err)
	}
	if n := len(srv.received()); n != 1 {
		t.Errorf("requests = %d, want 1 (refused redirects are not retried)", n)
	}
	if got := filesUnder(t, dir); len(got) != 0 {
		t.Errorf("files = %q, want none", got)
	}
}
