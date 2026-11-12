// Copyright (c) 2024 Neomantra Corp

// TODO: better state machine management (authenticated, started, stopped)

package dbn_live

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/NimbleMarkets/dbn-go"
)

const (
	DATABENTO_VERSION = "0.18.1"

	DATABENTO_API_ENV_KEY    = "DATABENTO_API_KEY"
	DATABENTO_CLIENT_ENV_KEY = "DATABENTO_CLIENT"

	LIVE_HOST_SUFFIX = ".lsg.databento.com"
	LIVE_API_PORT    = 13000

	SYSTEM_MSG_SIZE_V1 = 64
	SYSTEM_MSG_SIZE_V2 = 303

	ERROR_ERR_SIZE_V1 = 64
	ERROR_ERR_SIZE_V2 = 302

	BUCKET_ID_LENGTH = 5

	API_VERSION     = 0
	API_VERSION_STR = "0"
	API_KEY_LENGTH  = 32

	MAX_STR_LENGTH = 24 * 1024

	DIAL_TIMEOUT = 10 * time.Second
	AUTH_TIMEOUT = 30 * time.Second // Default time allowed for the Authenticate handshake
)

// ErrControlMessageTooLong is returned when the gateway sends a control
// message longer than MAX_STR_LENGTH bytes.
var ErrControlMessageTooLong = errors.New("gateway control message too long")

// ErrNotConnected is returned when using a LiveClient whose connection was closed by Stop.
var ErrNotConnected = errors.New("live client is not connected")

type SystemMsgV1 struct {
	Msg [SYSTEM_MSG_SIZE_V1]byte `json:"msg"` // The message from the gateway
}

type ErrorMsgV1 struct {
	Err [ERROR_ERR_SIZE_V1]byte `json:"err"` // The error message
}

type SystemMsgV2 struct {
	Msg  [SYSTEM_MSG_SIZE_V2]byte `json:"msg"`  // The message from the gateway
	Code uint8                    `json:"code"` // Reserved for future use
}

type ErrorMsgV2 struct {
	Err    [ERROR_ERR_SIZE_V2]byte `json:"err"`     // The error message
	Code   uint8                   `json:"code"`    // Reserved for future use
	IsLast uint8                   `json:"is_last"` // Boolean flag indicating whther this is the last in a series of error records.
}

///////////////////////////////////////////////////////////////////////////////

type LiveConfig struct {
	Logger               *slog.Logger
	ApiKey               string
	Dataset              string
	Client               string
	Gateway              string       // Live gateway host; empty means derive from Dataset as "<dataset-host>.lsg.databento.com"
	Port                 uint16       // Live gateway port; 0 means LIVE_API_PORT
	Encoding             dbn.Encoding // nil mean Encoding_Dbn
	SendTsOut            bool
	HeartbeatInterval    time.Duration // Heartbeat interval; 0 means use server default
	AuthTimeout          time.Duration // Time allowed for the whole Authenticate handshake; 0 means AUTH_TIMEOUT
	SlowReaderBehavior   dbn.SlowReaderBehavior
	VersionUpgradePolicy dbn.VersionUpgradePolicy
	Verbose              bool
}

// SetFromEnv fills in the LiveConfig from environment variables.
// `DATABENTO_API_KEY` holds the Databento API key.
// `DATABENTO_CLIENT` holds the Client name.
func (c *LiveConfig) SetFromEnv() error {
	databentoApiKey := os.Getenv(DATABENTO_API_ENV_KEY)
	if databentoApiKey == "" {
		return errors.New("expected environment variable DATABENTO_API_KEY to be set")
	}
	c.ApiKey = databentoApiKey

	if c.Client == "" {
		c.Client = os.Getenv(DATABENTO_CLIENT_ENV_KEY)
	}
	return nil
}

func (c *LiveConfig) validate() error {
	if len(c.ApiKey) == 0 {
		return errors.New("field ApiKey is unset")
	}
	if len(c.ApiKey) != API_KEY_LENGTH {
		return fmt.Errorf("field ApiKey must contain %d characters", API_KEY_LENGTH)
	}
	if len(c.Dataset) == 0 {
		return errors.New("field Dataset is unset")
	}
	if err := validateControlValue("Dataset", c.Dataset); err != nil {
		return err
	}
	if err := validateControlValue("Client", c.Client); err != nil {
		return err
	}
	return nil
}

