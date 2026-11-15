// Package presenter maps wiki results and connections to JSON (#321).
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/wiki/usecase"
)

// PageDTO is one search result.
type PageDTO struct {
	Title   string     `json:"title"`
	Excerpt string     `json:"excerpt"`
	URL     string     `json:"url"`
	Space   string     `json:"space"`
	Wiki    string     `json:"wiki"`
	Updated *time.Time `json:"updated,omitempty"`
}

// SearchDTO is the merged answer; failed names wikis that did not answer.
type SearchDTO struct {
	Pages  []PageDTO `json:"pages"`
	Failed []string  `json:"failed"`
}

// WikiDTO is a connection; the token is never returned.
type WikiDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	URL      string `json:"url"`
	HasToken bool   `json:"hasToken"`
}

// Search maps a search result.
func Search(r usecase.SearchResult) SearchDTO {
	out := SearchDTO{Pages: make([]PageDTO, 0, len(r.Pages)), Failed: append([]string{}, r.Failed...)}
	for _, p := range r.Pages {
		d := PageDTO{Title: p.Title, Excerpt: p.Excerpt, URL: p.URL, Space: p.Space, Wiki: p.Wiki}
		if !p.Updated.IsZero() {
			u := p.Updated
			d.Updated = &u
		}
		out.Pages = append(out.Pages, d)
	}
	return out
}

// Wiki maps a connection.
func Wiki(v usecase.WikiView) WikiDTO {
	return WikiDTO{ID: v.ID, Name: v.Name, Kind: string(v.Kind), URL: v.URL, HasToken: v.HasToken}
}
