// Copyright (c) 2024 Neomantra Corp

package dbn_live

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	"github.com/NimbleMarkets/dbn-go/internal/version"
)

// validateControlValue returns an error if value contains a character that
// would delimit a field ('|') or a message ('\r', '\n') in the gateway line protocol.
func validateControlValue(name string, value string) error {
	if strings.ContainsAny(value, "|\r\n") {
		return fmt.Errorf("field %s must not contain '|' or newline", name)
	}
	return nil
}

// Returns a string key/value map from a Databento control message
// The format is: "k1=v1|k2=v2|k3=v3\n"
func parseControlMessage(b []byte) map[string]string {
	m := make(map[string]string)
	if len(b) == 0 {
		return m
	}
	if b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	if len(b) == 0 {
		return m
	}
	kvs := bytes.Split(b, []byte{'|'})
	for _, kv := range kvs {
		equals := bytes.IndexByte(kv, '=')
		if equals == -1 {
			continue
		}
		k := string(kv[:equals])
		v := string(kv[equals+1:])
		m[k] = v
	}
	return m
}

// GreetingMsg is a greeting message sent by the gateway upon connection.
type GreetingMsg struct {
	LsgVersion string // key: lsg_version
}

// NewGreetingMsgFromBytes parses a control message and returns a GreetingMsg
// Returns nil if required fields are missing.
func NewGreetingMsgFromBytes(b []byte) *GreetingMsg {
	m := parseControlMessage(b)
	version, ok := m["lsg_version"]
	if !ok {
		return nil // required
	}
	return &GreetingMsg{LsgVersion: version}
}

// Encode converts GreetingMsg to its line protocol representation.
func (m *GreetingMsg) Encode() []byte {
	return fmt.Appendf(nil, "lsg_version=%s\n", m.LsgVersion)
}

// ChallengeRequestMsg is sent by the gateway upon connection.
type ChallengeRequestMsg struct {
	Cram string // key: cram
}

// NewChallengeRequestMsgFromBytes parses a control message and returns a ChallengeRequestMsg
// Returns nil if required fields are missing.
func NewChallengeRequestMsgFromBytes(b []byte) *ChallengeRequestMsg {
	m := parseControlMessage(b)
	cram, ok := m["cram"]
	if !ok {
		return nil // required
	}
	return &ChallengeRequestMsg{Cram: cram}
}

// Encode converts ChallengeRequestMsg to its line protocol representation.
func (m *ChallengeRequestMsg) Encode() []byte {
	return fmt.Appendf(nil, "cram=%s\n", m.Cram)
}

// AuthenticationResponseMsg is an authentication response is sent by the gateway after a valid
// authentication request is sent to the gateway.
// https://databento.com/docs/api-reference-live/gateway-control-messages/authentication-response
type AuthenticationResponseMsg struct {
	Success   string // key: success
	Error     string // key: error
	SessionID string // key: session_id
}

// NewAuthenticationResponseMsgFromBytes parses a control message and returns a AuthenticationResponseMsg
// Returns nil if required fields are missing.
func NewAuthenticationResponseMsgFromBytes(b []byte) *AuthenticationResponseMsg {
	m := parseControlMessage(b)
	success, ok := m["success"]
	if !ok {
		return nil // required
	}
	return &AuthenticationResponseMsg{
		Success:   success,
		Error:     m["error"],
		SessionID: m["session_id"],
	}
}

// Encode converts AuthenticationResponseMsg to its line protocol representation.
// Empty Error and SessionID fields are omitted.
func (m *AuthenticationResponseMsg) Encode() []byte {
	b := fmt.Appendf(nil, "success=%s", m.Success)
	if m.Error != "" {
		b = fmt.Appendf(b, "|error=%s", m.Error)
	}
	if m.SessionID != "" {
		b = fmt.Appendf(b, "|session_id=%s", m.SessionID)
	}
	return append(b, '\n')
}

