package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"waylight/internal/modules/notifications/usecase"
	"waylight/internal/shared/httpapi"
)

// Timing of the live stream (#583).
const (
	// streamPing keeps proxies from closing an idle stream.
	streamPing = 25 * time.Second
	// streamLifetime ends a stream regularly, so the browser reconnects
	// and the session is checked again (a signed-out admin loses it).
	streamLifetime = 15 * time.Minute
	// streamWrite bounds one write; the server's write timeout would
	// otherwise end every stream after a few seconds.
	streamWrite = 10 * time.Second
	// streamRetry is the reconnect delay the browser uses, in ms.
	streamRetry = 5000
)

// LiveEventDTO is the data of one event on the stream.
type LiveEventDTO struct {
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	Severity string    `json:"severity"`
	Type     string    `json:"type"`
	At       time.Time `json:"at"`
}

// StreamHandlers serve the live notifications of the web interface.
type StreamHandlers struct{ live usecase.Live }

// NewStreamHandlers returns the handlers.
func NewStreamHandlers(live usecase.Live) *StreamHandlers { return &StreamHandlers{live: live} }

// Stream handles GET /api/notifications/stream: Server-Sent Events with
// the messages of the browser channel. A reconnect sends Last-Event-ID and
// gets the events it missed, as far as the backlog reaches.
func (h *StreamHandlers) Stream(w http.ResponseWriter, r *http.Request) {
	backlog, events, cancel, err := h.live.Subscribe(r.Context(), r.Header.Get("Last-Event-ID"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	defer cancel()

	rc := http.NewResponseController(w)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	// nginx and similar proxies buffer responses unless told otherwise.
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	write := func(chunk string) bool {
		_ = rc.SetWriteDeadline(time.Now().Add(streamWrite))
		if _, err := fmt.Fprint(w, chunk); err != nil {
			return false
		}
		return rc.Flush() == nil
	}
	send := func(e usecase.LiveEvent) bool {
		data, err := json.Marshal(LiveEventDTO{Title: e.Message.Title, Body: e.Message.Body,
			Severity: string(e.Message.Severity), Type: e.Message.Type, At: e.At})
		if err != nil {
			return false
		}
		return write(fmt.Sprintf("id: %s\nevent: notification\ndata: %s\n\n", e.ID, data))
	}

	if !write(fmt.Sprintf("retry: %d\n\n", streamRetry)) {
		return
	}
	for _, e := range backlog {
		if !send(e) {
			return
		}
	}
	ping := time.NewTicker(streamPing)
	defer ping.Stop()
	end := time.NewTimer(streamLifetime)
	defer end.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-end.C:
			return
		case <-ping.C:
			if !write(": ping\n\n") {
				return
			}
		case e, ok := <-events:
			if !ok || !send(e) {
				return
			}
		}
	}
}
