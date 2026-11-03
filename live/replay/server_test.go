// Copyright (c) 2026 Neomantra Corp

package dbn_live_replay_test

import (
	"bufio"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NimbleMarkets/dbn-go"
	dbn_live "github.com/NimbleMarkets/dbn-go/live"
	dbn_live_replay "github.com/NimbleMarkets/dbn-go/live/replay"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDbnLiveReplay(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dbn-go live replay suite")
}

var (
	ohlcvFile  = filepath.Join("..", "..", "tests", "data", "test_data.ohlcv-1s.v3.dbn.zst")
	tradesFile = filepath.Join("..", "..", "tests", "data", "test_data.trades.v3.dbn.zst")
	testApiKey = "db-" + strings.Repeat("x", 24) + "ABCDE" // 32 chars
	discardLog = slog.New(slog.DiscardHandler)
)

// readFile reads the metadata and OHLCV records of a DBN file directly.
func readOhlcvFile(path string) ([]dbn.OhlcvMsg, *dbn.Metadata) {
	reader, closer, err := dbn.MakeCompressedReader(path, false)
	Expect(err).ToNot(HaveOccurred())
	defer closer.Close()
	records, metadata, err := dbn.ReadDBNToSlice[dbn.OhlcvMsg](reader)
	Expect(err).ToNot(HaveOccurred())
	Expect(records).ToNot(BeEmpty())
	return records, metadata
}

// startServer starts a replay server on a loopback port and registers its cleanup.
func startServer(cfg dbn_live_replay.ServerConfig) *dbn_live_replay.Server {
	if cfg.Logger == nil {
		cfg.Logger = discardLog
	}
	server, err := dbn_live_replay.NewServer(cfg)
	Expect(err).ToNot(HaveOccurred())
	Expect(server.Listen("127.0.0.1:0")).To(Succeed())
	DeferCleanup(func() {
		Expect(closeWithin(server, 5*time.Second)).To(BeTrue(), "Server.Close hung")
	})
	return server
}

// closeWithin reports whether server.Close returns within timeout.
func closeWithin(server *dbn_live_replay.Server, timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		_ = server.Close()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

// newClient connects a real LiveClient to server.
func newClient(server *dbn_live_replay.Server, dataset string, encoding dbn.Encoding) *dbn_live.LiveClient {
	addr := server.Addr().(*net.TCPAddr)
	client, err := dbn_live.NewLiveClient(dbn_live.LiveConfig{
		Logger:   discardLog,
		ApiKey:   testApiKey,
		Dataset:  dataset,
		Gateway:  "127.0.0.1",
		Port:     uint16(addr.Port),
		Encoding: encoding,
	})
	Expect(err).ToNot(HaveOccurred())
	DeferCleanup(func() { _ = client.Stop() })
	return client
}

func subscribe(client *dbn_live.LiveClient, schema string) {
	Expect(client.Subscribe(dbn_live.SubscriptionRequestMsg{
		Schema:  schema,
		StypeIn: dbn.SType_RawSymbol,
		Symbols: []string{"ESH1"},
	})).To(Succeed())
}

// readOhlcvStream reads OHLCV records from the client until the stream ends.
func readOhlcvStream(client *dbn_live.LiveClient) []dbn.OhlcvMsg {
	scanner := client.GetDbnScanner()
	Expect(scanner).ToNot(BeNil())
	var records []dbn.OhlcvMsg
	for scanner.Next() {
		record, err := dbn.DbnScannerDecode[dbn.OhlcvMsg](scanner)
		Expect(err).ToNot(HaveOccurred())
		records = append(records, *record)
	}
	Expect(scanner.Error()).To(Equal(io.EOF))
	return records
}

var _ = Describe("Server", func() {
	Context("NewServer", func() {
		It("should reject invalid configs", func() {
			_, err := dbn_live_replay.NewServer(dbn_live_replay.ServerConfig{})
			Expect(err).To(HaveOccurred())

			_, err = dbn_live_replay.NewServer(dbn_live_replay.ServerConfig{Files: []string{"does-not-exist.dbn"}})
			Expect(err).To(HaveOccurred())

			_, err = dbn_live_replay.NewServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}, ApiKey: "short"})
			Expect(err).To(HaveOccurred())

			_, err = dbn_live_replay.NewServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}, Speed: -1})
			Expect(err).To(HaveOccurred())
		})

		It("should have no address until listening", func() {
			server, err := dbn_live_replay.NewServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}, Logger: discardLog})
			Expect(err).ToNot(HaveOccurred())
			Expect(server.Addr()).To(BeNil())
			Expect(server.Close()).To(Succeed())
		})
	})

	Context("replay", func() {
		It("should replay a file's metadata and records to a LiveClient", func() {
			expected, expectedMeta := readOhlcvFile(ohlcvFile)
			server := startServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}})

			client := newClient(server, expectedMeta.Dataset, dbn.Encoding_Dbn)
			sessionID, err := client.Authenticate(testApiKey)
			Expect(err).ToNot(HaveOccurred())
			Expect(sessionID).ToNot(BeEmpty())
			Expect(client.GetLsgVersion()).ToNot(BeEmpty())

			subscribe(client, "ohlcv-1s")
			Expect(client.Start()).To(Succeed())

			metadata, err := client.GetDbnScanner().Metadata()
			Expect(err).ToNot(HaveOccurred())
			Expect(*metadata).To(Equal(*expectedMeta))

			records := readOhlcvStream(client)
			Expect(records).To(HaveLen(len(expected)))
			Expect(records[0]).To(Equal(expected[0]))
			Expect(records[len(records)-1]).To(Equal(expected[len(expected)-1]))
			Expect(records).To(Equal(expected))
		})

		It("should replay the file matching the subscribed schema", func() {
			expected, expectedMeta := readOhlcvFile(ohlcvFile)
			server := startServer(dbn_live_replay.ServerConfig{Files: []string{tradesFile, ohlcvFile}})

			client := newClient(server, expectedMeta.Dataset, dbn.Encoding_Dbn)
			_, err := client.Authenticate(testApiKey)
			Expect(err).ToNot(HaveOccurred())
			subscribe(client, "ohlcv-1s")
			Expect(client.Start()).To(Succeed())

			metadata, err := client.GetDbnScanner().Metadata()
			Expect(err).ToNot(HaveOccurred())
			Expect(metadata.Schema).To(Equal(dbn.Schema_Ohlcv1S))
			Expect(readOhlcvStream(client)).To(Equal(expected))
		})

		It("should serve consecutive sessions", func() {
			expected, expectedMeta := readOhlcvFile(ohlcvFile)
			server := startServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}})
			sessionIDs := map[string]bool{}
			for range 3 {
				client := newClient(server, expectedMeta.Dataset, dbn.Encoding_Dbn)
				sessionID, err := client.Authenticate(testApiKey)
				Expect(err).ToNot(HaveOccurred())
				sessionIDs[sessionID] = true
				subscribe(client, "ohlcv-1s")
				Expect(client.Start()).To(Succeed())
				Expect(readOhlcvStream(client)).To(Equal(expected))
				Expect(client.Stop()).To(Succeed())
			}
			Expect(sessionIDs).To(HaveLen(3))
		})

		It("should pace records by ts_event when Speed is set", func() {
			expected, expectedMeta := readOhlcvFile(ohlcvFile)
			Expect(len(expected)).To(BeNumerically(">=", 2))
			span := time.Duration(expected[len(expected)-1].Header.TsEvent - expected[0].Header.TsEvent)
			Expect(span).To(BeNumerically(">", 0))
			const speed = 4.0
			server := startServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}, Speed: speed})

			client := newClient(server, expectedMeta.Dataset, dbn.Encoding_Dbn)
			_, err := client.Authenticate(testApiKey)
			Expect(err).ToNot(HaveOccurred())
			subscribe(client, "ohlcv-1s")
			began := time.Now()
			Expect(client.Start()).To(Succeed())
			Expect(readOhlcvStream(client)).To(Equal(expected))
			minimum := time.Duration(float64(span) / speed)
			Expect(time.Since(began)).To(BeNumerically(">=", minimum*9/10))
		})
	})

	Context("authentication", func() {
		It("should accept the configured API key", func() {
			_, meta := readOhlcvFile(ohlcvFile)
			server := startServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}, ApiKey: testApiKey})
			client := newClient(server, meta.Dataset, dbn.Encoding_Dbn)
			_, err := client.Authenticate(testApiKey)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reject a wrong API key", func() {
			_, meta := readOhlcvFile(ohlcvFile)
			otherKey := "db-" + strings.Repeat("y", 24) + "ABCDE"
			server := startServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}, ApiKey: otherKey})
			client := newClient(server, meta.Dataset, dbn.Encoding_Dbn)
			_, err := client.Authenticate(testApiKey)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to authenticate"))
		})

		It("should reject non-DBN encodings", func() {
			_, meta := readOhlcvFile(ohlcvFile)
			server := startServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}})
			client := newClient(server, meta.Dataset, dbn.Encoding_Json)
			_, err := client.Authenticate(testApiKey)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("dbn"))
		})

		It("should reject a dataset with no replay file", func() {
			server := startServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}})
			client := newClient(server, "XNAS.ITCH", dbn.Encoding_Dbn)
			_, err := client.Authenticate(testApiKey)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("XNAS.ITCH"))
		})

		It("should reject a malformed authentication request", func() {
			server := startServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}})
			conn, err := net.Dial("tcp", server.Addr().String())
			Expect(err).ToNot(HaveOccurred())
			defer conn.Close()
			Expect(conn.SetDeadline(time.Now().Add(5 * time.Second))).To(Succeed())
			reader := bufio.NewReader(conn)
			greeting, err := reader.ReadBytes('\n')
			Expect(err).ToNot(HaveOccurred())
			Expect(dbn_live.NewGreetingMsgFromBytes(greeting)).ToNot(BeNil())
			challenge, err := reader.ReadBytes('\n')
			Expect(err).ToNot(HaveOccurred())
			Expect(dbn_live.NewChallengeRequestMsgFromBytes(challenge)).ToNot(BeNil())

			_, err = conn.Write([]byte("hello=world\n"))
			Expect(err).ToNot(HaveOccurred())
			line, err := reader.ReadBytes('\n')
			Expect(err).ToNot(HaveOccurred())
			resp := dbn_live.NewAuthenticationResponseMsgFromBytes(line)
			Expect(resp).ToNot(BeNil())
			Expect(resp.Success).To(Equal("0"))
			Expect(resp.Error).ToNot(BeEmpty())
		})
	})

	Context("shutdown", func() {
		It("should not hang when a client disconnects mid-stream", func() {
			_, meta := readOhlcvFile(ohlcvFile)
			// ohlcv-1s records are a second apart, so this waits ~1000s between records.
			server, err := dbn_live_replay.NewServer(dbn_live_replay.ServerConfig{
				Files: []string{ohlcvFile}, Speed: 0.001, Logger: discardLog,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(server.Listen("127.0.0.1:0")).To(Succeed())

			client := newClient(server, meta.Dataset, dbn.Encoding_Dbn)
			_, err = client.Authenticate(testApiKey)
			Expect(err).ToNot(HaveOccurred())
			subscribe(client, "ohlcv-1s")
			Expect(client.Start()).To(Succeed())
			Expect(client.GetDbnScanner().Next()).To(BeTrue())
			Expect(client.Stop()).To(Succeed())

			Expect(closeWithin(server, 2*time.Second)).To(BeTrue(), "Server.Close hung after client disconnect")
		})

		It("should end active sessions on Close", func() {
			_, meta := readOhlcvFile(ohlcvFile)
			server, err := dbn_live_replay.NewServer(dbn_live_replay.ServerConfig{
				Files: []string{ohlcvFile}, Speed: 0.001, Logger: discardLog,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(server.Listen("127.0.0.1:0")).To(Succeed())

			client := newClient(server, meta.Dataset, dbn.Encoding_Dbn)
			_, err = client.Authenticate(testApiKey)
			Expect(err).ToNot(HaveOccurred())
			subscribe(client, "ohlcv-1s")
			Expect(client.Start()).To(Succeed())
			scanner := client.GetDbnScanner()
			Expect(scanner.Next()).To(BeTrue())

			Expect(closeWithin(server, 2*time.Second)).To(BeTrue(), "Server.Close hung with an active session")
			Expect(scanner.Next()).To(BeFalse())
			Expect(server.Close()).To(Succeed()) // idempotent
		})

		It("should return nil from Serve after Close and refuse to listen again", func() {
			server, err := dbn_live_replay.NewServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}, Logger: discardLog})
			Expect(err).ToNot(HaveOccurred())
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			Expect(err).ToNot(HaveOccurred())
			served := make(chan error, 1)
			go func() { served <- server.Serve(ln) }()
			Eventually(server.Addr, 2*time.Second).ShouldNot(BeNil())

			Expect(closeWithin(server, 2*time.Second)).To(BeTrue(), "Server.Close hung")
			Eventually(served, 2*time.Second).Should(Receive(BeNil()))
			Expect(server.Listen("127.0.0.1:0")).To(MatchError(dbn_live_replay.ErrServerClosed))
		})

		It("should close a connection that never authenticates", func() {
			server, err := dbn_live_replay.NewServer(dbn_live_replay.ServerConfig{Files: []string{ohlcvFile}, Logger: discardLog})
			Expect(err).ToNot(HaveOccurred())
			Expect(server.Listen("127.0.0.1:0")).To(Succeed())
			conn, err := net.Dial("tcp", server.Addr().String())
			Expect(err).ToNot(HaveOccurred())
			defer conn.Close()
			// Wait for the greeting so the server has accepted the connection;
			// closing the listener with it still in the backlog would reset it.
			Expect(conn.SetReadDeadline(time.Now().Add(2 * time.Second))).To(Succeed())
			reader := bufio.NewReader(conn)
			_, err = reader.ReadBytes('\n')
			Expect(err).ToNot(HaveOccurred())
			Expect(closeWithin(server, 2*time.Second)).To(BeTrue(), "Server.Close hung with an idle connection")
			_, err = io.ReadAll(reader)
			Expect(err).ToNot(HaveOccurred()) // EOF, not a timeout
		})
	})
})
