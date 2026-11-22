package usecase

import (
	"context"
	"crypto/subtle"
	"sync"
	"time"

	"waylight/internal/modules/setupsignin/domain"
	"waylight/internal/shared/kernel"
)

// PasteDeps bundles the output ports of the pastes.
type PasteDeps struct {
	Clock kernel.Clock
	// Token returns an unguessable random string for codes.
	Token func() string
}

// Pastes is the input port of Paste from phone: the assistant shows a code
// as a QR code, the signed-in phone sends the copied key with it, and the
// assistant reads the key once.
type Pastes interface {
	// Create starts a code for the caller and replaces an earlier one.
	Create(ctx context.Context) (code string, expires time.Time, err error)
	// Submit stores the phone's value; only the account that created the
	// code may, and only once.
	Submit(ctx context.Context, code, value string) error
	// Take returns the value once it arrived, and forgets the code. ready
	// stays false while the phone has not sent it.
	Take(ctx context.Context, code string) (value string, ready bool, err error)
	// Cancel forgets a code, e.g. when the assistant closes.
	Cancel(ctx context.Context, code string) error
	// Check reports whether the code still waits for a value. It serves the
	// page of a phone without the app, which has no account: the code is
	// its only credential, so wrong codes count against the client.
	Check(client, code string) error
	// Fill is Submit for that page.
	Fill(client, code, value string) error
}

type pastes struct {
	d       PasteDeps
	mu      sync.Mutex
	pending map[string]*domain.Paste
	misses  *missLimiter
}

// NewPastes returns the paste use cases.
func NewPastes(d PasteDeps) Pastes {
	return &pastes{d: d, pending: map[string]*domain.Paste{}, misses: newMissLimiter()}
}

func (uc *pastes) Create(ctx context.Context) (string, time.Time, error) {
	acc, err := account(ctx)
	if err != nil {
		return "", time.Time{}, err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.sweep()
	// One code per account: a new one replaces an abandoned one.
	for code, p := range uc.pending {
		if p.Account == acc {
			delete(uc.pending, code)
		}
	}
	if len(uc.pending) >= domain.MaxPastes {
		return "", time.Time{}, errTooMany
	}
	p := domain.NewPaste(uc.d.Token(), acc, uc.d.Clock.Now())
	uc.pending[p.Code] = p
	return p.Code, p.Expires, nil
}

// sweep drops expired codes; callers hold mu.
func (uc *pastes) sweep() {
	now := uc.d.Clock.Now()
	for code, p := range uc.pending {
		if p.Expired(now) {
			delete(uc.pending, code)
		}
	}
}

// own returns the caller's live code; callers hold mu.
func (uc *pastes) own(acc, code string) *domain.Paste {
	uc.sweep()
	p := uc.pending[code]
	if p == nil || p.Account != acc {
		return nil
	}
	return p
}

func (uc *pastes) Submit(ctx context.Context, code, value string) error {
	acc, err := account(ctx)
	if err != nil {
		return err
	}
	v, err := domain.CleanPasteValue(value)
	if err != nil {
		return err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	p := uc.own(acc, code)
	if p == nil || p.Ready() {
		return domain.ErrPasteUnknown
	}
	p.Value = v
	return nil
}

func (uc *pastes) Take(ctx context.Context, code string) (string, bool, error) {
	acc, err := account(ctx)
	if err != nil {
		return "", false, err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	p := uc.own(acc, code)
	if p == nil {
		return "", false, domain.ErrPasteUnknown
	}
	if !p.Ready() {
		return "", false, nil
	}
	delete(uc.pending, code)
	return p.Value, true, nil
}

func (uc *pastes) Cancel(ctx context.Context, code string) error {
	acc, err := account(ctx)
	if err != nil {
		return err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if uc.own(acc, code) != nil {
		delete(uc.pending, code)
	}
	return nil
}

// waiting returns the live code that has no value yet. Every code is
// compared in constant time so the lookup time does not reveal how much of
// a guess was right. The client's wrong guesses are counted; callers hold mu.
func (uc *pastes) waiting(client, code string) (*domain.Paste, error) {
	now := uc.d.Clock.Now()
	if !uc.misses.allow(client, now) {
		return nil, errTooManyGuesses
	}
	uc.sweep()
	var found *domain.Paste
	for c, p := range uc.pending {
		if subtle.ConstantTimeCompare([]byte(c), []byte(code)) == 1 {
			found = p
		}
	}
	if found == nil || found.Ready() {
		uc.misses.fail(client, now)
		return nil, domain.ErrPasteUnknown
	}
	return found, nil
}

func (uc *pastes) Check(client, code string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	_, err := uc.waiting(client, code)
	return err
}

func (uc *pastes) Fill(client, code, value string) error {
	v, err := domain.CleanPasteValue(value)
	if err != nil {
		return err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	p, err := uc.waiting(client, code)
	if err != nil {
		return err
	}
	p.Value = v
	return nil
}
