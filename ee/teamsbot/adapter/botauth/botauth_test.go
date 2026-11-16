package botauth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type fakeConnector struct {
	priv      *rsa.PrivateKey
	srv       *httptest.Server
	keyFetchs atomic.Int32
}

func newConnector(t *testing.T) *fakeConnector {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeConnector{priv: priv}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /meta", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"issuer": Issuer, "jwks_uri": f.srv.URL + "/keys"})
	})
	mux.HandleFunc("GET /keys", func(w http.ResponseWriter, _ *http.Request) {
		f.keyFetchs.Add(1)
		b64 := base64.RawURLEncoding.EncodeToString
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]any{{
			"kty": "RSA", "kid": "k1", "n": b64(priv.N.Bytes()), "e": b64(big.NewInt(int64(priv.E)).Bytes()),
			"endorsements": []string{"msteams", "webchat"},
		}}})
	})
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeConnector) token(t *testing.T, header, claims map[string]any) string {
	t.Helper()
	enc := func(v any) string { b, _ := json.Marshal(v); return base64.RawURLEncoding.EncodeToString(b) }
	signed := enc(header) + "." + enc(claims)
	digest := sha256.Sum256([]byte(signed))
	sig, err := rsa.SignPKCS1v15(rand.Reader, f.priv, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + signed + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func TestVerify(t *testing.T) {
	f := newConnector(t)
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	v := &Verifier{HTTP: f.srv.Client(), MetadataURL: f.srv.URL + "/meta", AppID: "app-1", Now: func() time.Time { return now }}
	act := Activity{ServiceURL: "https://smba.trafficmanager.net/emea/", ChannelID: "msteams"}
	hdr := map[string]any{"alg": "RS256", "kid": "k1", "typ": "JWT"}
	good := func() map[string]any {
		return map[string]any{"iss": Issuer, "aud": "app-1", "exp": now.Add(time.Hour).Unix(), "nbf": now.Add(-time.Minute).Unix(),
			"serviceUrl": act.ServiceURL}
	}
	if err := v.Verify(context.Background(), f.token(t, hdr, good()), act); err != nil {
		t.Fatalf("valid token refused: %v", err)
	}
	bad := map[string]func(c map[string]any){
		"wrong audience":   func(c map[string]any) { c["aud"] = "someone-else" },
		"wrong issuer":     func(c map[string]any) { c["iss"] = "https://sts.windows.net/x/" },
		"expired":          func(c map[string]any) { c["exp"] = now.Add(-10 * time.Minute).Unix() },
		"not yet valid":    func(c map[string]any) { c["nbf"] = now.Add(10 * time.Minute).Unix() },
		"other serviceUrl": func(c map[string]any) { c["serviceUrl"] = "https://evil.example/" },
		"no expiry":        func(c map[string]any) { delete(c, "exp") },
	}
	for name, change := range bad {
		c := good()
		change(c)
		if err := v.Verify(context.Background(), f.token(t, hdr, c), act); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	// Within the five minutes of clock skew it still passes.
	c := good()
	c["exp"] = now.Add(-4 * time.Minute).Unix()
	if err := v.Verify(context.Background(), f.token(t, hdr, c), act); err != nil {
		t.Errorf("clock skew: %v", err)
	}
	if err := v.Verify(context.Background(), f.token(t, hdr, good()), Activity{ServiceURL: act.ServiceURL, ChannelID: "slack"}); err == nil {
		t.Error("channel without endorsement accepted")
	}
	tok := f.token(t, hdr, good())
	if err := v.Verify(context.Background(), tok[:len(tok)-4]+"AAAA", act); err == nil {
		t.Error("tampered signature accepted")
	}
	none := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","kid":"k1"}`)) + "." +
		base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"`+Issuer+`","aud":"app-1"}`)) + "."
	if err := v.Verify(context.Background(), "Bearer "+none, act); err == nil {
		t.Error("alg none accepted")
	}
	if err := v.Verify(context.Background(), "", act); err == nil {
		t.Error("no header accepted")
	}
}

func TestUnknownKeysDoNotHammerMicrosoft(t *testing.T) {
	f := newConnector(t)
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	v := &Verifier{HTTP: f.srv.Client(), MetadataURL: f.srv.URL + "/meta", AppID: "app-1", Now: func() time.Time { return now }}
	act := Activity{ServiceURL: "https://s/", ChannelID: "msteams"}
	claims := map[string]any{"iss": Issuer, "aud": "app-1", "exp": now.Add(time.Hour).Unix(), "serviceUrl": "https://s/"}
	for range 20 {
		_ = v.Verify(context.Background(), f.token(t, map[string]any{"alg": "RS256", "kid": "forged"}, claims), act)
	}
	if n := f.keyFetchs.Load(); n != 1 {
		t.Fatalf("key list fetched %d times for unknown kids", n)
	}
	now = now.Add(6 * time.Minute)
	if err := v.Verify(context.Background(), f.token(t, map[string]any{"alg": "RS256", "kid": "k1"}, map[string]any{
		"iss": Issuer, "aud": "app-1", "exp": now.Add(time.Hour).Unix(), "serviceUrl": "https://s/"}), act); err != nil {
		t.Fatalf("known key after the pause: %v", err)
	}
}
