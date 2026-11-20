package httpapi

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"waylight/internal/pushrelay"
	"waylight/internal/pushrelay/stream"
)

const (
	// heartbeat keeps proxies and mobile networks from closing an idle
	// stream and lets the device notice a dead connection.
	heartbeat = 25 * time.Second
	// writeWait bounds one write; a device that does not read is dropped.
	writeWait = 10 * time.Second
	// retryMillis is the reconnect delay the stream suggests to clients.
	retryMillis = 5000
)

// Streams are the parts of the Server-Sent Events delivery (ADR 0027).
type Streams struct {
	Mailbox stream.Mailbox
	Hub     *stream.Hub
	Gate    *stream.Gate
	// Opens limits how often one client address may open a stream.
	Opens pushrelay.Limiter
}

// stream handles GET /v1/stream: the device authenticates, gets the
// messages after Last-Event-ID and then every new one until it disconnects.
func (h *Handler) stream(w http.ResponseWriter, r *http.Request) {
	start := h.Now()
	status := h.serveStream(w, r)
	// Only the outcome is logged, never device IDs or payloads (ADR 0025).
	h.Log.Info("stream", "status", status, "duration", h.Now().Sub(start))
}

func (h *Handler) serveStream(w http.ResponseWriter, r *http.Request) int {
	device, err := stream.Verify(r.Header.Get("Authorization"), h.Now())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return http.StatusUnauthorized
	}
	client := h.ClientIP(r)
	if !h.Streams.Opens.Allow(client, h.Now()) || !h.Streams.Gate.Acquire(client) {
		w.Header().Set("Retry-After", "60")
		http.Error(w, "too many streams", http.StatusTooManyRequests)
		return http.StatusTooManyRequests
	}
	defer h.Streams.Gate.Release(client)

	ctx := r.Context()
	last, _ := strconv.ParseInt(r.Header.Get("Last-Event-ID"), 10, 64)
	if last > 0 {
		// The device saw everything up to last; those messages are done.
		if err := h.Streams.Mailbox.Ack(ctx, device, last); err != nil {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return http.StatusServiceUnavailable
		}
	}
	sub, cancel := h.Streams.Hub.Subscribe(device)
	defer cancel()

	rc := http.NewResponseController(w)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	if err := write(rc, w, fmt.Sprintf("retry: %d\n\n", retryMillis)); err != nil {
		return http.StatusOK
	}

	tick := time.NewTicker(heartbeat)
	defer tick.Stop()
	for {
		events, err := h.Streams.Mailbox.After(ctx, device, last)
		if err != nil {
			return http.StatusOK
		}
		for _, e := range events {
			if err := write(rc, w, fmt.Sprintf("id: %d\ndata: %s\n\n", e.ID, base64.StdEncoding.EncodeToString(e.Payload))); err != nil {
				return http.StatusOK
			}
			last = e.ID
		}
		select {
		case <-ctx.Done():
			return http.StatusOK
		case <-sub.Replaced:
			return http.StatusOK
		case <-sub.Wake:
		case <-tick.C:
			if err := write(rc, w, ": ping\n\n"); err != nil {
				return http.StatusOK
			}
		}
	}
}

// write sends one chunk with its own deadline, which replaces the server's
// write timeout for this long-lived response, and flushes it.
func write(rc *http.ResponseController, w http.ResponseWriter, chunk string) error {
	if err := rc.SetWriteDeadline(time.Now().Add(writeWait)); err != nil && !errors.Is(err, http.ErrNotSupported) {
		return err
	}
	if _, err := w.Write([]byte(chunk)); err != nil {
		return err
	}
	return rc.Flush()
}
