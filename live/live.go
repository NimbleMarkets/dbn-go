// Copyright (c) 2024 Neomantra Corp

// TODO: better state machine management (authenticated, started, stopped)

package dbn_live

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

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
)

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
	Encoding             dbn.Encoding // nil mean Encoding_Dbn
	SendTsOut            bool
	VersionUpgradePolicy dbn.VersionUpgradePolicy
	Verbose              bool
}

// SetFromEnv fills in the LiveConfig from environment variables.
// `DATABENTO_API_KEY` holds the DataBento API key.
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
	return nil
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

	c := &LiveClient{
		config:  config,
		gateway: dbn.DatasetToHostname(config.Dataset) + LIVE_HOST_SUFFIX,
		port:    LIVE_API_PORT,
		logger:  config.Logger,
	}

	if c.logger == nil {
		c.logger = slog.Default()
	}

	if c.config.Client == "" {
		c.config.Client = "Go " + DATABENTO_VERSION
	}

	// Connect to server
	hostPort := fmt.Sprintf("%s:%d", c.gateway, c.port)
	if conn, err := net.Dial("tcp", hostPort); err != nil {
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
	return c.lsgVersion
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
// subscriptions. Note there is no unsubscribe method. Subscriptions end
// when the client disconnects with Stop or the LiveClient instance is garbage collected.
func (c *LiveClient) Subscribe(sub SubscriptionRequestMsg) error {
	if len(sub.Symbols) == 0 {
		return errors.New("subscribe request must contain at least one symbol")
	}
	requestBytes := sub.Encode()
	if n, err := c.conn.Write(requestBytes); err != nil {
		return fmt.Errorf("failed to send subscribe request: %v", err)
	} else if n != len(requestBytes) {
		return fmt.Errorf("failed to send subscribe request: wanted %d sent %d", len(requestBytes), n)
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

// Notifies the gateway to start sending messages for all subscriptions.
// This method should only be called once per instance.
func (c *LiveClient) Start() error {
	// TODO: don't start twice, etc
	// Send start_session
	msg := SessionStartMsg{}
	startBytes := msg.Encode()
	if n, err := c.conn.Write(startBytes); err != nil {
		return fmt.Errorf("failed to send start: %v", err)
	} else if n != len(startBytes) {
		return fmt.Errorf("failed to send start: wanted %d sent %d", len(startBytes), n)
	}
	if c.config.Verbose {
		c.logger.Info("[LiveClient.Start] sent start_session")
	}

	// Create a DbnScanner and ensure we get the metadata
	if c.config.Encoding == dbn.Encoding_Json {
		c.jsonScanner = dbn.NewJsonScanner(c.conn)
	} else {
		c.dbnScanner = dbn.NewDbnScanner(c.conn)
		_, err := c.dbnScanner.Metadata()
		if err != nil {
			return fmt.Errorf("failed to get metadata: %v", err)
		}
	}
	if c.config.Verbose {
		c.logger.Info("[LiveClient.Start] read metadata susccessfully")
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
	// Read challege from socket and calcluate reply
	challengeKey, err := c.decodeChallenge()
	if err != nil {
		return "", err
	}

	auth := generateCramReply(apiKey, challengeKey)

	// Write out the auth request
	request := AuthenticationRequestMsg{
		Auth:     auth,
		Dataset:  c.config.Dataset,
		Encoding: c.config.Encoding,
		TsOut:    c.config.SendTsOut,
		Client:   "Go " + DATABENTO_VERSION,
	}
	requestBytes := request.Encode()
	if n, err := c.conn.Write(requestBytes); err != nil {
		return "", fmt.Errorf("failed to send auth request: %v", err)
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

// https://databento.com/docs/api-reference-live/message-flows?historical=http&live=raw
func (c *LiveClient) decodeChallenge() (string, error) {
	// first line is version
	line, err := c.bufReader.ReadBytes('\n')
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
	line, err = c.bufReader.ReadBytes('\n')
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
	line, err := c.bufReader.ReadBytes('\n')
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

	if resp.Success == "0" {
		return "", fmt.Errorf("failed to authenticate: error: %s", resp.Error)
	}

	return resp.SessionID, nil
}

func generateCramReply(apiKey string, challengeKey string) string {
	request := fmt.Sprintf("%s|%s", challengeKey, apiKey)

	hasher := sha256.New()
	hasher.Write([]byte(request))
	checksum := hasher.Sum(nil)

	firstKeyIndex := API_KEY_LENGTH - BUCKET_ID_LENGTH
	bucketID := apiKey[firstKeyIndex:]
	return fmt.Sprintf("%x-%s", checksum, bucketID)
}
