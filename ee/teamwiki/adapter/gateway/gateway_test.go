package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func fake(t *testing.T, routes map[string]func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for p, h := range routes {
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer member-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			h(w, r)
		})
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestConfluenceSearchesEverySiteWithTheMembersToken(t *testing.T) {
	var gotCQL string
	srv := fake(t, map[string]func(http.ResponseWriter, *http.Request){
		"GET /oauth/token/accessible-resources": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`[{"id":"site-1","name":"Example"}]`))
		},
		"GET /ex/confluence/site-1/wiki/rest/api/search": func(w http.ResponseWriter, r *http.Request) {
			gotCQL = r.URL.Query().Get("cql")
			_, _ = w.Write([]byte(`{"results":[{"content":{"title":"Holiday rules"},"excerpt":"how to @@@hl@@@book@@@endhl@@@ days",
				"url":"/spaces/HR/pages/1","lastModified":"2026-11-10T08:00:00.000Z","resultGlobalContainer":{"title":"HR"}}],
				"_links":{"base":"https://example.atlassian.net/wiki"}}`))
		},
	})
	pages, err := Confluence{Client: srv.Client(), API: srv.URL}.Search(context.Background(), "member-token", `book "days"`, 5)
	if err != nil || len(pages) != 1 {
		t.Fatalf("%v %v", pages, err)
	}
	p := pages[0]
	if p.Title != "Holiday rules" || p.URL != "https://example.atlassian.net/wiki/spaces/HR/pages/1" || p.Space != "HR" ||
		p.Excerpt != "how to book days" || p.Updated.IsZero() {
		t.Fatalf("page %+v", p)
	}
	if gotCQL != `siteSearch ~ "book \"days\"" AND type in (page, blogpost)` {
		t.Fatalf("cql %q", gotCQL)
	}
}

func TestGraphSearchesDriveAndListItems(t *testing.T) {
	var body map[string]any
	srv := fake(t, map[string]func(http.ResponseWriter, *http.Request){
		"POST /search/query": func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&body)
			_, _ = w.Write([]byte(`{"value":[{"hitsContainers":[{"hits":[{"summary":"the <c0>budget</c0> for 2027",
				"resource":{"name":"Budget 2027.xlsx","webUrl":"https://example.sharepoint.com/b","lastModifiedDateTime":"2026-11-01T10:00:00Z"}}]}]}]}`))
		},
	})
	pages, err := Graph{Client: srv.Client(), API: srv.URL}.Search(context.Background(), "member-token", "budget", 5)
	if err != nil || len(pages) != 1 || pages[0].Title != "Budget 2027.xlsx" || pages[0].Excerpt != "the budget for 2027" {
		t.Fatalf("%+v %v", pages, err)
	}
	if !strings.Contains(mustJSON(body), `"driveItem"`) {
		t.Fatalf("request %v", body)
	}
}

func TestNotionReadsTheTitleProperty(t *testing.T) {
	var version string
	srv := fake(t, map[string]func(http.ResponseWriter, *http.Request){
		"POST /search": func(w http.ResponseWriter, r *http.Request) {
			version = r.Header.Get("Notion-Version")
			_, _ = w.Write([]byte(`{"results":[{"url":"https://notion.so/p1","last_edited_time":"2026-11-12T09:00:00.000Z",
				"properties":{"Name":{"type":"title","title":[{"plain_text":"Onboarding "},{"plain_text":"guide"}]},"Tags":{"type":"multi_select"}}}]}`))
		},
	})
	pages, err := Notion{Client: srv.Client(), API: srv.URL}.Search(context.Background(), "member-token", "onboarding", 5)
	if err != nil || len(pages) != 1 || pages[0].Title != "Onboarding guide" || version != notionVersion {
		t.Fatalf("%+v %v (version %q)", pages, err, version)
	}
}

func TestAnUnauthorisedAnswerIsAnError(t *testing.T) {
	srv := fake(t, map[string]func(http.ResponseWriter, *http.Request){"POST /search": func(http.ResponseWriter, *http.Request) {}})
	if _, err := (Notion{Client: srv.Client(), API: srv.URL}).Search(context.Background(), "someone-else", "x", 5); err == nil {
		t.Fatal("no error for 401")
	}
}

func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }
