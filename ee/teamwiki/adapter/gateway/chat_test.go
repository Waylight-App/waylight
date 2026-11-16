package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestSlackSearchesWithTheMembersTokenAndFallsBackToPublic(t *testing.T) {
	var asked []string
	srv := fake(t, map[string]func(http.ResponseWriter, *http.Request){
		"POST /assistant.search.context": func(w http.ResponseWriter, r *http.Request) {
			_ = r.ParseForm()
			asked = append(asked, r.Form.Get("channel_types"))
			if strings.Contains(r.Form.Get("channel_types"), "im") {
				_, _ = w.Write([]byte(`{"ok":false,"error":"missing_scope"}`))
				return
			}
			_, _ = w.Write([]byte(`{"ok":true,"results":{"messages":[{"author_name":"Ben","channel_name":"release",
				"message_ts":"1763290000.000100","content":"we ship the  release on Friday","permalink":"https://x.slack.com/archives/C1/p1"}]}}`))
		},
	})
	pages, err := Slack{Client: srv.Client(), API: srv.URL}.Search(context.Background(), "member-token", "release", 5)
	if err != nil || len(pages) != 1 {
		t.Fatalf("%+v %v", pages, err)
	}
	p := pages[0]
	if p.Title != "#release · Ben" || p.Excerpt != "we ship the release on Friday" || p.URL != "https://x.slack.com/archives/C1/p1" ||
		p.Updated.Unix() != 1763290000 {
		t.Fatalf("page %+v", p)
	}
	if len(asked) != 2 || asked[0] != allChannels || asked[1] != "public_channel" {
		t.Fatalf("channel types asked: %v", asked)
	}
}

func TestSlackErrorsAreFailures(t *testing.T) {
	srv := fake(t, map[string]func(http.ResponseWriter, *http.Request){
		"POST /assistant.search.context": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"ok":false,"error":"ratelimited"}`))
		},
	})
	if _, err := (Slack{Client: srv.Client(), API: srv.URL}).Search(context.Background(), "member-token", "x", 5); err == nil {
		t.Fatal("no error for ok:false")
	}
}

func TestTeamsSearchesChatMessagesOnly(t *testing.T) {
	var body map[string]any
	srv := fake(t, map[string]func(http.ResponseWriter, *http.Request){
		"POST /search/query": func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&body)
			_, _ = w.Write([]byte(`{"value":[{"hitsContainers":[{"hits":[
				{"summary":"...the <c0>budget</c0> is approved","resource":{"subject":"","webLink":"https://teams.example/m1",
				 "createdDateTime":"2026-11-14T07:01:01Z","from":{"emailAddress":{"name":"Goncalo Torres"}},"channelIdentity":{}}},
				{"summary":"budget thread","resource":{"subject":"Q1","webUrl":"https://teams.example/m2",
				 "from":{"emailAddress":{"name":"Anna"}},"channelIdentity":{"channelId":"19:abc"}}}]}]}]}`))
		},
	})
	pages, err := Teams{Client: srv.Client(), API: srv.URL}.Search(context.Background(), "member-token", "budget", 5)
	if err != nil || len(pages) != 2 {
		t.Fatalf("%+v %v", pages, err)
	}
	if pages[0].Title != "Goncalo Torres" || pages[0].Space != "Chat" || pages[0].Excerpt != "...the budget is approved" || pages[0].URL != "https://teams.example/m1" {
		t.Fatalf("chat hit %+v", pages[0])
	}
	if pages[1].Title != "Q1 · Anna" || pages[1].Space != "Channel" || pages[1].URL != "https://teams.example/m2" {
		t.Fatalf("channel hit %+v", pages[1])
	}
	if got := mustJSON(body); !strings.Contains(got, `"entityTypes":["chatMessage"]`) {
		t.Fatalf("request %s", got)
	}
}
