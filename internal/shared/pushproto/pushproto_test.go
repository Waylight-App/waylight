package pushproto

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSignAndVerify(t *testing.T) {
	_, key, _ := ed25519.GenerateKey(rand.Reader)
	now := time.Unix(1_800_000_000, 0)
	body := []byte(`{"platform":"sse"}`)
	h := Authorization(key, body, now)
	pub, err := Verify(h, body, now.Add(time.Minute))
	if err != nil || !pub.Equal(key.Public()) {
		t.Fatal(pub, err)
	}
	if _, err := Verify(h, []byte(`{"platform":"apns"}`), now); !errors.Is(err, ErrSignature) {
		t.Fatal("other body:", err)
	}
	if _, err := Verify(h, body, now.Add(6*time.Minute)); !errors.Is(err, ErrExpired) {
		t.Fatal("late:", err)
	}
	if _, err := Verify(h, body, now.Add(-6*time.Minute)); !errors.Is(err, ErrExpired) {
		t.Fatal("early:", err)
	}
	for _, bad := range []string{"", "Bearer x", Scheme + " key=a", strings.Replace(h, "ts=", "ts=x", 1)} {
		if _, err := Verify(bad, body, now); !errors.Is(err, ErrMalformed) {
			t.Fatalf("%q: %v", bad, err)
		}
	}
}

func TestSealAndOpen(t *testing.T) {
	device, _ := ecdh.X25519().GenerateKey(rand.Reader)
	plain := []byte(`{"title":"Backup failed","body":"nas","link":"waylight://open"}`)
	sealed, err := Seal(device.PublicKey(), plain)
	if err != nil || sealed[0] != version || len(sealed) != 1+32+12+len(plain)+16 {
		t.Fatal(len(sealed), err)
	}
	got, err := Open(device, sealed)
	if err != nil || string(got) != string(plain) {
		t.Fatal(string(got), err)
	}
	other, _ := ecdh.X25519().GenerateKey(rand.Reader)
	if _, err := Open(other, sealed); !errors.Is(err, ErrPayload) {
		t.Fatal("other key:", err)
	}
	sealed[len(sealed)-1] ^= 1
	if _, err := Open(device, sealed); !errors.Is(err, ErrPayload) {
		t.Fatal("tampered:", err)
	}
	if _, err := Open(device, []byte{2}); !errors.Is(err, ErrPayload) {
		t.Fatal("short:", err)
	}
}
