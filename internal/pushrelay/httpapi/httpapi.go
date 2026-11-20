// Package httpapi is the HTTP interface of the push relay: POST /v1/push for
// instances, GET /v1/stream for devices and a health check.
package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"waylight/internal/pushrelay"
	"waylight/internal/shared/pushproto"
)

// maxBody bounds a request: a base64 payload of pushproto.MaxPayload bytes
// plus a device token and the field names.
const maxBody = 8 << 10

// Handler serves the relay.
type Handler struct {
	Relay *pushrelay.Relay
	// ClientIP returns the client address, behind a proxy the forwarded one.
	ClientIP func(*http.Request) string
	Now      func() time.Time
	Log      *slog.Logger
	// Streams serves GET /v1/stream; nil turns the endpoint off.
	Streams *Streams
}

type request struct {
	Platform pushrelay.Platform `json:"platform"`
	Token    string             `json:"token"`
	Payload  string             `json:"payload"`
}

// Routes returns the mux.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/push", h.push)
	if h.Streams != nil {
		mux.HandleFunc("GET /v1/stream", h.stream)
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	return mux
}

func (h *Handler) push(w http.ResponseWriter, r *http.Request) {
	start := h.Now()
	status := h.serve(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": http.StatusText(status)})
	// Only the outcome is logged, never keys, tokens or payloads (ADR 0025).
	h.Log.Info("push", "status", status, "duration", h.Now().Sub(start))
}

func (h *Handler) serve(r *http.Request) int {
	body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, maxBody))
	if err != nil {
		return http.StatusRequestEntityTooLarge
	}
	key, err := pushproto.Verify(r.Header.Get("Authorization"), body, h.Now())
	if err != nil {
		return http.StatusUnauthorized
	}
	var req request
	if json.Unmarshal(body, &req) != nil {
		return http.StatusBadRequest
	}
	payload, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		return http.StatusBadRequest
	}
	err = h.Relay.Send(r.Context(), pushproto.KeyID(key), h.ClientIP(r), pushrelay.Message{Platform: req.Platform, Token: req.Token, Payload: payload})
	switch {
	case err == nil:
		return http.StatusAccepted
	case errors.Is(err, pushrelay.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, pushrelay.ErrTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, pushrelay.ErrForeignToken):
		return http.StatusForbidden
	case errors.Is(err, pushrelay.ErrRateLimited):
		return http.StatusTooManyRequests
	case errors.Is(err, pushrelay.ErrUnregistered):
		return http.StatusGone
	default:
		h.Log.Warn("push failed", "error", errors.Unwrap(err))
		return http.StatusBadGateway
	}
}