// AuthenticationRequestMsg is an authentication request is sent to the gateway after a challenge response is received.
// This is required to authenticate a user.
// https://databento.com/docs/api-reference-live/client-control-messages/authentication-request
type AuthenticationRequestMsg struct {
	Auth               string                 // key: auth, required
	Dataset            string                 // key: dataset, required
	Client             string                 // key: client
	Encoding           dbn.Encoding           // key: encoding
	Compression        dbn.Compression        // key: compression
	TsOut              bool                   // key: ts_out
	PrettyPx           bool                   // key: pretty_px
	PrettyTs           bool                   // key: pretty_ts
	HeartbeatIntervalS uint32                 // key: heartbeat_interval_s (in seconds); 0 means use server default
	SlowReaderBehavior dbn.SlowReaderBehavior // key: slow_reader_behavior, default "warn"
}

// NewAuthenticationRequestMsg parses a control message and returns a AuthenticationRequestMsg
// Returns nil if required fields are missing.
func NewAuthenticationRequestMsg() AuthenticationRequestMsg {
	return AuthenticationRequestMsg{
		Encoding: dbn.Encoding_Dbn,
		Client:   "dbn-go_" + version.Get(),
	}
}

const defaultGatewayEncoding = dbn.Encoding_Dbn
const defaultGatewayCompression = dbn.Compress_None
const defaultGatewaySlowReaderBehavior = dbn.SlowReaderBehavior_Warn

// Encode converts AuthenticationRequestMsg to its line protocol representation.
func (m *AuthenticationRequestMsg) Encode() []byte {
	b := fmt.Appendf(nil, "auth=%s|dataset=%s|client=%s",
		m.Auth, m.Dataset, m.Client)
	if m.Encoding != defaultGatewayEncoding {
		if str := m.Encoding.String(); str != "" {
			b = fmt.Appendf(b, "|encoding=%s", str)
		}
	}
	if m.Compression != defaultGatewayCompression {
		if str := m.Compression.String(); str != "" {
			b = fmt.Appendf(b, "|compression=%s", str)
		}
	}
	if m.TsOut {
		b = fmt.Appendf(b, "|ts_out=1")
	}
	if m.PrettyPx {
		b = fmt.Appendf(b, "|pretty_px=1")
	}
	if m.PrettyTs {
		b = fmt.Appendf(b, "|pretty_ts=1")
	}
	if m.HeartbeatIntervalS != 0 {
		b = fmt.Appendf(b, "|heartbeat_interval_s=%d", m.HeartbeatIntervalS)
	}
	if m.SlowReaderBehavior != defaultGatewaySlowReaderBehavior {
		if str := m.SlowReaderBehavior.String(); str != "" {
			b = fmt.Appendf(b, "|slow_reader_behavior=%s", str)
		}
	}
	b = append(b, '\n')
	return b
}

// NewAuthenticationRequestMsgFromBytes parses an authentication request control
// message, as produced by AuthenticationRequestMsg.Encode. It is the gateway-side
// counterpart of Encode. Absent optional fields take Encode's defaults.
// Returns an error if a required field (auth, dataset) is missing or a field value is invalid.
func NewAuthenticationRequestMsgFromBytes(b []byte) (*AuthenticationRequestMsg, error) {
	m := parseControlMessage(b)
	msg := &AuthenticationRequestMsg{
		Encoding:           defaultGatewayEncoding,
		Compression:        defaultGatewayCompression,
		SlowReaderBehavior: defaultGatewaySlowReaderBehavior,
	}
	var ok bool
	if msg.Auth, ok = m["auth"]; !ok {
		return nil, errors.New("authentication request missing auth")
	}
	if msg.Dataset, ok = m["dataset"]; !ok {
		return nil, errors.New("authentication request missing dataset")
	}
	msg.Client = m["client"]
	var err error
	if str, ok := m["encoding"]; ok {
		if msg.Encoding, err = dbn.EncodingFromString(str); err != nil {
			return nil, err
		}
	}
	if str, ok := m["compression"]; ok {
		if msg.Compression, err = dbn.CompressionFromString(str); err != nil {
			return nil, err
		}
	}
	if msg.TsOut, err = parseControlBool(m, "ts_out"); err != nil {
		return nil, err
	}
	if msg.PrettyPx, err = parseControlBool(m, "pretty_px"); err != nil {
		return nil, err
	}
	if msg.PrettyTs, err = parseControlBool(m, "pretty_ts"); err != nil {
		return nil, err
	}
	if str, ok := m["heartbeat_interval_s"]; ok {
		v, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid heartbeat_interval_s: %w", err)
		}
		msg.HeartbeatIntervalS = uint32(v)
	}
	if str, ok := m["slow_reader_behavior"]; ok {
		if msg.SlowReaderBehavior, err = dbn.SlowReaderBehaviorFromString(str); err != nil {
			return nil, err
		}
	}
	return msg, nil
}

