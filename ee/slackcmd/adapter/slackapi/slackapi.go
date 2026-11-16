// Package slackapi calls the Slack Web API for the slash command (#333).
//
// Ring: Interface adapters (gateway).
package slackapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client calls Slack. API is https://slack.com/api; tests point it at a fake.
type Client struct {
	HTTP *http.Client
	API  string
}

type answer struct {
	OK     bool   `json:"ok"`
	Error  string `json:"error"`
	URL    string `json:"url"`
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
}

func (c Client) post(ctx context.Context, method, token string) (answer, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.API+"/"+method, strings.NewReader(""))
	if err != nil {
		return answer{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return answer{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return answer{}, fmt.Errorf("slack %s answered %s", method, res.Status)
	}
	var a answer
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&a); err != nil {
		return answer{}, err
	}
	if !a.OK {
		return answer{}, fmt.Errorf("slack %s: %s", method, a.Error)
	}
	return a, nil
}

// Identity implements usecase.Slack with auth.test.
func (c Client) Identity(ctx context.Context, userToken string) (string, string, error) {
	a, err := c.post(ctx, "auth.test", userToken)
	if err != nil {
		return "", "", err
	}
	if a.TeamID == "" || a.UserID == "" {
		return "", "", errors.New("slack auth.test: no team or user")
	}
	return a.TeamID, a.UserID, nil
}

// OpenSocket asks for a Socket Mode address with the app-level token
// (apps.connections.open).
func (c Client) OpenSocket(ctx context.Context, appToken string) (string, error) {
	a, err := c.post(ctx, "apps.connections.open", appToken)
	if err != nil {
		return "", err
	}
	// Only TLS, except against a local fake in tests (an http:// API).
	plain := strings.HasPrefix(c.API, "http://") && strings.HasPrefix(a.URL, "ws://")
	if !strings.HasPrefix(a.URL, "wss://") && !plain {
		return "", errors.New("slack apps.connections.open: no WebSocket address")
	}
	return a.URL, nil
}
