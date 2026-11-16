// Package graph asks Microsoft Graph whose a member's token is (#334).
//
// Ring: Interface adapters (gateway).
package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client calls Graph. API is https://graph.microsoft.com/v1.0; tests use a
// fake.
type Client struct {
	HTTP *http.Client
	API  string
}

// ObjectID implements usecase.Directory with GET /me.
func (c Client) ObjectID(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.API+"/me?$select=id", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("graph /me answered %s", res.Status)
	}
	var me struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<16)).Decode(&me); err != nil {
		return "", err
	}
	return me.ID, nil
}
