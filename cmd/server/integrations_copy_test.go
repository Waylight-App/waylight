package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"testing"

	"waylight/internal/modules/integrations/adapter/gateway"
	"waylight/internal/modules/integrations/domain"
	"waylight/internal/platform/crypto"
)

// TestCopyModuleSettings runs the copy of #610 against the real module
// tables: settings as the modules store them (encrypted with SECRET_KEY,
// with their empty optional fields) become connections with the same
// address and credential, the module tables stay byte for byte the same,
// and a second run changes nothing.
func TestCopyModuleSettings(t *testing.T) {
	ctx := context.Background()
	up := fakeUpstream(t)
	cfg := newTestConfig(t, up.URL)
	a, err := build(ctx, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer a.db.Close()
	cipher, err := crypto.NewAESGCM(cfg.SecretKey)
	if err != nil {
		t.Fatal(err)
	}
	enc := func(s string) string {
		out, err := cipher.Encrypt(s)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	fixtures := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO immich_settings VALUES (1, 'http://192.168.178.20:2283', ?)`, []any{enc("imm-Key-AbC123")}},
		{`INSERT INTO grafana_settings VALUES (1, 'https://grafana.example.org', ?)`, []any{enc("glsa_token_1")}},
		{`INSERT INTO nextcloud_settings VALUES (1, 'https://cloud.example.org/nextcloud', 'anna', ?)`, []any{enc("Xk3pQ-app-pass")}},
		{`INSERT INTO proxmox_settings VALUES (1, 'https://192.168.178.5:8006', 'waylight@pve!dash', ?, '')`, []any{enc("1b2c3d4e-uuid")}},
		{`INSERT INTO smarthome_settings VALUES (1, 'http://homeassistant.local:8123', ?)`, []any{enc("eyJ.long.lived")}},
		{`INSERT INTO speedtest_settings VALUES (1, 'http://speedtest.lan', ?, 100, 20)`, []any{enc("st-token")}},
		{`INSERT INTO fritzbox_settings VALUES (1, 'http://fritz.box:49000', '', '')`, nil},
		{`INSERT INTO media_settings VALUES (1, 'jellyfin', 'http://jellyfin.lan:8096', ?)`, []any{enc("jf-key")}},
		{`INSERT INTO adguard_connections VALUES ('adguard', 'http://adguard.lan', 'admin', ?)`, []any{enc("ag-pass")}},
		{`INSERT INTO adguard_connections VALUES ('pihole', 'http://pi.hole', '', ?)`, []any{enc("ph-pass")}},
		{`INSERT INTO arr_instances VALUES ('s1', 'sonarr', 'Sonarr', 'http://nas:8989', ?, 0)`, []any{enc("son-key")}},
		{`INSERT INTO arr_instances VALUES ('s2', 'sonarr', 'Sonarr 4K', 'http://nas:8990', ?, 1)`, []any{enc("son4k-key")}},
		{`INSERT INTO fuel_settings VALUES (1, ?, 52.5, 13.4, 'Berlin', 5, 'e5', '[]')`, []any{enc("00000000-0000-0000-0000-000000000002")}},
	}
	for _, f := range fixtures {
		if _, err := a.db.ExecContext(ctx, f.query, f.args...); err != nil {
			t.Fatalf("%s: %v", f.query, err)
		}
	}
	tables := []string{"immich_settings", "grafana_settings", "nextcloud_settings", "proxmox_settings", "smarthome_settings",
		"speedtest_settings", "fritzbox_settings", "media_settings", "adguard_connections", "arr_instances", "fuel_settings"}
	before := dumpTables(t, a, tables)

	run := copyJob(t, a)
	if err := run(ctx); err != nil {
		t.Fatal(err)
	}
	repo := gateway.NewRepository(a.db, cipher)
	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]domain.Connection{}
	for _, c := range list {
		got[c.Service+"/"+c.Name] = c
	}
	want := map[string]struct {
		url  string
		cred domain.Credential
	}{
		"immich/Immich":                {"http://192.168.178.20:2283", domain.Credential{"apiKey": "imm-Key-AbC123"}},
		"grafana/Grafana":              {"https://grafana.example.org", domain.Credential{"token": "glsa_token_1"}},
		"nextcloud/Nextcloud":          {"https://cloud.example.org/nextcloud", domain.Credential{"username": "anna", "appPassword": "Xk3pQ-app-pass"}},
		"proxmox/Proxmox VE":           {"https://192.168.178.5:8006", domain.Credential{"tokenId": "waylight@pve!dash", "secret": "1b2c3d4e-uuid"}},
		"homeassistant/Home Assistant": {"http://homeassistant.local:8123", domain.Credential{"token": "eyJ.long.lived"}},
		"speedtest/Speedtest Tracker":  {"http://speedtest.lan", domain.Credential{"token": "st-token"}},
		"fritzbox/FRITZ!Box":           {"http://fritz.box:49000", domain.Credential{}},
		"jellyfin/Jellyfin":            {"http://jellyfin.lan:8096", domain.Credential{"token": "jf-key"}},
		"adguard/AdGuard Home":         {"http://adguard.lan", domain.Credential{"username": "admin", "password": "ag-pass"}},
		"pihole/Pi-hole":               {"http://pi.hole", domain.Credential{"password": "ph-pass"}},
		"sonarr/Sonarr":                {"http://nas:8989", domain.Credential{"apiKey": "son-key"}},
		"sonarr/Sonarr 4K":             {"http://nas:8990", domain.Credential{"apiKey": "son4k-key"}},
		"tankerkoenig/Tankerkönig":     {"", domain.Credential{"apiKey": "00000000-0000-0000-0000-000000000002"}},
	}
	if len(got) != len(want) {
		t.Errorf("got %d connections, want %d: %v", len(got), len(want), slices.Sorted(maps.Keys(got)))
	}
	for k, w := range want {
		c, ok := got[k]
		if !ok {
			t.Errorf("%s: missing", k)
			continue
		}
		if (w.url != "" && c.URL != w.url) || !maps.Equal(c.Credential, w.cred) {
			t.Errorf("%s: url %q, fields %v", k, c.URL, c.Credential.Fields())
		}
	}
	if after := dumpTables(t, a, tables); !maps.Equal(before, after) {
		t.Error("the module settings must stay untouched")
	}
	var stored string
	_ = a.db.QueryRowContext(ctx, `SELECT group_concat(credential_encrypted) FROM integrations_connections`).Scan(&stored)
	if strings.Contains(stored, "imm-Key-AbC123") || strings.Contains(stored, "ph-pass") {
		t.Error("credentials must be stored encrypted")
	}

	// The next start copies nothing again.
	if err := run(ctx); err != nil {
		t.Fatal(err)
	}
	if again, _ := repo.List(ctx); len(again) != len(list) {
		t.Fatalf("second run: %d connections, want %d", len(again), len(list))
	}
}

func copyJob(t *testing.T, a *app) func(context.Context) error {
	t.Helper()
	for _, m := range a.registry.All() {
		if m.Name() != "integrations" {
			continue
		}
		for _, j := range m.Jobs() {
			if j.Name == "integrations-copy-settings" {
				return j.Run
			}
		}
	}
	t.Fatal("the integrations module has no copy job")
	return nil
}

// dumpTables reads every row of the tables as text, to compare before and
// after the copy.
func dumpTables(t *testing.T, a *app, tables []string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, tb := range tables {
		out[tb] = dumpTable(t, a, tb)
	}
	return out
}

func dumpTable(t *testing.T, a *app, table string) string {
	t.Helper()
	rows, err := a.db.Query(`SELECT * FROM ` + table + ` ORDER BY rowid`)
	if err != nil {
		t.Fatalf("%s: %v", table, err)
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var b strings.Builder
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		fmt.Fprintf(&b, "%v\n", vals)
	}
	if err := rows.Err(); err != nil || b.Len() == 0 {
		t.Fatalf("%s: %v, or the fixture is missing", table, err)
	}
	return b.String()
}
