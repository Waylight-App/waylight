package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"waylight/internal/platform/demo"
)

// The demo shows its showcase data, and the hourly reset brings it back
// after a visitor removed it, also in modules without an export.
func TestDemoShowcaseDataSurvivesTheReset(t *testing.T) {
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
	var session obj
	c.json(http.MethodPost, "/api/auth/login", obj{"password": demo.Password}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	var habits, drops []obj
	c.json(http.MethodGet, "/api/habits", nil, http.StatusOK, &habits)
	c.json(http.MethodGet, "/api/drop", nil, http.StatusOK, &drops)
	if len(habits) != 4 || len(drops) != 1 {
		t.Fatalf("showcase data missing: %d habits, %d drops", len(habits), len(drops))
	}

	// A visitor empties the drop zone, which has no export.
	c.do(http.MethodDelete, "/api/drop/"+drops[0]["id"].(string), nil, http.StatusNoContent)
	for range 2 { // the first run (at start) is skipped
		if err := reset.Run(ctx); err != nil {
			t.Fatal(err)
		}
	}
	c.json(http.MethodGet, "/api/drop", nil, http.StatusOK, &drops)
	c.json(http.MethodGet, "/api/habits", nil, http.StatusOK, &habits)
	if len(drops) != 1 || len(habits) != 4 {
		t.Fatalf("reset did not restore the showcase: %d drops, %d habits", len(drops), len(habits))
	}
}
