package stream

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"strings"
	"testing"
	"time"

	"waylight/internal/pushrelay"
)

func newKey(t *testing.T, curve elliptic.Curve) *ecdsa.PrivateKey {
	t.Helper()
	k, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func TestVerify(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	key := newKey(t, elliptic.P256())
	header, err := Authorization(key, now)
	if err != nil {
		t.Fatal(err)
	}
	id, err := Verify(header, now.Add(30*time.Second))
	spki, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil || id != DeviceID(spki) || !ValidDeviceID(id) {
		t.Fatalf("id = %q, %v", id, err)
	}

	stale, _ := Authorization(key, now.Add(-2*time.Minute))
	other, _ := Authorization(newKey(t, elliptic.P384()), now)
	tampered := strings.Replace(header, "ts=", "ts=1", 1)
	for name, h := range map[string]string{"stale": stale, "P-384": other, "tampered": tampered, "empty": "", "scheme": "Bearer x"} {
		if _, err := Verify(h, now); !errors.Is(err, ErrUnauthorized) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestHubReplacesAndWakes(t *testing.T) {
	h := NewHub()
	first, cancelFirst := h.Subscribe("d")
	second, cancelSecond := h.Subscribe("d")
	select {
	case <-first.Replaced:
	default:
		t.Fatal("the older stream must be replaced")
	}
	h.Notify("d")
	h.Notify("d") // a pending signal is enough; this must not block
	select {
	case <-second.Wake:
	default:
		t.Fatal("no wake signal")
	}
	cancelFirst() // removing the replaced stream keeps the newer one
	if h.Open() != 1 {
		t.Fatalf("open = %d", h.Open())
	}
	cancelSecond()
	if h.Open() != 0 {
		t.Fatalf("open = %d", h.Open())
	}
	h.Notify("nobody")
}

func TestGate(t *testing.T) {
	g := &Gate{MaxPerClient: 2, MaxTotal: 3}
	for range 2 {
		if !g.Acquire("a") {
			t.Fatal("a free stream was refused")
		}
	}
	if g.Acquire("a") {
		t.Fatal("per-client limit")
	}
	if !g.Acquire("b") || g.Acquire("c") {
		t.Fatal("total limit")
	}
	g.Release("a")
	if !g.Acquire("c") {
		t.Fatal("released stream not free")
	}
}

type memBox struct {
	put   []string
	limit int
}

func (m *memBox) Put(_ context.Context, device string, _ []byte, _ time.Time, limit int) (int64, error) {
	m.put = append(m.put, device)
	m.limit = limit
	return int64(len(m.put)), nil
}
func (m *memBox) After(context.Context, string, int64) ([]Event, error) { return nil, nil }
func (m *memBox) Ack(context.Context, string, int64) error              { return nil }

func TestSender(t *testing.T) {
	box, hub := &memBox{}, NewHub()
	s := &Sender{Mailbox: box, Hub: hub, Now: time.Now}
	spki, _ := x509.MarshalPKIXPublicKey(&newKey(t, elliptic.P256()).PublicKey)
	id := DeviceID(spki)
	sub, cancel := hub.Subscribe(id)
	defer cancel()
	if err := s.Send(context.Background(), id, []byte{1}); err != nil || box.limit != MaxPerDevice {
		t.Fatalf("send: %v, limit %d", err, box.limit)
	}
	select {
	case <-sub.Wake:
	default:
		t.Fatal("stream not woken")
	}
	if s.ValidToken("not-an-id") || !errors.Is(s.Send(context.Background(), "x", []byte{1}), pushrelay.ErrInvalid) {
		t.Fatal("a malformed token must be invalid")
	}
}
