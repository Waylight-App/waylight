package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"

	"waylight/internal/platform/demo"
)

// demoSettings set up the example integrations of the showcase. They are
// sent through the app's own API, so every module validates and stores them
// the way it would for a user; they point at the fixtures of
// demo.Transport. Saving the same settings again changes nothing, so the
// demo can send them after every reset.
var demoSettings = []struct {
	path string
	body any
}{
	{"/api/adguard/settings", map[string]any{"url": demo.AdGuardURL, "username": "admin", "password": "demo-only"}},
	{"/api/media/settings", map[string]any{"provider": "jellyfin", "url": demo.JellyfinURL, "token": "demo-only-token"}},
	{"/api/proxmox/settings", map[string]any{"url": demo.ProxmoxURL, "tokenId": "monitor@pve!waylight",
		"secret": "00000000-0000-4000-8000-000000000000"}},
	{"/api/immich/settings", map[string]any{"url": demo.ImmichURL, "apiKey": "demo-only-key"}},
	{"/api/smarthome/settings", map[string]any{"url": demo.HomeAssistantURL, "token": "demo-only-token"}},
	{"/api/nextcloud/settings", map[string]any{"url": demo.NextcloudURL, "username": "anna", "appPassword": "demo-only-app-password"}},
	{"/api/fritzbox/settings", map[string]any{"url": demo.FritzBoxURL, "username": "waylight", "password": "demo-only"}},
	{"/api/fuel/settings", map[string]any{"apiKey": demo.TankerkoenigKey, "lat": 53.1435, "lon": 8.2146,
		"place": "Oldenburg", "radiusKm": 5, "fuel": "e10"}},
	{"/api/speedtest/settings", map[string]any{"url": demo.SpeedtestURL, "token": "demo-only-token", "minDownMbps": 150, "minUpMbps": 20}},
	{"/api/space/settings", map[string]any{"launches": true, "picture": true}},
	{"/api/gamedeals/settings", map[string]any{"steamId": demo.DemoSteamID, "country": "DE", "minDiscount": 20, "epic": true}},
	{"/api/parcels/settings", map[string]any{"dhlApiKey": demo.DHLAPIKey, "hideDeliveredDays": 7}},
}

// apiClient sends requests through the app's handler as the signed-in demo
// account.
type apiClient struct {
	ctx     context.Context
	handler http.Handler
	cookies []*http.Cookie
	csrf    string
}

func (c *apiClient) call(method, path string, body any) ([]byte, error) {
	payload := bytes.NewReader(nil)
	if body != nil {
		b, _ := json.Marshal(body) // plain maps, cannot fail
		payload = bytes.NewReader(b)
	}
	req := httptest.NewRequestWithContext(c.ctx, method, path, payload)
	req.RemoteAddr = "127.0.0.1:1"
	req.Header.Set("Content-Type", "application/json")
	if c.csrf != "" {
		req.Header.Set("X-CSRF-Token", c.csrf)
	}
	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}
	rec := httptest.NewRecorder()
	c.handler.ServeHTTP(rec, req)
	if rec.Code >= 300 {
		return nil, fmt.Errorf("%s %s: %d %s", method, path, rec.Code, bytes.TrimSpace(rec.Body.Bytes()))
	}
	c.cookies = append(c.cookies, rec.Result().Cookies()...)
	return rec.Body.Bytes(), nil
}

