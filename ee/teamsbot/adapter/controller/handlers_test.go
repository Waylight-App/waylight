package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/ee/teamsbot/adapter/botauth"
	"trisse-home/ee/teamsbot/domain"
)

type verifier struct {
	ok  bool
	got botauth.Activity
}

func (v *verifier) Verify(_ context.Context, auth string, a botauth.Activity) error {
	v.got = a
	if !v.ok || auth == "" {
		return errors.New("no")
	}
	return nil
}

type querier struct{ calls int }

func (q *querier) Query(_ context.Context, oid, query string) domain.Response {
	q.calls++
	return domain.Message(oid + ":" + query)
}

const invoke = `{"type":"invoke","name":"composeExtension/query","serviceUrl":"https://smba/","channelId":"msteams",
	"from":{"aadObjectId":"oid-1"},"value":{"commandId":"search","parameters":[{"name":"q","value":"vpn"}]}}`

func post(h *Handlers, body, auth string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/teams/messages", strings.NewReader(body))
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	h.Messages(rec, req)
	return rec
}

func TestNothingRunsWithoutAValidToken(t *testing.T) {
	q := &querier{}
	h := NewHandlers(&verifier{ok: false}, q)
	if rec := post(h, invoke, "Bearer forged"); rec.Code != http.StatusForbidden || q.calls != 0 {
		t.Fatalf("%d, %d queries", rec.Code, q.calls)
	}
	if rec := post(h, "{not json", "Bearer x"); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad json: %d", rec.Code)
	}
}

func TestQueryIsAnswered(t *testing.T) {
	q, v := &querier{}, &verifier{ok: true}
	h := NewHandlers(v, q)
	rec := post(h, invoke, "Bearer good")
	var res domain.Response
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if rec.Code != http.StatusOK || res.ComposeExtension.Text != "oid-1:vpn" || v.got.ServiceURL != "https://smba/" || v.got.ChannelID != "msteams" {
		t.Fatalf("%d %s %+v", rec.Code, rec.Body, v.got)
	}
	// Other activities are acknowledged without a search.
	if rec := post(h, `{"type":"message","serviceUrl":"https://smba/","channelId":"msteams"}`, "Bearer good"); rec.Code != http.StatusOK || q.calls != 1 {
		t.Fatalf("message: %d, %d queries", rec.Code, q.calls)
	}
}
