package gateway

import (
	"context"
	"net/http"

	"trisse-home/internal/modules/wiki/domain"
)

// Teams searches the chat and channel messages the member is part of through
// Microsoft Graph search (#331), with the delegated Chat.Read and
// ChannelMessage.Read.All of their connected Microsoft 365 account.
// Graph cannot mix messages with other entity types in one request, so this
// is a source of its own next to SharePoint and OneDrive.
type Teams struct {
	Client *http.Client
	// API is https://graph.microsoft.com/v1.0; tests point it at a fake.
	API string
}

type teamsSearch struct {
	Value []struct {
		HitsContainers []struct {
			Hits []struct {
				Summary  string `json:"summary"`
				Resource struct {
					Subject         string `json:"subject"`
					WebLink         string `json:"webLink"`
					WebURL          string `json:"webUrl"`
					CreatedDateTime string `json:"createdDateTime"`
					From            struct {
						EmailAddress struct {
							Name string `json:"name"`
						} `json:"emailAddress"`
					} `json:"from"`
					ChannelIdentity struct {
						ChannelID string `json:"channelId"`
					} `json:"channelIdentity"`
				} `json:"resource"`
			} `json:"hits"`
		} `json:"hitsContainers"`
	} `json:"value"`
}

// Search implements usecase.Searcher.
func (t Teams) Search(ctx context.Context, token, query string, limit int) ([]domain.Page, error) {
	body := map[string]any{"requests": []map[string]any{{
		"entityTypes": []string{"chatMessage"},
		"query":       map[string]string{"queryString": query},
		"from":        0,
		"size":        limit,
	}}}
	var res teamsSearch
	if err := call(ctx, t.Client, http.MethodPost, t.API+"/search/query", token, body, nil, &res); err != nil {
		return nil, err
	}
	var out []domain.Page
	for _, v := range res.Value {
		for _, c := range v.HitsContainers {
			for _, h := range c.Hits {
				r := h.Resource
				title := r.From.EmailAddress.Name
				if r.Subject != "" {
					title = r.Subject + " · " + title
				}
				space := "Chat"
				if r.ChannelIdentity.ChannelID != "" {
					space = "Channel"
				}
				link := r.WebLink
				if link == "" {
					link = r.WebURL
				}
				out = append(out, domain.Page{Title: title, Excerpt: excerpt(h.Summary), URL: link, Space: space,
					Updated: parseTime(r.CreatedDateTime)})
			}
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
