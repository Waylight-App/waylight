// Package usecase connects members' accounts at external services and hands
// valid tokens to the team integrations (#326).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/ee/connections/domain"
	licence "trisse-home/ee/licence/domain"
	"trisse-home/internal/shared/kernel"
)

// Tokens stores connections, encrypted per row (output port).
type Tokens interface {
	Get(ctx context.Context, accountID, service string) (domain.Connection, bool, error)
	ListFor(ctx context.Context, accountID string) ([]domain.Connection, error)
	Save(ctx context.Context, c domain.Connection) error
	Delete(ctx context.Context, accountID, service string) error
	DeleteAccount(ctx context.Context, accountID string) error
}

// Clients stores the admins' app registrations (output port).
type Clients interface {
	Get(ctx context.Context, service string) (domain.Client, bool, error)
	List(ctx context.Context) ([]domain.Client, error)
	Save(ctx context.Context, c domain.Client) error
}

// OAuth runs the authorization code flow at a service (output port).
type OAuth interface {
	// Services lists the services this version knows, with display names.
	Services() []ServiceInfo
	// RedirectURL is the callback to register; "" without PUBLIC_URL.
	RedirectURL() string
	AuthCodeURL(c domain.Client, state, verifier string) (string, error)
	Exchange(ctx context.Context, c domain.Client, code, verifier string) (domain.Token, error)
	// Refresh fails with an error wrapping ErrRejected when the member has to
	// connect again.
	Refresh(ctx context.Context, c domain.Client, refresh string) (domain.Token, error)
	Revoke(ctx context.Context, c domain.Client, token string) error
}

// ServiceInfo names a service.
type ServiceInfo struct {
	ID   string
	Name string
}

// ErrRejected is wrapped by OAuth.Refresh for refused grants.
var ErrRejected = errors.New("the service rejected the grant")

// Licensing tells whether connected accounts are licensed (output port).
type Licensing interface {
	Level(ctx context.Context, feature string) licence.Level
}

// Audit records connecting and disconnecting (output port, the audit log).
type Audit func(ctx context.Context, action, service string)

// Errors shown to people.
var (
	ErrUnknownService = kernel.NewPublicError(kernel.ErrNotFound, "This service is not known.")
	ErrNotConfigured  = kernel.NewPublicError(kernel.ErrConflict, "An admin has to register trisse-home at this service first.")
	ErrNoPublicURL    = kernel.NewPublicError(kernel.ErrConflict, "Connecting accounts needs PUBLIC_URL, so the service can send the browser back.")
	ErrNotConnected   = kernel.NewPublicError(kernel.ErrNotFound, "This account is not connected.")
	ErrReconnect      = kernel.NewPublicError(kernel.ErrConflict, "The connection has expired; connect the account again.")
	ErrSignInExpired  = kernel.NewPublicError(kernel.ErrInvalid, "The sign-in has expired or does not belong to you; start again.")
	errAdminOnly      = kernel.NewPublicError(kernel.ErrForbidden, "Only an admin can register services.")
)

// Deps are the ports.
type Deps struct {
	Tokens    Tokens
	Clients   Clients
	OAuth     OAuth
	Clock     kernel.Clock
	Licensing Licensing
	Audit     Audit
}

// pending is a started sign-in, kept in memory for ten minutes.
type pending struct {
	accountID, service, verifier string
	expires                      time.Time
}

const signInTTL = 10 * time.Minute

// Service implements the use cases.
type Service struct {
	d       Deps
	mu      sync.Mutex
	pending map[string]pending // state -> sign-in
	refresh sync.Mutex         // one refresh at a time, so two requests do not race for the same token
}

// New returns the service.
func New(d Deps) *Service { return &Service{d: d, pending: map[string]pending{}} }

func (s *Service) known(service string) bool {
	for _, info := range s.d.OAuth.Services() {
		if info.ID == service {
			return true
		}
	}
	return false
}

// View is one service as a member sees it.
type View struct {
	Service    ServiceInfo
	Configured bool
	Status     domain.Status // "" when not connected
	Since      time.Time
}

// List returns every known service with the member's connection status.
func (s *Service) List(ctx context.Context) ([]View, error) {
	me := kernel.PrincipalFrom(ctx).AccountID
	conns, err := s.d.Tokens.ListFor(ctx, me)
	if err != nil {
		return nil, err
	}
	var out []View
	for _, info := range s.d.OAuth.Services() {
		_, configured, err := s.d.Clients.Get(ctx, info.ID)
		if err != nil {
			return nil, err
		}
		v := View{Service: info, Configured: configured}
		for _, c := range conns {
			if c.Service == info.ID {
				v.Status, v.Since = c.Status, c.ConnectedAt
			}
		}
		out = append(out, v)
	}
	return out, nil
}

func randomState() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Start begins connecting the signed-in account and returns where the
// browser signs in at the service.
func (s *Service) Start(ctx context.Context, service string) (string, error) {
	if !s.known(service) {
		return "", ErrUnknownService
	}
	if s.d.OAuth.RedirectURL() == "" {
		return "", ErrNoPublicURL
	}
	c, ok, err := s.d.Clients.Get(ctx, service)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrNotConfigured
	}
	state, verifier := randomState(), randomState()
	url, err := s.d.OAuth.AuthCodeURL(c, state, verifier)
	if err != nil {
		return "", err
	}
	now := s.d.Clock.Now()
	s.mu.Lock()
	for k, p := range s.pending {
		if now.After(p.expires) {
			delete(s.pending, k)
		}
	}
	s.pending[state] = pending{accountID: kernel.PrincipalFrom(ctx).AccountID, service: service, verifier: verifier, expires: now.Add(signInTTL)}
	s.mu.Unlock()
	return url, nil
}

