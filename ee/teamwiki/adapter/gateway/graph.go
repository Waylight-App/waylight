package gateway

import (
	"context"
	"net/http"

	"trisse-home/internal/modules/wiki/domain"
)

// Graph searches SharePoint and OneDrive through Microsoft Graph with the
// member's delegated permissions (#328).
type Graph struct {
	Client *http.Client
	// API is https://graph.microsoft.com/v1.0; tests point it at a fake.
	API string
}

type graphSearch struct {
	Value []struct {
		HitsContainers []struct {
			Hits []struct {
				Summary  string `json:"summary"`
				Resource struct {
					Name                 string `json:"name"`
					WebURL               string `json:"webUrl"`
					LastModifiedDateTime string `json:"lastModifiedDateTime"`
					ParentReference      struct {
						SiteID string `json:"siteId"`
					} `json:"parentReference"`
				} `json:"resource"`
			} `json:"hits"`
		} `json:"hitsContainers"`
	} `json:"value"`
}

// Search implements usecase.Searcher.
func (g Graph) Search(ctx context.Context, token, query string, limit int) ([]domain.Page, error) {
	body := map[string]any{"requests": []map[string]any{{
		"entityTypes": []string{"driveItem", "listItem"},
		"query":       map[string]string{"queryString": query},
		"size":        limit,
	}}}
	var res graphSearch
	if err := call(ctx, g.Client, http.MethodPost, g.API+"/search/query", token, body, nil, &res); err != nil {
		return nil, err
	}
	var out []domain.Page
	for _, v := range res.Value {
		for _, c := range v.HitsContainers {
			for _, h := range c.Hits {
				out = append(out, domain.Page{Title: h.Resource.Name, Excerpt: excerpt(h.Summary), URL: h.Resource.WebURL,
					Updated: parseTime(h.Resource.LastModifiedDateTime)})
			}
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
