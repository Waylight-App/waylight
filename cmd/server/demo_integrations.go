package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"waylight/internal/platform/demo"
)

// demoIntegration is one request that sets up an example integration of the
// showcase, sent through the app's own API so every module validates and
// stores it the way it would for a user.
type demoIntegration struct {
	method, path string
	body         any
}

// demoIntegrations point at the fixtures of demo.Transport; the example
// credentials are stored only while demo.Seed runs.
var demoIntegrations = []demoIntegration{
	{http.MethodPut, "/api/adguard/settings", map[string]any{"url": demo.AdGuardURL, "username": "admin", "password": "demo-only"}},
	{http.MethodPut, "/api/media/settings", map[string]any{"provider": "jellyfin", "url": demo.JellyfinURL, "token": "demo-only-token"}},
	{http.MethodPut, "/api/proxmox/settings", map[string]any{"url": demo.ProxmoxURL, "tokenId": "monitor@pve!waylight",
		"secret": "00000000-0000-4000-8000-000000000000"}},
	{http.MethodPut, "/api/immich/settings", map[string]any{"url": demo.ImmichURL, "apiKey": "demo-only-key"}},
	{http.MethodPut, "/api/smarthome/settings", map[string]any{"url": demo.HomeAssistantURL, "token": "demo-only-token"}},
	{http.MethodPut, "/api/nextcloud/settings", map[string]any{"url": demo.NextcloudURL, "username": "anna", "appPassword": "demo-only-app-password"}},
	{http.MethodPost, "/api/arr/instances", map[string]any{"app": "sonarr", "name": "Sonarr", "url": demo.SonarrURL, "apiKey": "0123456789abcdef0123456789abcdef"}},
	{http.MethodPost, "/api/arr/instances", map[string]any{"app": "radarr", "name": "Radarr", "url": demo.RadarrURL, "apiKey": "fedcba9876543210fedcba9876543210"}},
	{http.MethodPut, "/api/fritzbox/settings", map[string]any{"url": demo.FritzBoxURL, "username": "waylight", "password": "demo-only"}},
	{http.MethodPut, "/api/fuel/settings", map[string]any{"apiKey": demo.TankerkoenigKey, "lat": 53.1435, "lon": 8.2146,
		"place": "Oldenburg", "radiusKm": 5, "fuel": "e10"}},
	{http.MethodPut, "/api/speedtest/settings", map[string]any{"url": demo.SpeedtestURL, "token": "demo-only-token", "minDownMbps": 150, "minUpMbps": 20}},
	{http.MethodPut, "/api/space/settings", map[string]any{"launches": true, "picture": true}},
	{http.MethodPut, "/api/gamedeals/settings", map[string]any{"steamId": demo.DemoSteamID, "country": "DE", "minDiscount": 20, "epic": true}},
	{http.MethodPost, "/api/releases/artists", map[string]any{"id": demo.DemoArtists[0].ID, "name": demo.DemoArtists[0].Name}},
	{http.MethodPost, "/api/releases/artists", map[string]any{"id": demo.DemoArtists[1].ID, "name": demo.DemoArtists[1].Name}},
	{http.MethodPost, "/api/releases/artists", map[string]any{"id": demo.DemoArtists[2].ID, "name": demo.DemoArtists[2].Name}},
	{http.MethodPost, "/api/apiwidget/endpoints", map[string]any{"name": "Balcony solar", "url": demo.SolarURL, "intervalSeconds": 60,
		"extractions": []map[string]any{
			{"label": "Power now", "path": "inverter.power_w", "unit": "W"},
			{"label": "Today", "path": "inverter.today_kwh", "unit": "kWh"},
			{"label": "Total", "path": "inverter.total_kwh", "unit": "kWh"},
			{"label": "Inverter", "path": "inverter.temperature_c", "unit": "°C", "warn": 55.0, "crit": 70.0},
		}}},
}

// seedDemoIntegrations signs in to the demo account and sends
// demoIntegrations through handler.
func seedDemoIntegrations(ctx context.Context, handler http.Handler) error {
	var cookies []*http.Cookie
	csrf := ""
	call := func(method, path string, body any) ([]byte, error) {
		var payload *bytes.Reader
		if body != nil {
			b, _ := json.Marshal(body) // plain maps, cannot fail
			payload = bytes.NewReader(b)
		} else {
			payload = bytes.NewReader(nil)
		}
		req := httptest.NewRequestWithContext(ctx, method, path, payload)
		req.RemoteAddr = "127.0.0.1:1"
		req.Header.Set("Content-Type", "application/json")
		if csrf != "" {
			req.Header.Set("X-CSRF-Token", csrf)
		}
		for _, c := range cookies {
			req.AddCookie(c)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code >= 300 {
			return nil, fmt.Errorf("%s %s: %d %s", method, path, rec.Code, bytes.TrimSpace(rec.Body.Bytes()))
		}
		cookies = append(cookies, rec.Result().Cookies()...)
		return rec.Body.Bytes(), nil
	}
	out, err := call(http.MethodPost, "/api/auth/login", map[string]string{"password": demo.Password})
	if err != nil {
		return err
	}
	var session struct {
		CSRFToken string `json:"csrfToken"`
	}
	if err := json.Unmarshal(out, &session); err != nil {
		return fmt.Errorf("demo sign-in: %w", err)
	}
	csrf = session.CSRFToken
	endpointID := ""
	for _, in := range demoIntegrations {
		out, err := call(in.method, in.path, in.body)
		if err != nil {
			return fmt.Errorf("demo integration: %w", err)
		}
		if in.path == "/api/apiwidget/endpoints" {
			var created struct {
				ID string `json:"id"`
			}
			_ = json.Unmarshal(out, &created) // checked below
			endpointID = created.ID
		}
	}
	if endpointID == "" {
		return fmt.Errorf("demo integration: the API widget endpoint has no ID")
	}
	return pointWidgets(call, "apiwidget", map[string]string{"endpointId": endpointID})
}

// pointWidgets sets the configuration of every showcase widget of kind; the
// API widget needs the ID of the endpoint created above, which the layout
// cannot know in advance.
func pointWidgets(call func(method, path string, body any) ([]byte, error), kind string, config any) error {
	out, err := call(http.MethodGet, "/api/dashboard", nil)
	if err != nil {
		return err
	}
	var dash struct {
		Tabs []struct {
			Widgets []struct {
				ID   string `json:"id"`
				Kind string `json:"kind"`
			} `json:"widgets"`
		} `json:"tabs"`
	}
	if err := json.Unmarshal(out, &dash); err != nil {
		return fmt.Errorf("read the demo dashboard: %w", err)
	}
	for _, t := range dash.Tabs {
		for _, w := range t.Widgets {
			if w.Kind != kind {
				continue
			}
			if _, err := call(http.MethodPatch, "/api/dashboard/widgets/"+w.ID, map[string]any{"config": config}); err != nil {
				return err
			}
		}
	}
	return nil
}
