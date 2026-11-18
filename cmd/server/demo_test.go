package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"waylight/internal/platform/config"
	"waylight/internal/platform/demo"
)

func demoConfig(t *testing.T, dbPath string) config.Config {
	t.Helper()
	return config.Config{
		DatabasePath: dbPath, SecretKey: make([]byte, config.SecretKeyBytes), SessionTTL: time.Hour,
		DemoMode: true, DemoResetInterval: time.Hour, SampleData: true,
		SampleWeather: config.SampleLocation{Name: "Berlin", Lat: 52.52, Lon: 13.41},
		SampleFeeds:   []string{"https://go.dev/blog/feed.atom"},
		HostProcPath:  fakeProc(t), HostSysPath: t.TempDir(), DiskPaths: []string{"Temp=" + t.TempDir()},
		DockerProxyURL: "http://docker-socket-proxy:2375",
		BackupDir:      t.TempDir(), BackupInterval: 24 * time.Hour, BackupRetention: 1,
	}
}

func TestDemoMode(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	a, err := build(ctx, demoConfig(t, filepath.Join(t.TempDir(), "demo.db")), log)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.db.Close() })
	reset, err := prepareDemo(ctx, a)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(newHandler(a))
	t.Cleanup(srv.Close)
	c := newClient(t, srv.URL)

	var status obj
	c.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &status)
	if status["demo"] != true || status["setupRequired"] != false || status["demoPassword"] != demo.Password {
		t.Fatalf("status: %v", status)
	}
	// Visitors sign in on the sign-in page with the password it shows.
	var session obj
	c.json(http.MethodPost, "/api/auth/login", obj{"password": demo.Password}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	// Visitors share the account, so nothing about them is logged for others.
	var events []obj
	c.json(http.MethodGet, "/api/auth/events", nil, http.StatusOK, &events)
	if len(events) != 0 {
		t.Fatalf("the demo keeps a security log: %v", events)
	}

	var dash struct {
		Tabs []obj `json:"tabs"`
	}
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	if len(dash.Tabs) == 0 {
		t.Fatal("sample data missing in demo")
	}
	// Fixture adapters answer instead of the internet and the Docker socket.
	c.do(http.MethodGet, "/api/docker/containers", nil, http.StatusOK)

	// Account changes are refused (credentials: see demo.ReadOnlySecrets).
	c.do(http.MethodPut, "/api/auth/password", obj{"currentPassword": "x", "newPassword": "y"}, http.StatusForbidden)
	c.do(http.MethodPost, "/api/auth/totp/setup", nil, http.StatusForbidden)

	// Visitors' changes disappear with the next reset.
	c.do(http.MethodPost, "/api/dashboard/tabs", obj{"name": "Visitor"}, http.StatusCreated)
	for range 2 { // the first run (at start) is skipped
		if err := reset.Run(ctx); err != nil {
			t.Fatal(err)
		}
	}
	var after struct {
		Tabs []obj `json:"tabs"`
	}
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &after)
	if len(after.Tabs) != len(dash.Tabs) {
		t.Fatalf("reset did not restore the tabs: %d -> %d", len(dash.Tabs), len(after.Tabs))
	}

	// The example integrations are set up once; resets add no duplicates.
	for path, want := range map[string]int{"/api/arr/instances": 2, "/api/releases/artists": len(demo.Artists),
		"/api/apiwidget/endpoints": 1, "/api/calendar/subscriptions": 1} {
		var list []obj
		c.json(http.MethodGet, path, nil, http.StatusOK, &list)
		if len(list) != want {
			t.Errorf("%s: %d entries after the resets, want %d", path, len(list), want)
		}
	}
	// Visitors still cannot store credentials of their own.
	c.do(http.MethodPut, "/api/immich/settings", obj{"url": demo.ImmichURL, "apiKey": "visitor-key"}, http.StatusForbidden)

	// Signing out hands the demo to the next visitor.
	c.do(http.MethodPost, "/api/auth/logout", nil, http.StatusNoContent)
}

func TestDemoModeRefusesProductionDatabase(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	path := filepath.Join(t.TempDir(), "prod.db")
	prod := demoConfig(t, path)
	prod.DemoMode = false
	a, err := build(ctx, prod, log)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.db.ExecContext(ctx, `INSERT INTO auth_accounts (id, role, name, password_hash, created_at, updated_at) VALUES ('owner', 'owner', 'owner', 'x', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	_ = a.db.Close()
	if _, err := build(ctx, demoConfig(t, path), log); err == nil {
		t.Fatal("demo mode started on a production database")
	}
}
