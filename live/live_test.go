// Copyright (c) 2024 Neomantra Corp

package dbn_live

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test Launcher
func TestDbnLive(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dbn-go live suite")
}

var _ = Describe("DbnLive", func() {
	Context("auth", func() {
		It("should generate CRAM response properly", func() {
			// https://databento.com/docs/api-reference-live/message-flows/authentication/example?historical=http&live=raw
			apiKey := "db-89s9vCvwDDKPdQJ5Pb30Fyj9mNUM6"
			cram := "j5pwMHz6vwXruJM4cOwQrQeQE0bImIzT"
			expected := "6d3c875bb9f8cf503c3ed83ee5f476a3ad53f0c67706c51cf42d2db5ad8ff5a9-mNUM6"

			resp := generateCramReply(apiKey, cram)
			Expect(resp).To(Equal(expected))
		})
	})

	Context("AuthenticationRequestMsg.Encode", func() {
		It("should encode only required fields and client when all defaults", func() {
			msg := AuthenticationRequestMsg{
				Auth:    "my-auth",
				Dataset: "XNAS.ITCH",
				Client:  "test-client",
			}
			encoded := string(msg.Encode())
			Expect(encoded).To(Equal("auth=my-auth|dataset=XNAS.ITCH|client=test-client\n"))
		})

		It("should omit encoding when dbn (default)", func() {
			msg := AuthenticationRequestMsg{
				Auth:     "a",
				Dataset:  "d",
				Client:   "c",
				Encoding: dbn.Encoding_Dbn,
			}
			Expect(string(msg.Encode())).ToNot(ContainSubstring("encoding="))
		})

		It("should include encoding when non-default", func() {
			msg := AuthenticationRequestMsg{
				Auth:     "a",
				Dataset:  "d",
				Client:   "c",
				Encoding: dbn.Encoding_Json,
			}
			Expect(string(msg.Encode())).To(ContainSubstring("|encoding=json"))
		})

		It("should include compression when non-default", func() {
			msg := AuthenticationRequestMsg{
				Auth:        "a",
				Dataset:     "d",
				Client:      "c",
				Compression: dbn.Compress_ZStd,
			}
			Expect(string(msg.Encode())).To(ContainSubstring("|compression=zstd"))
		})

		It("should include boolean flags only when true", func() {
			msg := AuthenticationRequestMsg{
				Auth:     "a",
				Dataset:  "d",
				Client:   "c",
				TsOut:    true,
				PrettyPx: true,
				PrettyTs: true,
			}
			encoded := string(msg.Encode())
			Expect(encoded).To(ContainSubstring("|ts_out=1"))
			Expect(encoded).To(ContainSubstring("|pretty_px=1"))
			Expect(encoded).To(ContainSubstring("|pretty_ts=1"))
		})

		It("should omit boolean flags when false", func() {
			msg := AuthenticationRequestMsg{
				Auth:    "a",
				Dataset: "d",
				Client:  "c",
			}
			encoded := string(msg.Encode())
			Expect(encoded).ToNot(ContainSubstring("ts_out"))
			Expect(encoded).ToNot(ContainSubstring("pretty_px"))
			Expect(encoded).ToNot(ContainSubstring("pretty_ts"))
		})

		It("should include heartbeat_interval_s when non-zero", func() {
			msg := AuthenticationRequestMsg{
				Auth:               "a",
				Dataset:            "d",
				Client:             "c",
				HeartbeatIntervalS: 10,
			}
			Expect(string(msg.Encode())).To(ContainSubstring("|heartbeat_interval_s=10"))
		})

		It("should omit heartbeat_interval_s when zero", func() {
			msg := AuthenticationRequestMsg{
				Auth:    "a",
				Dataset: "d",
				Client:  "c",
			}
			Expect(string(msg.Encode())).ToNot(ContainSubstring("heartbeat_interval_s"))
		})

		It("should include slow_reader_behavior when skip (non-default)", func() {
			msg := AuthenticationRequestMsg{
				Auth:               "a",
				Dataset:            "d",
				Client:             "c",
				SlowReaderBehavior: dbn.SlowReaderBehavior_Skip,
			}
			Expect(string(msg.Encode())).To(ContainSubstring("|slow_reader_behavior=skip"))
		})

		It("should omit slow_reader_behavior when warn (default)", func() {
			msg := AuthenticationRequestMsg{
				Auth:               "a",
				Dataset:            "d",
				Client:             "c",
				SlowReaderBehavior: dbn.SlowReaderBehavior_Warn,
			}
			Expect(string(msg.Encode())).ToNot(ContainSubstring("slow_reader_behavior"))
		})

		It("should encode all non-default fields together", func() {
			msg := AuthenticationRequestMsg{
				Auth:               "my-auth",
				Dataset:            "GLBX.MDP3",
				Client:             "my-client",
				Encoding:           dbn.Encoding_Json,
				Compression:        dbn.Compress_ZStd,
				TsOut:              true,
				PrettyPx:           true,
				PrettyTs:           true,
				HeartbeatIntervalS: 30,
				SlowReaderBehavior: dbn.SlowReaderBehavior_Skip,
			}
			encoded := string(msg.Encode())
			Expect(encoded).To(HavePrefix("auth=my-auth|dataset=GLBX.MDP3|client=my-client"))
			Expect(encoded).To(ContainSubstring("|encoding=json"))
			Expect(encoded).To(ContainSubstring("|compression=zstd"))
			Expect(encoded).To(ContainSubstring("|ts_out=1"))
			Expect(encoded).To(ContainSubstring("|pretty_px=1"))
			Expect(encoded).To(ContainSubstring("|pretty_ts=1"))
			Expect(encoded).To(ContainSubstring("|heartbeat_interval_s=30"))
			Expect(encoded).To(ContainSubstring("|slow_reader_behavior=skip"))
			Expect(encoded).To(HaveSuffix("\n"))
		})
	})

	Context("control messages", func() {
		It("should not panic on empty input", func() {
			Expect(func() {
				_ = NewGreetingMsgFromBytes([]byte{})
			}).ToNot(Panic())
			Expect(NewGreetingMsgFromBytes([]byte{})).To(BeNil())
		})
	})

	Context("getters", func() {
		It("should return session ID from GetSessionID", func() {
			client := &LiveClient{
				lsgVersion: "1.2.3",
				sessionID:  "sess-123",
			}
			Expect(client.GetSessionID()).To(Equal("sess-123"))
		})
	})

	Context("authenticate", func() {
		It("should send configured client in auth request and store session ID", func() {
			clientConn, serverConn := net.Pipe()
			defer clientConn.Close()
			defer serverConn.Close()

			client := &LiveClient{
				config: LiveConfig{
					Dataset:  "XNAS.ITCH",
					Encoding: dbn.Encoding_Dbn,
					Client:   "dbn-go-live-test",
				},
				conn:      clientConn,
				bufReader: bufio.NewReaderSize(clientConn, MAX_STR_LENGTH),
			}

			serverErr := make(chan error, 1)
			authLine := make(chan string, 1)
			go func() {
				reader := bufio.NewReader(serverConn)
				if _, err := serverConn.Write([]byte("lsg_version=1.0\n")); err != nil {
					serverErr <- err
					return
				}
				if _, err := serverConn.Write([]byte("cram=challenge-value\n")); err != nil {
					serverErr <- err
					return
				}
				line, err := reader.ReadString('\n')
				if err != nil {
					serverErr <- err
					return
				}
				authLine <- line
				if _, err := serverConn.Write([]byte("success=1|session_id=sess-42\n")); err != nil {
					serverErr <- err
					return
				}
				serverErr <- nil
			}()

			sessionID, err := client.Authenticate("db-89s9vCvwDDKPdQJ5Pb30Fyj9mNUM6")
			Expect(err).To(BeNil())
			Expect(sessionID).To(Equal("sess-42"))
			Expect(client.GetSessionID()).To(Equal("sess-42"))

			line := <-authLine
			Expect(line).To(ContainSubstring("dataset=XNAS.ITCH"))
			Expect(line).To(ContainSubstring("client=dbn-go-live-test"))
			Expect(line).To(ContainSubstring("auth="))

			Expect(<-serverErr).To(BeNil())
		})

		It("should return error on failed authentication response", func() {
			clientConn, serverConn := net.Pipe()
			defer clientConn.Close()
			defer serverConn.Close()

			client := &LiveClient{
				config: LiveConfig{
					Dataset:  "XNAS.ITCH",
					Encoding: dbn.Encoding_Dbn,
					Client:   "dbn-go-live-test",
				},
				conn:      clientConn,
				bufReader: bufio.NewReaderSize(clientConn, MAX_STR_LENGTH),
			}

			go func() {
				reader := bufio.NewReader(serverConn)
				serverConn.Write([]byte("lsg_version=1.0\n"))
				serverConn.Write([]byte("cram=challenge-value\n"))
				_, _ = reader.ReadString('\n') // consume auth request
				serverConn.Write([]byte("success=0|error=bad credentials\n"))
			}()

			_, err := client.Authenticate("db-89s9vCvwDDKPdQJ5Pb30Fyj9mNUM6")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to authenticate"))
		})
	})

	Context("start", func() {
		It("should preserve buffered DBN bytes after authentication", func() {
			dbnBytes, err := os.ReadFile(filepath.Join("..", "tests", "data", "test_data.ohlcv-1s.dbn"))
			Expect(err).To(BeNil())

			conn := &scriptedConn{
				reads: [][]byte{
					[]byte("lsg_version=1.0\n"),
					[]byte("cram=challenge-value\n"),
					append([]byte("success=1|session_id=sess-42\n"), dbnBytes...),
				},
			}

			client := &LiveClient{
				config: LiveConfig{
					Dataset:  "GLBX.MDP3",
					Encoding: dbn.Encoding_Dbn,
				},
				conn:      conn,
				bufReader: bufio.NewReaderSize(conn, MAX_STR_LENGTH),
			}

			sessionID, err := client.Authenticate("db-89s9vCvwDDKPdQJ5Pb30Fyj9mNUM6")
			Expect(err).To(BeNil())
			Expect(sessionID).To(Equal("sess-42"))

			err = client.Start()
			Expect(err).To(BeNil())
			Expect(conn.writes.String()).To(ContainSubstring("start_session="))

			scanner := client.GetDbnScanner()
			Expect(scanner).ToNot(BeNil())
			Expect(scanner.Next()).To(BeTrue())

			record, err := dbn.DbnScannerDecode[dbn.OhlcvMsg](scanner)
			Expect(err).To(BeNil())
			Expect(record.Header.InstrumentID).To(Equal(uint32(5482)))
		})

		It("should preserve buffered JSON bytes after authentication", func() {
			jsonBytes, err := os.ReadFile(filepath.Join("..", "tests", "data", "test_data.ohlcv-1s.json"))
			Expect(err).To(BeNil())

			firstLine, _, ok := bytes.Cut(jsonBytes, []byte{'\n'})
			Expect(ok).To(BeTrue())

			conn := &scriptedConn{
				reads: [][]byte{
					[]byte("lsg_version=1.0\n"),
					[]byte("cram=challenge-value\n"),
					append([]byte("success=1|session_id=sess-43\n"), append(firstLine, '\n')...),
				},
			}

			client := &LiveClient{
				config: LiveConfig{
					Dataset:  "GLBX.MDP3",
					Encoding: dbn.Encoding_Json,
				},
				conn:      conn,
				bufReader: bufio.NewReaderSize(conn, MAX_STR_LENGTH),
			}

			sessionID, err := client.Authenticate("db-89s9vCvwDDKPdQJ5Pb30Fyj9mNUM6")
			Expect(err).To(BeNil())
			Expect(sessionID).To(Equal("sess-43"))

			err = client.Start()
			Expect(err).To(BeNil())

			scanner := client.GetJsonScanner()
			Expect(scanner).ToNot(BeNil())
			Expect(scanner.Next()).To(BeTrue())

			record, err := dbn.JsonScannerDecode[dbn.OhlcvMsg](scanner)
			Expect(err).To(BeNil())
			Expect(record.Header.InstrumentID).To(Equal(uint32(5482)))
		})
	})
})

type scriptedConn struct {
	reads   [][]byte
	pending []byte
	writes  bytes.Buffer
	closed  bool
}

func (c *scriptedConn) Read(p []byte) (int, error) {
	if len(c.pending) == 0 {
		if len(c.reads) == 0 {
			return 0, io.EOF
		}
		c.pending = c.reads[0]
		c.reads = c.reads[1:]
	}

	n := copy(p, c.pending)
	c.pending = c.pending[n:]
	return n, nil
}

func (c *scriptedConn) Write(p []byte) (int, error) {
	if c.closed {
		return 0, net.ErrClosed
	}
	return c.writes.Write(p)
}

func (c *scriptedConn) Close() error {
	c.closed = true
	return nil
}

func (c *scriptedConn) LocalAddr() net.Addr {
	return dummyAddr("local")
}

func (c *scriptedConn) RemoteAddr() net.Addr {
	return dummyAddr("remote")
}

func (c *scriptedConn) SetDeadline(time.Time) error {
	return nil
}

func (c *scriptedConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *scriptedConn) SetWriteDeadline(time.Time) error {
	return nil
}

type dummyAddr string

func (a dummyAddr) Network() string {
	return string(a)
}

func (a dummyAddr) String() string {
	return string(a)
}
