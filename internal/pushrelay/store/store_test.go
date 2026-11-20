package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"waylight/internal/pushrelay"
)

func TestBindings(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "relay.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h := pushrelay.TokenHash{1}
	t0 := time.Unix(1_800_000_000, 0)
	if k, err := s.Bind(ctx, h, "a", t0); k != "a" || err != nil {
		t.Fatal(k, err)
	}
	if k, _ := s.Bind(ctx, h, "b", t0.Add(time.Hour)); k != "a" {
		t.Fatal("first instance must stay bound:", k)
	}
	// The use by b still counts as use, so the binding is not pruned.
	if n, _ := s.Prune(ctx, t0.Add(30*time.Minute)); n != 0 {
		t.Fatal("pruned a used binding")
	}
	if n, _ := s.Prune(ctx, t0.Add(2*time.Hour)); n != 1 {
		t.Fatal("unused binding not pruned")
	}
	_, _ = s.Bind(ctx, h, "c", t0)
	if err := s.Delete(ctx, h); err != nil {
		t.Fatal(err)
	}
	if k, _ := s.Bind(ctx, h, "d", t0); k != "d" {
		t.Fatal("deleted binding must be free:", k)
	}
}

func TestMailbox(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "relay.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Unix(1_800_000_000, 0)
	for i := range 4 {
		if _, err := s.Put(ctx, "dev", []byte{byte(i)}, now, 3); err != nil {
			t.Fatal(err)
		}
	}
	_, _ = s.Put(ctx, "other", []byte{9}, now.Add(-100*time.Hour), 3)
	events, err := s.After(ctx, "dev", 0)
	if err != nil || len(events) != 3 || events[0].Payload[0] != 1 {
		t.Fatalf("only the newest three are kept: %+v, %v", events, err)
	}
	if err := s.Ack(ctx, "dev", events[1].ID); err != nil {
		t.Fatal(err)
	}
	rest, _ := s.After(ctx, "dev", 0)
	if len(rest) != 1 || rest[0].Payload[0] != 3 {
		t.Fatalf("after ack: %+v", rest)
	}
	n, err := s.PruneMailbox(ctx, now.Add(-72*time.Hour))
	if err != nil || n != 1 {
		t.Fatalf("pruned %d, %v", n, err)
	}
}
