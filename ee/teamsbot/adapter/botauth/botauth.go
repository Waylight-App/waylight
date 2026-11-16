// Package botauth checks the token Microsoft's Bot Connector sends with
// every request to the Teams command box search (#334, ADR 0023). The
// route is reachable from the internet, so nothing runs before this check
// has passed; it follows "Authenticate requests from the Bot Connector
// service to your bot" of the Bot Framework documentation.
//
// Ring: Interface adapters.
package botauth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
)

// Defaults of the Bot Framework.
const (
	MetadataURL = "https://login.botframework.com/v1/.well-known/openidconfiguration"
	Issuer      = "https://api.botframework.com"
	// clockSkew is the tolerance the documentation names for exp and nbf.
	clockSkew = 5 * time.Minute
	// keysMaxAge: the documentation asks for a refresh at least daily.
	keysMaxAge = 24 * time.Hour
	// refetchPause bounds how often an unknown key ID makes us refetch, so
	// forged tokens cannot turn the bot into a request amplifier.
	refetchPause = 5 * time.Minute
)

// ErrUnauthorized is any failed check; the reason is not told to the caller.
var ErrUnauthorized = errors.New("bot connector token not valid")

// Verifier checks Bot Connector tokens.
type Verifier struct {
	HTTP        *http.Client
	MetadataURL string
	AppID       string
	Now         func() time.Time

	mu      sync.Mutex
	keys    map[string]key
	fetched time.Time
	tried   time.Time
}

type key struct {
	pub          *rsa.PublicKey
	endorsements []string
}

// Activity is what the checks need from the request body.
type Activity struct {
	ServiceURL string
	ChannelID  string
}

// Verify checks the Authorization header against the activity.
func (v *Verifier) Verify(ctx context.Context, authorization string, a Activity) error {
	raw, ok := strings.CutPrefix(authorization, "Bearer ")
	if !ok || v.AppID == "" {
		return ErrUnauthorized
	}
	parts := strings.Split(strings.TrimSpace(raw), ".")
	if len(parts) != 3 {
		return ErrUnauthorized
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if decodePart(parts[0], &header) != nil || header.Alg != "RS256" || header.Kid == "" {
		return ErrUnauthorized
	}
	k, err := v.key(ctx, header.Kid)
	if err != nil {
		return ErrUnauthorized
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return ErrUnauthorized
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if rsa.VerifyPKCS1v15(k.pub, crypto.SHA256, digest[:], sig) != nil {
		return ErrUnauthorized
	}
	var claims struct {
		Iss string          `json:"iss"`
		Aud json.RawMessage `json:"aud"`
		Exp int64           `json:"exp"`
		Nbf int64           `json:"nbf"`
		// Go matches JSON keys without regard to case, so this also reads
		// the "serviceUrl" the Bot Connector sends.
		ServiceURL string `json:"serviceurl"`
	}
	if decodePart(parts[1], &claims) != nil {
		return ErrUnauthorized
	}
	now := v.Now()
	service := claims.ServiceURL
	switch {
	case claims.Iss != Issuer:
		return ErrUnauthorized
	case !audience(claims.Aud, v.AppID):
		return ErrUnauthorized
	case claims.Exp == 0 || now.After(time.Unix(claims.Exp, 0).Add(clockSkew)):
		return ErrUnauthorized
	case claims.Nbf != 0 && now.Add(clockSkew).Before(time.Unix(claims.Nbf, 0)):
		return ErrUnauthorized
	case service == "" || service != a.ServiceURL:
		return ErrUnauthorized
	case a.ChannelID == "" || !slices.Contains(k.endorsements, a.ChannelID):
		return ErrUnauthorized
	}
	return nil
}

func decodePart(s string, dst any) error {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// audience accepts "aud" as a string or a list.
func audience(raw json.RawMessage, appID string) bool {
	var one string
	if json.Unmarshal(raw, &one) == nil {
		return one == appID
	}
	var many []string
	return json.Unmarshal(raw, &many) == nil && slices.Contains(many, appID)
}

// key returns the signing key with kid, fetching the key list when it is
// older than a day or, at most every five minutes, when kid is unknown.
func (v *Verifier) key(ctx context.Context, kid string) (key, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	now := v.Now()
	k, ok := v.keys[kid]
	stale := now.Sub(v.fetched) > keysMaxAge
	if ok && !stale {
		return k, nil
	}
	if !stale && now.Sub(v.tried) < refetchPause {
		return key{}, ErrUnauthorized
	}
	v.tried = now
	keys, err := v.fetch(ctx)
	if err != nil {
		if ok {
			return k, nil // keep using a known key while Microsoft is unreachable
		}
		return key{}, err
	}
	v.keys, v.fetched = keys, now
	if k, ok = keys[kid]; !ok {
		return key{}, ErrUnauthorized
	}
	return k, nil
}

func (v *Verifier) getJSON(ctx context.Context, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := v.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %s", url, res.Status)
	}
	return json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(dst)
}

func (v *Verifier) fetch(ctx context.Context) (map[string]key, error) {
	var meta struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := v.getJSON(ctx, v.MetadataURL, &meta); err != nil {
		return nil, err
	}
	if !strings.HasPrefix(meta.JWKSURI, "https://") && !strings.HasPrefix(v.MetadataURL, "http://") {
		return nil, errors.New("bot connector keys not served over https")
	}
	var set struct {
		Keys []struct {
			Kty          string   `json:"kty"`
			Kid          string   `json:"kid"`
			N            string   `json:"n"`
			E            string   `json:"e"`
			Endorsements []string `json:"endorsements"`
		} `json:"keys"`
	}
	if err := v.getJSON(ctx, meta.JWKSURI, &set); err != nil {
		return nil, err
	}
	out := map[string]key{}
	for _, k := range set.Keys {
		if k.Kty != "RSA" || k.Kid == "" {
			continue
		}
		n, err1 := base64.RawURLEncoding.DecodeString(k.N)
		e, err2 := base64.RawURLEncoding.DecodeString(k.E)
		if err1 != nil || err2 != nil || len(e) == 0 || len(e) > 4 {
			continue
		}
		exp := 0
		for _, b := range e {
			exp = exp<<8 | int(b)
		}
		pub := &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exp}
		if pub.N.BitLen() < 2048 {
			continue
		}
		out[k.Kid] = key{pub: pub, endorsements: k.Endorsements}
	}
	if len(out) == 0 {
		return nil, errors.New("bot connector published no usable keys")
	}
	return out, nil
}
