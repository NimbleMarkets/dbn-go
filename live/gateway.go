// Copyright (c) 2024 Neomantra Corp

package dbn_live

import (
	"bytes"
	"fmt"
	"time"

	"github.com/NimbleMarkets/dbn-go"
)

// Returns a string key/value map from a Databento control message
// The format is: "k1=v1|k2=v2|k3=v3\n"
func parseControlMessage(b []byte) map[string]string {
	m := make(map[string]string)
	kvs := bytes.Split(b[:len(b)-1], []byte{'|'})
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

// AuthenticationResponseMsg is an authentication response is sent by the gateway after a valid
// authentication request is sent to the gateway.
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

// AuthenticationRequestMsg is an authentication request is sent to the gateway after a challenge response is received.
// This is required to authenticate a user.
type AuthenticationRequestMsg struct {
	Auth     string       // key: auth
	Dataset  string       // key: dataset
	Encoding dbn.Encoding // key: encoding
	Details  string       // key: details
	TsOut    bool         // key: ts_out
	Client   string       // key: client
}

// NewAuthenticationRequestMsg parses a control message and returns a AuthenticationRequestMsg
// Returns nil if required fields are missing.
func NewAuthenticationRequestMsg() AuthenticationRequestMsg {
	return AuthenticationRequestMsg{
		Encoding: dbn.Encoding_Dbn,
		Client:   "USER_AGENT", // TODO
	}
}

// Encode converts AuthenticationRequestMsg to its line protocol representation.
func (m *AuthenticationRequestMsg) Encode() []byte {
	tsOutStr := "0"
	if m.TsOut {
		tsOutStr = "1"
	}
	return fmt.Appendf(nil, "auth=%s|dataset=%s|encoding=%s|ts_out=%s|client=%s\n",
		m.Auth, m.Dataset, m.Encoding.String(), tsOutStr, m.Client)
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
		b = fmt.Appendf(b, "|time=%d", m.Start.UnixNano())
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

// A session start message is sent to the gateway upon request from the client.
type SessionStartMsg struct {
	StartSession string // key: start_session
}

// Encode converts SessionStartMsg to its line protocol representation.
func (m *SessionStartMsg) Encode() []byte {
	return fmt.Appendf(nil, "start_session=%s\n", m.StartSession)
}
