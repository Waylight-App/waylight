package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
)

func TestSignES256(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tok, err := Sign(key, map[string]any{"kid": "K"}, map[string]any{"iss": "T"})
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(tok, ".")
	sig, _ := b64.DecodeString(parts[2])
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	r, s := new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:])
	if len(sig) != 64 || !ecdsa.Verify(&key.PublicKey, sum[:], r, s) {
		t.Fatal("bad ES256 signature")
	}
	if h, _ := b64.DecodeString(parts[0]); !strings.Contains(string(h), `"alg":"ES256"`) {
		t.Fatal(string(h))
	}
}

func TestParseKeyAndRS256(t *testing.T) {
	rk, _ := rsa.GenerateKey(rand.Reader, 2048)
	der, _ := x509.MarshalPKCS8PrivateKey(rk)
	key, err := ParseKey(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Sign(key, map[string]any{}, map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseKey([]byte("nope")); err == nil {
		t.Fatal("no PEM must fail")
	}
}
