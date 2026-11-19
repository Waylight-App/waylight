package main

import (
	"net/http"
	"testing"
)

// TestSmartSearchAnswers checks the answers of the smart search on the
// server (ADR 0026): ranked, in the requested language, and only from
// modules the account may read (members never see the host or containers).
func TestSmartSearchAnswers(t *testing.T) {
	saved := membersEnabled
	membersEnabled = true
	t.Cleanup(func() { membersEnabled = saved })

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	type answersBody struct {
		Answers []obj `json:"answers"`
	}
	var res answersBody
	owner.json(http.MethodGet, "/api/search/answers?q=arbeitsspeicher&lang=de", nil, http.StatusOK, &res)
	if len(res.Answers) == 0 || res.Answers[0]["intent"] != "system.memory" || res.Answers[0]["title"] != "Arbeitsspeicher" ||
		res.Answers[0]["kind"] != "metric" || res.Answers[0]["widgetKind"] != "system" {
		t.Fatalf("memory answer: %v", res)
	}
	owner.do(http.MethodPost, "/api/services", obj{"name": "NAS", "url": "http://nas.local"}, http.StatusCreated)
	res = answersBody{}
	owner.json(http.MethodGet, "/api/search/answers?q=nas", nil, http.StatusOK, &res)
	if len(res.Answers) == 0 || res.Answers[0]["intent"] != "services.status" || res.Answers[0]["title"] != "NAS" {
		t.Fatalf("service answer: %v", res)
	}
	res = answersBody{}
	owner.json(http.MethodGet, "/api/search/answers?q=hallo", nil, http.StatusOK, &res)
	if len(res.Answers) != 0 {
		t.Fatalf("no answer expected: %v", res)
	}

	owner.do(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-2"}, http.StatusCreated)
	member := newClient(t, srv.URL)
	var ms obj
	member.json(http.MethodPost, "/api/auth/login", obj{"username": "ben", "password": "member-password-2"}, http.StatusOK, &ms)
	member.csrf = ms["csrfToken"].(string)
	res = answersBody{}
	member.json(http.MethodGet, "/api/search/answers?q=cpu", nil, http.StatusOK, &res)
	for _, a := range res.Answers {
		if a["intent"] == "system.cpu" {
			t.Fatalf("a member got the host's CPU: %v", res)
		}
	}
	res = answersBody{}
	member.json(http.MethodGet, "/api/search/answers?q=dienste&lang=en", nil, http.StatusOK, &res)
	if len(res.Answers) == 0 || res.Answers[0]["intent"] != "services.overview" {
		t.Fatalf("member services answer: %v", res)
	}
}
