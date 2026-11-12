// Copyright (c) 2026 Neomantra Corp

// Package dbn_live_replay provides a local Databento Live gateway that replays
// DBN files, so Live clients can be developed and tested without a real gateway.
//
// It speaks the Live line protocol as implemented by dbn_live.LiveClient:
// greeting, CRAM challenge, authentication, subscriptions, start_session, then
// a DBN stream. Point a client at it with LiveConfig.Gateway and LiveConfig.Port
// (or dbn-go-live --gateway/--port).
//
// Limitations, by design:
//   - Each session replays exactly one file. Among the files whose metadata
//     dataset matches the requested dataset, the first whose schema matches a
//     subscribed schema is chosen (a mixed-schema file matches any schema);
//     if none matches, the first file of that dataset is replayed.
//   - The file's metadata is sent byte-for-byte, followed by all of its records,
//     regardless of subscribed symbols, stype_in, start, or snapshot.
//   - Only DBN encoding, no compression, and ts_out=0 are supported; other
//     requests fail authentication with an error.
//   - After the last record the server half-closes the connection and waits
//     briefly for the client to disconnect. No heartbeats are sent.
package dbn_live_replay

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	"github.com/NimbleMarkets/dbn-go/internal/version"
	dbn_live "github.com/NimbleMarkets/dbn-go/live"
)

// ErrServerClosed is returned by Listen and Serve after Close has been called.
var ErrServerClosed = errors.New("replay server closed")

const (
	cramLength     = 32
	lingerTimeout  = 5 * time.Second
	maxMetadataLen = 64 << 20 // sanity bound for a metadata header
)

// ServerConfig configures a replay Server.
type ServerConfig struct {
	// Files are the DBN files (.dbn, or zstd-compressed .dbn.zst) to replay.
	// At least one is required. Each session replays one of them; see the package docs.
	Files []string
	// ApiKey, if non-empty, must be the 32-character key clients authenticate with.
	// If empty, any key is accepted.
	ApiKey string
	// Speed paces records by their ts_event: 1 replays in real time, 2 at double
	// speed, and so on. 0 (the default) sends records as fast as possible.
	Speed float64
	// Logger receives connection and session logs. nil means slog.Default().
	Logger *slog.Logger
}

// replayFile is a validated replay file and the metadata fields used to select it.
type replayFile struct {
	path    string
	dataset string
	schema  dbn.Schema
}

// Server is a Databento Live gateway that replays DBN files.
// Create one with NewServer, start it with Listen or Serve, and stop it with Close.
type Server struct {
	config ServerConfig
	logger *slog.Logger
	files  []replayFile

	mu       sync.Mutex
	listener net.Listener
	conns    map[net.Conn]struct{}
	closed   bool
	done     chan struct{} // closed by Close
	wg       sync.WaitGroup

	sessionCounter atomic.Uint64
}

// NewServer validates cfg and returns a Server. Each file must exist and begin
// with valid DBN metadata.
func NewServer(cfg ServerConfig) (*Server, error) {
	if len(cfg.Files) == 0 {
		return nil, errors.New("at least one replay file is required")
	}
	if cfg.ApiKey != "" && len(cfg.ApiKey) != dbn_live.API_KEY_LENGTH {
		return nil, fmt.Errorf("ApiKey must contain %d characters", dbn_live.API_KEY_LENGTH)
	}
	if cfg.Speed < 0 {
		return nil, errors.New("speed must not be negative")
	}
	s := &Server{
		config: cfg,
		logger: cfg.Logger,
		conns:  make(map[net.Conn]struct{}),
		done:   make(chan struct{}),
	}
	if s.logger == nil {
		s.logger = slog.Default()
	}
	for _, path := range cfg.Files {
		file, err := inspectFile(path)
		if err != nil {
			return nil, err
		}
		s.files = append(s.files, file)
	}
	return s, nil
}

