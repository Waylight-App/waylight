package stream

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Scheme is the authorization scheme of the stream request.
const Scheme = "Waylight-Device"

// MaxSkew is how far the time stamp of a stream request may differ from the
// relay's clock.
const MaxSkew = time.Minute

// ErrUnauthorized means the device signature is missing, malformed, stale
// or wrong.
var ErrUnauthorized = errors.New("device not authorized")

var b64 = base64.RawURLEncoding

// DeviceID is the device token instances send to: the base64url SHA-256 of
// the device's public key in SubjectPublicKeyInfo form.
func DeviceID(spki []byte) string {
	sum := sha256.Sum256(spki)
	return b64.EncodeToString(sum[:])
}

// signedText is what the device signs: a fixed context and the time stamp.
func signedText(ts int64) []byte {
	return []byte("waylight-stream-v1\n" + strconv.FormatInt(ts, 10))
}

// Authorization builds the header value for key at now. Devices sign with a
// P-256 key in their secure storage; this function serves tests and tools.
func Authorization(key *ecdsa.PrivateKey, now time.Time) (string, error) {
	spki, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return "", fmt.Errorf("encode public key: %w", err)
	}
	ts := now.Unix()
	digest := sha256.Sum256(signedText(ts))
	sig, err := ecdsa.SignASN1(rand.Reader, key, digest[:])
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}
	return fmt.Sprintf("%s key=%s, ts=%d, sig=%s", Scheme, b64.EncodeToString(spki), ts, b64.EncodeToString(sig)), nil
}

// Verify checks the header of a stream request and returns the device ID.
// The key must be an ECDSA P-256 key; the signature covers the time stamp,
// which must lie within MaxSkew of now.
func Verify(header string, now time.Time) (string, error) {
	params, ok := strings.CutPrefix(header, Scheme+" ")
	if !ok {
		return "", ErrUnauthorized
	}
	var keyText, tsText, sigText string
	for part := range strings.SplitSeq(params, ",") {
		name, value, _ := strings.Cut(strings.TrimSpace(part), "=")
		switch name {
		case "key":
			keyText = value
		case "ts":
			tsText = value
		case "sig":
			sigText = value
		}
	}
	spki, err1 := b64.DecodeString(keyText)
	sig, err2 := b64.DecodeString(sigText)
	ts, err3 := strconv.ParseInt(tsText, 10, 64)
	if err := errors.Join(err1, err2, err3); err != nil {
		return "", ErrUnauthorized
	}
	if d := now.Sub(time.Unix(ts, 0)); d > MaxSkew || d < -MaxSkew {
		return "", ErrUnauthorized
	}
	parsed, err := x509.ParsePKIXPublicKey(spki)
	if err != nil {
		return "", ErrUnauthorized
	}
	pub, ok := parsed.(*ecdsa.PublicKey)
	if !ok || pub.Curve.Params().Name != "P-256" {
		return "", ErrUnauthorized
	}
	digest := sha256.Sum256(signedText(ts))
	if !ecdsa.VerifyASN1(pub, digest[:], sig) {
		return "", ErrUnauthorized
	}
	return DeviceID(spki), nil
}

// ValidDeviceID reports whether s has the form of a device ID.
func ValidDeviceID(s string) bool {
	raw, err := b64.DecodeString(s)
	return err == nil && len(raw) == sha256.Size
}
