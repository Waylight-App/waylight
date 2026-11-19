package pushproto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Scheme is the authorization scheme of signed relay requests.
const Scheme = "Waylight-Push"

// MaxSkew is how far a request's time stamp may be from the relay's clock.
const MaxSkew = 5 * time.Minute

var (
	// ErrMalformed means the Authorization header cannot be parsed.
	ErrMalformed = errors.New("pushproto: malformed authorization")
	// ErrExpired means the time stamp is outside MaxSkew.
	ErrExpired = errors.New("pushproto: request expired")
	// ErrSignature means the signature does not match key and body.
	ErrSignature = errors.New("pushproto: bad signature")
)

var b64 = base64.RawURLEncoding

// signed is the byte string an instance signs: a context label, the time
// stamp and the body's hash, so a signature cannot be reused for another
// body or later than MaxSkew.
func signed(ts int64, body []byte) []byte {
	sum := sha256.Sum256(body)
	return []byte("waylight-push-v1\n" + strconv.FormatInt(ts, 10) + "\n" + hex.EncodeToString(sum[:]))
}

// Authorization returns the header value that signs body at now.
func Authorization(key ed25519.PrivateKey, body []byte, now time.Time) string {
	ts := now.Unix()
	pub := key.Public().(ed25519.PublicKey)
	sig := ed25519.Sign(key, signed(ts, body))
	return fmt.Sprintf("%s key=%s, ts=%d, sig=%s", Scheme, b64.EncodeToString(pub), ts, b64.EncodeToString(sig))
}

// Verify checks the header against body and now and returns the signing
// instance's public key, which is its identity.
func Verify(header string, body []byte, now time.Time) (ed25519.PublicKey, error) {
	rest, ok := strings.CutPrefix(header, Scheme+" ")
	if !ok {
		return nil, ErrMalformed
	}
	fields := map[string]string{}
	for part := range strings.SplitSeq(rest, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			return nil, ErrMalformed
		}
		fields[k] = v
	}
	pub, err1 := b64.DecodeString(fields["key"])
	sig, err2 := b64.DecodeString(fields["sig"])
	ts, err3 := strconv.ParseInt(fields["ts"], 10, 64)
	if err := errors.Join(err1, err2, err3); err != nil || len(pub) != ed25519.PublicKeySize || len(sig) != ed25519.SignatureSize {
		return nil, ErrMalformed
	}
	if d := now.Sub(time.Unix(ts, 0)); d > MaxSkew || d < -MaxSkew {
		return nil, ErrExpired
	}
	if !ed25519.Verify(pub, signed(ts, body), sig) {
		return nil, ErrSignature
	}
	return pub, nil
}

// KeyID is the printable identity of an instance key.
func KeyID(pub ed25519.PublicKey) string { return b64.EncodeToString(pub) }
