package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"waylight/internal/modules/mail/domain"
	"waylight/internal/shared/kernel"
)

type repo struct{ m map[string]domain.Account }

func (r *repo) List(context.Context) ([]domain.Account, error) {
	var out []domain.Account
	for _, a := range r.m {
		out = append(out, a)
	}
	return out, nil
}
func (r *repo) Get(_ context.Context, id string) (domain.Account, error) {
	a, ok := r.m[id]
	if !ok {
		return a, domain.ErrAccountNotFound
	}
	return a, nil
}
func (r *repo) Save(_ context.Context, a domain.Account) error { r.m[a.ID] = a; return nil }
func (r *repo) Delete(_ context.Context, id string) error      { delete(r.m, id); return nil }

type reader struct {
	calls atomic.Int32
	fail  map[string]bool
}

func (f *reader) Test(_ context.Context, a domain.Account) error {
	if a.Password != "pw" {
		return errors.New("login failed")
	}
	return nil
}
func (f *reader) Unread(_ context.Context, a domain.Account, _ int) (domain.Inbox, error) {
	f.calls.Add(1)
	if f.fail[a.Name] {
		return domain.Inbox{}, errors.New("timeout")
	}
	base := time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)
	if a.Name == "Arbeit" {
		base = base.Add(time.Hour)
	}
	return domain.Inbox{Unread: 2, Messages: []domain.Message{{From: `"Anna" <a@x.de>`, Subject: a.Name + " 1", Date: base}}}, nil
}

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return fmt.Sprintf("id%d", i.n) }

func TestMail(t *testing.T) {
	ctx := context.Background()
	r, rd, c := &repo{m: map[string]domain.Account{}}, &reader{fail: map[string]bool{}}, &clk{t: time.Now()}
	uc := New(r, rd, &ids{}, c)
	if _, err := uc.Inbox(ctx, ""); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal("no account yet", err)
	}
	if _, err := uc.SaveAccount(ctx, AccountInput{Name: "Privat", Host: "imap.x.de", Username: "u", Password: "wrong"}); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatal("login is tested before saving", err)
	}
	p, _ := uc.SaveAccount(ctx, AccountInput{Name: "Privat", Host: "imap.x.de", Username: "u", Password: "pw", WebmailURL: "https://mail.x.de"})
	if _, err := uc.SaveAccount(ctx, AccountInput{ID: p.ID, Name: "Privat", Host: "imap.x.de", Username: "u", WebmailURL: "https://mail.x.de"}); err != nil {
		t.Fatal("editing keeps the password", err)
	}
	_, _ = uc.SaveAccount(ctx, AccountInput{Name: "Arbeit", Host: "imap.y.de", Username: "u", Password: "pw"})

	all, err := uc.Inbox(ctx, "")
	if err != nil || all.Unread != 4 || len(all.Messages) != 2 || all.Messages[0].Subject != "Arbeit 1" || all.Messages[0].From != "Anna" || all.WebmailURL != "" {
		t.Fatalf("combined, newest first: %+v %v", all, err)
	}
	one, _ := uc.Inbox(ctx, p.ID)
	if one.Unread != 2 || one.WebmailURL != "https://mail.x.de/" && one.WebmailURL != "https://mail.x.de" {
		t.Fatalf("%+v", one)
	}
	if rd.calls.Load() != 2 {
		t.Fatal("cached per account:", rd.calls.Load())
	}
	// One failing account is named; all failing is an error.
	c.t = c.t.Add(time.Hour)
	rd.fail["Arbeit"] = true
	r.m["id9"] = domain.Account{ID: "id9", Name: "Neu"}
	rd.fail["Neu"] = true
	if v, err := uc.Inbox(ctx, ""); err != nil || len(v.Failed) != 1 || v.Failed[0] != "Neu" {
		t.Fatalf("Arbeit keeps its last state, Neu failed: %+v %v", v, err)
	}
	if _, err := uc.Inbox(ctx, "id9"); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatal(err)
	}
	if err := uc.DeleteAccount(ctx, p.ID); err != nil || len(r.m) != 2 {
		t.Fatal(err)
	}
}
