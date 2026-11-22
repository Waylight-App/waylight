package usecase

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/setupsignin/domain"
	"waylight/internal/shared/kernel"
)

func newPastes() (Pastes, *clock) {
	c := &clock{now: time.Unix(1_700_000_000, 0)}
	n := 0
	return NewPastes(PasteDeps{Clock: c, Token: func() string {
		n++
		return fmt.Sprintf("code-%d", n)
	}}), c
}

// TestPasteFlow: the phone sends the key, the assistant takes it once, and
// only the account that created the code may do either.
func TestPasteFlow(t *testing.T) {
	uc, _ := newPastes()
	anna, ben := as("anna", "admin"), as("ben", "member")
	code, _, err := uc.Create(anna)
	if err != nil {
		t.Fatal(err)
	}
	if _, ready, err := uc.Take(anna, code); err != nil || ready {
		t.Fatalf("before the phone: ready=%v err=%v", ready, err)
	}
	if err := uc.Submit(ben, code, "secret"); !errors.Is(err, domain.ErrPasteUnknown) {
		t.Fatalf("another account must not submit: %v", err)
	}
	if err := uc.Submit(anna, code, "  secret\n"); err != nil {
		t.Fatal(err)
	}
	if err := uc.Submit(anna, code, "other"); !errors.Is(err, domain.ErrPasteUnknown) {
		t.Fatalf("a code takes one value: %v", err)
	}
	if _, _, err := uc.Take(ben, code); !errors.Is(err, domain.ErrPasteUnknown) {
		t.Fatalf("another account must not read: %v", err)
	}
	if v, ready, err := uc.Take(anna, code); err != nil || !ready || v != "secret" {
		t.Fatalf("take: %q %v %v", v, ready, err)
	}
	if _, _, err := uc.Take(anna, code); !errors.Is(err, domain.ErrPasteUnknown) {
		t.Fatalf("the value is read once: %v", err)
	}
}

func TestPasteExpiresAndReplaces(t *testing.T) {
	uc, c := newPastes()
	anna := as("anna", "admin")
	first, _, _ := uc.Create(anna)
	second, _, _ := uc.Create(anna)
	if err := uc.Submit(anna, first, "x"); !errors.Is(err, domain.ErrPasteUnknown) {
		t.Fatalf("a new code replaces the old one: %v", err)
	}
	c.now = c.now.Add(domain.PasteTTL)
	if _, _, err := uc.Take(anna, second); !errors.Is(err, domain.ErrPasteUnknown) {
		t.Fatalf("an expired code is gone: %v", err)
	}
}

func TestPasteCancelAndValue(t *testing.T) {
	uc, _ := newPastes()
	anna := as("anna", "admin")
	code, _, _ := uc.Create(anna)
	for _, bad := range []string{"", "   ", strings.Repeat("x", domain.MaxPasteValue+1)} {
		if err := uc.Submit(anna, code, bad); err == nil {
			t.Errorf("accepted %d characters", len(bad))
		}
	}
	if err := uc.Cancel(as("ben", "member"), code); err != nil {
		t.Fatal(err)
	}
	if _, _, err := uc.Take(anna, code); err != nil {
		t.Fatalf("another account cancelled the code: %v", err)
	}
	_ = uc.Cancel(anna, code)
	if _, _, err := uc.Take(anna, code); !errors.Is(err, domain.ErrPasteUnknown) {
		t.Fatalf("cancelled: %v", err)
	}
}

func TestPasteLimits(t *testing.T) {
	uc, _ := newPastes()
	for i := 0; i < domain.MaxPastes; i++ {
		if _, _, err := uc.Create(as(fmt.Sprintf("a%d", i), "member")); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := uc.Create(as("late", "member")); !errors.Is(err, kernel.ErrTooManyRequests) {
		t.Fatalf("too many codes: %v", err)
	}
	for _, role := range []string{"guest", "kiosk", "token"} {
		if _, _, err := uc.Create(as("x", role)); !errors.Is(err, kernel.ErrForbidden) {
			t.Errorf("%s: %v", role, err)
		}
	}
}

// TestPasteFillFromBrowser: a phone without the app fills a code once, and
// wrong codes lock that client out without affecting others.
func TestPasteFillFromBrowser(t *testing.T) {
	uc, c := newPastes()
	code, _, _ := uc.Create(as("anna", "admin"))
	if err := uc.Check("1.1.1.1", code); err != nil {
		t.Fatal(err)
	}
	if err := uc.Fill("1.1.1.1", code, " "); !errors.Is(err, domain.ErrPasteValue) {
		t.Fatalf("empty value: %v", err)
	}
	if err := uc.Fill("1.1.1.1", code, "secret"); err != nil {
		t.Fatal(err)
	}
	if err := uc.Fill("1.1.1.1", code, "again"); !errors.Is(err, domain.ErrPasteUnknown) {
		t.Fatalf("single use: %v", err)
	}
	if v, ready, _ := uc.Take(as("anna", "admin"), code); !ready || v != "secret" {
		t.Fatalf("take: %q %v", v, ready)
	}

	code, _, _ = uc.Create(as("anna", "admin"))
	for i := 0; i < missLimit; i++ {
		if err := uc.Check("2.2.2.2", "wrong"); !errors.Is(err, domain.ErrPasteUnknown) {
			t.Fatalf("guess %d: %v", i, err)
		}
	}
	if err := uc.Fill("2.2.2.2", code, "x"); !errors.Is(err, kernel.ErrTooManyRequests) {
		t.Fatalf("locked client with the right code: %v", err)
	}
	if err := uc.Check("3.3.3.3", code); err != nil {
		t.Fatalf("other client: %v", err)
	}
	c.now = c.now.Add(missWindow)
	code, _, _ = uc.Create(as("anna", "admin"))
	if err := uc.Check("2.2.2.2", code); err != nil {
		t.Fatalf("after the window: %v", err)
	}
}
