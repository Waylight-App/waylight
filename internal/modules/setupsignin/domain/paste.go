package domain

import (
	"strings"
	"time"

	"waylight/internal/shared/kernel"
)

// Limits of pastes from the phone. Like sign-ins they live in memory only.
const (
	// PasteTTL bounds a code from its creation; scanning and copying a key
	// takes a minute, not more.
	PasteTTL = 5 * time.Minute
	// MaxPastes bounds the codes of all accounts together.
	MaxPastes = 50
	// MaxPasteValue bounds a pasted key; real tokens are far shorter.
	MaxPasteValue = 4096
)

// ErrPasteUnknown is returned for a code that does not exist (any more) for
// the asking account, or that already holds a value.
var ErrPasteUnknown = kernel.NewPublicError(kernel.ErrInvalid, "The code has expired. Show a new code and scan it again.")

// ErrPasteValue is returned for an empty or oversized value.
var ErrPasteValue = kernel.NewValidationError("value", "The value must not be empty or longer than 4096 characters.")

// Paste is a single-use code that carries one copied key from the phone to
// the setup assistant of the account that created it.
type Paste struct {
	Code    string
	Account string
	Expires time.Time
	// Value is empty until the phone has sent it.
	Value string
}

// NewPaste creates a code that ends after PasteTTL.
func NewPaste(code, account string, now time.Time) *Paste {
	return &Paste{Code: code, Account: account, Expires: now.Add(PasteTTL)}
}

// Expired reports whether the code is over.
func (p *Paste) Expired(now time.Time) bool { return !now.Before(p.Expires) }

// Ready reports whether the phone has sent the value.
func (p *Paste) Ready() bool { return p.Value != "" }

// CleanPasteValue trims a value from the phone and rejects an empty or
// oversized one.
func CleanPasteValue(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" || len(v) > MaxPasteValue {
		return "", ErrPasteValue
	}
	return v, nil
}
