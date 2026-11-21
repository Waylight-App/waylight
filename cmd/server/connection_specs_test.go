package main

import (
	"context"
	"io"
	"log/slog"
	"regexp"
	"slices"
	"testing"

	"waylight/internal/modules"
)

// TestConnectionSpecs keeps the services of named connections (#610) in
// step with the modules: the identifiers are stored with every connection
// and copied from the module settings, so renaming one is a migration.
func TestConnectionSpecs(t *testing.T) {
	up := fakeUpstream(t)
	a, err := build(context.Background(), newTestConfig(t, up.URL), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer a.db.Close()
	field := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]{0,31}$`)
	var got []string
	for _, s := range modules.ConnectionSpecsOf(a.registry.All()) {
		if slices.Contains(got, s.Service) {
			t.Errorf("%s: declared twice", s.Service)
		}
		got = append(got, s.Service)
		if s.Title == "" || s.Check == nil || len(s.Fields) == 0 || len(s.Fields) > 4 {
			t.Errorf("%s: title, check and one to four fields are required: %+v", s.Service, s)
		}
		secret := false
		for _, f := range s.Fields {
			secret = secret || f.Secret
			if !field.MatchString(f.Name) {
				t.Errorf("%s: invalid field %q", s.Service, f.Name)
			}
		}
		if !secret {
			t.Errorf("%s: a credential without a secret field belongs in the module settings", s.Service)
		}
	}
	slices.Sort(got)
	want := []string{"adguard", "fritzbox", "grafana", "homeassistant", "immich", "jellyfin", "nextcloud", "pihole",
		"plex", "prowlarr", "proxmox", "radarr", "sonarr", "speedtest", "tankerkoenig"}
	if !slices.Equal(got, want) {
		t.Errorf("services:\n got  %v\n want %v", got, want)
	}
}
