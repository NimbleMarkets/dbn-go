// Copyright (c) 2026 Neomantra Corp

package dbn_hist

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Databento Batch download:
//  https://databento.com/docs/api-reference-historical/batch/batch-download

// downloadRetryMax is how many times a failed file download is retried, and
// downloadRetryWait how long to wait before the first retry (it then doubles).
var (
	downloadRetryMax  = 10
	downloadRetryWait = time.Second
)

// errRefusedRedirect is returned when a download redirect is not followed.
var errRefusedRedirect = errors.New("refusing redirect")

// DownloadJobParams are the parameters for DownloadJob.
type DownloadJobParams struct {
	JobID              string // The batch job ID. Required.
	OutputDir          string // The directory to download into; files go to OutputDir/JobID/. Defaults to the current directory.
	FilenameToDownload string // If set, only this file of the job is downloaded. Otherwise all of its files are.
}

// DownloadJob downloads the files of the batch job params.JobID into
// params.OutputDir/params.JobID/, like batch.download in Databento's clients.
// If params.FilenameToDownload is set, only that file is downloaded, and it
// is an error if the job has no such file.
//
// Each file is downloaded with DownloadJobFile, so a file already present
// with the listed size and hash is not downloaded again.
// Returns the paths of the job's files, in listing order. On error, the
// paths of the files completed so far are returned with it.
//
// # Errors
// This function returns an error when it fails to communicate with the Databento API,
// the API indicates there's an issue with the request, or a file fails verification.
// Returns an error wrapping JobExpiredError if the job has expired.
func DownloadJob(apiKey string, params DownloadJobParams) ([]string, error) {
	if err := checkDownloadFilename(params.JobID); err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}
	outputDir := params.OutputDir
	if outputDir == "" {
		outputDir = "."
	}

	files, err := ListFiles(apiKey, params.JobID)
	if err != nil {
		return nil, fmt.Errorf("failed to list files for job %s: %w", params.JobID, err)
	}

	if params.FilenameToDownload != "" {
		var selected []BatchFileDesc
		for _, file := range files {
			if file.Filename == params.FilenameToDownload {
				selected = append(selected, file)
				break
			}
		}
		if len(selected) == 0 {
			return nil, fmt.Errorf("%q is not a file of batch job %s", params.FilenameToDownload, params.JobID)
		}
		files = selected
	}
	// Check every listed name before downloading anything
	for _, file := range files {
		if err := checkDownloadFilename(file.Filename); err != nil {
			return nil, err
		}
	}

	jobDir := filepath.Join(outputDir, params.JobID)
	paths := make([]string, 0, len(files))
	for _, file := range files {
		path, err := DownloadJobFile(apiKey, file, jobDir)
		if err != nil {
			return paths, fmt.Errorf("failed to download %s: %w", file.Filename, err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

// DownloadJobFile downloads one batch file, as listed by ListFiles, to
// destDir/file.Filename and returns that path. destDir is created if needed.
//
// The listing comes from the network, so none of it is trusted: the file name
// must be a plain name (no directory part), the API key is only sent to a
// Databento HTTPS URL, and the file is only kept if its size, and its SHA-256
// hash when one is listed, match the listing. The download is written to a
// temporary file in destDir that is renamed into place once verified, so a
// failed download never leaves a partial or unverified file at the path.
//
// If a file with the listed size and hash is already at the path, it is kept
// and nothing is downloaded. Any other existing file is replaced.
func DownloadJobFile(apiKey string, file BatchFileDesc, destDir string) (string, error) {
	// Check everything from the listing before touching the disk or network
	if err := checkDownloadFilename(file.Filename); err != nil {
		return "", err
	}
	wantHash, err := parseSHA256Hash(file.Hash)
	if err != nil {
		return "", err
	}
	if file.Size >= math.MaxInt64 {
		return "", fmt.Errorf("refusing download with listed size %d", file.Size)
	}
	httpsURL := file.Urls["https"]
	if httpsURL == "" {
		return "", fmt.Errorf("no https URL listed for %q", file.Filename)
	}
	downloadURL, err := url.Parse(httpsURL)
	if err != nil {
		return "", fmt.Errorf("invalid download URL: %w", err)
	}
	if err := databentoDownloadURLAllowed(downloadURL); err != nil {
		return "", fmt.Errorf("refusing download URL: %w", err)
	}

	if destDir == "" {
		destDir = "."
	}
	destPath := filepath.Join(destDir, file.Filename)
	if filepath.Dir(destPath) != filepath.Clean(destDir) {
		return "", fmt.Errorf("refusing unsafe file name %q", file.Filename)
	}

	if complete, err := fileMatches(destPath, file.Size, wantHash); err != nil {
		return "", err
	} else if complete {
		return destPath, nil
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	if err := downloadVerified(apiKey, downloadURL, destPath, file.Size, wantHash); err != nil {
		return "", err
	}
	return destPath, nil
}

// downloadVerified downloads u to a temporary file next to destPath, checks
// its size and its hash (if wantHash is non-nil), then renames it to destPath.
// Network errors and HTTP 429 and 5xx responses are retried, with backoff, up
// to downloadRetryMax times.
func downloadVerified(apiKey string, u *url.URL, destPath string, wantSize uint64, wantHash []byte) (err error) {
	tmpFile, err := os.CreateTemp(filepath.Dir(destPath), "."+filepath.Base(destPath)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	wait := downloadRetryWait
	for attempt := 0; ; attempt++ {
		var retryable bool
		retryable, err = downloadAttempt(apiKey, u, tmpFile, wantSize, wantHash)
		if err == nil {
			break
		}
		if !retryable || attempt >= downloadRetryMax {
			if attempt > 0 {
				err = fmt.Errorf("giving up after %d attempts: %w", attempt+1, err)
			}
			return err
		}
		time.Sleep(wait)
		wait = min(2*wait, 30*time.Second)
	}

	if err := tmpFile.Chmod(0o644); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, destPath)
}

// downloadAttempt makes one request for u and writes the body to tmpFile,
// replacing anything already in it, then checks its size and hash.
// It reports whether a failure is worth retrying.
func downloadAttempt(apiKey string, u *url.URL, tmpFile *os.File, wantSize uint64, wantHash []byte) (retryable bool, err error) {
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return false, err
	}
	if err := tmpFile.Truncate(0); err != nil {
		return false, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return false, err
	}
	auth := base64.StdEncoding.EncodeToString([]byte(apiKey + ":"))
	req.Header.Add("Authorization", "Basic "+auth)

	client := *httpClient
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("%w: stopped after 10 redirects", errRefusedRedirect)
		}
		if req.URL.Scheme != "https" {
			return fmt.Errorf("%w to non-https URL %q", errRefusedRedirect, req.URL.Redacted())
		}
		if databentoDownloadURLAllowed(req.URL) != nil {
			req.Header.Del("Authorization") // the key stays with Databento
		}
		return nil
	}

	resp, err := client.Do(req)
	if err != nil {
		return !errors.Is(err, errRefusedRedirect), err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		retryable = resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		return retryable, fmt.Errorf("HTTP %s", resp.Status)
	}

	// Never write more than the listing declared (one extra byte detects an
	// overrun). This holds for every listed size, including zero.
	body := io.LimitReader(resp.Body, int64(wantSize)+1)
	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(tmpFile, hasher), body)
	if err != nil {
		return true, err
	}
	if uint64(written) > wantSize {
		return false, fmt.Errorf("download is larger than the %d bytes listed", wantSize)
	} else if uint64(written) < wantSize {
		return false, fmt.Errorf("download is %d bytes, but %d were listed", written, wantSize)
	}
	if wantHash != nil && !bytes.Equal(hasher.Sum(nil), wantHash) {
		return false, errors.New("download does not match the listed SHA-256 hash")
	}
	return false, nil
}

// fileMatches reports whether path is a regular file with the given size and,
// if wantHash is non-nil, SHA-256 hash. A missing file does not match.
func fileMatches(path string, size uint64, wantHash []byte) (bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() || uint64(info.Size()) != size {
		return false, nil
	}
	if wantHash == nil {
		return true, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return false, err
	}
	return bytes.Equal(hasher.Sum(nil), wantHash), nil
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
// An empty hash returns nil, meaning there is nothing to verify; any other
// form is an error.
func parseSHA256Hash(hash string) ([]byte, error) {
	if hash == "" {
		return nil, nil
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
