// Copyright (c) 2024 Neomantra Corp

package dbn_live

import (
	"bufio"
	"net"
	"testing"

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
			Expect(line).To(ContainSubstring("encoding=dbn"))
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
})
