//go:build ee

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"trisse-home/ee/licence/domain"
)

// TestTeamCalendar checks the team calendar of #299: a member shares a
// calendar subscription with a group, the group sees its events, nobody
// else does, the feed address never leaks, and without the licence the
// shared calendar disappears again.
func TestTeamCalendar(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	saved, savedMembers := licencePublicKey, membersEnabled
	licencePublicKey, membersEnabled = pub, true
	t.Cleanup(func() { licencePublicKey, membersEnabled = saved, savedMembers })

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)
	var anna, ben, carl obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "anna", "password": "member-password-1"}, http.StatusCreated, &anna)
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-2"}, http.StatusCreated, &ben)
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "carl", "password": "member-password-3"}, http.StatusCreated, &carl)
	var office obj
	owner.json(http.MethodPost, "/api/auth/groups", obj{"name": "Office", "accounts": []string{anna["id"].(string), ben["id"].(string)}},
		http.StatusCreated, &office)
	login := func(name, pw string) *client {
		c := newClient(t, srv.URL)
		var s obj
		c.json(http.MethodPost, "/api/auth/login", obj{"username": name, "password": pw}, http.StatusOK, &s)
		c.csrf = s["csrfToken"].(string)
		return c
	}
	a, b, c := login("anna", "member-password-1"), login("ben", "member-password-2"), login("carl", "member-password-3")

	var sub obj
	a.json(http.MethodPost, "/api/calendar/subscriptions", obj{"name": "Office holidays", "url": up.URL + "/holidays.ics?token=secret"},
		http.StatusCreated, &sub)
	hasHoliday := func(cl *client) bool {
		res, err := cl.http.Get(srv.URL + "/api/calendar/events?days=7")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode == http.StatusNotFound {
			return false // no calendar at all
		}
		var agenda struct{ Events []obj }
		if err := json.NewDecoder(res.Body).Decode(&agenda); err != nil || res.StatusCode != http.StatusOK {
			t.Fatalf("agenda: %d %v", res.StatusCode, err)
		}
		for _, e := range agenda.Events {
			if e["calendar"] == "Office holidays" {
				return true
			}
		}
		return false
	}

	// Without the licence: no sharing.
	a.do(http.MethodPut, "/api/teamcalendar/subscriptions/"+sub["id"].(string), obj{"visibility": "group:" + office["id"].(string)},
		http.StatusForbidden)
	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 5,
		Features: []string{"absences"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)

	// Only with one's own groups; not someone else's subscription.
	c.do(http.MethodPut, "/api/teamcalendar/subscriptions/"+sub["id"].(string), obj{"visibility": "everyone"}, http.StatusNotFound)
	a.do(http.MethodPut, "/api/teamcalendar/subscriptions/"+sub["id"].(string), obj{"visibility": "group:nope"}, http.StatusUnprocessableEntity)
	a.do(http.MethodPut, "/api/teamcalendar/subscriptions/"+sub["id"].(string), obj{"visibility": "group:" + office["id"].(string)},
		http.StatusNoContent)
	var own []obj
	a.json(http.MethodGet, "/api/teamcalendar/subscriptions", nil, http.StatusOK, &own)
	if len(own) != 1 || !strings.HasPrefix(own[0]["visibility"].(string), "group:") {
		t.Fatalf("own subscriptions: %v", own)
	}

	var accounts []obj
	a.json(http.MethodGet, "/api/teamcalendar/accounts", nil, http.StatusOK, &accounts)
	if len(accounts) != 0 {
		t.Fatalf("connected accounts: %v", accounts)
	}
	a.do(http.MethodPut, "/api/teamcalendar/accounts/nope", obj{"visibility": "everyone"}, http.StatusNotFound)

	if !hasHoliday(b) {
		t.Fatal("ben (Office) does not see the shared calendar")
	}
	if hasHoliday(c) {
		t.Fatal("carl (not in Office) sees the shared calendar")
	}
	if body := string(b.do(http.MethodGet, "/api/calendar/subscriptions", nil, http.StatusOK)); strings.Contains(body, "secret") ||
		strings.Contains(body, "Office holidays") {
		t.Fatalf("ben's own subscription list shows anna's: %s", body)
	}

	// Removing the key hides it at once: shared calendars are never part of
	// the cached agenda.
	owner.do(http.MethodDelete, "/api/licence", nil, http.StatusNoContent)
	if hasHoliday(b) {
		t.Fatal("the shared calendar still shows without the licence")
	}
}
