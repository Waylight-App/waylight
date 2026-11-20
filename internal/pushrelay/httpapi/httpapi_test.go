package httpapi

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"waylight/internal/pushrelay"
	"waylight/internal/shared/pushproto"
)

type sender struct{ err error }

func (s sender) Send(context.Context, string, []byte) error { return s.err }

type bindings map[pushrelay.TokenHash]string

func (b bindings) Bind(_ context.Context, h pushrelay.TokenHash, k string, _ time.Time) (string, error) {
	if v, ok := b[h]; ok {
		return v, nil
	}
	b[h] = k
	return k, nil
}
func (b bindings) Delete(_ context.Context, h pushrelay.TokenHash) error { delete(b, h); return nil }

func TestPush(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	clock := func() time.Time { return now }
	win := func() *pushrelay.Window { return &pushrelay.Window{Max: 100, Period: time.Hour} }
	relay := pushrelay.New(map[pushrelay.Platform]pushrelay.Sender{pushrelay.SSE: sender{}, pushrelay.APNs: sender{err: pushrelay.ErrUnregistered}},
		bindings{}, pushrelay.Limits{Instance: win(), Token: win(), Client: win()}, clock)
	h := (&Handler{Relay: relay, ClientIP: func(*http.Request) string { return "ip" }, Now: clock,
		Log: slog.New(slog.NewTextHandler(io.Discard, nil))}).Routes()
	_, keyA, _ := ed25519.GenerateKey(rand.Reader)
	_, keyB, _ := ed25519.GenerateKey(rand.Reader)
	post := func(key ed25519.PrivateKey, body map[string]any, sign bool) int {
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/v1/push", bytes.NewReader(b))
		if sign {
			req.Header.Set("Authorization", pushproto.Authorization(key, b, now))
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	p := base64.StdEncoding.EncodeToString([]byte{1, 2, 3})
	cases := []struct {
		name string
		key  ed25519.PrivateKey
		body map[string]any
		sign bool
		want int
	}{
		{"sent", keyA, map[string]any{"platform": "sse", "token": "t1", "payload": p}, true, http.StatusAccepted},
		{"unsigned", keyA, map[string]any{"platform": "sse", "token": "t1", "payload": p}, false, http.StatusUnauthorized},
		{"other instance", keyB, map[string]any{"platform": "sse", "token": "t1", "payload": p}, true, http.StatusForbidden},
		{"unknown platform", keyA, map[string]any{"platform": "mail", "token": "t1", "payload": p}, true, http.StatusBadRequest},
		{"bad payload", keyA, map[string]any{"platform": "sse", "token": "t1", "payload": "%%"}, true, http.StatusBadRequest},
		{"too large", keyA, map[string]any{"platform": "sse", "token": "t1", "payload": base64.StdEncoding.EncodeToString(make([]byte, 3001))}, true, http.StatusRequestEntityTooLarge},
		{"unregistered", keyA, map[string]any{"platform": "apns", "token": "t2", "payload": p}, true, http.StatusGone},
	}
	for _, c := range cases {
		if got := post(c.key, c.body, c.sign); got != c.want {
			t.Errorf("%s: %d, want %d", c.name, got, c.want)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/push", bytes.NewReader(make([]byte, maxBody+1))))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Error("oversized body:", rec.Code)
	}
}
