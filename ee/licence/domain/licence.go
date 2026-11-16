// Package domain holds licence keys and the rules that decide what a key
// unlocks (#307).
//
// Ring: Entities.
package domain

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"
)

// Prefix starts every key of format version 1. It is part of the signed
// bytes, so a key cannot be moved to another format version.
const Prefix = "THL1."

// Errors of a key that cannot be used. They never say which part was wrong
// beyond this, so a forged key teaches nothing.
var (
	ErrMalformed = errors.New("the licence key is not in the expected format")
	ErrSignature = errors.New("the licence key's signature is not valid")
)

// Features that can be licensed. A key may name others (from a newer
// version); they are kept but unlock nothing here.
const (
	FeatureAnnouncements     = "announcements"      // #297
	FeatureAbsences          = "absences"           // #299
	FeatureTeamTabs          = "team-tabs"          // #295
	FeatureSSO               = "sso"                // #302
	FeatureAuditLog          = "audit-log"          // #303
	FeaturePolicies          = "policies"           // #304
	FeatureBranding          = "branding"           // #305
	FeatureConnectedAccounts = "connected-accounts" // #326
	FeatureWorkplaceSearch   = "workplace-search"   // #327–#331
	FeatureChatApps          = "chat-apps"          // #332–#334
	FeaturePresence          = "presence"           // #346
	FeatureDirectorySync     = "directory-sync"     // #396
)

// KnownFeatures lists the features this version can unlock.
var KnownFeatures = []string{
	FeatureAnnouncements, FeatureAbsences, FeatureTeamTabs, FeatureSSO, FeatureAuditLog, FeaturePolicies,
	FeatureBranding, FeatureConnectedAccounts, FeatureWorkplaceSearch, FeatureChatApps, FeaturePresence,
	FeatureDirectorySync,
}

// Licence is the signed content of a key.
type Licence struct {
	ID       string    `json:"id"`
	Licensee string    `json:"licensee"`
	Issued   time.Time `json:"issued"`
	// Until is the last day the key is valid (UTC); it runs until the end of it.
	Until time.Time `json:"until"`
	// Seats is the number of accounts the key covers: the owner, members and
	// guests, without kiosk displays.
	Seats    int      `json:"seats"`
	Features []string `json:"features"`
}

// payload is the JSON form; dates are plain days so a signing tool cannot
// get time zones wrong.
type payload struct {
	ID       string   `json:"id"`
	Licensee string   `json:"licensee"`
	Issued   string   `json:"issued"`
	Until    string   `json:"until"`
	Seats    int      `json:"seats"`
	Features []string `json:"features"`
}

var b64 = base64.RawURLEncoding

// Parse checks the signature of key against the public key and returns its
// content. The format is Prefix + base64url(JSON) + "." + base64url(signature
// over Prefix + base64url(JSON)).
func Parse(key string, public ed25519.PublicKey) (Licence, error) {
	key = strings.TrimSpace(key)
	body, ok := strings.CutPrefix(key, Prefix)
	if !ok {
		return Licence{}, ErrMalformed
	}
	encoded, sigText, ok := strings.Cut(body, ".")
	if !ok {
		return Licence{}, ErrMalformed
	}
	sig, err := b64.DecodeString(sigText)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return Licence{}, ErrMalformed
	}
	if len(public) != ed25519.PublicKeySize || !ed25519.Verify(public, []byte(Prefix+encoded), sig) {
		return Licence{}, ErrSignature
	}
	raw, err := b64.DecodeString(encoded)
	if err != nil {
		return Licence{}, ErrMalformed
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var p payload
	if err := dec.Decode(&p); err != nil {
		return Licence{}, ErrMalformed
	}
	issued, err1 := time.Parse(time.DateOnly, p.Issued)
	until, err2 := time.Parse(time.DateOnly, p.Until)
	if err1 != nil || err2 != nil || p.Seats < 1 || strings.TrimSpace(p.Licensee) == "" || until.Before(issued) {
		return Licence{}, ErrMalformed
	}
	return Licence{ID: p.ID, Licensee: p.Licensee, Issued: issued, Until: until, Seats: p.Seats, Features: p.Features}, nil
}

// Sign builds a key; the real signing tool lives outside the repository with
// the private key, this is the same format for tests.
func Sign(l Licence, private ed25519.PrivateKey) string {
	raw, _ := json.Marshal(payload{
		ID: l.ID, Licensee: l.Licensee, Issued: l.Issued.Format(time.DateOnly), Until: l.Until.Format(time.DateOnly),
		Seats: l.Seats, Features: l.Features,
	})
	encoded := b64.EncodeToString(raw)
	return Prefix + encoded + "." + b64.EncodeToString(ed25519.Sign(private, []byte(Prefix+encoded)))
}

// Expired reports whether now is after the last valid day.
func (l Licence) Expired(now time.Time) bool {
	return !now.UTC().Before(l.Until.AddDate(0, 0, 1))
}

// Includes reports whether the key names a feature.
func (l Licence) Includes(feature string) bool { return slices.Contains(l.Features, feature) }

// UnknownFeatures lists the features of the key this version does not know.
func (l Licence) UnknownFeatures() []string {
	var out []string
	for _, f := range l.Features {
		if !slices.Contains(KnownFeatures, f) {
			out = append(out, f)
		}
	}
	return out
}

// Status describes the installed key as a whole.
type Status string

// The statuses of a key.
const (
	StatusMissing   Status = "missing"
	StatusInvalid   Status = "invalid"
	StatusValid     Status = "valid"
	StatusExpired   Status = "expired"
	StatusOverSeats Status = "over-seats"
)

// Level is what a feature may do, from nothing to full use.
type Level int

// The levels, from nothing to full use.
const (
	LevelNone Level = iota
	// LevelRead keeps data readable and refuses changes.
	LevelRead
	LevelFull
)

// Evaluate decides the status of a key and how far a feature may be used.
// An expired key, or one covering fewer accounts than the instance has,
// keeps the features it names readable: nothing is lost, but changes need a
// valid key ("a key for every account").
func Evaluate(l *Licence, now time.Time, accounts int) Status {
	switch {
	case l == nil:
		return StatusMissing
	case l.Expired(now):
		return StatusExpired
	case accounts > l.Seats:
		return StatusOverSeats
	default:
		return StatusValid
	}
}

// LevelOf is the level of a feature under a key with the given status.
func LevelOf(l *Licence, status Status, feature string) Level {
	if l == nil || !slices.Contains(KnownFeatures, feature) || !l.Includes(feature) {
		return LevelNone
	}
	switch status {
	case StatusValid:
		return LevelFull
	case StatusExpired, StatusOverSeats:
		return LevelRead
	default:
		return LevelNone
	}
}
