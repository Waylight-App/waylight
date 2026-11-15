package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestAccountRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "calendar", Migrations())
	r := NewAccountRepository(db, upperCipher{})
	exp := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	a := domain.Account{ID: "a1", Provider: domain.ProviderGoogle, ClientID: "cid", Name: "me@example.org",
		Grant:     domain.Grant{ClientSecret: "sec", Access: "acc", Refresh: "ref", Expiry: exp},
		Calendars: []domain.AccountCalendar{{ID: "primary", Name: "Me", Color: "#112233", Writable: true, Enabled: true}}}
	if err := r.Save(ctx, a); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT refresh_encrypted FROM calendar_accounts`).Scan(&stored)
	if !strings.HasPrefix(stored, "enc:") {
		t.Fatalf("tokens must be stored encrypted, got %q", stored)
	}
	got, err := r.Get(ctx, "a1")
	if err != nil || got.Grant != a.Grant || got.Calendars[0] != a.Calendars[0] || got.Provider != domain.ProviderGoogle {
		t.Fatalf("%+v %v", got, err)
	}
	got.Grant.Access = "new"
	_ = r.Save(ctx, got)
	if list, _ := r.List(ctx); len(list) != 1 || list[0].Grant.Access != "new" {
		t.Fatalf("%+v", list)
	}
	if err := r.Delete(ctx, "a1"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Get(ctx, "a1"); !errors.Is(err, domain.ErrAccountNotFound) {
		t.Fatal(err)
	}
}

// fakeProviders answers like Google Calendar v3 and Microsoft Graph.
func fakeProviders(t *testing.T, created *[]string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	auth := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer tok" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			h(w, r)
		}
	}
	mux.HandleFunc("GET /google/calendars/primary", auth(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"id":"me@gmail.example"}`)
	}))
	mux.HandleFunc("GET /google/users/me/calendarList", auth(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"items":[{"id":"me@gmail.example","summary":"Me","backgroundColor":"#9FE1E7","accessRole":"owner"},
			{"id":"de.german#holiday@group.v.calendar.google.com","summary":"Holidays","backgroundColor":"red","accessRole":"reader"}]}`)
	}))
	mux.HandleFunc("GET /google/calendars/{cal}/events", auth(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("pageToken") == "" {
			_, _ = io.WriteString(w, `{"items":[{"id":"e1","summary":"Dentist","start":{"dateTime":"2026-11-16T09:00:00+01:00"},"end":{"dateTime":"2026-11-16T10:00:00+01:00"},"htmlLink":"https://calendar.google.com/x"},
				{"id":"e2","status":"cancelled","summary":"Gone","start":{"date":"2026-11-17"},"end":{"date":"2026-11-18"}}],"nextPageToken":"p2"}`)
			return
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"e3","summary":"Holiday","start":{"date":"2026-11-18"},"end":{"date":"2026-11-19"}}]}`)
	}))
	mux.HandleFunc("POST /google/calendars/{cal}/events", auth(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*created = append(*created, r.PathValue("cal")+" "+string(b))
	}))
	mux.HandleFunc("GET /graph/me", auth(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"mail":"","userPrincipalName":"me@contoso.example","displayName":"Me"}`)
	}))
	mux.HandleFunc("GET /graph/me/calendars", auth(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"value":[{"id":"AAMk=","name":"Calendar","hexColor":"#E07907","canEdit":true}]}`)
	}))
	mux.HandleFunc("GET /graph/me/calendars/{cal}/calendarView", auth(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Prefer") != `outlook.timezone="UTC"` {
			t.Errorf("Graph must be asked for UTC times")
		}
		_, _ = io.WriteString(w, `{"value":[{"id":"m1","subject":"Standup","start":{"dateTime":"2026-11-16T08:00:00.0000000","timeZone":"UTC"},"end":{"dateTime":"2026-11-16T08:15:00.0000000","timeZone":"UTC"},"isAllDay":false,"location":{"displayName":"Room 1"}},
			{"id":"m2","subject":"Off","isCancelled":true,"start":{"dateTime":"2026-11-17T00:00:00.0000000"},"end":{"dateTime":"2026-11-18T00:00:00.0000000"}}],
			"@odata.nextLink":"https://evil.example/steal"}`)
	}))
	mux.HandleFunc("POST /graph/me/calendars/{cal}/events", auth(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*created = append(*created, r.PathValue("cal")+" "+string(b))
	}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestAccountAPI(t *testing.T) {
	ctx := context.Background()
	var created []string
	srv := fakeProviders(t, &created)
	api := &AccountAPI{client: srv.Client(), GoogleURL: srv.URL + "/google", GraphURL: srv.URL + "/graph"}
	from, to := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC), time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC)

	if name, err := api.Profile(ctx, domain.ProviderGoogle, "tok"); err != nil || name != "me@gmail.example" {
		t.Fatalf("%q %v", name, err)
	}
	cals, err := api.Calendars(ctx, domain.ProviderGoogle, "tok")
	if err != nil || len(cals) != 2 || !cals[0].Writable || cals[1].Writable || cals[0].Color != "#9fe1e7" || cals[1].Color != "" {
		t.Fatalf("%+v %v", cals, err)
	}
	events, err := api.Events(ctx, domain.ProviderGoogle, "tok", cals[1].ID, from, to)
	if err != nil || len(events) != 2 || events[0].Summary != "Dentist" || !events[1].AllDay || events[1].Start.Format(time.DateOnly) != "2026-11-18" {
		t.Fatalf("pages, cancelled and all-day: %+v %v", events, err)
	}
	if _, err := api.Events(ctx, domain.ProviderGoogle, "wrong", "x", from, to); err == nil {
		t.Fatal("a refused token must be an error")
	}

	if name, _ := api.Profile(ctx, domain.ProviderMicrosoft, "tok"); name != "me@contoso.example" {
		t.Fatalf("%q", name)
	}
	ms, _ := api.Events(ctx, domain.ProviderMicrosoft, "tok", "AAMk=", from, to)
	if len(ms) != 1 || ms[0].Location != "Room 1" || ms[0].Start != time.Date(2026, 11, 16, 8, 0, 0, 0, time.UTC) {
		t.Fatalf("graph events (and the foreign nextLink must not be followed): %+v", ms)
	}

	day := domain.Event{Summary: "Trip", Start: time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC), AllDay: true}
	if err := api.Create(ctx, domain.ProviderGoogle, "tok", "me@gmail.example", day); err != nil {
		t.Fatal(err)
	}
	if err := api.Create(ctx, domain.ProviderMicrosoft, "tok", "AAMk=", day); err != nil {
		t.Fatal(err)
	}
	var g map[string]map[string]string
	_ = json.Unmarshal([]byte(strings.SplitN(created[0], " ", 2)[1]), &g)
	if g["start"]["date"] != "2026-11-20" || g["end"]["date"] != "2026-11-22" {
		t.Fatalf("google all-day body: %s", created[0])
	}
	if !strings.Contains(created[1], `"isAllDay":true`) || !strings.Contains(created[1], `"dateTime":"2026-11-20T00:00:00"`) {
		t.Fatalf("graph body: %s", created[1])
	}
}
