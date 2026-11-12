// Copyright (c) 2025 Neomantra Corp

package dbn_live

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/NimbleMarkets/dbn-go"
)

// is_last defaults to 1 on the gateway, so only a partial message carries it.
func TestSubscriptionRequestMsgEncodePartial(t *testing.T) {
	msg := SubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_Parent, Symbols: []string{"NQ.FUT", "ES.FUT"}, Partial: true}
	const want = "schema=trades|stype_in=parent|is_last=0|symbols=NQ.FUT,ES.FUT\n"
	if got := string(msg.Encode()); got != want {
		t.Errorf("Encode() = %q, want %q", got, want)
	}
}

// Like the Rust and C++ clients, a long symbol list is sent as messages of at
// most 500 symbols, with is_last=0 on all but the last. Every message repeats
// the subscription's other fields.
func TestSubscribeSplitsLongSymbolLists(t *testing.T) {
	start := time.Unix(0, 1700000000000000000).UTC()
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
			err := unsubscribeClient(conn).Subscribe(SubscriptionRequestMsg{
				Schema: "trades", StypeIn: dbn.SType_RawSymbol, Symbols: symbols, Start: start,
			})
			if err != nil {
				t.Fatalf("Subscribe: %v", err)
			}
			lines := strings.Split(strings.TrimSuffix(conn.writes.String(), "\n"), "\n")
			if len(lines) != len(tc.wantSizes) {
				t.Fatalf("sent %d messages, want %d", len(lines), len(tc.wantSizes))
			}
			var sent []string
			for i, line := range lines {
				prefix := "schema=trades|stype_in=raw_symbol|start=1700000000000000000|symbols="
				if i < len(lines)-1 {
					prefix = "schema=trades|stype_in=raw_symbol|start=1700000000000000000|is_last=0|symbols="
				}
				symbolList, ok := strings.CutPrefix(line, prefix)
				if !ok {
					t.Fatalf("message %d = %.120q, want prefix %q", i, line, prefix)
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

func TestSubscribeRepeatsSnapshotOnEveryMessage(t *testing.T) {
	conn := &scriptedConn{}
	err := unsubscribeClient(conn).Subscribe(SubscriptionRequestMsg{
		Schema: "mbo", StypeIn: dbn.SType_RawSymbol, Symbols: numberedSymbols(501), Snapshot: true,
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(conn.writes.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("sent %d messages, want 2", len(lines))
	}
	for i, line := range lines {
		if !strings.HasPrefix(line, "schema=mbo|stype_in=raw_symbol|snapshot=1|") {
			t.Errorf("message %d = %.80q, want snapshot=1", i, line)
		}
	}
}

// A caller that splits a request itself can mark its own messages partial.
func TestSubscribeKeepsCallerPartialOnFinalMessage(t *testing.T) {
	conn := &scriptedConn{}
	err := unsubscribeClient(conn).Subscribe(SubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_RawSymbol, Symbols: []string{"AAPL"}, Partial: true})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	const want = "schema=trades|stype_in=raw_symbol|is_last=0|symbols=AAPL\n"
	if got := conn.writes.String(); got != want {
		t.Errorf("sent %q, want %q", got, want)
	}
}

func TestSubscribeDoesNotModifyCallerSymbols(t *testing.T) {
	symbols := numberedSymbols(501)
	want := strings.Join(symbols, ",")
	sub := SubscriptionRequestMsg{Schema: "trades", StypeIn: dbn.SType_RawSymbol, Symbols: symbols}
	if err := unsubscribeClient(&scriptedConn{}).Subscribe(sub); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if strings.Join(symbols, ",") != want || len(sub.Symbols) != 501 || sub.Partial {
		t.Error("Subscribe modified the caller's request")
	}
}

func TestSubscribeWrapsWriteError(t *testing.T) {
	conn := &scriptedConn{}
	conn.Close() // the connection fails underneath a client that still holds it
	err := unsubscribeClient(conn).Subscribe(SubscriptionRequestMsg{Schema: "trades", Symbols: []string{"AAPL"}})
	if !errors.Is(err, net.ErrClosed) {
		t.Errorf("Subscribe error = %v, want it to wrap net.ErrClosed", err)
	}
}
