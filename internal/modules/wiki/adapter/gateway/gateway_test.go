package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"waylight/internal/modules/wiki/domain"
	"waylight/internal/platform/database/dbtest"
)

// The fixtures follow the documented response shapes of BookStack's
// /api/search, Outline's documents.search and collections.list, and the
// Wiki.js 2 GraphQL search.

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "wiki", Migrations())
	r := NewRepository(db, prefixCipher{})
	w := domain.Wiki{ID: "w1", Name: "Handbook", Kind: domain.KindBookStack, URL: "https://wiki.example.org", Token: "id:secret"}
	if err := r.Save(ctx, w); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT token_encrypted FROM wiki_connections`).Scan(&stored)
	if stored != "enc:id:secret" {
		t.Fatalf("token must be stored encrypted, got %q", stored)
	}
	if got, err := r.Get(ctx, "w1"); err != nil || got != w {
		t.Fatalf("%+v %v", got, err)
	}
	if err := r.Delete(ctx, "w1"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Get(ctx, "w1"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBookStack(t *testing.T) {
	var auth, query string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, query = r.Header.Get("Authorization"), r.URL.Query().Get("query")
		// Links in the fixture point to wiki.example.org; rewrite them to this server.
		_, _ = w.Write([]byte(strings.ReplaceAll(string(fixture(t, "bookstack-search.json")), "https://wiki.example.org", "http://"+r.Host)))
	}))
	defer srv.Close()
	pages, err := BookStack{client: srv.Client()}.Search(context.Background(), domain.Wiki{URL: srv.URL, Token: "abc:def"}, "vpn", 5)
	if err != nil {
		t.Fatal(err)
	}
	if auth != "Token abc:def" || query != "vpn" {
		t.Fatalf("request: %q %q", auth, query)
	}
	// The second result points to another host and is dropped.
	if len(pages) != 1 || pages[0].Title != "VPN setup" || pages[0].Space != "IT handbook" ||
		pages[0].Excerpt != "How to connect to the office VPN from home & abroad" || pages[0].Updated.IsZero() {
		t.Fatalf("%+v", pages)
	}
}

func TestOutline(t *testing.T) {
	var collectionCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer ol_key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/documents.search":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["query"] != "printer" {
				t.Errorf("query %v", body)
			}
			_, _ = w.Write(fixture(t, "outline-search.json"))
		case "/api/collections.list":
			collectionCalls.Add(1)
			_, _ = w.Write(fixture(t, "outline-collections.json"))
		}
	}))
	defer srv.Close()
	now := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	s := &Outline{client: srv.Client(), now: func() time.Time { return now }, collections: map[string]collectionCache{}}
	w := domain.Wiki{ID: "o1", URL: srv.URL, Token: "ol_key"}
	for range 2 {
		pages, err := s.Search(context.Background(), w, "printer", 5)
		if err != nil || len(pages) != 1 || pages[0].Space != "Office" || pages[0].URL != srv.URL+"/doc/printers-ab12cd" || pages[0].Excerpt != "Printer on the second floor needs A4" {
			t.Fatalf("%+v %v", pages, err)
		}
	}
	if collectionCalls.Load() != 1 {
		t.Fatalf("collection names are cached, got %d calls", collectionCalls.Load())
	}
	if _, err := s.Search(context.Background(), domain.Wiki{ID: "o2", URL: srv.URL, Token: "wrong"}, "x", 5); err == nil || !strings.Contains(err.Error(), "refused") {
		t.Fatalf("refused key: %v", err)
	}
}

func TestWikiJS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if r.URL.Path != "/graphql" || !strings.Contains(body.Query, "search(query: $q)") || body.Variables["q"] == "" {
			t.Errorf("request %s %+v", r.URL.Path, body)
		}
		if body.Variables["q"] == "boom" {
			_, _ = w.Write([]byte(`{"errors":[{"message":"Forbidden"}]}`))
			return
		}
		_, _ = w.Write(fixture(t, "wikijs-search.json"))
	}))
	defer srv.Close()
	pages, err := WikiJS{client: srv.Client()}.Search(context.Background(), domain.Wiki{URL: srv.URL, Token: "k"}, "backup", 5)
	if err != nil || len(pages) != 1 || pages[0].URL != srv.URL+"/en/ops/backups" || pages[0].Space != "en" || pages[0].Excerpt != "Where the backups go" {
		t.Fatalf("%+v %v", pages, err)
	}
	if _, err := (WikiJS{client: srv.Client()}).Search(context.Background(), domain.Wiki{URL: srv.URL, Token: "k"}, "boom", 5); err == nil {
		t.Fatal("GraphQL errors must be reported")
	}
}

func TestAbsoluteKeepsTheWikiHost(t *testing.T) {
	base := "https://wiki.example.org/sub"
	for link, want := range map[string]string{
		"/doc/a":                         "https://wiki.example.org/doc/a",
		"https://wiki.example.org/doc/b": "https://wiki.example.org/doc/b",
		"https://evil.example/doc":       "",
		"javascript:alert(1)":            "",
	} {
		if got := absolute(base, link); got != want {
			t.Errorf("%q: got %q, want %q", link, got, want)
		}
	}
}