// Finish completes the sign-in from the callback. The state must be one this
// server issued to the same account within ten minutes; it is used once.
func (s *Service) Finish(ctx context.Context, state, code, providerError string) (string, error) {
	now := s.d.Clock.Now()
	s.mu.Lock()
	p, ok := s.pending[state]
	delete(s.pending, state)
	s.mu.Unlock()
	if !ok || state == "" || now.After(p.expires) || p.accountID != kernel.PrincipalFrom(ctx).AccountID {
		return "", ErrSignInExpired
	}
	if providerError != "" || code == "" {
		return p.service, kernel.NewPublicErrorf(kernel.ErrInvalid, "The service did not confirm the sign-in: {detail}", "detail", providerError)
	}
	c, ok, err := s.d.Clients.Get(ctx, p.service)
	if err != nil || !ok {
		return p.service, ErrNotConfigured
	}
	tok, err := s.d.OAuth.Exchange(ctx, c, code, p.verifier)
	if err != nil {
		return p.service, kernel.NewPublicErrorf(kernel.ErrInvalid, "The service did not confirm the sign-in: {detail}", "detail", err.Error())
	}
	conn := domain.Connection{AccountID: p.accountID, Service: p.service, Token: tok, Status: domain.StatusConnected, ConnectedAt: now}
	if err := s.d.Tokens.Save(ctx, conn); err != nil {
		return p.service, err
	}
	s.audit(ctx, "connected", p.service)
	return p.service, nil
}

func (s *Service) audit(ctx context.Context, action, service string) {
	if s.d.Audit != nil {
		s.d.Audit(ctx, action, service)
	}
}

// Disconnect revokes the token at the service where possible and deletes it.
func (s *Service) Disconnect(ctx context.Context, service string) error {
	me := kernel.PrincipalFrom(ctx).AccountID
	conn, ok, err := s.d.Tokens.Get(ctx, me, service)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotConnected
	}
	if c, ok, _ := s.d.Clients.Get(ctx, service); ok {
		token := conn.Token.Refresh
		if token == "" {
			token = conn.Token.Access
		}
		// A failed revocation must not keep the token here; it expires at the
		// service on its own.
		_ = s.d.OAuth.Revoke(ctx, c, token)
	}
	if err := s.d.Tokens.Delete(ctx, me, service); err != nil {
		return err
	}
	s.audit(ctx, "disconnected", service)
	return nil
}

// AccessToken is what integrations ask: a valid access token of a member at
// a service, refreshed when needed. It fails with ErrNotConnected,
// ErrReconnect or an error without a licence key.
func (s *Service) AccessToken(ctx context.Context, accountID, service string) (string, error) {
	if s.d.Licensing.Level(ctx, licence.FeatureConnectedAccounts) == licence.LevelNone {
		return "", kernel.NewPublicError(kernel.ErrForbidden, "Connected accounts need a valid licence key.")
	}
	s.refresh.Lock()
	defer s.refresh.Unlock()
	conn, ok, err := s.d.Tokens.Get(ctx, accountID, service)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrNotConnected
	}
	if conn.Status == domain.StatusExpired {
		return "", ErrReconnect
	}
	if !conn.NeedsRefresh(s.d.Clock.Now()) {
		return conn.Token.Access, nil
	}
	expire := func() (string, error) {
		conn.Status = domain.StatusExpired
		if err := s.d.Tokens.Save(ctx, conn); err != nil {
			return "", err
		}
		return "", ErrReconnect
	}
	if conn.Token.Refresh == "" {
		// Services without refresh tokens: the member reconnects when it runs out.
		return expire()
	}
	c, ok, err := s.d.Clients.Get(ctx, service)
	if err != nil || !ok {
		return "", ErrNotConfigured
	}
	tok, err := s.d.OAuth.Refresh(ctx, c, conn.Token.Refresh)
	if errors.Is(err, ErrRejected) || (err == nil && tok.Access == "") {
		return expire()
	}
	if err != nil {
		return "", fmt.Errorf("refresh %s token: %w", service, err)
	}
	conn.Token = tok
	if err := s.d.Tokens.Save(ctx, conn); err != nil {
		return "", err
	}
	return tok.Access, nil
}

// ClientView is an app registration as admins see it: never the secret.
type ClientView struct {
	Service   ServiceInfo
	ClientID  string
	Tenant    string
	HasSecret bool
}

// ClientsList returns every known service with its registration (admins).
func (s *Service) ClientsList(ctx context.Context) ([]ClientView, string, error) {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return nil, "", errAdminOnly
	}
	var out []ClientView
	for _, info := range s.d.OAuth.Services() {
		c, _, err := s.d.Clients.Get(ctx, info.ID)
		if err != nil {
			return nil, "", err
		}
		out = append(out, ClientView{Service: info, ClientID: c.ClientID, Tenant: c.Tenant, HasSecret: c.Secret != ""})
	}
	return out, s.d.OAuth.RedirectURL(), nil
}

// SaveClient registers trisse-home at a service (admins). An empty secret
// keeps the stored one.
func (s *Service) SaveClient(ctx context.Context, in domain.Client) error {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return errAdminOnly
	}
	if !s.known(in.Service) {
		return ErrUnknownService
	}
	if in.Secret == "" {
		if old, ok, err := s.d.Clients.Get(ctx, in.Service); err != nil {
			return err
		} else if ok {
			in.Secret = old.Secret
		}
	}
	return s.d.Clients.Save(ctx, in)
}

// AccountDeleted removes a deleted account's tokens.
func (s *Service) AccountDeleted(ctx context.Context, accountID string) error {
	return s.d.Tokens.DeleteAccount(ctx, accountID)
}