// ensure creates an entry with POST unless the list at path already has one
// whose field key equals want; it returns the entry's ID either way.
func (c *apiClient) ensure(path, key, want string, body any) (string, error) {
	out, err := c.call(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	var list []map[string]any
	if err := json.Unmarshal(out, &list); err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	for _, e := range list {
		if e[key] == want {
			id, _ := e["id"].(string)
			return id, nil
		}
	}
	out, err = c.call(http.MethodPost, path, body)
	if err != nil {
		return "", err
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(out, &created) // an empty ID is reported by the caller
	return created.ID, nil
}

// seedDemoIntegrations signs in to the demo account and sets up the example
// integrations; it only adds what is missing.
func seedDemoIntegrations(ctx context.Context, handler http.Handler) error {
	c := &apiClient{ctx: ctx, handler: handler}
	out, err := c.call(http.MethodPost, "/api/auth/login", map[string]string{"password": demo.Password})
	if err != nil {
		return err
	}
	var session struct {
		CSRFToken string `json:"csrfToken"`
	}
	if err := json.Unmarshal(out, &session); err != nil {
		return fmt.Errorf("demo sign-in: %w", err)
	}
	c.csrf = session.CSRFToken
	for _, s := range demoSettings {
		if _, err := c.call(http.MethodPut, s.path, s.body); err != nil {
			return fmt.Errorf("demo integration: %w", err)
		}
	}
	for _, in := range []struct{ app, name, url, key string }{
		{"sonarr", "Sonarr", demo.SonarrURL, "0123456789abcdef0123456789abcdef"},
		{"radarr", "Radarr", demo.RadarrURL, "fedcba9876543210fedcba9876543210"},
	} {
		if _, err := c.ensure("/api/arr/instances", "name", in.name,
			map[string]any{"app": in.app, "name": in.name, "url": in.url, "apiKey": in.key}); err != nil {
			return fmt.Errorf("demo integration: %w", err)
		}
	}
	for _, a := range demo.DemoArtists {
		if _, err := c.ensure("/api/releases/artists", "id", a.ID, map[string]any{"id": a.ID, "name": a.Name}); err != nil {
			return fmt.Errorf("demo integration: %w", err)
		}
	}
	if _, err := c.ensure("/api/calendar/subscriptions", "name", "Family", map[string]any{
		"name": "Family", "url": demo.FamilyCalendarURL, "color": "#e9a41f", "refreshMinutes": 60}); err != nil {
		return fmt.Errorf("demo integration: %w", err)
	}
	for _, p := range demo.DHLParcels {
		if _, err := c.ensure("/api/parcels", "trackingNumber", p.Number,
			map[string]any{"trackingNumber": p.Number, "carrier": "dhl", "label": p.Label}); err != nil {
			return fmt.Errorf("demo integration: %w", err)
		}
	}
	if err := c.ensurePhotos(); err != nil {
		return fmt.Errorf("demo integration: %w", err)
	}
	endpointID, err := c.ensure("/api/apiwidget/endpoints", "name", "Balcony solar", map[string]any{
		"name": "Balcony solar", "url": demo.SolarURL, "intervalSeconds": 60,
		"extractions": []map[string]any{
			{"label": "Power now", "path": "inverter.power_w", "unit": "W"},
			{"label": "Today", "path": "inverter.today_kwh", "unit": "kWh"},
			{"label": "Total", "path": "inverter.total_kwh", "unit": "kWh"},
			{"label": "Inverter", "path": "inverter.temperature_c", "unit": "°C", "warn": 55.0, "crit": 70.0},
		}})
	if err != nil {
		return fmt.Errorf("demo integration: %w", err)
	}
	if endpointID == "" {
		return fmt.Errorf("demo integration: the API widget endpoint has no ID")
	}
	return c.pointWidgets("apiwidget", map[string]string{"endpointId": endpointID})
}

// ensurePhotos uploads the example pictures of the photo frame when its
// album is empty.
func (c *apiClient) ensurePhotos() error {
	out, err := c.call(http.MethodGet, "/api/photos", nil)
	if err != nil {
		return err
	}
	var list []map[string]any
	if err := json.Unmarshal(out, &list); err != nil {
		return fmt.Errorf("read photos: %w", err)
	}
	if len(list) > 0 {
		return nil
	}
	for i, pic := range demo.Pictures() {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		part, _ := mw.CreateFormFile("file", fmt.Sprintf("demo-%d.png", i+1)) // in-memory buffer, cannot fail
		_, _ = part.Write(pic)
		_ = mw.Close()
		req := httptest.NewRequestWithContext(c.ctx, http.MethodPost, "/api/photos", &body)
		req.RemoteAddr = "127.0.0.1:1"
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req.Header.Set("X-CSRF-Token", c.csrf)
		for _, ck := range c.cookies {
			req.AddCookie(ck)
		}
		rec := httptest.NewRecorder()
		c.handler.ServeHTTP(rec, req)
		if rec.Code >= 300 {
			return fmt.Errorf("upload a demo photo: %d %s", rec.Code, bytes.TrimSpace(rec.Body.Bytes()))
		}
	}
	return nil
}

// pointWidgets sets the configuration of every showcase widget of kind; the
// API widget needs the ID of the endpoint created above, which the layout
// cannot know in advance.
func (c *apiClient) pointWidgets(kind string, config any) error {
	out, err := c.call(http.MethodGet, "/api/dashboard", nil)
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
			if _, err := c.call(http.MethodPatch, "/api/dashboard/widgets/"+w.ID, map[string]any{"config": config}); err != nil {
				return err
			}
		}
	}
	return nil
}
