package usecase

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

type memEmails map[string]string

func (m memEmails) Get(_ context.Context, id string) (string, error) { return m[id], nil }
func (m memEmails) Set(_ context.Context, id, e string) error {
	if e == "" {
		delete(m, id)
	} else {
		m[id] = e
	}
	return nil
}

type fakeMailer struct {
	ready bool
	sent  []kernel.Mail
}

func (f *fakeMailer) CanSend(context.Context) bool { return f.ready }
func (f *fakeMailer) Send(_ context.Context, m kernel.Mail) error {
	f.sent = append(f.sent, m)
	return nil
}

func newReset(t *testing.T, e *env, mailer *fakeMailer, links *memLinks) (*passwordReset, memEmails) {
	t.Helper()
	emails := memEmails{}
	pr := NewPasswordReset(e.deps, links, emails, mailer, &url.URL{Scheme: "https", Host: "home.example.org"}).(*passwordReset)
	pr.send = func(f func()) { f() }
	return pr, emails
}

func TestResetByEmail(t *testing.T) {
	e := newEnv().withAccount()
	e.limiter.max = 100
	mailer := &fakeMailer{ready: true}
	links := &memLinks{m: map[string]domain.AccountLink{}}
	pr, _ := newReset(t, e, mailer, links)
	g, err := NewGuests(e.deps).Create(ctx, NewGuestRequest{Name: "oma", Password: goodPassword, Tabs: []string{"home"}})
	if err != nil {
		t.Fatal(err)
	}
	oma := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: g.ID, Role: "guest"})
	if _, err := pr.SetMyEmail(oma, "Oma <oma@example.org>"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("only a bare address", err)
	}
	if _, err := pr.SetMyEmail(oma, " oma@example.org "); err != nil {
		t.Fatal(err)
	}
	if got, _ := pr.MyEmail(oma); got != "oma@example.org" {
		t.Fatal(got)
	}
	owner := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: domain.OwnerID, Role: "owner"})
	if _, err := pr.SetMyEmail(owner, "admin@example.org"); !errors.Is(err, domain.ErrAdminResetByShell) {
		t.Fatal("the admin resets on the server", err)
	}

	// Unknown names and accounts without address answer the same, silently.
	if err := pr.Request(ctx, "nobody", "c1"); err != nil || len(mailer.sent) != 0 {
		t.Fatal(err, mailer.sent)
	}
	if err := pr.Request(ctx, "owner", "c1"); err != nil || len(mailer.sent) != 0 {
		t.Fatal("no mail for the admin", err)
	}

	if err := pr.Request(ctx, "oma", "c1"); err != nil {
		t.Fatal(err)
	}
	if len(mailer.sent) != 1 || mailer.sent[0].To != "oma@example.org" ||
		!strings.Contains(mailer.sent[0].Body, "https://home.example.org/account-link#") {
		t.Fatalf("mail %+v", mailer.sent)
	}
	token := mailer.sent[0].Body[strings.Index(mailer.sent[0].Body, "#")+1:]
	token = token[:strings.Index(token, "\n")]
	// A second request right away sends nothing new.
	if err := pr.Request(ctx, "oma", "c2"); err != nil || len(mailer.sent) != 1 {
		t.Fatal("a mailbox must not be flooded", err)
	}

	accept := NewAccountLinks(e.deps, links)
	if _, err := accept.Accept(ctx, AcceptLinkRequest{Token: token, Password: "another-good-password", ClientKey: "c3"}); err != nil {
		t.Fatal(err)
	}
	// The e-mail link is valid for an hour only.
	for _, l := range links.m {
		t.Fatalf("link not consumed: %+v", l)
	}
}

func TestResetByEmailNeedsMailAndPublicURL(t *testing.T) {
	e := newEnv().withAccount()
	links := &memLinks{m: map[string]domain.AccountLink{}}
	pr, _ := newReset(t, e, &fakeMailer{ready: false}, links)
	if pr.Available(ctx) {
		t.Fatal("no e-mail channel")
	}
	if err := pr.Request(ctx, "oma", "c"); !errors.Is(err, domain.ErrResetUnavailable) {
		t.Fatal(err)
	}
	noURL := NewPasswordReset(e.deps, links, memEmails{}, &fakeMailer{ready: true}, nil)
	if noURL.Available(ctx) {
		t.Fatal("without PUBLIC_URL there is no safe link to send")
	}
}

func TestResetRequestsAreThrottled(t *testing.T) {
	e := newEnv().withAccount()
	pr, _ := newReset(t, e, &fakeMailer{ready: true}, &memLinks{m: map[string]domain.AccountLink{}})
	for range 3 {
		_ = pr.Request(ctx, "x", "c")
	}
	if err := pr.Request(ctx, "x", "c"); !errors.Is(err, domain.ErrTooManyAttempts) {
		t.Fatal(err)
	}
}