// inspectFile reads the metadata of a replay file.
func inspectFile(path string) (replayFile, error) {
	if path == "-" {
		return replayFile{}, errors.New("replay files must be paths, not stdin")
	}
	f, err := openDbnFile(path)
	if err != nil {
		return replayFile{}, err
	}
	defer f.Close()
	if _, err := f.readMetadata(); err != nil {
		return replayFile{}, fmt.Errorf("failed to read metadata of %s: %w", path, err)
	}
	return replayFile{path: path, dataset: f.metadata.Dataset, schema: f.metadata.Schema}, nil
}

///////////////////////////////////////////////////////////////////////////////

// Listen binds addr (for example "127.0.0.1:0") and serves on it in the background.
// Use Addr to learn the bound address and Close to stop.
func (s *Server) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// The accept goroutine joins the WaitGroup under the lock, so Close waits for it.
	if err := s.setListener(ln, true); err != nil {
		ln.Close()
		return err
	}
	go func() {
		defer s.wg.Done()
		if err := s.acceptLoop(ln); err != nil {
			s.logger.Error("replay server stopped accepting", "error", err)
		}
	}()
	return nil
}

// Serve accepts connections on ln until Close is called, then returns nil.
// It returns an error if accepting fails for another reason.
func (s *Server) Serve(ln net.Listener) error {
	if err := s.setListener(ln, false); err != nil {
		return err
	}
	return s.acceptLoop(ln)
}

// Addr returns the listening address, or nil if the server is not listening.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// Close stops listening, ends all sessions, and waits for their goroutines to exit.
// It is safe to call more than once.
func (s *Server) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		s.wg.Wait()
		return nil
	}
	s.closed = true
	close(s.done)
	var err error
	if s.listener != nil {
		err = s.listener.Close()
	}
	for conn := range s.conns {
		conn.Close()
	}
	s.mu.Unlock()
	s.wg.Wait()
	if errors.Is(err, net.ErrClosed) {
		err = nil
	}
	return err
}

// setListener records ln as the server's listener. If addToWaitGroup, it also
// adds one to the WaitGroup for the caller's accept goroutine.
func (s *Server) setListener(ln net.Listener, addToWaitGroup bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrServerClosed
	}
	if s.listener != nil {
		return errors.New("replay server is already listening")
	}
	s.listener = ln
	if addToWaitGroup {
		s.wg.Add(1)
	}
	return nil
}

func (s *Server) acceptLoop(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-s.done:
				return nil
			default:
				return err
			}
		}
		if !s.track(conn) {
			conn.Close()
			continue
		}
		go func() {
			defer s.wg.Done()
			defer s.untrack(conn)
			s.handle(conn)
		}()
	}
}

// track registers conn and adds it to the WaitGroup, unless the server is closed.
func (s *Server) track(conn net.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.conns[conn] = struct{}{}
	s.wg.Add(1)
	return true
}

func (s *Server) untrack(conn net.Conn) {
	conn.Close()
	s.mu.Lock()
	delete(s.conns, conn)
	s.mu.Unlock()
}

///////////////////////////////////////////////////////////////////////////////

