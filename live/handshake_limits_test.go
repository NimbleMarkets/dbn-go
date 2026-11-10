// Copyright (c) 2025 Neomantra Corp

package dbn_live

import (
	"bufio"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/NimbleMarkets/dbn-go"
)

const testApiKey = "db-89s9vCvwDDKPdQJ5Pb30Fyj9mNUM6"

// handshakeClient returns a LiveClient connected to an in-memory gateway and
// the gateway's end of the connection.
func handshakeClient(t *testing.T, authTimeout time.Duration) (*LiveClient, net.Conn) {
	t.Helper()
	clientConn, serverConn := net.Pipe()
	t.Cleanup(func() { clientConn.Close(); serverConn.Close() })
	return &LiveClient{
		config:    LiveConfig{Dataset: "XNAS.ITCH", Encoding: dbn.Encoding_Dbn, Client: "test", AuthTimeout: authTimeout},
		conn:      clientConn,
		bufReader: bufio.NewReaderSize(clientConn, MAX_STR_LENGTH),
	}, serverConn
}

// authenticate runs Authenticate and fails the test if it does not return in time.
func authenticate(t *testing.T, client *LiveClient) error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		_, err := client.Authenticate(testApiKey)
		done <- err
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("Authenticate did not return within 3s")
		return nil
	}
}

// A gateway that streams bytes without ever ending the line must not make the
// client buffer them indefinitely.
func TestAuthenticateRejectsOverlongControlLines(t *testing.T) {
	endless := strings.Repeat("a", 1<<20) // 1 MiB with no newline; the gateway then stays silent
	for _, tc := range []struct {
		name   string
		script []string // lines the gateway sends before the endless one
		drain  bool     // whether the gateway must first read the client's auth request
	}{
		{"greeting", nil, false},
		{"challenge", []string{"lsg_version=1.0\n"}, false},
		{"auth_response", []string{"lsg_version=1.0\n", "cram=challenge-value\n"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, serverConn := handshakeClient(t, 0)
			go func() {
				for _, line := range tc.script {
					serverConn.Write([]byte(line))
				}
				if tc.drain {
					bufio.NewReader(serverConn).ReadString('\n')
				}
				serverConn.Write([]byte("lsg_version=" + endless))
			}()
			err := authenticate(t, client)
			if !errors.Is(err, ErrControlMessageTooLong) {
				t.Fatalf("Authenticate error = %v, want ErrControlMessageTooLong", err)
			}
		})
	}
}

// A gateway that accepts the connection and then says nothing must not hang
// Authenticate forever.
func TestAuthenticateTimesOutOnSilentGateway(t *testing.T) {
	client, _ := handshakeClient(t, 100*time.Millisecond)
	err := authenticate(t, client)
	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("Authenticate error = %v, want a deadline error", err)
	}
}

// The handshake deadline must not carry over to the data stream.
func TestAuthenticateClearsDeadlineOnSuccess(t *testing.T) {
	client, serverConn := handshakeClient(t, 200*time.Millisecond)
	go func() {
		serverConn.Write([]byte("lsg_version=1.0\ncram=challenge-value\n"))
		bufio.NewReader(serverConn).ReadString('\n')
		serverConn.Write([]byte("success=1|session_id=sess-42\n"))
		time.Sleep(400 * time.Millisecond) // well past the handshake deadline
		serverConn.Write([]byte("x"))
	}()
	if err := authenticate(t, client); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if _, err := client.bufReader.ReadByte(); err != nil {
		t.Fatalf("read after the handshake deadline passed: %v", err)
	}
}

// Text the gateway puts in its error field must not reach the caller's
// terminal or logs with control characters intact.
func TestAuthenticateErrorEscapesGatewayText(t *testing.T) {
	client, serverConn := handshakeClient(t, 0)
	go func() {
		serverConn.Write([]byte("lsg_version=1.0\ncram=challenge-value\n"))
		bufio.NewReader(serverConn).ReadString('\n')
		serverConn.Write([]byte("success=0|error=bad key\x1b[2J\x1b]0;owned\x07\r\n"))
	}()
	err := authenticate(t, client)
	if err == nil || !strings.Contains(err.Error(), "bad key") {
		t.Fatalf("Authenticate error = %v, want the gateway's message", err)
	}
	for _, r := range err.Error() {
		if r < 0x20 || r == 0x7f {
			t.Fatalf("error text contains control character %q: %q", r, err.Error())
		}
	}
}
