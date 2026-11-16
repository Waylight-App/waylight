package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/wiki/domain"
)

// Slack searches messages with the member's user token through the
// Real-time Search API, assistant.search.context (#330). Slack applies the
// person's own access: channels and conversations they are not in never
// show up. Nothing is stored; Slack's terms do not allow keeping results.
type Slack struct {
	Client *http.Client
	// API is https://slack.com/api; tests point it at a fake.
	API string
}

// allChannels asks for every conversation type the person may search; a
// token without the private scopes falls back to public channels.
const allChannels = "public_channel,private_channel,mpim,im"

type slackSearch struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error"`
	Results struct {
		Messages []struct {
			AuthorName  string `json:"author_name"`
			ChannelName string `json:"channel_name"`
			MessageTS   string `json:"message_ts"`
			Content     string `json:"content"`
			Permalink   string `json:"permalink"`
		} `json:"messages"`
	} `json:"results"`
}

// errMissingScope is Slack's answer for a channel type the token may not search.
var errMissingScope = errors.New("slack: missing_scope")

// Search implements usecase.Searcher.
func (s Slack) Search(ctx context.Context, token, query string, limit int) ([]domain.Page, error) {
	res, err := s.search(ctx, token, query, allChannels, limit)
	if errors.Is(err, errMissingScope) {
		res, err = s.search(ctx, token, query, "public_channel", limit)
	}
	if err != nil {
		return nil, err
	}
	out := make([]domain.Page, 0, len(res.Results.Messages))
	for _, m := range res.Results.Messages {
		title := m.AuthorName
		if m.ChannelName != "" {
			title = "#" + m.ChannelName + " · " + m.AuthorName
		}
		out = append(out, domain.Page{Title: title, Excerpt: excerpt(m.Content), URL: m.Permalink, Space: m.ChannelName,
			Updated: slackTime(m.MessageTS)})
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s Slack) search(ctx context.Context, token, query, channels string, limit int) (slackSearch, error) {
	form := url.Values{"query": {query}, "channel_types": {channels}, "limit": {strconv.Itoa(min(limit, 20))}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.API+"/assistant.search.context", strings.NewReader(form.Encode()))
	if err != nil {
		return slackSearch{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.Client.Do(req)
	if err != nil {
		return slackSearch{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return slackSearch{}, errors.New("slack: rate limited")
	}
	if resp.StatusCode != http.StatusOK {
		return slackSearch{}, fmt.Errorf("slack answered %s", resp.Status)
	}
	var res slackSearch
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponse)).Decode(&res); err != nil {
		return slackSearch{}, err
	}
	switch {
	case res.OK:
		return res, nil
	case res.Error == "missing_scope":
		return slackSearch{}, errMissingScope
	default:
		return slackSearch{}, fmt.Errorf("slack: %s", res.Error)
	}
}

// slackTime reads a message timestamp such as "1733260762.123456".
func slackTime(ts string) time.Time {
	sec, _, _ := strings.Cut(ts, ".")
	n, err := strconv.ParseInt(sec, 10, 64)
	if err != nil || n <= 0 {
		return time.Time{}
	}
	return time.Unix(n, 0).UTC()
}
