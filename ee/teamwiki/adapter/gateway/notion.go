package gateway

import (
	"context"
	"net/http"
	"strings"

	"trisse-home/internal/modules/wiki/domain"
)

// notionVersion is the API version the requests are written for.
const notionVersion = "2022-06-28"

// Notion searches the pages the member's integration was given (#329).
// Notion's public search matches titles, not full text.
type Notion struct {
	Client *http.Client
	// API is https://api.notion.com/v1; tests point it at a fake.
	API string
}

type notionSearch struct {
	Results []struct {
		URL            string `json:"url"`
		LastEditedTime string `json:"last_edited_time"`
		Properties     map[string]struct {
			Type  string `json:"type"`
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"properties"`
	} `json:"results"`
}

// Search implements usecase.Searcher.
func (n Notion) Search(ctx context.Context, token, query string, limit int) ([]domain.Page, error) {
	body := map[string]any{"query": query, "page_size": limit, "filter": map[string]string{"value": "page", "property": "object"}}
	var res notionSearch
	if err := call(ctx, n.Client, http.MethodPost, n.API+"/search", token, body, map[string]string{"Notion-Version": notionVersion}, &res); err != nil {
		return nil, err
	}
	var out []domain.Page
	for _, r := range res.Results {
		var title []string
		for _, p := range r.Properties {
			if p.Type != "title" {
				continue
			}
			for _, t := range p.Title {
				title = append(title, t.PlainText)
			}
		}
		out = append(out, domain.Page{Title: strings.Join(title, ""), URL: r.URL, Updated: parseTime(r.LastEditedTime)})
	}
	return out, nil
}