// parseControlBool parses an optional boolean control field ("0"/"1"/"false"/"true").
// An absent field is false.
func parseControlBool(m map[string]string, key string) (bool, error) {
	str, ok := m[key]
	if !ok {
		return false, nil
	}
	v, err := strconv.ParseBool(str)
	if err != nil {
		return false, fmt.Errorf("invalid %s: %w", key, err)
	}
	return v, nil
}

// A subscription request is sent to the gateway upon request from the client.
type SubscriptionRequestMsg struct {
	Schema   string    // key: schema
	StypeIn  dbn.SType // key: stype_in
	Symbols  []string  // key: symbols (comma separated)
	Start    time.Time // key: time (nanoseconds since epoch)
	Snapshot bool      // key: snapshot (int)
}

// Encode converts SubscriptionRequestMsg to its line protocol representation.
func (m *SubscriptionRequestMsg) Encode() []byte {
	b := fmt.Appendf(nil, "schema=%s|stype_in=%s", m.Schema, m.StypeIn.String())

	if !m.Start.IsZero() {
		b = fmt.Appendf(b, "|start=%d", m.Start.UnixNano())
	}
	if m.Snapshot {
		b = append(b, "|snapshot=1"...)
	}

	b = append(b, "|symbols="...)
	isFirst := true
	for _, symbol := range m.Symbols {
		if !isFirst {
			b = append(b, ',')
		}
		b = append(b, symbol...)
		isFirst = false
	}
	b = append(b, '\n')
	return b
}

// NewSubscriptionRequestMsgFromBytes parses a subscription request control message,
// as produced by SubscriptionRequestMsg.Encode. It is the gateway-side counterpart of Encode.
// Returns an error if the schema field is missing or a field value is invalid.
// An empty symbols field yields nil Symbols; an absent start yields a zero Start.
func NewSubscriptionRequestMsgFromBytes(b []byte) (*SubscriptionRequestMsg, error) {
	m := parseControlMessage(b)
	schema, ok := m["schema"]
	if !ok {
		return nil, errors.New("subscription request missing schema")
	}
	msg := &SubscriptionRequestMsg{Schema: schema}
	var err error
	if str, ok := m["stype_in"]; ok {
		if msg.StypeIn, err = dbn.STypeFromString(str); err != nil {
			return nil, err
		}
	}
	if str, ok := m["start"]; ok {
		nanos, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid start: %w", err)
		}
		msg.Start = time.Unix(0, nanos)
	}
	if msg.Snapshot, err = parseControlBool(m, "snapshot"); err != nil {
		return nil, err
	}
	if str := m["symbols"]; str != "" {
		msg.Symbols = strings.Split(str, ",")
	}
	return msg, nil
}

// A session start message is sent to the gateway upon request from the client.
type SessionStartMsg struct {
	StartSession string // key: start_session
}

// Encode converts SessionStartMsg to its line protocol representation.
func (m *SessionStartMsg) Encode() []byte {
	return fmt.Appendf(nil, "start_session=%s\n", m.StartSession)
}

// NewSessionStartMsgFromBytes parses a session start control message, as produced by
// SessionStartMsg.Encode. It is the gateway-side counterpart of Encode.
// Returns an error if the start_session key is missing (its value may be empty).
func NewSessionStartMsgFromBytes(b []byte) (*SessionStartMsg, error) {
	m := parseControlMessage(b)
	value, ok := m["start_session"]
	if !ok {
		return nil, errors.New("not a session start message")
	}
	return &SessionStartMsg{StartSession: value}, nil
}