// resolveGateway returns the effective gateway host and port for config,
// applying the defaults for any unset fields.
func resolveGateway(config LiveConfig) (string, uint16) {
	gateway, port := config.Gateway, config.Port
	if gateway == "" {
		gateway = dbn.DatasetToHostname(config.Dataset) + LIVE_HOST_SUFFIX
	}
	if port == 0 {
		port = LIVE_API_PORT
	}
	return gateway, port
}

///////////////////////////////////////////////////////////////////////////////

// LiveClient interfaces with Databento's real-time and intraday replay
// market data API. This client provides a blocking API for getting the next
// record. Unlike Historical, each instance of LiveClient is associated with a
// particular dataset.
type LiveClient struct {
	config  LiveConfig
	gateway string
	port    uint16

	logger *slog.Logger

	conn      net.Conn
	bufReader *bufio.Reader
	writeMu   sync.Mutex // serializes control messages written to conn

	dbnScanner  *dbn.DbnScanner
	jsonScanner *dbn.JsonScanner

	lsgVersion string
	sessionID  string
}

// NewLiveClient takes a LiveConfig, creates a LiveClient and tries to connect.
// Returns an error if connection fails.
func NewLiveClient(config LiveConfig) (*LiveClient, error) {
	if config.validate() != nil {
		return nil, fmt.Errorf("invalid config: %v", config.validate())
	}

	gateway, port := resolveGateway(config)
	c := &LiveClient{
		config:  config,
		gateway: gateway,
		port:    port,
		logger:  config.Logger,
	}

	if c.logger == nil {
		c.logger = slog.Default()
	}

	if c.config.Client == "" {
		c.config.Client = "Go " + DATABENTO_VERSION
	}

	// Connect to server
	hostPort := net.JoinHostPort(c.gateway, strconv.FormatUint(uint64(c.port), 10))
	if conn, err := net.DialTimeout("tcp", hostPort, DIAL_TIMEOUT); err != nil {
		return nil, err
	} else {
		c.conn = conn
	}
	c.bufReader = bufio.NewReaderSize(c.conn, MAX_STR_LENGTH)
	if c.config.Verbose {
		c.logger.Info("[LiveClient.NewLiveClient] connected", "dataset", config.Dataset, "hostport", hostPort)
	}
	return c, nil
}

// GetConfig returns the LiveConfig used to create the LiveClient.
func (c *LiveClient) GetConfig() LiveConfig {
	return c.config
}

// GetGateway returns the gateway host for connection.
func (c *LiveClient) GetGateway() string {
	return c.gateway
}

// GetPort returns the port for connection.
func (c *LiveClient) GetPort() uint16 {
	return c.port
}

// GetLsgVersion returns the version of the LSG server
func (c *LiveClient) GetLsgVersion() string {
	return c.lsgVersion
}

// GetSessionID returns the session ID
func (c *LiveClient) GetSessionID() string {
	return c.sessionID
}

// GetDbnScanner returns the DbnScanner
// Returns nil if the encoding is JSON
func (c *LiveClient) GetDbnScanner() *dbn.DbnScanner {
	return c.dbnScanner
}

// GetJsonScanner returns a JsonScanner
// Returns nil if the encoding is DBN
func (c *LiveClient) GetJsonScanner() *dbn.JsonScanner {
	return c.jsonScanner
}

///////////////////////////////////////////////////////////////////////////////

