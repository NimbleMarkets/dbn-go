// Copyright (c) 2024-2025 Neomantra Corp

package dbn_live

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/NimbleMarkets/dbn-go"
)

func FuzzLiveGatewayControl(f *testing.F) {
	f.Add([]byte("lsg_version=0.18.1\n"), "db-89s9vCvwDDKPdQJ5Pb30Fyj9mNUM6", "XNAS.ITCH", "dbn-go", "trades", "AAPL", uint8(0))
	f.Add([]byte("cram=j5pwMHz6vwXruJM4cOwQrQeQE0bImIzT\n"), "0123456789abcdef0123456789abcdef", "GLBX.MDP3", "client=1", "mbp-1", "ESH1", uint8(255))
	f.Add([]byte("success=1|session_id=12345\n"), "short-key", "bad|dataset", "bad\nclient", "ohlcv-1s", "SPY,QQQ", uint8(7))
	f.Add([]byte("success=0|error=Authentication failed: bad key\r\n"), "", "", "", "", "", uint8(1))

	f.Fuzz(func(t *testing.T, raw []byte, apiKey, dataset, client, schema, symbol string, flags uint8) {
		if len(raw) > 8192 {
			t.Skip()
		}
		before := bytes.Clone(raw)

		parsed := parseControlMessage(raw)
		greeting := NewGreetingMsgFromBytes(raw)
		challenge := NewChallengeRequestMsgFromBytes(raw)
		authResp := NewAuthenticationResponseMsgFromBytes(raw)
		if !bytes.Equal(raw, before) {
			t.Fatal("control message parser mutated input")
		}

		if v, ok := parsed["lsg_version"]; ok {
			if greeting == nil || greeting.LsgVersion != v {
				t.Fatalf("greeting mismatch: got %#v, want %q", greeting, v)
			}
		} else if greeting != nil {
			t.Fatalf("unexpected non-nil greeting: %#v", greeting)
		}

		if v, ok := parsed["cram"]; ok {
			if challenge == nil || challenge.Cram != v {
				t.Fatalf("challenge mismatch: got %#v, want %q", challenge, v)
			}
		} else if challenge != nil {
			t.Fatalf("unexpected non-nil challenge: %#v", challenge)
		}

		if v, ok := parsed["success"]; ok {
			if authResp == nil || authResp.Success != v || authResp.Error != parsed["error"] || authResp.SessionID != parsed["session_id"] {
				t.Fatalf("authResp mismatch: got %#v, parsed %v", authResp, parsed)
			}
		} else if authResp != nil {
			t.Fatalf("unexpected non-nil authResp: %#v", authResp)
		}

		if len(apiKey) == API_KEY_LENGTH && validateControlValue("apiKey", apiKey) == nil && challenge != nil && validateControlValue("cram", challenge.Cram) == nil {
			reply := generateCramReply(apiKey, challenge.Cram)
			// 64 hex chars + '-' + 5-char bucket ID
			if len(reply) != 64+1+BUCKET_ID_LENGTH || !strings.HasSuffix(reply, "-"+apiKey[API_KEY_LENGTH-BUCKET_ID_LENGTH:]) {
				t.Fatalf("invalid CRAM reply format: %q", reply)
			}
		}

		if validateControlValue("Dataset", dataset) == nil && validateControlValue("Client", client) == nil && validateControlValue("Auth", apiKey) == nil {
			authMsg := AuthenticationRequestMsg{
				Auth:               apiKey,
				Dataset:            dataset,
				Client:             client,
				Encoding:           dbn.Encoding(flags % 4),
				Compression:        dbn.Compression((flags >> 2) % 3),
				TsOut:              flags&0x10 != 0,
				PrettyPx:           flags&0x20 != 0,
				PrettyTs:           flags&0x40 != 0,
				HeartbeatIntervalS: uint32(flags),
				SlowReaderBehavior: dbn.SlowReaderBehavior((flags >> 4) % 3),
			}
			encoded := authMsg.Encode()
			if len(encoded) == 0 || encoded[len(encoded)-1] != '\n' {
				t.Fatalf("AuthenticationRequestMsg.Encode missing trailing newline: %q", encoded)
			}
			if bytes.ContainsAny(encoded[:len(encoded)-1], "\r\n") {
				t.Fatalf("AuthenticationRequestMsg.Encode produced embedded newline: %q", encoded)
			}
			rt := parseControlMessage(encoded)
			if rt["auth"] != apiKey || rt["dataset"] != dataset || rt["client"] != client {
				t.Fatalf("AuthenticationRequestMsg round-trip mismatch: got %v", rt)
			}
		}

		if validateControlValue("Schema", schema) == nil && validateControlValue("symbol", symbol) == nil && !strings.Contains(symbol, ",") {
			subMsg := SubscriptionRequestMsg{
				Schema:   schema,
				StypeIn:  dbn.SType(flags % 12),
				Symbols:  []string{symbol},
				Snapshot: flags&0x01 != 0,
			}
			if flags&0x02 != 0 {
				subMsg.Start = time.Unix(0, int64(flags)*1e9).UTC()
			}
			encoded := subMsg.Encode()
			if len(encoded) == 0 || encoded[len(encoded)-1] != '\n' {
				t.Fatalf("SubscriptionRequestMsg.Encode missing trailing newline: %q", encoded)
			}
			if bytes.ContainsAny(encoded[:len(encoded)-1], "\r\n") {
				t.Fatalf("SubscriptionRequestMsg.Encode produced embedded newline: %q", encoded)
			}
			rt := parseControlMessage(encoded)
			if rt["schema"] != schema || rt["symbols"] != symbol {
				t.Fatalf("SubscriptionRequestMsg round-trip mismatch: got %v", rt)
			}
		}
	})
}
