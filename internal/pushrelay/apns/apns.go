// Package apns sends sealed payloads through the Apple Push Notification
// service with token-based authentication (a .p8 key).
package apns

import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"waylight/internal/pushrelay"
	"waylight/internal/pushrelay/jwt"
)

// Production and Sandbox are the APNs hosts.
const (
	Production = "https://api.push.apple.com"
	Sandbox    = "https://api.sandbox.push.apple.com"
)

const (
	// tokenAge stays below the hour after which APNs rejects a provider
	// token, and above the 20 minutes it may at most be renewed.
	tokenAge = 40 * time.Minute
)

// Config names the key and the app.
type Config struct {
	KeyID, TeamID, Topic string
	Key                  []byte // the .p8 file
}

// Sender implements pushrelay.Sender for APNs.
type Sender struct {
	client *http.Client
	base   string
	cfg    Config
	key    crypto.Signer
	now    func() time.Time

	mu      sync.Mutex
	bearer  string
	created time.Time
}

// New parses the key. base is Production or Sandbox outside tests.
func New(client *http.Client, base string, cfg Config, now func() time.Time) (*Sender, error) {
	if cfg.KeyID == "" || cfg.TeamID == "" || cfg.Topic == "" {
		return nil, fmt.Errorf("apns: key id, team id and topic are required")
	}
	key, err := jwt.ParseKey(cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("apns: %w", err)
	}
	return &Sender{client: client, base: base, cfg: cfg, key: key, now: now}, nil
}

func (s *Sender) token() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if s.bearer != "" && now.Sub(s.created) < tokenAge {
		return s.bearer, nil
	}
	t, err := jwt.Sign(s.key, map[string]any{"kid": s.cfg.KeyID}, map[string]any{"iss": s.cfg.TeamID, "iat": now.Unix()})
	if err != nil {
		return "", err
	}
	s.bearer, s.created = t, now
	return t, nil
}

// Send posts an alert with a generic text and the sealed payload in "p";
// mutable-content lets the app's notification service extension decrypt it
// and replace the text before it is shown.
func (s *Sender) Send(ctx context.Context, token string, payload []byte) error {
	bearer, err := s.token()
	if err != nil {
		return err
	}
	body, err := json.Marshal(map[string]any{
		"aps": map[string]any{"alert": map[string]string{"title": "Waylight", "body": "New notification"}, "mutable-content": 1},
		"p":   base64.StdEncoding.EncodeToString(payload),
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.base+"/3/device/"+url.PathEscape(token), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "bearer "+bearer)
	req.Header.Set("apns-topic", s.cfg.Topic)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("apns-priority", "10")
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		return nil
	}
	var out struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(io.LimitReader(res.Body, 1<<12)).Decode(&out)
	if res.StatusCode == http.StatusGone || out.Reason == "BadDeviceToken" || out.Reason == "Unregistered" {
		return pushrelay.ErrUnregistered
	}
	return fmt.Errorf("apns: answered %d %s", res.StatusCode, out.Reason)
}
