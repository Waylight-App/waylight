package controller

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/notifications/domain"
	"waylight/internal/modules/notifications/usecase"
	"waylight/internal/shared/kernel"
)

type nowClock struct{}

func (nowClock) Now() time.Time { return time.Unix(1_800_000_000, 0) }

func TestStream(t *testing.T) {
	live := usecase.NewLive(nowClock{})
	h := NewStreamHandlers(live)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := kernel.WithPrincipal(r.Context(), kernel.Principal{AccountID: "acc", Role: "admin"})
		h.Stream(w, r.WithContext(ctx))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK || res.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("%d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
	lines := bufio.NewScanner(res.Body)
	next := func() string {
		if !lines.Scan() {
			t.Fatal("stream ended", lines.Err())
		}
		return lines.Text()
	}
	if got := next(); got != "retry: 5000" {
		t.Fatalf("first line %q", got)
	}
	next()

	// The subscription exists once the first chunk arrived.
	_ = live.Publish(context.Background(), domain.Channel{}, domain.Message{Title: "Disk full", Body: "nas", Severity: "critical", Type: "service"})
	if id := next(); !strings.HasPrefix(id, "id: ") {
		t.Fatalf("id line %q", id)
	}
	if ev := next(); ev != "event: notification" {
		t.Fatalf("event line %q", ev)
	}
	if data := next(); !strings.Contains(data, `"title":"Disk full"`) || !strings.Contains(data, `"severity":"critical"`) {
		t.Fatalf("data line %q", data)
	}

	rec := httptest.NewRecorder()
	member := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "acc", Role: "member"})
	h.Stream(rec, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(member))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("member: %d", rec.Code)
	}
}
