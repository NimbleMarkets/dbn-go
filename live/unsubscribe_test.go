// Copyright (c) 2025 Neomantra Corp

package dbn_live

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/NimbleMarkets/dbn-go"
)

func unsubscribeClient(conn net.Conn) *LiveClient {
	return &LiveClient{
		config:    LiveConfig{Dataset: "GLBX.MDP3", Client: "dbn-go-live-test"},
		conn:      conn,
		bufReader: bufio.NewReaderSize(conn, MAX_STR_LENGTH),
	}
}

func numberedSymbols(n int) []string {
	symbols := make([]string, n)
	for i := range symbols {
		symbols[i] = fmt.Sprintf("S%04d", i)
	}
	return symbols
}

// is_last defaults to 1 on the gateway, so only a partial message carries it.
func TestUnsubscriptionRequestMsgEncodePartial(t *testing.T) {
	msg := UnsubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_Parent, Symbols: []string{"NQ.FUT", "ES.FUT"}, Partial: true}
	const want = "unsubscribe|schema=trades|stype_in=parent|symbols=NQ.FUT,ES.FUT|is_last=0\n"
	if got := string(msg.Encode()); got != want {
		t.Errorf("Encode() = %q, want %q", got, want)
	}
}

// Like the Rust and C++ clients, a long symbol list is sent as messages of at
// most 500 symbols, with is_last=0 on all but the last.
func TestUnsubscribeSplitsLongSymbolLists(t *testing.T) {
	for _, tc := range []struct {
		numSymbols int
		wantSizes  []int
	}{
		{1, []int{1}},
		{500, []int{500}},
		{501, []int{500, 1}},
		{1000, []int{500, 500}},
		{1201, []int{500, 500, 201}},
	} {
		t.Run(fmt.Sprint(tc.numSymbols), func(t *testing.T) {
			conn := &scriptedConn{}
			symbols := numberedSymbols(tc.numSymbols)
			err := unsubscribeClient(conn).Unsubscribe(UnsubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_RawSymbol, Symbols: symbols})
			if err != nil {
				t.Fatalf("Unsubscribe: %v", err)
			}
			lines := strings.Split(strings.TrimSuffix(conn.writes.String(), "\n"), "\n")
			if len(lines) != len(tc.wantSizes) {
				t.Fatalf("sent %d messages, want %d", len(lines), len(tc.wantSizes))
			}
			var sent []string
			for i, line := range lines {
				const prefix = "unsubscribe|schema=trades|stype_in=raw_symbol|symbols="
				rest, ok := strings.CutPrefix(line, prefix)
				if !ok {
					t.Fatalf("message %d = %q, want prefix %q", i, line, prefix)
				}
				isLast := i == len(lines)-1
				symbolList, partial := strings.CutSuffix(rest, "|is_last=0")
				if partial == isLast {
					t.Errorf("message %d: is_last=0 present = %v, want %v", i, partial, !isLast)
				}
				chunk := strings.Split(symbolList, ",")
				if len(chunk) != tc.wantSizes[i] {
					t.Errorf("message %d has %d symbols, want %d", i, len(chunk), tc.wantSizes[i])
				}
				sent = append(sent, chunk...)
			}
			if strings.Join(sent, ",") != strings.Join(symbols, ",") {
				t.Error("symbols sent do not match the request, in order")
			}
		})
	}
}

// A caller that splits a request itself can mark its own messages partial.
func TestUnsubscribeKeepsCallerPartialOnFinalMessage(t *testing.T) {
	conn := &scriptedConn{}
	err := unsubscribeClient(conn).Unsubscribe(UnsubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_RawSymbol, Symbols: []string{"AAPL"}, Partial: true})
	if err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
	const want = "unsubscribe|schema=trades|stype_in=raw_symbol|symbols=AAPL|is_last=0\n"
	if got := conn.writes.String(); got != want {
		t.Errorf("sent %q, want %q", got, want)
	}
}

func TestUnsubscribeDoesNotModifyCallerSymbols(t *testing.T) {
	symbols := numberedSymbols(501)
	want := strings.Join(symbols, ",")
	unsub := UnsubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_RawSymbol, Symbols: symbols}
	if err := unsubscribeClient(&scriptedConn{}).Unsubscribe(unsub); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
	if strings.Join(symbols, ",") != want || len(unsub.Symbols) != 501 || unsub.Partial {
		t.Error("Unsubscribe modified the caller's request")
	}
}

func TestUnsubscribeWrapsWriteError(t *testing.T) {
	conn := &scriptedConn{}
	conn.Close() // the connection fails underneath a client that still holds it
	err := unsubscribeClient(conn).Unsubscribe(UnsubscriptionRequestMsg{Schema: "trades", Symbols: []string{"AAPL"}})
	if !errors.Is(err, net.ErrClosed) {
		t.Errorf("Unsubscribe error = %v, want it to wrap net.ErrClosed", err)
	}
}
