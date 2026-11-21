package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/notifications/domain"
	"waylight/internal/shared/kernel"
)

func adminCtx() context.Context {
	return kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "acc", Role: "admin"})
}

func publish(t *testing.T, l Live, titles ...string) {
	t.Helper()
	for _, title := range titles {
		if err := l.Publish(context.Background(), domain.Channel{}, domain.Message{Title: title}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLiveOnlyForAdmins(t *testing.T) {
	l := NewLive(clk{})
	for _, role := range []string{"member", "guest", "kiosk", "token", "system"} {
		ctx := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "acc", Role: role})
		if _, _, _, err := l.Subscribe(ctx, ""); !errors.Is(err, kernel.ErrForbidden) && !errors.Is(err, kernel.ErrAdminOnly) {
			t.Errorf("%s: %v", role, err)
		}
	}
}

func TestLiveDeliversAndCatchesUp(t *testing.T) {
	l := NewLive(clk{})
	publish(t, l, "before")
	backlog, events, cancel, err := l.Subscribe(adminCtx(), "")
	if err != nil || len(backlog) != 0 {
		t.Fatalf("a new stream starts without backlog: %v %v", backlog, err)
	}
	publish(t, l, "a", "b")
	first := <-events
	second := <-events
	if first.Message.Title != "a" || second.Message.Title != "b" {
		t.Fatalf("got %q %q", first.Message.Title, second.Message.Title)
	}
	cancel()
	cancel() // idempotent
	if _, ok := <-events; ok {
		t.Fatal("cancel closes the stream")
	}

	publish(t, l, "c")
	backlog, _, cancel, _ = l.Subscribe(adminCtx(), first.ID)
	defer cancel()
	if len(backlog) != 2 || backlog[0].Message.Title != "b" || backlog[1].Message.Title != "c" {
		t.Fatalf("catch up after %s: %+v", first.ID, backlog)
	}
	if b, _, c, _ := l.Subscribe(adminCtx(), "otherboot-1"); len(b) != 0 {
		t.Fatalf("IDs of another run replay nothing: %+v", b)
	} else {
		c()
	}
}

func TestLiveClosesSlowStreamsAndBounds(t *testing.T) {
	l := NewLive(clk{})
	_, events, cancel, _ := l.Subscribe(adminCtx(), "")
	defer cancel()
	for range liveBuffer + 1 {
		publish(t, l, "x")
	}
	n := 0
	for range events {
		n++
	}
	if n != liveBuffer {
		t.Fatalf("a slow stream gets its buffer, then closes: %d", n)
	}

	var cancels []func()
	for range MaxLiveStreams {
		_, _, c, err := l.Subscribe(adminCtx(), "")
		if err != nil {
			t.Fatal(err)
		}
		cancels = append(cancels, c)
	}
	if _, _, _, err := l.Subscribe(adminCtx(), ""); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("beyond the bound: %v", err)
	}
	for _, c := range cancels {
		c()
	}

	// The backlog keeps the last liveBacklog events.
	_, _, c, _ := l.Subscribe(adminCtx(), "")
	c()
	for range liveBacklog + 5 {
		publish(t, l, "y")
	}
	h := l.(*liveHub)
	if len(h.ring) != liveBacklog {
		t.Fatalf("ring %d", len(h.ring))
	}
}