// handle runs one client session.
func (s *Server) handle(conn net.Conn) {
	logger := s.logger.With("remote", conn.RemoteAddr().String())
	logger.Info("client connected")
	defer logger.Info("client disconnected")

	cram, err := newCram()
	if err != nil {
		logger.Error("failed to generate challenge", "error", err)
		return
	}
	greeting := dbn_live.GreetingMsg{LsgVersion: version.Get()}
	challenge := dbn_live.ChallengeRequestMsg{Cram: cram}
	if _, err := conn.Write(append(greeting.Encode(), challenge.Encode()...)); err != nil {
		logger.Debug("failed to send greeting", "error", err)
		return
	}

	reader := bufio.NewReaderSize(conn, dbn_live.MAX_STR_LENGTH)
	line, err := readLine(reader)
	if err != nil {
		logger.Debug("failed to read authentication request", "error", err)
		return
	}
	candidates, authErr := s.authenticate(line, cram)
	if authErr != nil {
		logger.Info("authentication failed", "error", authErr)
		resp := dbn_live.AuthenticationResponseMsg{Success: "0", Error: sanitize(authErr.Error())}
		_, _ = conn.Write(resp.Encode())
		return
	}
	sessionID := strconv.FormatUint(s.sessionCounter.Add(1), 10)
	logger = logger.With("session_id", sessionID)
	resp := dbn_live.AuthenticationResponseMsg{Success: "1", SessionID: sessionID}
	if _, err := conn.Write(resp.Encode()); err != nil {
		logger.Debug("failed to send authentication response", "error", err)
		return
	}

	// Collect subscriptions until the client starts the session.
	var schemas []string
	for {
		line, err := readLine(reader)
		if err != nil {
			logger.Debug("client left before starting the session", "error", err)
			return
		}
		if _, err := dbn_live.NewSessionStartMsgFromBytes(line); err == nil {
			break
		}
		sub, err := dbn_live.NewSubscriptionRequestMsgFromBytes(line)
		if err != nil {
			logger.Warn("ignoring unrecognized control message", "line", strings.TrimSpace(string(line)), "error", err)
			continue
		}
		logger.Info("subscription", "schema", sub.Schema, "stype_in", sub.StypeIn.String(), "symbols", strings.Join(sub.Symbols, ","))
		schemas = append(schemas, sub.Schema)
	}

	file := selectFile(candidates, schemas)
	logger.Info("starting replay", "file", file.path, "schema", file.schema.String())

	// Watch for the client disconnecting so that pacing waits end promptly.
	clientGone := make(chan struct{})
	go func() {
		defer close(clientGone)
		_, _ = io.Copy(io.Discard, reader)
	}()
	defer func() {
		conn.Close()
		<-clientGone
	}()

	count, err := s.stream(conn, file.path, clientGone)
	if err != nil {
		logger.Info("replay ended early", "records", count, "error", err)
		return
	}
	logger.Info("replay finished", "records", count)

	// Half-close so the client reads EOF after the last record, then give it
	// a moment to hang up before closing.
	if cw, ok := conn.(interface{ CloseWrite() error }); ok {
		if cw.CloseWrite() == nil {
			select {
			case <-clientGone:
			case <-s.done:
			case <-time.After(lingerTimeout):
			}
		}
	}
}

