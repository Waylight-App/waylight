package httpapi

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"waylight/internal/pushrelay"
	"waylight/internal/pushrelay/store"
	"waylight/internal/pushrelay/stream"
	"waylight/internal/shared/pushproto"
)

// relayServer runs the relay with the real store and stream delivery.
func relayServer(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "relay.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	hub := stream.NewHub()
	win := func() *pushrelay.Window { return &pushrelay.Window{Max: 100, Period: time.Hour} }
	relay := pushrelay.New(map[pushrelay.Platform]pushrelay.Sender{pushrelay.SSE: &stream.Sender{Mailbox: st, Hub: hub, Now: time.Now}},
		st, pushrelay.Limits{Instance: win(), Token: win(), Client: win()}, time.Now)
	h := &Handler{
		Relay: relay, ClientIP: func(*http.Request) string { return "ip" }, Now: time.Now,
		Log:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		Streams: &Streams{Mailbox: st, Hub: hub, Gate: &stream.Gate{MaxPerClient: 2, MaxTotal: 10}, Opens: win()},
	}
	srv := httptest.NewServer(h.Routes())
	t.Cleanup(srv.Close)
	return srv, st
}

// openStream connects as the device and returns a reader of its events.
func openStream(t *testing.T, srv *httptest.Server, key *ecdsa.PrivateKey, lastID string) (*bufio.Reader, func()) {
	t.Helper()
	auth, err := stream.Authorization(key, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/v1/stream", nil)
	req.Header.Set("Authorization", auth)
	if lastID != "" {
		req.Header.Set("Last-Event-ID", lastID)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK || res.Header.Get("Content-Type") != "text/event-stream" {
		cancel()
		t.Fatalf("stream: %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
	return bufio.NewReader(res.Body), func() { cancel(); _ = res.Body.Close() }
}

// nextEvent reads lines until a complete event with data arrives.
func nextEvent(t *testing.T, r *bufio.Reader) (id, data string) {
	t.Helper()
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		line = strings.TrimRight(line, "\n")
		switch {
		case strings.HasPrefix(line, "id: "):
			id = strings.TrimPrefix(line, "id: ")
		case strings.HasPrefix(line, "data: "):
			data = strings.TrimPrefix(line, "data: ")
		case line == "" && data != "":
			return id, data
		}
	}
}

func TestStreamDeliversAndAcknowledges(t *testing.T) {
	srv, st := relayServer(t)
	device, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	spki, _ := x509.MarshalPKIXPublicKey(&device.PublicKey)
	token := stream.DeviceID(spki)
	_, instance, _ := ed25519.GenerateKey(rand.Reader)
	push := func(payload []byte) int {
		body, _ := json.Marshal(map[string]any{"platform": "sse", "token": token, "payload": base64.StdEncoding.EncodeToString(payload)})
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/push", bytes.NewReader(body))
		req.Header.Set("Authorization", pushproto.Authorization(instance, body, time.Now()))
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		return res.StatusCode
	}

	// A message sent while the device is offline waits in the mailbox.
	if code := push([]byte("first")); code != http.StatusAccepted {
		t.Fatalf("push: %d", code)
	}
	r, closeStream := openStream(t, srv, device, "")
	id1, data := nextEvent(t, r)
	if want := base64.StdEncoding.EncodeToString([]byte("first")); data != want {
		t.Fatalf("data = %q", data)
	}
	// A message sent while the stream is open arrives at once.
	if code := push([]byte("second")); code != http.StatusAccepted {
		t.Fatalf("push: %d", code)
	}
	id2, _ := nextEvent(t, r)
	closeStream()

	// Reconnecting with the last ID acknowledges both; nothing is left.
	r2, closeSecond := openStream(t, srv, device, id2)
	defer closeSecond()
	_ = r2
	waitFor(t, func() bool {
		left, _ := st.After(context.Background(), token, 0)
		return len(left) == 0
	})
	if id1 == id2 {
		t.Fatal("event IDs must grow")
	}
}

func TestStreamRefusesBadSignatures(t *testing.T) {
	srv, _ := relayServer(t)
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/v1/stream", nil)
	req.Header.Set("Authorization", "Waylight-Device key=AA, ts=1, sig=AA")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func waitFor(t *testing.T, ok func() bool) {
	t.Helper()
	for range 100 {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met")
}
