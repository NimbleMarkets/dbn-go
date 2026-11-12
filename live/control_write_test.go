// Copyright (c) 2025 Neomantra Corp

package dbn_live

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NimbleMarkets/dbn-go"
)

// shortWriteConn accepts at most max bytes per Write without reporting an error.
type shortWriteConn struct {
	scriptedConn
	max int
}

func (c *shortWriteConn) Write(p []byte) (int, error) {
	return c.scriptedConn.Write(p[:min(len(p), c.max)])
}

// failingConn fails every Write from call number failAt (1-based) onward.
// When writeBeforeErr is true, the failing Write still accepts len(p) bytes
// alongside net.ErrClosed.
type failingConn struct {
	scriptedConn
	calls, failAt  int
	writeBeforeErr bool
}

func (c *failingConn) Write(p []byte) (int, error) {
	c.calls++
	if c.calls >= c.failAt {
		if c.writeBeforeErr {
			_, _ = c.scriptedConn.Write(p)
			return len(p), net.ErrClosed
		}
		return 0, net.ErrClosed
	}
	return c.scriptedConn.Write(p)
}

// slowConn is safe for concurrent use and pauses inside Write, so writers
// that are not serialized interleave.
type slowConn struct {
	scriptedConn
	mu    sync.Mutex
	lines []string
}

func (c *slowConn) Write(p []byte) (int, error) {
	time.Sleep(time.Millisecond)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lines = append(c.lines, string(p))
	return len(p), nil
}

// A Write that accepts only part of a line must not leave a truncated control
// message on the wire.
func TestSubscribeCompletesShortWrites(t *testing.T) {
	sub := SubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_RawSymbol, Symbols: numberedSymbols(501)}
	reference := &scriptedConn{}
	if err := unsubscribeClient(reference).Subscribe(sub); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	conn := &shortWriteConn{max: 7}
	if err := unsubscribeClient(conn).Subscribe(sub); err != nil {
		t.Fatalf("Subscribe over short writes: %v", err)
	}
	if conn.writes.String() != reference.writes.String() {
		t.Error("bytes sent over short writes differ from a normal send")
	}
}

func TestSubscribeFailsWhenWriteMakesNoProgress(t *testing.T) {
	conn := &shortWriteConn{max: 0}
	done := make(chan error, 1)
	go func() {
		done <- unsubscribeClient(conn).Subscribe(SubscriptionRequestMsg{Schema: "trades", Symbols: []string{"AAPL"}})
	}()
	select {
	case err := <-done:
		if !errors.Is(err, io.ErrShortWrite) {
			t.Errorf("Subscribe error = %v, want io.ErrShortWrite", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Subscribe did not return")
	}
}

// When a later message of a split request fails, the gateway is left holding
// earlier is_last=0 messages. The error must say so, including when a Write
// returns len(p) alongside a non-nil error.
func TestChunkedSendReportsSymbolsAlreadySent(t *testing.T) {
	for _, tc := range []struct {
		name           string
		failAt         int
		writeBeforeErr bool
		wantSent       int
	}{
		{"first_message_zero_bytes", 1, false, 0},
		{"second_message_zero_bytes", 2, false, 500},
		{"third_message_zero_bytes", 3, false, 1000},
		{"first_message_full_bytes_with_err", 1, true, 500},
		{"second_message_full_bytes_with_err", 2, true, 1000},
		{"third_message_full_bytes_with_err", 3, true, 0},
	} {
		for _, op := range []string{"subscribe", "unsubscribe"} {
			t.Run(tc.name+"_"+op, func(t *testing.T) {
				conn := &failingConn{failAt: tc.failAt, writeBeforeErr: tc.writeBeforeErr}
				client := unsubscribeClient(conn)
				symbols := numberedSymbols(1201)
				var err error
				if op == "subscribe" {
					err = client.Subscribe(SubscriptionRequestMsg{Schema: "trades", Symbols: symbols})
				} else {
					err = client.Unsubscribe(UnsubscriptionRequestMsg{Schema: "trades", Symbols: symbols})
				}
				if !errors.Is(err, net.ErrClosed) {
					t.Fatalf("error = %v, want it to wrap net.ErrClosed", err)
				}
				var partial *PartialSendError
				if tc.wantSent == 0 {
					if errors.As(err, &partial) {
						t.Errorf("error = %v, want no PartialSendError", err)
					}
					return
				}
				if !errors.As(err, &partial) {
					t.Fatalf("error = %v, want a PartialSendError", err)
				}
				if partial.SymbolsSent != tc.wantSent || partial.SymbolsTotal != 1201 {
					t.Errorf("PartialSendError = %d of %d symbols, want %d of 1201", partial.SymbolsSent, partial.SymbolsTotal, tc.wantSent)
				}
				if !strings.Contains(err.Error(), fmt.Sprintf("%d of 1201 symbols", tc.wantSent)) {
					t.Errorf("error text %q does not say how many symbols were sent", err.Error())
				}
			})
		}
	}
}

// The messages of one split request must reach the gateway back to back: it
// joins an is_last=0 message with whatever control message comes next.
func TestConcurrentControlMessagesDoNotInterleave(t *testing.T) {
	conn := &slowConn{}
	client := &LiveClient{
		config:    LiveConfig{Dataset: "GLBX.MDP3", Encoding: dbn.Encoding_Json},
		conn:      conn,
		bufReader: bufio.NewReaderSize(conn, MAX_STR_LENGTH),
	}
	symbols := numberedSymbols(1201) // three messages per request

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			schema := fmt.Sprintf("schema%d", i)
			var err error
			if i%2 == 0 {
				err = client.Subscribe(SubscriptionRequestMsg{Schema: schema, Symbols: symbols})
			} else {
				err = client.Unsubscribe(UnsubscriptionRequestMsg{Schema: schema, Symbols: symbols})
			}
			if err != nil {
				t.Errorf("request %d: %v", i, err)
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(2 * time.Millisecond)
		if err := client.Start(); err != nil {
			t.Errorf("Start: %v", err)
		}
	}()
	wg.Wait()

	schemaOf := func(line string) string {
		for _, field := range strings.Split(line, "|") {
			if value, ok := strings.CutPrefix(field, "schema="); ok {
				return value
			}
		}
		return ""
	}
	if len(conn.lines) != 6*3+1 {
		t.Fatalf("sent %d messages, want %d", len(conn.lines), 6*3+1)
	}
	for i, line := range conn.lines {
		if !strings.Contains(line, "is_last=0") {
			continue
		}
		if i+1 == len(conn.lines) {
			t.Fatalf("message %d is partial but nothing follows it", i)
		}
		if next := conn.lines[i+1]; schemaOf(next) != schemaOf(line) {
			t.Fatalf("partial %s message %d was followed by %.40q", schemaOf(line), i, next)
		}
	}
}
