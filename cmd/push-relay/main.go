// Command push-relay delivers end-to-end encrypted notifications from
// Waylight instances to the mobile app: to Android over its own
// Server-Sent Events stream, to iOS through Apple APNs (ADR 0025, ADR 0027,
// docs/push-relay.md). It is configured through the environment.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"waylight/internal/platform/httpserver"
	"waylight/internal/platform/logging"
	"waylight/internal/platform/middleware"
	"waylight/internal/pushrelay"
	"waylight/internal/pushrelay/apns"
	"waylight/internal/pushrelay/httpapi"
	"waylight/internal/pushrelay/store"
	"waylight/internal/pushrelay/stream"
)

// bindingTTL is how long an unused token binding is kept (ADR 0025).
const bindingTTL = 90 * 24 * time.Hour

func main() {
	log := logging.New(os.Stderr, slog.LevelInfo, env("LOG_FORMAT", "json"))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, log)
	stop()
	if err != nil {
		log.Error("push relay stopped", "error", err)
		os.Exit(1)
	}
}

func env(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func run(ctx context.Context, log *slog.Logger) error {
	proxies, err := prefixes(os.Getenv("TRUSTED_PROXIES"))
	if err != nil {
		return err
	}
	st, err := store.Open(ctx, env("PUSH_RELAY_DB", "/data/push-relay.db"))
	if err != nil {
		return err
	}
	defer st.Close()
	go prune(ctx, st, log)

	hub := stream.NewHub()
	senders := map[pushrelay.Platform]pushrelay.Sender{
		pushrelay.SSE: &stream.Sender{Mailbox: st, Hub: hub, Now: time.Now},
	}
	if err := addAPNs(senders, &http.Client{Timeout: 10 * time.Second}); err != nil {
		return err
	}

	limit := func(max int) *pushrelay.Window { return &pushrelay.Window{Max: max, Period: time.Hour} }
	relay := pushrelay.New(senders, st, pushrelay.Limits{Instance: limit(600), Token: limit(120), Client: limit(1200)}, time.Now)
	h := &httpapi.Handler{
		Relay:    relay,
		ClientIP: func(r *http.Request) string { return middleware.ResolveClientIP(r, proxies).String() },
		Now:      time.Now,
		Log:      log,
		Streams: &httpapi.Streams{
			Mailbox: st,
			Hub:     hub,
			Gate:    &stream.Gate{MaxPerClient: 20, MaxTotal: 20000},
			// Reconnects back off on the device; this only stops floods.
			Opens: &pushrelay.Window{Max: 600, Period: time.Hour},
		},
	}
	platforms := make([]string, 0, len(senders))
	for p := range senders {
		platforms = append(platforms, string(p))
	}
	log.Info("push relay starting", "platforms", platforms)
	return httpserver.Run(ctx, env("PUSH_RELAY_ADDR", ":8080"), h.Routes(), 10*time.Second, log)
}

// addAPNs adds the APNs sender for the iOS app when a key is configured.
func addAPNs(senders map[pushrelay.Platform]pushrelay.Sender, client *http.Client) error {
	path := os.Getenv("APNS_KEY_FILE")
	if path == "" {
		return nil
	}
	key, err := os.ReadFile(path) //nolint:gosec // path set by the operator
	if err != nil {
		return err
	}
	base := apns.Production
	if os.Getenv("APNS_SANDBOX") == "true" {
		base = apns.Sandbox
	}
	s, err := apns.New(client, base, apns.Config{KeyID: os.Getenv("APNS_KEY_ID"), TeamID: os.Getenv("APNS_TEAM_ID"), Topic: os.Getenv("APNS_TOPIC"), Key: key}, time.Now)
	if err != nil {
		return err
	}
	senders[pushrelay.APNs] = s
	return nil
}

func prefixes(list string) ([]netip.Prefix, error) {
	var out []netip.Prefix
	for s := range strings.FieldsSeq(strings.ReplaceAll(list, ",", " ")) {
		p, err := netip.ParsePrefix(s)
		if err != nil {
			return nil, fmt.Errorf("TRUSTED_PROXIES: %w", err)
		}
		out = append(out, p)
	}
	return out, nil
}

// prune deletes unused bindings and expired mailbox messages every hour.
func prune(ctx context.Context, st *store.Store, log *slog.Logger) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		now := time.Now()
		if n, err := st.Prune(ctx, now.Add(-bindingTTL)); err != nil {
			log.Warn("prune bindings", "error", err)
		} else if n > 0 {
			log.Info("pruned bindings", "count", n)
		}
		if n, err := st.PruneMailbox(ctx, now.Add(-stream.TTL)); err != nil {
			log.Warn("prune mailbox", "error", err)
		} else if n > 0 {
			log.Info("pruned messages", "count", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
