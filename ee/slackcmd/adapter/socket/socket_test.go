package socket

import (
	"context"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

// fakeSlack plays Socket Mode: hello, one slash command, then a
// disconnect on the first connection; the reply comes back as the ack.
func TestSocketModeAnswersAndReconnects(t *testing.T) {
	acks := make(chan ack, 4)
	var conns atomic.Int32
	srv := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		n := conns.Add(1)
		_ = websocket.JSON.Send(ws, map[string]any{"type": "hello"})
		if n == 1 {
			_ = websocket.JSON.Send(ws, map[string]any{"type": "slash_commands", "envelope_id": "e1", "accepts_response_payload": true,
				"payload": map[string]any{"team_id": "T1", "user_id": "U1", "command": "/home", "text": "vpn"}})
			var a struct {
				EnvelopeID string  `json:"envelope_id"`
				Payload    message `json:"payload"`
			}
			if err := websocket.JSON.Receive(ws, &a); err == nil {
				acks <- ack{EnvelopeID: a.EnvelopeID, Payload: a.Payload}
			}
			_ = websocket.JSON.Send(ws, map[string]any{"type": "disconnect", "reason": "refresh_requested"})
			return
		}
		var ignore map[string]any
		_ = websocket.JSON.Receive(ws, &ignore) // hold the second connection open
	}))
	defer srv.Close()
	addr := "ws" + strings.TrimPrefix(srv.URL, "http")

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	var warmed atomic.Int32
	go func() {
		_ = Run(ctx, func(context.Context) (string, error) { return addr, nil },
			func(_ context.Context, c Command) string { return "found " + c.Text + " for " + c.UserID },
			func(context.Context) { warmed.Add(1) })
		close(done)
	}()
	select {
	case a := <-acks:
		m := a.Payload.(message)
		if a.EnvelopeID != "e1" || m.Text != "found vpn for U1" || m.ResponseType != "ephemeral" {
			t.Fatalf("ack %+v", a)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no ack")
	}
	deadline := time.Now().Add(5 * time.Second)
	for conns.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if conns.Load() < 2 {
		t.Fatal("no reconnect after the disconnect")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop with its context")
	}
	if warmed.Load() < 1 {
		t.Fatal("onConnect not called")
	}
}
