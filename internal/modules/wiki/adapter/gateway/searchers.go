package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"trisse-home/internal/modules/wiki/domain"
	"trisse-home/internal/modules/wiki/usecase"
)

// maxResponseBytes bounds a search answer; a page of results is far smaller.
const maxResponseBytes = 2 << 20

// statusError keeps the wiki's status code readable in the settings.
type statusError struct{ code int }

func (e statusError) Error() string {
	if e.code == http.StatusUnauthorized || e.code == http.StatusForbidden {
		return "the wiki refused the API token"
	}
	return fmt.Sprintf("the wiki answered %d", e.code)
}

func call(ctx context.Context, client *http.Client, method, rawURL string, header http.Header, body any, dst any) error {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, r)
	if err != nil {
		return err
	}
	req.Header = header
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return statusError{res.StatusCode}
	}
	return json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(dst)
}

// absolute resolves a link a wiki returns against the wiki's address and
// keeps only links on that same host, so a result can never point the
// browser somewhere else.
func absolute(base, link string) string {
	b, err := url.Parse(base + "/")
	if err != nil {
		return ""
	}
	u, err := b.Parse(link)
	if err != nil || u.Host != b.Host || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	return u.String()
}

// Searchers returns one PageSearcher per supported wiki kind.
func Searchers(client *http.Client, now func() time.Time) map[domain.Kind]usecase.PageSearcher {
	return map[domain.Kind]usecase.PageSearcher{
		domain.KindBookStack: BookStack{client: client},
		domain.KindOutline:   &Outline{client: client, now: now, collections: map[string]collectionCache{}},
		domain.KindWikiJS:    WikiJS{client: client},
	}
}

// BookStack searches with GET /api/search (API token "id:secret").
type BookStack struct{ client *http.Client }

// Search implements usecase.PageSearcher.
func (s BookStack) Search(ctx context.Context, w domain.Wiki, query string, limit int) ([]domain.Page, error) {
	var res struct {
		Data []struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			URL         string `json:"url"`
			UpdatedAt   string `json:"updated_at"`
			PreviewHTML struct {
				Content string `json:"content"`
			} `json:"preview_html"`
			Book *struct {
				Name string `json:"name"`
			} `json:"book"`
		} `json:"data"`
	}
	q := url.Values{"query": {query}, "count": {strconv.Itoa(limit)}}
	h := http.Header{"Authorization": {"Token " + w.Token}}
	if err := call(ctx, s.client, http.MethodGet, w.URL+"/api/search?"+q.Encode(), h, nil, &res); err != nil {
		return nil, err
	}
	var out []domain.Page
	for _, d := range res.Data {
		link := absolute(w.URL, d.URL)
		if link == "" {
			continue
		}
		p := domain.Page{Title: d.Name, Excerpt: domain.Excerpt(d.PreviewHTML.Content), URL: link}
		if d.Book != nil {
			p.Space = d.Book.Name
		}
		p.Updated, _ = time.Parse(time.RFC3339Nano, d.UpdatedAt)
		out = append(out, p)
	}
	return out, nil
}

// collectionTTL is how long Outline's collection names are reused; they
// rarely change and would otherwise cost a second request per search.
const collectionTTL = 10 * time.Minute

type collectionCache struct {
	at    time.Time
	names map[string]string
}

// Outline searches with POST /api/documents.search (API key as bearer).
type Outline struct {
	client      *http.Client
	now         func() time.Time
	mu          sync.Mutex
	collections map[string]collectionCache
}

func (s *Outline) collectionNames(ctx context.Context, w domain.Wiki, h http.Header) map[string]string {
	s.mu.Lock()
	c, ok := s.collections[w.ID]
	s.mu.Unlock()
	if ok && s.now().Sub(c.at) < collectionTTL {
		return c.names
	}
	var res struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	names := map[string]string{}
	// Without collection names the results are still useful, so a failure
	// here only leaves the space empty.
	if err := call(ctx, s.client, http.MethodPost, w.URL+"/api/collections.list", h.Clone(), map[string]any{"limit": 100}, &res); err == nil {
		for _, d := range res.Data {
			names[d.ID] = d.Name
		}
		s.mu.Lock()
		s.collections[w.ID] = collectionCache{at: s.now(), names: names}
		s.mu.Unlock()
	}
	return names
}

// Search implements usecase.PageSearcher.
func (s *Outline) Search(ctx context.Context, w domain.Wiki, query string, limit int) ([]domain.Page, error) {
	h := http.Header{"Authorization": {"Bearer " + w.Token}}
	var res struct {
		Data []struct {
			Context  string `json:"context"`
			Document struct {
				Title        string `json:"title"`
				URL          string `json:"url"`
				UpdatedAt    string `json:"updatedAt"`
				CollectionID string `json:"collectionId"`
			} `json:"document"`
		} `json:"data"`
	}
	if err := call(ctx, s.client, http.MethodPost, w.URL+"/api/documents.search", h.Clone(), map[string]any{"query": query, "limit": limit}, &res); err != nil {
		return nil, err
	}
	names := s.collectionNames(ctx, w, h)
	var out []domain.Page
	for _, d := range res.Data {
		link := absolute(w.URL, d.Document.URL)
		if link == "" {
			continue
		}
		p := domain.Page{Title: d.Document.Title, Excerpt: domain.Excerpt(d.Context), URL: link, Space: names[d.Document.CollectionID]}
		p.Updated, _ = time.Parse(time.RFC3339Nano, d.Document.UpdatedAt)
		out = append(out, p)
	}
	return out, nil
}

// wikiJSQuery is the GraphQL search of Wiki.js 2.x; sent with plain
// net/http, no GraphQL library is needed for one fixed query.
const wikiJSQuery = `query ($q: String!) { pages { search(query: $q) { results { title description path locale } } } }`

// WikiJS searches through the GraphQL API (API key as bearer).
type WikiJS struct{ client *http.Client }

// Search implements usecase.PageSearcher.
func (s WikiJS) Search(ctx context.Context, w domain.Wiki, query string, limit int) ([]domain.Page, error) {
	var res struct {
		Data struct {
			Pages struct {
				Search struct {
					Results []struct {
						Title       string `json:"title"`
						Description string `json:"description"`
						Path        string `json:"path"`
						Locale      string `json:"locale"`
					} `json:"results"`
				} `json:"search"`
			} `json:"pages"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	h := http.Header{"Authorization": {"Bearer " + w.Token}}
	body := map[string]any{"query": wikiJSQuery, "variables": map[string]string{"q": query}}
	if err := call(ctx, s.client, http.MethodPost, w.URL+"/graphql", h, body, &res); err != nil {
		return nil, err
	}
	if len(res.Errors) > 0 {
		return nil, fmt.Errorf("the wiki answered: %s", res.Errors[0].Message)
	}
	var out []domain.Page
	for i, r := range res.Data.Pages.Search.Results {
		if i == limit {
			break
		}
		path := strings.TrimPrefix(r.Path, "/")
		link := absolute(w.URL, "/"+url.PathEscape(r.Locale)+"/"+path)
		if link == "" {
			continue
		}
		out = append(out, domain.Page{Title: r.Title, Excerpt: domain.Excerpt(r.Description), URL: link, Space: r.Locale})
	}
	return out, nil
}
