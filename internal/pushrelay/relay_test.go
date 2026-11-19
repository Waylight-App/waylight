package pushrelay

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeSender struct {
	sent []string
	err  error
}

func (f *fakeSender) Send(_ context.Context, token string, _ []byte) error {
	f.sent = append(f.sent, token)
	return f.err
}

type memBindings map[TokenHash]string

func (m memBindings) Bind(_ context.Context, h TokenHash, key string, _ time.Time) (string, error) {
	if k, ok := m[h]; ok {
		return k, nil
	}
	m[h] = key
	return key, nil
}

func (m memBindings) Delete(_ context.Context, h TokenHash) error { delete(m, h); return nil }

func newRelay(s Sender, b Bindings, max int) *Relay {
	w := func() *Window { return &Window{Max: max, Period: time.Hour} }
	now := time.Unix(1_800_000_000, 0)
	return New(map[Platform]Sender{FCM: s}, b, Limits{Instance: w(), Token: w(), Client: w()}, func() time.Time { return now })
}

func TestSend(t *testing.T) {
	ctx := context.Background()
	s, b := &fakeSender{}, memBindings{}
	r := newRelay(s, b, 10)
	msg := Message{Platform: FCM, Token: "tok", Payload: []byte{1}}
	if err := r.Send(ctx, "a", "1.2.3.4", msg); err != nil || len(s.sent) != 1 {
		t.Fatal(err, s.sent)
	}
	// The token is bound to instance a now.
	if err := r.Send(ctx, "b", "1.2.3.4", msg); !errors.Is(err, ErrForeignToken) {
		t.Fatal(err)
	}
	for _, m := range []Message{
		{Platform: APNs, Token: "tok", Payload: []byte{1}}, // no APNs sender configured
		{Platform: FCM, Payload: []byte{1}},
		{Platform: FCM, Token: "tok"},
	} {
		if err := r.Send(ctx, "a", "1.2.3.4", m); !errors.Is(err, ErrInvalid) {
			t.Fatalf("%+v: %v", m, err)
		}
	}
	if err := r.Send(ctx, "a", "1.2.3.4", Message{Platform: FCM, Token: "tok", Payload: make([]byte, 3001)}); !errors.Is(err, ErrTooLarge) {
		t.Fatal(err)
	}
}

func TestSendUnregisteredDropsBinding(t *testing.T) {
	ctx := context.Background()
	s, b := &fakeSender{err: ErrUnregistered}, memBindings{}
	r := newRelay(s, b, 10)
	if err := r.Send(ctx, "a", "ip", Message{Platform: FCM, Token: "tok", Payload: []byte{1}}); !errors.Is(err, ErrUnregistered) || len(b) != 0 {
		t.Fatal(err, len(b))
	}
	s.err = errors.New("timeout")
	if err := r.Send(ctx, "a", "ip", Message{Platform: FCM, Token: "tok", Payload: []byte{1}}); !errors.Is(err, ErrUpstream) {
		t.Fatal(err)
	}
}

func TestSendRateLimited(t *testing.T) {
	r := newRelay(&fakeSender{}, memBindings{}, 2)
	msg := Message{Platform: FCM, Token: "tok", Payload: []byte{1}}
	for i := range 3 {
		err := r.Send(context.Background(), "a", "ip", msg)
		if (i < 2) != (err == nil) || (i == 2 && !errors.Is(err, ErrRateLimited)) {
			t.Fatal(i, err)
		}
	}
}

func TestWindow(t *testing.T) {
	w := &Window{Max: 1, Period: time.Hour}
	t0 := time.Unix(0, 0)
	if !w.Allow("k", t0) || w.Allow("k", t0.Add(time.Minute)) || !w.Allow("other", t0) {
		t.Fatal("within the window")
	}
	if !w.Allow("k", t0.Add(time.Hour)) {
		t.Fatal("a new window starts after the period")
	}
}
