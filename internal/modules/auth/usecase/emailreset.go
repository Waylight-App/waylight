package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

// EmailRepository stores the optional e-mail address of each account (#296).
type EmailRepository interface {
	// Get returns "" when the account has no address.
	Get(ctx context.Context, accountID string) (string, error)
	// Set stores the address; "" removes it.
	Set(ctx context.Context, accountID, email string) error
}

// PasswordReset is the self-service reset by e-mail (input port, #296).
// A member or guest stores an address; whoever knows the account name can
// ask for a link, which goes only to that address.
type PasswordReset interface {
	// Available reports whether reset mails can be sent: an e-mail
	// channel is set up and PUBLIC_URL is known.
	Available(ctx context.Context) bool
	// MyEmail and SetMyEmail are the signed-in account's own address.
	MyEmail(ctx context.Context) (string, error)
	SetMyEmail(ctx context.Context, email string) (string, error)
	// Request sends a reset link if the name belongs to a member or guest
	// with an address. It answers the same in every case, so it does not
	// reveal which accounts exist or have an address.
	Request(ctx context.Context, name, clientKey string) error
}

type passwordReset struct {
	deps   Deps
	links  AccountLinkRepository
	emails EmailRepository
	mailer kernel.Mailer
	// publicURL is the configured origin for links in mails. The Host
	// header of the request is never used: an attacker could set it and
	// receive the victim's reset link (reset poisoning).
	publicURL *url.URL
	// send runs the delivery; tests replace it to run synchronously.
	send func(func())
}

// NewPasswordReset returns the self-service reset; mailer or publicURL nil
// switches it off.
func NewPasswordReset(deps Deps, links AccountLinkRepository, emails EmailRepository, mailer kernel.Mailer, publicURL *url.URL) PasswordReset {
	if mailer == nil {
		mailer = kernel.NoMailer{}
	}
	return &passwordReset{deps: deps, links: links, emails: emails, mailer: mailer, publicURL: publicURL, send: func(f func()) { go f() }}
}

func (uc *passwordReset) Available(ctx context.Context) bool {
	return uc.publicURL != nil && uc.mailer.CanSend(ctx)
}

// self returns the signed-in account that may keep an address: members
// and guests. The admin resets on the server and kiosks have no password.
func self(ctx context.Context) (string, error) {
	p := kernel.PrincipalFrom(ctx)
	if p.AccountID == "" || p.IsSystem() || p.IsKiosk() {
		return "", domain.ErrNoPassword
	}
	switch p.Role {
	case string(domain.RoleOwner):
		return "", domain.ErrAdminResetByShell
	case string(domain.RoleAdmin):
		return "", domain.ErrAdminResetByAdmin
	}
	return p.AccountID, nil
}

func (uc *passwordReset) MyEmail(ctx context.Context) (string, error) {
	id, err := self(ctx)
	if err != nil {
		return "", err
	}
	return uc.emails.Get(ctx, id)
}

func (uc *passwordReset) SetMyEmail(ctx context.Context, raw string) (string, error) {
	id, err := self(ctx)
	if err != nil {
		return "", err
	}
	email, err := domain.NormaliseEmail(raw)
	if err != nil {
		return "", err
	}
	if err := uc.emails.Set(ctx, id, email); err != nil {
		return "", err
	}
	// The address itself is personal data and stays out of the log.
	uc.deps.Events.record(ctx, domain.EventEmailChanged, id, domain.Client{}, "")
	return email, nil
}

func (uc *passwordReset) Request(ctx context.Context, name, clientKey string) error {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(clientKey, now) {
		return domain.ErrTooManyAttempts
	}
	// Every request counts against the client, found or not, so the
	// endpoint cannot be used to send mail in bulk.
	uc.deps.Limiter.RecordFailure(clientKey, now)
	if !uc.Available(ctx) {
		return domain.ErrResetUnavailable
	}
	a, err := uc.deps.Accounts.FindByName(ctx, name)
	if errors.Is(err, domain.ErrGuestNotFound) || errors.Is(err, domain.ErrAccountNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if (a.Role() != domain.RoleMember && a.Role() != domain.RoleGuest) || a.Disabled() {
		return nil
	}
	email, err := uc.emails.Get(ctx, a.ID())
	if err != nil || email == "" {
		return err
	}
	if recent, err := uc.recentlySent(ctx, a.ID(), now); err != nil || recent {
		return err
	}
	raw, err := uc.deps.Tokens.NewToken()
	if err != nil {
		return fmt.Errorf("new token: %w", err)
	}
	l, err := domain.NewResetLink(uc.deps.IDs.NewID(), domain.HashToken(raw), a, now, domain.EmailResetLinkTTL)
	if err != nil {
		return err
	}
	if err := uc.links.DeleteForAccount(ctx, a.ID()); err != nil {
		return err
	}
	if err := uc.links.Save(ctx, l); err != nil {
		return err
	}
	uc.deps.Events.record(ctx, domain.EventResetRequested, a.ID(), domain.NewClient("", clientKey), "")
	link := uc.publicURL.JoinPath("account-link").String() + "#" + url.PathEscape(raw)
	m := kernel.Mail{To: email, Subject: "Reset your password", Body: resetMailBody(a.Name(), uc.deps.InstanceName.Of(ctx), link)}
	// Delivered in the background, so the answer time does not reveal
	// whether a mail was sent.
	detached := context.WithoutCancel(ctx)
	uc.send(func() {
		sendCtx, cancel := context.WithTimeout(detached, 30*time.Second)
		defer cancel()
		_ = uc.mailer.Send(sendCtx, m) // a failure shows in the mail server's log; the user can ask again
	})
	return nil
}

// recentlySent reports whether the account got a reset link moments ago.
func (uc *passwordReset) recentlySent(ctx context.Context, accountID string, now time.Time) (bool, error) {
	list, err := uc.links.List(ctx, domain.LinkReset)
	if err != nil {
		return false, err
	}
	for _, l := range list {
		if l.AccountID == accountID && now.Sub(l.CreatedAt) < domain.ResetRequestInterval {
			return true, nil
		}
	}
	return false, nil
}

func resetMailBody(name, instance, link string) string {
	return "Someone asked to reset the password of the account \"" + name + "\" on " + instance + ".\n\n" +
		"To choose a new password, open this link within one hour:\n\n" + link + "\n\n" +
		"If you did not ask for this, ignore this mail; your password stays as it is."
}