// authenticate validates an authentication request line and returns the
// replay files for its dataset.
func (s *Server) authenticate(line []byte, cram string) ([]replayFile, error) {
	req, err := dbn_live.NewAuthenticationRequestMsgFromBytes(line)
	if err != nil {
		return nil, fmt.Errorf("invalid authentication request: %w", err)
	}
	if s.config.ApiKey != "" {
		expected := dbn_live.GenerateCramReply(s.config.ApiKey, cram)
		if subtle.ConstantTimeCompare([]byte(expected), []byte(req.Auth)) != 1 {
			return nil, errors.New("authentication failed: invalid API key")
		}
	}
	if req.Encoding != dbn.Encoding_Dbn {
		return nil, fmt.Errorf("unsupported encoding %q: replay server only supports dbn", req.Encoding.String())
	}
	if req.Compression != dbn.Compress_None {
		return nil, fmt.Errorf("unsupported compression %q: replay server only supports none", req.Compression.String())
	}
	if req.TsOut {
		return nil, errors.New("ts_out is not supported by the replay server")
	}
	var candidates []replayFile
	for _, f := range s.files {
		if strings.EqualFold(f.dataset, req.Dataset) {
			candidates = append(candidates, f)
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no replay data for dataset %s", req.Dataset)
	}
	return candidates, nil
}

// selectFile returns the first candidate whose schema matches a subscribed
// schema, or the first candidate if none matches.
func selectFile(candidates []replayFile, schemas []string) replayFile {
	for _, f := range candidates {
		if f.schema == dbn.Schema_Mixed {
			return f
		}
		for _, str := range schemas {
			if schema, err := dbn.SchemaFromString(str); err == nil && schema == f.schema {
				return f
			}
		}
	}
	return candidates[0]
}

// stream writes the file's metadata and records to w, pacing them by ts_event
// if configured. It stops early if clientGone or the server is closed.
// Returns the number of records written.
func (s *Server) stream(w io.Writer, path string, clientGone <-chan struct{}) (int, error) {
	f, err := openDbnFile(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	rawMetadata, err := f.readMetadata()
	if err != nil {
		return 0, err
	}

	bw := bufio.NewWriter(w)
	if _, err := bw.Write(rawMetadata); err != nil {
		return 0, err
	}

	scanner := dbn.NewDbnScanner(io.MultiReader(bytes.NewReader(rawMetadata), f.reader))
	count := 0
	var lastTs uint64
	haveLastTs := false
	for scanner.Next() {
		if s.config.Speed > 0 {
			header, err := scanner.GetLastHeader()
			if err != nil {
				return count, err
			}
			if ts := header.TsEvent; ts != dbn.UNDEF_TIMESTAMP {
				if haveLastTs && ts > lastTs {
					if err := bw.Flush(); err != nil {
						return count, err
					}
					delay := time.Duration(float64(ts-lastTs) / s.config.Speed)
					if !s.wait(delay, clientGone) {
						return count, errors.New("session ended while pacing")
					}
				}
				if !haveLastTs || ts > lastTs {
					lastTs, haveLastTs = ts, true
				}
			}
		}
		if _, err := bw.Write(scanner.GetLastRecord()[:scanner.GetLastSize()]); err != nil {
			return count, err
		}
		count++
	}
	if err := scanner.Error(); err != nil && err != io.EOF {
		return count, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return count, bw.Flush()
}

// wait sleeps for d, returning false if the client leaves or the server closes first.
func (s *Server) wait(d time.Duration, clientGone <-chan struct{}) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-clientGone:
		return false
	case <-s.done:
		return false
	}
}

///////////////////////////////////////////////////////////////////////////////

// dbnFile is an open, possibly zstd-compressed, DBN file.
type dbnFile struct {
	reader   *bufio.Reader
	closers  []func()
	metadata *dbn.Metadata
}

func openDbnFile(path string) (*dbnFile, error) {
	reader, closer, err := dbn.MakeCompressedReader(path, false)
	if err != nil {
		return nil, err
	}
	f := &dbnFile{reader: bufio.NewReader(reader)}
	// A zstd decoder must be closed to release its goroutines.
	if decoder, ok := reader.(interface{ Close() }); ok {
		f.closers = append(f.closers, decoder.Close)
	}
	if closer != nil {
		f.closers = append(f.closers, func() { _ = closer.Close() })
	}
	return f, nil
}

func (f *dbnFile) Close() {
	for _, closer := range f.closers {
		closer()
	}
}

// readMetadata reads the raw metadata header, parses it into f.metadata, and
// returns its bytes unchanged.
func (f *dbnFile) readMetadata() ([]byte, error) {
	prefix := make([]byte, 8)
	if _, err := io.ReadFull(f.reader, prefix); err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(prefix, []byte("DBN")) {
		return nil, dbn.ErrInvalidDBNFile
	}
	length := binary.LittleEndian.Uint32(prefix[4:])
	if length > maxMetadataLen {
		return nil, fmt.Errorf("metadata length %d is too large", length)
	}
	raw := make([]byte, 8+int(length))
	copy(raw, prefix)
	if _, err := io.ReadFull(f.reader, raw[8:]); err != nil {
		return nil, err
	}
	metadata, err := dbn.ReadMetadata(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	f.metadata = metadata
	return raw, nil
}

///////////////////////////////////////////////////////////////////////////////

// readLine reads one control message line, failing on lines longer than the reader's buffer.
func readLine(reader *bufio.Reader) ([]byte, error) {
	line, err := reader.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	return bytes.Clone(line), nil
}

// newCram returns a random alphanumeric challenge.
func newCram() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, cramLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}

// sanitize makes text safe to use as a control message value.
func sanitize(text string) string {
	return strings.NewReplacer("|", "/", "\n", " ", "\r", " ").Replace(text)
}