// Subscribe adds a new subscription for a set of symbols with a given schema and stype.
// Returns an error if any.
// A single client instance supports multiple
// subscriptions. Use Unsubscribe to remove symbols. Subscriptions end
// when the client disconnects with Stop or the LiveClient instance is garbage collected.
//
// More than MaxSymbolsPerControlMessage symbols are sent as several messages.
func (c *LiveClient) Subscribe(sub SubscriptionRequestMsg) error {
	if c.conn == nil {
		return ErrNotConnected
	}
	if len(sub.Symbols) == 0 {
		return errors.New("subscribe request must contain at least one symbol")
	}
	if err := validateControlValue("Schema", sub.Schema); err != nil {
		return err
	}
	for _, symbol := range sub.Symbols {
		if err := validateControlValue("symbol", symbol); err != nil {
			return err
		}
	}
	err := c.sendChunkedSymbols("subscribe", sub.Symbols, func(symbols []string, partial bool) []byte {
		chunk := sub // copy; the caller's request is left as is
		chunk.Symbols, chunk.Partial = symbols, partial || sub.Partial
		return chunk.Encode()
	})
	if err != nil {
		return err
	}
	if c.config.Verbose {
		symbols := strings.Join(sub.Symbols, ",")
		if c.config.Verbose {
			c.logger.Info("[LiveClient.Subscribe]",
				"schema", sub.Schema, "start", sub.Start,
				"stype_in", sub.StypeIn.String(), "symbols", symbols,
			)
		}
	}

	return nil
}

// Unsubscribe removes a set of symbols with a given schema and stype from the session's subscriptions.
// It can be called before or after Start. See UnsubscriptionRequestMsg for
// which symbols and stypes can be removed.
//
// Returns an error if any. A nil error only means the request was sent.
// The gateway acknowledges it with a SystemMsg with code SystemCode_UnsubscribeAck.
// Symbols that are not subscribed, or cannot be removed with the given stype,
// stay subscribed and are reported in an ErrorMsg; the session continues.
//
// More than MaxSymbolsPerControlMessage symbols are sent as several messages.
func (c *LiveClient) Unsubscribe(unsub UnsubscriptionRequestMsg) error {
	if c.conn == nil {
		return ErrNotConnected
	}
	if len(unsub.Symbols) == 0 {
		return errors.New("unsubscribe request must contain at least one symbol")
	}
	if err := validateControlValue("Schema", unsub.Schema); err != nil {
		return err
	}
	for _, symbol := range unsub.Symbols {
		if err := validateControlValue("symbol", symbol); err != nil {
			return err
		}
	}
	err := c.sendChunkedSymbols("unsubscribe", unsub.Symbols, func(symbols []string, partial bool) []byte {
		chunk := unsub // copy; the caller's request is left as is
		chunk.Symbols, chunk.Partial = symbols, partial || unsub.Partial
		return chunk.Encode()
	})
	if err != nil {
		return err
	}
	if c.config.Verbose {
		c.logger.Info("[LiveClient.Unsubscribe]",
			"schema", unsub.Schema, "stype_in", unsub.StypeIn.String(),
			"symbols", strings.Join(unsub.Symbols, ","),
		)
	}

	return nil
}

// sendChunkedSymbols sends a subscribe or unsubscribe request as one control
// message per MaxSymbolsPerControlMessage symbols. encode builds the message
// for one chunk; partial is true for every chunk but the last (is_last=0).
func (c *LiveClient) sendChunkedSymbols(what string, symbols []string, encode func(symbols []string, partial bool) []byte) error {
	// The gateway joins an is_last=0 message with the next control message it
	// receives, so nothing else may be written until the request is complete.
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	conn := c.conn
	if conn == nil {
		return ErrNotConnected
	}

	total := len(symbols)
	for len(symbols) > 0 {
		chunk := symbols[:min(len(symbols), MaxSymbolsPerControlMessage)]
		sent := total - len(symbols)
		symbols = symbols[len(chunk):]
		msg := encode(chunk, len(symbols) > 0)
		if n, err := writeFull(conn, msg); err != nil {
			if n == len(msg) {
				sent += len(chunk)
			}
			if sent > 0 && sent < total {
				err = &PartialSendError{SymbolsSent: sent, SymbolsTotal: total, Err: err}
			}
			return fmt.Errorf("failed to send %s request: %w", what, err)
		}
	}
	return nil
}

// writeFull writes all of b to w, continuing after a short write, and returns
// the total number of bytes written. A write that makes no progress without an
// error returns io.ErrShortWrite.
func writeFull(w io.Writer, b []byte) (int, error) {
	written := 0
	for len(b) > 0 {
		n, err := w.Write(b)
		if n > 0 {
			n = min(n, len(b))
			written += n
			b = b[n:]
		}
		if err != nil {
			return written, err
		}
		if n <= 0 {
			return written, io.ErrShortWrite
		}
	}
	return written, nil
}

