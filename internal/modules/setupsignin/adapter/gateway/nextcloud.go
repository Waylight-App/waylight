package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"waylight/internal/shared/kernel"
)

// maxAnswerBytes bounds the answers of the login flow; they are a few
// hundred bytes.
const maxAnswerBytes = 64 << 10

// appName is what Nextcloud calls the app password it creates; it reads
// the name from the User-Agent.
const appName = "Waylight"

// NextcloudLogin implements usecase.NextcloudLogin with Login Flow v2
// (https://docs.nextcloud.com/server/latest/developer_manual/client_apis/LoginFlow/).
type NextcloudLogin struct{ client *http.Client }

// NewNextcloudLogin returns the gateway; client must follow the
// integration URL rules.
func NewNextcloudLogin(client *http.Client) *NextcloudLogin { return &NextcloudLogin{client: client} }

// Start implements usecase.NextcloudLogin: POST /index.php/login/v2.
func (n *NextcloudLogin) Start(ctx context.Context, base string) (string, string, error) {
	var body struct {
		Poll struct {
			Token string `json:"token"`
		} `json:"poll"`
		Login string `json:"login"`
	}
	res, err := n.post(ctx, base+"/index.php/login/v2", nil)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	if err := status(res); err != nil {
		return "", "", err
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxAnswerBytes)).Decode(&body); err != nil {
		return "", "", fmt.Errorf("%w: %w", kernel.ErrTargetWrongService, err)
	}
	if body.Poll.Token == "" || body.Login == "" {
		return "", "", fmt.Errorf("%w: the login flow answer has no token", kernel.ErrTargetWrongService)
	}
	return body.Login, body.Poll.Token, nil
}

// Poll implements usecase.NextcloudLogin. Nextcloud answers 404 until the
// user has granted access, then once with the app password.
//
// The poll goes to the entered address, not to the endpoint Nextcloud
// names: behind a reverse proxy that endpoint may carry another host, and
// following it would let the answer pick where the server sends requests.
func (n *NextcloudLogin) Poll(ctx context.Context, base, token string) (string, string, bool, error) {
	res, err := n.post(ctx, base+"/index.php/login/v2/poll", url.Values{"token": {token}})
	if err != nil {
		return "", "", false, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return "", "", false, nil
	}
	if err := status(res); err != nil {
		return "", "", false, err
	}
	var body struct {
		LoginName   string `json:"loginName"`
		AppPassword string `json:"appPassword"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxAnswerBytes)).Decode(&body); err != nil {
		return "", "", false, fmt.Errorf("%w: %w", kernel.ErrTargetWrongService, err)
	}
	if body.LoginName == "" || body.AppPassword == "" {
		return "", "", false, fmt.Errorf("%w: the login flow answer has no app password", kernel.ErrTargetWrongService)
	}
	return body.LoginName, body.AppPassword, true, nil
}

func (n *NextcloudLogin) post(ctx context.Context, target string, form url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", appName)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return n.client.Do(req)
}

// status maps answers other than 200 to the setup assistant's causes.
func status(res *http.Response) error {
	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("%w: %s", kernel.ErrTargetWrongService, res.Status)
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("%w: %s", kernel.ErrTargetAuth, res.Status)
	default:
		return fmt.Errorf("unexpected answer from Nextcloud: %s", res.Status)
	}
}
