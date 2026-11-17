package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// The device authorization grant (RFC 8628, #535): the person confirms the
// sign-in on the provider's own page with a short code, and the server asks
// the provider until it is done. Nothing is redirected back to the server, so
// it works on any address, also without HTTPS and inside a home network.

// DeviceAuth is what the provider returns when a device flow starts.
type DeviceAuth struct {
	DeviceCode      string
	UserCode        string
	VerificationURI string
	ExpiresIn       time.Duration
	// Interval is how long to wait between two polls.
	Interval time.Duration
}

// Errors of PollDevice that the caller handles.
var (
	// ErrDevicePending: the person has not confirmed yet; poll again.
	ErrDevicePending = errors.New("device flow: authorization pending")
	// ErrDeviceSlowDown: polled too often; wait longer before the next poll.
	ErrDeviceSlowDown = errors.New("device flow: slow down")
	// ErrDeviceExpired: the code ran out before it was confirmed.
	ErrDeviceExpired = errors.New("device flow: code expired")
	// ErrDeviceDenied: the person declined at the provider.
	ErrDeviceDenied = errors.New("device flow: declined")
)

// deviceScope asks for the identity only: no access to mail, files or anything else.
const deviceScope = "openid profile email"

// StartDevice starts a device flow.
func (p *Provider) StartDevice(ctx context.Context) (DeviceAuth, error) {
	d, err := p.discover(ctx)
	if err != nil {
		return DeviceAuth{}, err
	}
	if d.DeviceEndpoint == "" {
		return DeviceAuth{}, errors.New("oidc: the provider offers no device authorization")
	}
	var body struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURI string `json:"verification_uri"`
		ExpiresIn       int    `json:"expires_in"`
		Interval        int    `json:"interval"`
		Error           string `json:"error"`
	}
	status, err := p.postForm(ctx, d.DeviceEndpoint, url.Values{"client_id": {p.cfg.ClientID}, "scope": {deviceScope}}, &body)
	if err != nil {
		return DeviceAuth{}, err
	}
	if status != http.StatusOK || body.DeviceCode == "" || body.UserCode == "" || body.VerificationURI == "" {
		return DeviceAuth{}, fmt.Errorf("oidc device authorization failed: %d %s", status, body.Error)
	}
	// Only the provider's own page is ever opened for the person.
	if u, err := url.Parse(body.VerificationURI); err != nil || u.Scheme != "https" && !strings.HasPrefix(p.cfg.Issuer, "http://") {
		return DeviceAuth{}, errors.New("oidc device authorization: unexpected verification address")
	}
	interval := time.Duration(max(body.Interval, 1)) * time.Second
	expires := time.Duration(max(body.ExpiresIn, 60)) * time.Second
	return DeviceAuth{DeviceCode: body.DeviceCode, UserCode: body.UserCode, VerificationURI: body.VerificationURI,
		ExpiresIn: expires, Interval: interval}, nil
}

// PollDevice asks whether the person confirmed; it returns the verified
// identity, or ErrDevicePending, ErrDeviceSlowDown, ErrDeviceExpired or
// ErrDeviceDenied.
func (p *Provider) PollDevice(ctx context.Context, deviceCode string) (Identity, error) {
	d, err := p.discover(ctx)
	if err != nil {
		return Identity{}, err
	}
	form := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"client_id":   {p.cfg.ClientID},
		"device_code": {deviceCode},
	}
	var body struct {
		IDToken string `json:"id_token"`
		Error   string `json:"error"`
	}
	status, err := p.postForm(ctx, d.TokenEndpoint, form, &body)
	if err != nil {
		return Identity{}, err
	}
	switch body.Error {
	case "":
	case "authorization_pending":
		return Identity{}, ErrDevicePending
	case "slow_down":
		return Identity{}, ErrDeviceSlowDown
	case "expired_token", "bad_verification_code":
		return Identity{}, ErrDeviceExpired
	case "access_denied", "authorization_declined":
		return Identity{}, ErrDeviceDenied
	default:
		return Identity{}, fmt.Errorf("oidc device token: %s", body.Error)
	}
	if status != http.StatusOK || body.IDToken == "" {
		return Identity{}, fmt.Errorf("oidc device token failed: %d", status)
	}
	// No nonce in this flow: the device code is the one-time secret.
	return p.verify(ctx, body.IDToken, "", false)
}

// postForm posts a form as a public client (with the secret when one is
// configured) and decodes the JSON answer whatever the status.
func (p *Provider) postForm(ctx context.Context, endpoint string, form url.Values, v any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if p.cfg.ClientSecret != "" {
		req.SetBasicAuth(url.QueryEscape(p.cfg.ClientID), url.QueryEscape(p.cfg.ClientSecret))
	}
	res, err := p.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("oidc request: %w", err)
	}
	defer res.Body.Close()
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(v); err != nil {
		return res.StatusCode, fmt.Errorf("oidc response: %w", err)
	}
	return res.StatusCode, nil
}
