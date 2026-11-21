package domain

import (
	"net/url"
	"strings"
	"time"

	"waylight/internal/shared/kernel"
)

// Service is a service the setup assistant signs in to.
type Service string

// The services with a sign-in.
const (
	Nextcloud     Service = kernel.SignInNextcloud
	HomeAssistant Service = kernel.SignInHomeAssistant
)

// Valid reports whether s has a sign-in.
func (s Service) Valid() bool { return s == Nextcloud || s == HomeAssistant }

// Limits of sign-ins in progress. They live in memory only: a restart
// forgets them and the user signs in again.
const (
	// TTL bounds a sign-in from its start. Nextcloud keeps a login flow for
	// 20 minutes; ten are plenty to sign in and keep abandoned flows short.
	TTL = 10 * time.Minute
	// RedeemWindow is how long a finished sign-in waits for the module
	// that stores it; the assistant redeems it right away.
	RedeemWindow = 2 * time.Minute
	// PollInterval is the least time between two polls of Nextcloud, however
	// often the browser asks.
	PollInterval = 2 * time.Second
	// MaxPollFailures ends a Nextcloud sign-in after this many failed polls
	// in a row; single failures (a restart of Nextcloud) are tolerated.
	MaxPollFailures = 3
	// MaxPending bounds the sign-ins of all accounts together.
	MaxPending = 50
)

// Status is where a sign-in stands.
type Status string

// The statuses of a sign-in.
const (
	Pending Status = "pending"
	Ready   Status = "ready"
	Failed  Status = "failed"
)

var (
	// ErrUnknownService is returned for a service without a sign-in.
	ErrUnknownService = kernel.NewValidationError("service", "This service has no sign-in.")
	// ErrURL is returned for an address that is not http(s).
	ErrURL = kernel.NewValidationError("url", "The address must start with http:// or https://.")
)

// SignIn is one sign-in between its start and the moment the module
// redeems it. The poll token and the state are known only to the server.
type SignIn struct {
	ID      string
	Account string
	Service Service
	URL     string
	// PollToken asks Nextcloud whether the user has signed in.
	PollToken string
	// State ties Home Assistant's redirect to this sign-in; it is cleared
	// on first use.
	State        string
	Expires      time.Time
	Status       Status
	Failure      error
	Grant        kernel.SignInGrant
	LastPoll     time.Time
	PollFailures int
}

// New starts a sign-in that ends after TTL.
func New(id, account string, s Service, address string, now time.Time) *SignIn {
	return &SignIn{ID: id, Account: account, Service: s, URL: address, Status: Pending, Expires: now.Add(TTL)}
}

// Expired reports whether the sign-in is over.
func (s *SignIn) Expired(now time.Time) bool { return !now.Before(s.Expires) }

// Finish stores the grant and forgets the secrets that led to it. The
// module then has RedeemWindow to pick it up, even late in the TTL.
func (s *SignIn) Finish(g kernel.SignInGrant, now time.Time) {
	s.Status, s.Grant, s.PollToken, s.State = Ready, g, "", ""
	if end := now.Add(RedeemWindow); end.After(s.Expires) {
		s.Expires = end
	}
}

// Fail ends the sign-in with a reason the assistant shows.
func (s *SignIn) Fail(err error) {
	s.Status, s.Failure, s.PollToken, s.State = Failed, err, "", ""
}

// PollDue reports whether Nextcloud may be asked again.
func (s *SignIn) PollDue(now time.Time) bool {
	return s.Status == Pending && s.PollToken != "" && now.Sub(s.LastPoll) >= PollInterval
}

// PollFailed counts a failed poll and reports whether the sign-in should
// give up.
func (s *SignIn) PollFailed() bool {
	s.PollFailures++
	return s.PollFailures >= MaxPollFailures
}

// NormaliseURL checks an entered service address and drops a trailing
// slash, so paths can be appended.
func NormaliseURL(raw string) (string, error) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil ||
		u.RawQuery != "" || u.Fragment != "" {
		return "", ErrURL
	}
	return raw, nil
}

// ValidPage reports whether a sign-in page named by a service may be
// opened by the browser: http(s) only, never javascript: or data:.
func ValidPage(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && u.User == nil
}
