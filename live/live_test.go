// Copyright (c) 2024 Neomantra Corp

package dbn_live

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
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

	Context("UnsubscriptionRequestMsg.Encode", func() {
		It("should encode the unsubscribe line", func() {
			msg := UnsubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_Parent, Symbols: []string{"NQ.FUT", "ES.FUT"}}
			Expect(string(msg.Encode())).To(Equal("unsubscribe|schema=trades|stype_in=parent|symbols=NQ.FUT,ES.FUT\n"))
		})

		It("should pass ALL_SYMBOLS through", func() {
			msg := UnsubscriptionRequestMsg{Schema: "mbp-1", StypeIn: dbn.SType_RawSymbol, Symbols: []string{"ALL_SYMBOLS"}}
			Expect(string(msg.Encode())).To(Equal("unsubscribe|schema=mbp-1|stype_in=raw_symbol|symbols=ALL_SYMBOLS\n"))
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

	Context("gateway message parsers", func() {
		It("should export the CRAM reply generator", func() {
			apiKey := "db-89s9vCvwDDKPdQJ5Pb30Fyj9mNUM6"
			cram := "j5pwMHz6vwXruJM4cOwQrQeQE0bImIzT"
			Expect(GenerateCramReply(apiKey, cram)).To(Equal(generateCramReply(apiKey, cram)))
		})

		It("should round-trip the gateway-side messages", func() {
			greeting := GreetingMsg{LsgVersion: "0.1.2"}
			Expect(string(greeting.Encode())).To(Equal("lsg_version=0.1.2\n"))
			Expect(NewGreetingMsgFromBytes(greeting.Encode())).To(Equal(&greeting))

			challenge := ChallengeRequestMsg{Cram: "abc123"}
			Expect(string(challenge.Encode())).To(Equal("cram=abc123\n"))
			Expect(NewChallengeRequestMsgFromBytes(challenge.Encode())).To(Equal(&challenge))

			ok := AuthenticationResponseMsg{Success: "1", SessionID: "42"}
			Expect(string(ok.Encode())).To(Equal("success=1|session_id=42\n"))
			Expect(NewAuthenticationResponseMsgFromBytes(ok.Encode())).To(Equal(&ok))

			bad := AuthenticationResponseMsg{Success: "0", Error: "bad key"}
			Expect(string(bad.Encode())).To(Equal("success=0|error=bad key\n"))
			Expect(NewAuthenticationResponseMsgFromBytes(bad.Encode())).To(Equal(&bad))
		})

		It("should round-trip a default AuthenticationRequestMsg", func() {
			// The zero SlowReaderBehavior ("") is encoded like the default "warn", so it decodes as "warn".
			msg := AuthenticationRequestMsg{Auth: "a-b", Dataset: "GLBX.MDP3", Client: "c", SlowReaderBehavior: dbn.SlowReaderBehavior_Warn}
			parsed, err := NewAuthenticationRequestMsgFromBytes(msg.Encode())
			Expect(err).ToNot(HaveOccurred())
			Expect(*parsed).To(Equal(msg))
			Expect(parsed.Encoding).To(Equal(dbn.Encoding_Dbn))
			Expect(parsed.Compression).To(Equal(dbn.Compress_None))
			Expect(parsed.SlowReaderBehavior).To(Equal(dbn.SlowReaderBehavior_Warn))
		})

		It("should round-trip an AuthenticationRequestMsg with all fields set", func() {
			msg := AuthenticationRequestMsg{
				Auth:               "my-auth",
				Dataset:            "XNAS.ITCH",
				Client:             "my-client",
				Encoding:           dbn.Encoding_Json,
				Compression:        dbn.Compress_ZStd,
				TsOut:              true,
				PrettyPx:           true,
				PrettyTs:           true,
				HeartbeatIntervalS: 30,
				SlowReaderBehavior: dbn.SlowReaderBehavior_Skip,
			}
			parsed, err := NewAuthenticationRequestMsgFromBytes(msg.Encode())
			Expect(err).ToNot(HaveOccurred())
			Expect(*parsed).To(Equal(msg))
		})

		It("should reject AuthenticationRequestMsg lines missing required fields or with bad values", func() {
			_, err := NewAuthenticationRequestMsgFromBytes([]byte("dataset=GLBX.MDP3\n"))
			Expect(err).To(HaveOccurred())
			_, err = NewAuthenticationRequestMsgFromBytes([]byte("auth=x\n"))
			Expect(err).To(HaveOccurred())
			_, err = NewAuthenticationRequestMsgFromBytes([]byte("auth=x|dataset=d|encoding=xml\n"))
			Expect(err).To(HaveOccurred())
			_, err = NewAuthenticationRequestMsgFromBytes([]byte("auth=x|dataset=d|heartbeat_interval_s=abc\n"))
			Expect(err).To(HaveOccurred())
		})

		It("should round-trip a SubscriptionRequestMsg", func() {
			start := time.Unix(1700000000, 123456789)
			msg := SubscriptionRequestMsg{
				Schema:   "ohlcv-1s",
				StypeIn:  dbn.SType_Parent,
				Symbols:  []string{"ES.FUT", "NQ.FUT"},
				Start:    start,
				Snapshot: true,
				Partial:  true,
			}
			parsed, err := NewSubscriptionRequestMsgFromBytes(msg.Encode())
			Expect(err).ToNot(HaveOccurred())
			Expect(parsed.Schema).To(Equal("ohlcv-1s"))
			Expect(parsed.StypeIn).To(Equal(dbn.SType_Parent))
			Expect(parsed.Symbols).To(Equal([]string{"ES.FUT", "NQ.FUT"}))
			Expect(parsed.Start.UnixNano()).To(Equal(start.UnixNano()))
			Expect(parsed.Snapshot).To(BeTrue())
			Expect(parsed.Partial).To(BeTrue())
			Expect(parsed.Encode()).To(Equal(msg.Encode()))
		})

		It("should parse a minimal SubscriptionRequestMsg", func() {
			msg := SubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_RawSymbol}
			parsed, err := NewSubscriptionRequestMsgFromBytes(msg.Encode())
			Expect(err).ToNot(HaveOccurred())
			Expect(parsed.Schema).To(Equal("trades"))
			Expect(parsed.StypeIn).To(Equal(dbn.SType_RawSymbol))
			Expect(parsed.Symbols).To(BeEmpty())
			Expect(parsed.Start.IsZero()).To(BeTrue())
			Expect(parsed.Snapshot).To(BeFalse())
			Expect(parsed.Partial).To(BeFalse())
		})

		It("should reject SubscriptionRequestMsg lines that are not subscriptions or have bad values", func() {
			_, err := NewSubscriptionRequestMsgFromBytes([]byte("start_session=\n"))
			Expect(err).To(HaveOccurred())
			_, err = NewSubscriptionRequestMsgFromBytes([]byte("schema=trades|stype_in=bogus|symbols=A\n"))
			Expect(err).To(HaveOccurred())
			_, err = NewSubscriptionRequestMsgFromBytes([]byte("schema=trades|stype_in=raw_symbol|start=notanumber|symbols=A\n"))
			Expect(err).To(HaveOccurred())
			_, err = NewSubscriptionRequestMsgFromBytes([]byte("schema=trades|stype_in=raw_symbol|symbols=A|is_last=maybe\n"))
			Expect(err).To(HaveOccurred())
			unsub := UnsubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_RawSymbol, Symbols: []string{"A"}}
			_, err = NewSubscriptionRequestMsgFromBytes(unsub.Encode())
			Expect(err).To(HaveOccurred())
		})

		It("should round-trip a SessionStartMsg", func() {
			msg := SessionStartMsg{}
			parsed, err := NewSessionStartMsgFromBytes(msg.Encode())
			Expect(err).ToNot(HaveOccurred())
			Expect(*parsed).To(Equal(msg))

			_, err = NewSessionStartMsgFromBytes([]byte("schema=trades|stype_in=raw_symbol|symbols=A\n"))
			Expect(err).To(HaveOccurred())
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

	Context("gateway", func() {
		It("should derive the gateway from the dataset by default", func() {
			host, port := resolveGateway(LiveConfig{Dataset: "XNAS.ITCH"})
			Expect(host).To(Equal("xnas-itch.lsg.databento.com"))
			Expect(port).To(Equal(uint16(LIVE_API_PORT)))
		})

		It("should use the configured gateway and port when set", func() {
			host, port := resolveGateway(LiveConfig{Dataset: "XNAS.ITCH", Gateway: "127.0.0.1", Port: 4242})
			Expect(host).To(Equal("127.0.0.1"))
			Expect(port).To(Equal(uint16(4242)))
		})

		It("should connect to the configured gateway and port", func() {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			Expect(err).To(BeNil())
			defer ln.Close()
			listenPort := uint16(ln.Addr().(*net.TCPAddr).Port)

			accepted := make(chan net.Conn, 1)
			go func() {
				if conn, err := ln.Accept(); err == nil {
					accepted <- conn
				}
			}()

			client, err := NewLiveClient(LiveConfig{
				ApiKey:  strings.Repeat("a", API_KEY_LENGTH),
				Dataset: "XNAS.ITCH",
				Gateway: "127.0.0.1",
				Port:    listenPort,
			})
			Expect(err).To(BeNil())
			defer func() { _ = client.Stop() }()

			Expect(client.GetGateway()).To(Equal("127.0.0.1"))
			Expect(client.GetPort()).To(Equal(listenPort))

			var serverConn net.Conn
			Eventually(accepted, 2*time.Second).Should(Receive(&serverConn))
			defer serverConn.Close()
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

	Context("robustness", func() {
		authConn := func(authResponse string) *scriptedConn {
			return &scriptedConn{
				reads: [][]byte{
					[]byte("lsg_version=1.0\n"),
					[]byte("cram=challenge-value\n"),
					[]byte(authResponse),
				},
			}
		}
		newClient := func(conn net.Conn) *LiveClient {
			return &LiveClient{
				config:    LiveConfig{Dataset: "XNAS.ITCH", Client: "dbn-go-live-test"},
				conn:      conn,
				bufReader: bufio.NewReaderSize(conn, MAX_STR_LENGTH),
			}
		}
		apiKey := "db-89s9vCvwDDKPdQJ5Pb30Fyj9mNUM6"

		It("should reject an auth response with an empty success value", func() {
			client := newClient(authConn("success=|session_id=sess-1\n"))
			_, err := client.Authenticate(apiKey)
			Expect(err).To(HaveOccurred())
			Expect(client.GetSessionID()).To(BeEmpty())
		})

		It("should reject an auth response with an unexpected success value", func() {
			client := newClient(authConn("success=true|session_id=sess-1\n"))
			_, err := client.Authenticate(apiKey)
			Expect(err).To(HaveOccurred())
		})

		It("should reject an auth response with no success field", func() {
			client := newClient(authConn("session_id=sess-1\n"))
			_, err := client.Authenticate(apiKey)
			Expect(err).To(HaveOccurred())
		})

		It("should return an error, not panic, on a short API key", func() {
			conn := authConn("success=1|session_id=sess-1\n")
			client := newClient(conn)
			var err error
			Expect(func() { _, err = client.Authenticate("short") }).ToNot(Panic())
			Expect(err).To(HaveOccurred())
			Expect(conn.writes.Len()).To(Equal(0))
		})

		It("should return errors, not panic, after Stop", func() {
			conn := authConn("success=1|session_id=sess-1\n")
			client := newClient(conn)
			Expect(client.Stop()).To(Succeed())

			var err error
			Expect(func() {
				err = client.Subscribe(SubscriptionRequestMsg{Schema: "trades", Symbols: []string{"AAPL"}})
			}).ToNot(Panic())
			Expect(err).To(HaveOccurred())

			Expect(func() { err = client.Start() }).ToNot(Panic())
			Expect(err).To(HaveOccurred())

			Expect(func() { _, err = client.Authenticate(apiKey) }).ToNot(Panic())
			Expect(err).To(HaveOccurred())
		})

		It("should return an error from Subscribe after Stop on a real connection", func() {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			Expect(err).To(BeNil())
			defer ln.Close()
			go func() {
				if conn, err := ln.Accept(); err == nil {
					defer conn.Close()
					_, _ = io.Copy(io.Discard, conn)
				}
			}()

			client, err := NewLiveClient(LiveConfig{
				ApiKey:  strings.Repeat("a", API_KEY_LENGTH),
				Dataset: "XNAS.ITCH",
				Gateway: "127.0.0.1",
				Port:    uint16(ln.Addr().(*net.TCPAddr).Port),
			})
			Expect(err).To(BeNil())
			Expect(client.Stop()).To(Succeed())

			Expect(func() {
				err = client.Subscribe(SubscriptionRequestMsg{Schema: "trades", Symbols: []string{"AAPL"}})
			}).ToNot(Panic())
			Expect(err).To(HaveOccurred())
		})

		DescribeTable("should reject control characters in subscription fields without writing",
			func(sub SubscriptionRequestMsg) {
				conn := &scriptedConn{}
				client := newClient(conn)
				Expect(client.Subscribe(sub)).ToNot(Succeed())
				Expect(conn.writes.Len()).To(Equal(0))
			},
			Entry("pipe in symbol", SubscriptionRequestMsg{Schema: "trades", Symbols: []string{"AAPL|snapshot=1"}}),
			Entry("newline in symbol", SubscriptionRequestMsg{Schema: "trades", Symbols: []string{"AAPL", "MSFT\nstart_session=0"}}),
			Entry("carriage return in symbol", SubscriptionRequestMsg{Schema: "trades", Symbols: []string{"AAPL\r"}}),
			Entry("pipe in schema", SubscriptionRequestMsg{Schema: "trades|snapshot=1", Symbols: []string{"AAPL"}}),
			Entry("newline in schema", SubscriptionRequestMsg{Schema: "trades\n", Symbols: []string{"AAPL"}}),
			Entry("carriage return in schema", SubscriptionRequestMsg{Schema: "trades\r", Symbols: []string{"AAPL"}}),
		)

		It("should write the encoded unsubscribe request", func() {
			conn := &scriptedConn{}
			client := newClient(conn)
			unsub := UnsubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_RawSymbol, Symbols: []string{"AAPL", "MSFT"}}
			Expect(client.Unsubscribe(unsub)).To(Succeed())
			Expect(conn.writes.String()).To(Equal("unsubscribe|schema=trades|stype_in=raw_symbol|symbols=AAPL,MSFT\n"))
		})

		It("should return ErrNotConnected from Unsubscribe after Stop", func() {
			client := newClient(&scriptedConn{})
			Expect(client.Stop()).To(Succeed())
			Expect(client.Unsubscribe(UnsubscriptionRequestMsg{Schema: "trades", Symbols: []string{"AAPL"}})).To(MatchError(ErrNotConnected))
		})

		DescribeTable("should reject invalid unsubscription fields without writing",
			func(unsub UnsubscriptionRequestMsg) {
				conn := &scriptedConn{}
				client := newClient(conn)
				Expect(client.Unsubscribe(unsub)).ToNot(Succeed())
				Expect(conn.writes.Len()).To(Equal(0))
			},
			Entry("no symbols", UnsubscriptionRequestMsg{Schema: "trades"}),
			Entry("pipe in symbol", UnsubscriptionRequestMsg{Schema: "trades", Symbols: []string{"AAPL|schema=mbo"}}),
			Entry("newline in symbol", UnsubscriptionRequestMsg{Schema: "trades", Symbols: []string{"AAPL\nstart_session=0"}}),
			Entry("pipe in schema", UnsubscriptionRequestMsg{Schema: "trades|stype_in=parent", Symbols: []string{"AAPL"}}),
		)

		DescribeTable("should reject control characters in auth fields without writing",
			func(dataset, clientName string) {
				conn := authConn("success=1|session_id=sess-1\n")
				client := newClient(conn)
				client.config.Dataset = dataset
				client.config.Client = clientName
				_, err := client.Authenticate(apiKey)
				Expect(err).To(HaveOccurred())
				Expect(conn.writes.Len()).To(Equal(0))
			},
			Entry("pipe in dataset", "XNAS.ITCH|ts_out=1", "c"),
			Entry("newline in dataset", "XNAS.ITCH\n", "c"),
			Entry("carriage return in dataset", "XNAS.ITCH\r", "c"),
			Entry("pipe in client", "XNAS.ITCH", "c|pretty_px=1"),
			Entry("newline in client", "XNAS.ITCH", "c\nstart_session=0"),
			Entry("carriage return in client", "XNAS.ITCH", "c\r"),
		)

		It("should reject control characters in the config before dialing", func() {
			_, err := NewLiveClient(LiveConfig{
				ApiKey:  strings.Repeat("a", API_KEY_LENGTH),
				Dataset: "XNAS.ITCH|ts_out=1",
				Gateway: "127.0.0.1",
				Port:    1,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Dataset"))
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