// Notifies the gateway to start sending messages for all subscriptions.
// This method should only be called once per instance.
func (c *LiveClient) Start() error {
	// TODO: don't start twice, etc
	if c.conn == nil {
		return ErrNotConnected
	}
	// Send start_session
	msg := SessionStartMsg{}
	c.writeMu.Lock() // never between the messages of a split request
	_, err := writeFull(c.conn, msg.Encode())
	c.writeMu.Unlock()
	if err != nil {
		return fmt.Errorf("failed to send start: %w", err)
	}
	if c.config.Verbose {
		c.logger.Info("[LiveClient.Start] sent start_session")
	}

	if c.bufReader == nil {
		c.bufReader = bufio.NewReaderSize(c.conn, MAX_STR_LENGTH)
	}

	// Create a DbnScanner and ensure we get the metadata
	if c.config.Encoding == dbn.Encoding_Json {
		c.jsonScanner = dbn.NewJsonScanner(c.bufReader)
	} else {
		c.dbnScanner = dbn.NewDbnScanner(c.bufReader)
		metadata, err := c.dbnScanner.Metadata()
		if err != nil {
			return fmt.Errorf("failed to get metadata: %v", err)
		}
		if c.config.Verbose {
			c.logger.Info("[LiveClient.Start] read metadata successfully",
				"version_num", metadata.VersionNum, "len_mappings", len(metadata.Mappings))
		}
	}

	return nil
}

// Stops the session with the gateway. Once stopped, the session cannot be restarted.
func (c *LiveClient) Stop() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		if err != nil {
			if c.config.Verbose {
				c.logger.Error("[LiveClient.Stop] error closing connection", "error", err.Error())
			}
			return err
		} else {
			if c.config.Verbose {
				c.logger.Info("[LiveClient.Stop] closed connection")
			}
		}
	}
	if c.config.Verbose {
		c.logger.Info("[LiveClient.Stop] stopped")
	}
	return nil
}

// Authenticate performs read/write with the server to authenticate.
// Returns a sessionID or an error.
func (c *LiveClient) Authenticate(apiKey string) (string, error) {
	if c.conn == nil {
		return "", ErrNotConnected
	}
	if len(apiKey) != API_KEY_LENGTH {
		return "", fmt.Errorf("apiKey must contain %d characters", API_KEY_LENGTH)
	}
	for _, field := range []struct{ name, value string }{
		{"apiKey", apiKey}, {"Dataset", c.config.Dataset}, {"Client", c.config.Client},
	} {
		if err := validateControlValue(field.name, field.value); err != nil {
			return "", err
		}
	}

	// Bound the whole handshake, so a gateway that stops responding cannot
	// hang the caller. The deadline is lifted before the data stream starts.
	authTimeout := c.config.AuthTimeout
	if authTimeout <= 0 {
		authTimeout = AUTH_TIMEOUT
	}
	conn := c.conn
	if err := conn.SetDeadline(time.Now().Add(authTimeout)); err != nil {
		return "", fmt.Errorf("failed to set handshake deadline: %w", err)
	}
	defer conn.SetDeadline(time.Time{})

	// Read challege from socket and calcluate reply
	challengeKey, err := c.decodeChallenge()
	if err != nil {
		return "", err
	}

	auth := generateCramReply(apiKey, challengeKey)

	// Write out the auth request
	request := AuthenticationRequestMsg{
		Auth:               auth,
		Dataset:            c.config.Dataset,
		Encoding:           c.config.Encoding,
		TsOut:              c.config.SendTsOut,
		Client:             c.config.Client,
		HeartbeatIntervalS: uint32(c.config.HeartbeatInterval.Seconds()),
		SlowReaderBehavior: c.config.SlowReaderBehavior,
	}
	requestBytes := request.Encode()
	if n, err := c.conn.Write(requestBytes); err != nil {
		return "", fmt.Errorf("failed to send auth request: %w", err)
	} else if n != len(requestBytes) {
		return "", fmt.Errorf("failed to send auth request: wanted %d sent %d", len(requestBytes), n)
	}

	// Read the response
	sessionID, err := c.decodeAuthResponse()
	if err != nil {
		return "", err
	}
	c.sessionID = sessionID
	if c.config.Verbose {
		c.logger.Info("[LiveClient.Authenticate] Successfully authenticated", "session_id", sessionID)
	}
	return sessionID, nil
}

