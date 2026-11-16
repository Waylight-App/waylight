package gateway

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"trisse-home/internal/modules/wiki/domain"
)

// Confluence searches every Confluence site the member's Atlassian account
// can reach (#327).
type Confluence struct {
	Client *http.Client
	// API is https://api.atlassian.com; tests point it at a fake.
	API string
}

type site struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type confluenceSearch struct {
	Results []struct {
		Content struct {
			Title string `json:"title"`
		} `json:"content"`
		Title                 string `json:"title"`
		Excerpt               string `json:"excerpt"`
		URL                   string `json:"url"`
		LastModified          string `json:"lastModified"`
		ResultGlobalContainer struct {
			Title string `json:"title"`
		} `json:"resultGlobalContainer"`
	} `json:"results"`
	Links struct {
		Base string `json:"base"`
	} `json:"_links"`
}

// cql quotes the query for siteSearch, so quotes and backslashes in it stay text.
func cql(query string) string {
	q := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(query)
	return `siteSearch ~ "` + q + `" AND type in (page, blogpost)`
}

// Search implements usecase.Searcher.
func (c Confluence) Search(ctx context.Context, token, query string, limit int) ([]domain.Page, error) {
	var sites []site
	if err := call(ctx, c.Client, http.MethodGet, c.API+"/oauth/token/accessible-resources", token, nil, nil, &sites); err != nil {
		return nil, err
	}
	var out []domain.Page
	for _, s := range sites {
		u := c.API + "/ex/confluence/" + url.PathEscape(s.ID) + "/wiki/rest/api/search?" +
			url.Values{"cql": {cql(query)}, "limit": {strconv.Itoa(limit)}}.Encode()
		var res confluenceSearch
		if err := call(ctx, c.Client, http.MethodGet, u, token, nil, nil, &res); err != nil {
			return nil, err
		}
		for _, r := range res.Results {
			title := r.Content.Title
			if title == "" {
				title = r.Title
			}
			out = append(out, domain.Page{Title: title, Excerpt: excerpt(r.Excerpt), URL: res.Links.Base + r.URL,
				Space: r.ResultGlobalContainer.Title, Updated: parseTime(r.LastModified)})
		}
		if len(out) >= limit {
			return out[:limit], nil
		}
	}
	return out, nil
}
