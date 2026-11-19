// Package fcm sends sealed payloads as data messages through the Firebase
// Cloud Messaging HTTP v1 API, authenticated with a service account.
package fcm

import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"waylight/internal/pushrelay"
	"waylight/internal/pushrelay/jwt"
)

const scope = "https://www.googleapis.com/auth/firebase.messaging"

// DefaultBase is the FCM API.
const DefaultBase = "https://fcm.googleapis.com"

// Account is the part of a Google service account file the sender needs.
type Account struct {
	ProjectID   string `json:"project_id"`
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

// Sender implements pushrelay.Sender for FCM.
type Sender struct {
	client  *http.Client
	base    string
	account Account
	key     crypto.Signer
	now     func() time.Time

	mu      sync.Mutex
	access  string
	expires time.Time
}

// New reads a service account file's content. base is DefaultBase outside
// tests.
func New(client *http.Client, base string, serviceAccount []byte, now func() time.Time) (*Sender, error) {
	var a Account
	if err := json.Unmarshal(serviceAccount, &a); err != nil {
		return nil, fmt.Errorf("fcm: service account: %w", err)
	}
	if a.ProjectID == "" || a.ClientEmail == "" || a.TokenURI == "" {
		return nil, errors.New("fcm: service account lacks project_id, client_email or token_uri")
	}
	key, err := jwt.ParseKey([]byte(a.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("fcm: %w", err)
	}
	return &Sender{client: client, base: base, account: a, key: key, now: now}, nil
}

// token returns an OAuth access token, fetched again five minutes before
// it expires (RFC 7523 JWT bearer grant).
func (s *Sender) token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if s.access != "" && now.Before(s.expires.Add(-5*time.Minute)) {
		return s.access, nil
	}
	assertion, err := jwt.Sign(s.key, map[string]any{}, map[string]any{
		"iss": s.account.ClientEmail, "scope": scope, "aud": s.account.TokenURI,
		"iat": now.Unix(), "exp": now.Add(time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}
	form := url.Values{"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"}, "assertion": {assertion}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.account.TokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fcm: token endpoint answered %d", res.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<16)).Decode(&out); err != nil || out.AccessToken == "" {
		return "", fmt.Errorf("fcm: token answer: %w", err)
	}
	s.access, s.expires = out.AccessToken, now.Add(time.Duration(out.ExpiresIn)*time.Second)
	return s.access, nil
}

// Send posts a high-priority data message {"p": payload}; the app decrypts
// it in its message handler.
func (s *Sender) Send(ctx context.Context, token string, payload []byte) error {
	access, err := s.token(ctx)
	if err != nil {
		return err
	}
	body, err := json.Marshal(map[string]any{"message": map[string]any{
		"token":   token,
		"data":    map[string]string{"p": base64.StdEncoding.EncodeToString(payload)},
		"android": map[string]any{"priority": "HIGH"},
	}})
	if err != nil {
		return err
	}
	u := s.base + "/v1/projects/" + url.PathEscape(s.account.ProjectID) + "/messages:send"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "application/json")
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		return nil
	}
	if unregistered(res) {
		return pushrelay.ErrUnregistered
	}
	return fmt.Errorf("fcm: answered %d", res.StatusCode)
}

// unregistered reports FCM's answer for a token that is gone: 404 with the
// error code UNREGISTERED.
func unregistered(res *http.Response) bool {
	if res.StatusCode != http.StatusNotFound {
		return false
	}
	var out struct {
		Error struct {
			Details []struct {
				ErrorCode string `json:"errorCode"`
			} `json:"details"`
		} `json:"error"`
	}
	if json.NewDecoder(io.LimitReader(res.Body, 1<<16)).Decode(&out) != nil {
		return true // a bare 404 for a message also means the token is gone
	}
	for _, d := range out.Error.Details {
		if d.ErrorCode != "" && d.ErrorCode != "UNREGISTERED" {
			return false
		}
	}
	return true
}
