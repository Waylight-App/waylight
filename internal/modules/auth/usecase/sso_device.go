package usecase

import (
	"context"
	"errors"
	"time"

	"waylight/internal/modules/auth/domain"
)

// Sign in with Microsoft by code (#535): the device authorization grant needs
// no redirect back to this server, so it works without an app registration
// per server, without HTTPS and inside a home network.

// DeviceStart is what the provider returns when a device flow starts.
type DeviceStart struct {
	DeviceCode      string
	UserCode        string
	VerificationURI string
	ExpiresIn       time.Duration
	Interval        time.Duration
}

// Errors a DeviceProvider returns from PollDevice besides a failure.
var (
	ErrDevicePending  = errors.New("device flow pending")
	ErrDeviceSlowDown = errors.New("device flow slow down")
	ErrDeviceGone     = errors.New("device flow expired or declined")
)

// DeviceProvider is the provider of the device flow (output port).
type DeviceProvider interface {
	StartDevice(ctx context.Context) (DeviceStart, error)
	// PollDevice returns the confirmed identity, or ErrDevicePending,
	// ErrDeviceSlowDown or ErrDeviceGone.
	PollDevice(ctx context.Context, deviceCode string) (domain.ExternalIdentity, error)
}

// DeviceFlow is a started flow as the page needs it.
type DeviceFlow struct {
	// ID identifies the flow when the page polls; the device code itself
	// never leaves the server.
	ID              string
	UserCode        string
	VerificationURI string
	ExpiresAt       time.Time
	Interval        time.Duration
}

// DeviceResult is the answer to a poll.
type DeviceResult struct {
	// Pending: not confirmed yet, poll again after the interval.
	Pending bool
	// Linked: the sign-in was linked to the account that started the flow.
	Linked bool
	// Session or Ticket after a sign-in, as for the other single sign-on.
	Session SessionResponse
	Ticket  string
}

// deviceFlow is a running flow; link is the account that links, "" to sign in.
type deviceFlow struct {
	code      string
	link      string
	expires   time.Time
	interval  time.Duration
	nextPoll  time.Time
	userCode  string
	verifyURI string
}

// maxDeviceFlows bounds the flows in memory, like the pending redirects.
const maxDeviceFlows = 64

// BeginDevice starts a device flow: to link a sign-in to linkAccount, or to
// sign in when linkAccount is "".
func (uc *sso) BeginDevice(ctx context.Context, linkAccount string) (DeviceFlow, error) {
	if uc.cfg.Device == nil || uc.deps.Identities == nil {
		return DeviceFlow{}, domain.ErrSSOUnavailable
	}
	if linkAccount != "" && !uc.mayLink(ctx, linkAccount) {
		return DeviceFlow{}, domain.ErrLinkNeedsLicence
	}
	start, err := uc.cfg.Device.StartDevice(ctx)
	if err != nil {
		return DeviceFlow{}, errProviderUnreachable(err)
	}
	id, err := uc.token()
	if err != nil {
		return DeviceFlow{}, err
	}
	now := uc.deps.Clock.Now()
	f := deviceFlow{code: start.DeviceCode, link: linkAccount, expires: now.Add(start.ExpiresIn), interval: start.Interval,
		nextPoll: now.Add(start.Interval), userCode: start.UserCode, verifyURI: start.VerificationURI}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	for k, d := range uc.devices {
		if now.After(d.expires) {
			delete(uc.devices, k)
		}
	}
	if len(uc.devices) >= maxDeviceFlows {
		return DeviceFlow{}, domain.ErrTooManyAttempts
	}
	uc.devices[id] = f
	return DeviceFlow{ID: id, UserCode: f.userCode, VerificationURI: f.verifyURI, ExpiresAt: f.expires, Interval: f.interval}, nil
}

// PollDevice asks the provider about a flow. A poll before the interval has
// passed does not reach the provider; it is answered as pending.
func (uc *sso) PollDevice(ctx context.Context, flowID string, client domain.Client) (DeviceResult, error) {
	if uc.cfg.Device == nil {
		return DeviceResult{}, domain.ErrSSOUnavailable
	}
	now := uc.deps.Clock.Now()
	uc.mu.Lock()
	f, ok := uc.devices[flowID]
	if ok && now.After(f.expires) {
		delete(uc.devices, flowID)
		ok = false
	}
	if ok && now.Before(f.nextPoll) {
		uc.mu.Unlock()
		return DeviceResult{Pending: true}, nil
	}
	if ok {
		// Reserve the next slot before asking, so parallel polls of the
		// same flow do not hit the provider twice.
		f.nextPoll = now.Add(f.interval)
		uc.devices[flowID] = f
	}
	uc.mu.Unlock()
	if !ok {
		return DeviceResult{}, domain.ErrDeviceExpired
	}

	id, err := uc.cfg.Device.PollDevice(ctx, f.code)
	switch {
	case errors.Is(err, ErrDevicePending):
		return DeviceResult{Pending: true}, nil
	case errors.Is(err, ErrDeviceSlowDown):
		uc.mu.Lock()
		if cur, still := uc.devices[flowID]; still {
			cur.interval += 5 * time.Second
			cur.nextPoll = now.Add(cur.interval)
			uc.devices[flowID] = cur
		}
		uc.mu.Unlock()
		return DeviceResult{Pending: true}, nil
	case errors.Is(err, ErrDeviceGone):
		uc.forgetDevice(flowID)
		return DeviceResult{}, domain.ErrDeviceExpired
	case errors.Is(err, domain.ErrSSODirectoryNotAllowed):
		uc.forgetDevice(flowID)
		return DeviceResult{}, err
	case err != nil:
		return DeviceResult{}, errProviderUnreachable(err)
	}
	uc.forgetDevice(flowID)
	// Work and school accounts carry the issuer of their directory, so the
	// same person in two directories stays two links (#537).
	id.Provider = "microsoft"
	if id.Issuer == "" {
		id.Issuer = uc.cfg.DeviceIssuer
	}
	if f.link != "" {
		if _, _, err := uc.completeLink(ctx, f.link, id, client); err != nil {
			return DeviceResult{}, err
		}
		return DeviceResult{Linked: true}, nil
	}
	res, ticket, err := uc.finish(ctx, id, client)
	if errors.Is(err, domain.ErrSSOSecondFactorNeed) {
		return DeviceResult{Ticket: ticket}, nil
	}
	if err != nil {
		return DeviceResult{}, err
	}
	return DeviceResult{Session: res}, nil
}

func (uc *sso) forgetDevice(flowID string) {
	uc.mu.Lock()
	delete(uc.devices, flowID)
	uc.mu.Unlock()
}