// readControlLine reads one newline-terminated control message from the
// gateway. The line may be at most the reader's buffer size (MAX_STR_LENGTH);
// a longer one returns ErrControlMessageTooLong rather than being buffered.
// The returned slice is only valid until the next read.
func (c *LiveClient) readControlLine() ([]byte, error) {
	line, err := c.bufReader.ReadSlice('\n')
	if errors.Is(err, bufio.ErrBufferFull) {
		return nil, ErrControlMessageTooLong
	}
	return line, err
}

// https://databento.com/docs/api-reference-live/message-flows?historical=http&live=raw
func (c *LiveClient) decodeChallenge() (string, error) {
	// first line is version
	line, err := c.readControlLine()
	if err != nil {
		return "", fmt.Errorf("failed to read Greeting: %w", err)
	}
	greeting := NewGreetingMsgFromBytes(line)
	if greeting == nil {
		return "", errors.New("failed to parse greeting")
	}
	c.lsgVersion = greeting.LsgVersion
	if c.config.Verbose {
		c.logger.Info("[LiveClient.decodeChallenge]", "version", greeting.LsgVersion)
	}

	// next is challenge request
	line, err = c.readControlLine()
	if err != nil {
		return "", fmt.Errorf("failed to read Challenge: %w", err)
	}
	challenge := NewChallengeRequestMsgFromBytes(line)
	if challenge == nil {
		return "", errors.New("failed to parse challenge")
	}
	if c.config.Verbose {
		c.logger.Info("[LiveClient.decodeChallenge]", "cram", challenge.Cram)
	}

	return challenge.Cram, nil
}

func (c *LiveClient) decodeAuthResponse() (string, error) {
	line, err := c.readControlLine()
	if err != nil {
		return "", fmt.Errorf("failed to read AuthResponse: %w", err)
	}
	resp := NewAuthenticationResponseMsgFromBytes(line)
	if resp == nil {
		return "", errors.New("failed to parse AuthResponse")
	}

	if c.config.Verbose {
		c.logger.Info("[LiveClient.decodeAuthResponse", "success", resp.Success, "error", resp.Error, "session_id", resp.SessionID)
	}

	if resp.Success != "1" {
		return "", fmt.Errorf("failed to authenticate: success: %q error: %q", resp.Success, resp.Error)
	}

	return resp.SessionID, nil
}

// GenerateCramReply returns the `auth` value of an authentication request:
// the hex SHA-256 of "<cram>|<apiKey>", a dash, and the last BUCKET_ID_LENGTH
// characters of apiKey (all of apiKey, if shorter).
// Gateways, such as a replay server, can use it to verify a client's reply.
func GenerateCramReply(apiKey string, challengeKey string) string {
	return generateCramReply(apiKey, challengeKey)
}

func generateCramReply(apiKey string, challengeKey string) string {
	request := fmt.Sprintf("%s|%s", challengeKey, apiKey)

	hasher := sha256.New()
	hasher.Write([]byte(request))
	checksum := hasher.Sum(nil)

	bucketID := apiKey[max(0, len(apiKey)-BUCKET_ID_LENGTH):]
	return fmt.Sprintf("%x-%s", checksum, bucketID)
}

// PartialSendError is returned by Subscribe and Unsubscribe when a request
// that was split across several control messages failed after some of them
// were sent. The gateway has received SymbolsSent symbols in messages marked
// is_last=0 and is waiting for the rest, so the connection should not be
// reused for further requests.
type PartialSendError struct {
	SymbolsSent  int   // Symbols in the messages that were fully sent
	SymbolsTotal int   // Symbols in the whole request
	Err          error // The underlying write error
}

func (e *PartialSendError) Error() string {
	return fmt.Sprintf("%d of %d symbols were already sent as partial messages: %v", e.SymbolsSent, e.SymbolsTotal, e.Err)
}

func (e *PartialSendError) Unwrap() error { return e.Err }
