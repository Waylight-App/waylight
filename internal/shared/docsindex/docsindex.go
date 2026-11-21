// Package docsindex serves a small index of the documentation (#678): per
// page its title, a short summary and its section headings, so the search
// can answer plain questions with the matching page without the server
// carrying the whole documentation.
package docsindex

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"waylight/internal/shared/httpapi"
)

// Section is a heading of a page with the anchor the published site gives it.
type Section struct {
	Title  string `json:"title"`
	Anchor string `json:"anchor"`
}

// Page is one documentation page; Path is relative to the published
// documentation, e.g. "modules/immich/".
type Page struct {
	Path     string    `json:"path"`
	Title    string    `json:"title"`
	Summary  string    `json:"summary"`
	Sections []Section `json:"sections"`
}

// Index is the DTO of GET /api/help/docs.
type Index struct {
	BaseURL string `json:"baseUrl"`
	Pages   []Page `json:"pages"`
}

//go:embed index.json
var raw []byte

// Handler serves the embedded index with the given base URL of the site.
// It panics on a broken index, which the tests of this package rule out.
func Handler(baseURL string) http.HandlerFunc {
	var pages []Page
	if err := json.Unmarshal(raw, &pages); err != nil {
		panic("docsindex: " + err.Error())
	}
	res := Index{BaseURL: baseURL, Pages: pages}
	return func(w http.ResponseWriter, _ *http.Request) {
		httpapi.WriteJSON(w, http.StatusOK, res)
	}
}
