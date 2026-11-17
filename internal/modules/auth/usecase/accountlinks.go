package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"waylight/internal/modules/auth/domain"
)

// AccountLinkRepository persists invitation and reset links (#296).
type AccountLinkRepository interface {
	Save(ctx context.Context, l domain.AccountLink) error
	// FindByTokenHash returns domain.ErrInvalidAccountLink for unknown links.
	FindByTokenHash(ctx context.Context, hash domain.TokenHash) (domain.AccountLink, error)
	List(ctx context.Context, purpose domain.LinkPurpose) ([]domain.AccountLink, error)
	// Delete reports whether the link existed, so it is used only once.
	Delete(ctx context.Context, id string) (bool, error)
	// DeleteForAccount removes the reset links of an account.
	DeleteForAccount(ctx context.Context, accountID string) error
	DeleteExpired(ctx context.Context, now time.Time) error
}

// InviteInput is the input of AccountLinks.CreateInvite.
type InviteInput struct {
	Name      string
	Role      string
	Tabs      []string
	ValidDays int
}

// InviteView describes an open invitation for the admin.
type InviteView struct {
	ID        string
	Name      string
	Role      string
	Tabs      []string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// LinkInfo is what the page behind a link may show before it is used.
type LinkInfo struct {
	Purpose domain.LinkPurpose
	Name    string
	// SSO: the invitation can be accepted through the provider (#528); set
	// by the controller, which knows single sign-on.
	SSO bool
}

// AcceptLinkRequest is the input of AccountLinks.Accept.
type AcceptLinkRequest struct {
	Token     string
	Password  string
	ClientKey string
	UserAgent string
}

// AccountLinks are invitation links and admin-triggered password reset
// links (input port, #296). Managing links is admin only; Inspect and
// Accept are public, because the link is the credential, and rate-limited
// like the login.
type AccountLinks interface {
	Invites(ctx context.Context) ([]InviteView, error)
	// CreateInvite returns the invitation and its token, shown once.
	CreateInvite(ctx context.Context, in InviteInput) (InviteView, string, error)
	RevokeInvite(ctx context.Context, id string) error
	// CreateResetLink replaces earlier reset links of the account and
	// returns the token, shown once, and when it expires.
	CreateResetLink(ctx context.Context, accountID string) (string, time.Time, error)
	Inspect(ctx context.Context, token, clientKey string) (LinkInfo, error)
	// Accept creates the invited account or sets the new password and
	// consumes the link. It does not sign in: the person signs in with the
	// new password, which also asks for a second factor if one is set up.
	Accept(ctx context.Context, req AcceptLinkRequest) (LinkInfo, error)
	// InviteSignUp creates invited accounts through the provider (#528).
	InviteSignUp
}

type accountLinks struct {
	deps   Deps
	repo   AccountLinkRepository
	guests *guests // shares password hashing
}

// NewAccountLinks returns the account link use cases.
func NewAccountLinks(deps Deps, repo AccountLinkRepository) AccountLinks {
	return &accountLinks{deps: deps, repo: repo, guests: &guests{deps: deps}}
}

func inviteView(l domain.AccountLink) InviteView {
	return InviteView{ID: l.ID, Name: l.Name, Role: string(l.Role), Tabs: append([]string{}, l.Tabs...), CreatedAt: l.CreatedAt, ExpiresAt: l.ExpiresAt}
}

func (uc *accountLinks) newToken() (string, domain.TokenHash, error) {
	raw, err := uc.deps.Tokens.NewToken()
	if err != nil {
		return "", "", fmt.Errorf("new token: %w", err)
	}
	return raw, domain.HashToken(raw), nil
}

// Invites implements AccountLinks; expired invitations are dropped first.
func (uc *accountLinks) Invites(ctx context.Context) ([]InviteView, error) {
	if err := adminOnly(ctx); err != nil {
		return nil, err
	}
	if err := uc.repo.DeleteExpired(ctx, uc.deps.Clock.Now()); err != nil {
		return nil, err
	}
	list, err := uc.repo.List(ctx, domain.LinkInvite)
	if err != nil {
		return nil, err
	}
	out := make([]InviteView, 0, len(list))
	for _, l := range list {
		out = append(out, inviteView(l))
	}
	return out, nil
}

// CreateInvite implements AccountLinks.
func (uc *accountLinks) CreateInvite(ctx context.Context, in InviteInput) (InviteView, string, error) {
	open, err := uc.Invites(ctx)
	if err != nil {
		return InviteView{}, "", err
	}
	if len(open) >= domain.MaxOpenInvites {
		return InviteView{}, "", domain.ErrTooManyInvites
	}
	role := domain.Role(in.Role)
	if role == domain.RoleMember && !uc.deps.MembersEnabled {
		return InviteView{}, "", domain.ErrMembersUnavailable
	}
	raw, hash, err := uc.newToken()
	if err != nil {
		return InviteView{}, "", err
	}
	l, err := domain.NewInvite(uc.deps.IDs.NewID(), hash, in.Name, role, in.Tabs, in.ValidDays, uc.deps.Clock.Now())
	if err != nil {
		return InviteView{}, "", err
	}
	if err := uc.nameFree(ctx, l.Name, open); err != nil {
		return InviteView{}, "", err
	}
	if err := uc.repo.Save(ctx, l); err != nil {
		return InviteView{}, "", err
	}
	uc.deps.Events.record(ctx, domain.EventInviteCreated, "", domain.Client{}, fmt.Sprintf("%s %s", l.Role, l.Name))
	return inviteView(l), raw, nil
}

// nameFree checks that no account and no other open invitation uses name.
func (uc *accountLinks) nameFree(ctx context.Context, name string, open []InviteView) error {
	for _, i := range open {
		if i.Name == name {
			return domain.ErrNameTaken
		}
	}
	_, err := uc.deps.Accounts.FindByName(ctx, name)
	switch {
	case err == nil:
		return domain.ErrNameTaken
	case errors.Is(err, domain.ErrGuestNotFound), errors.Is(err, domain.ErrAccountNotFound):
		return nil
	default:
		return err
	}
}

// RevokeInvite implements AccountLinks.
func (uc *accountLinks) RevokeInvite(ctx context.Context, id string) error {
	if err := adminOnly(ctx); err != nil {
		return err
	}
	if _, err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	uc.deps.Events.record(ctx, domain.EventInviteRevoked, "", domain.Client{}, "")
	return nil
}

// CreateResetLink implements AccountLinks.
func (uc *accountLinks) CreateResetLink(ctx context.Context, accountID string) (string, time.Time, error) {
	if err := adminOnly(ctx); err != nil {
		return "", time.Time{}, err
	}
	a, err := uc.deps.Accounts.Get(ctx, accountID)
	if err != nil {
		return "", time.Time{}, err
	}
	raw, hash, err := uc.newToken()
	if err != nil {
		return "", time.Time{}, err
	}
	l, err := domain.NewResetLink(uc.deps.IDs.NewID(), hash, a, uc.deps.Clock.Now(), domain.ResetLinkTTL)
	if err != nil {
		return "", time.Time{}, err
	}
	// Only the newest reset link of an account works.
	if err := uc.repo.DeleteForAccount(ctx, a.ID()); err != nil {
		return "", time.Time{}, err
	}
	if err := uc.repo.Save(ctx, l); err != nil {
		return "", time.Time{}, err
	}
	uc.deps.Events.record(ctx, domain.EventResetLinkCreated, a.ID(), domain.Client{}, a.Name())
	return raw, l.ExpiresAt, nil
}

// find resolves a token to a usable link, counting failures against the
// client like failed logins.
func (uc *accountLinks) find(ctx context.Context, token, clientKey string) (domain.AccountLink, error) {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(clientKey, now) {
		return domain.AccountLink{}, domain.ErrTooManyAttempts
	}
	l, err := uc.repo.FindByTokenHash(ctx, domain.HashToken(token))
	if err == nil && l.Expired(now) {
		err = domain.ErrInvalidAccountLink
	}
	if errors.Is(err, domain.ErrInvalidAccountLink) {
		uc.deps.Limiter.RecordFailure(clientKey, now)
	}
	return l, err
}

// Inspect implements AccountLinks.
func (uc *accountLinks) Inspect(ctx context.Context, token, clientKey string) (LinkInfo, error) {
	l, err := uc.find(ctx, token, clientKey)
	if err != nil {
		return LinkInfo{}, err
	}
	return LinkInfo{Purpose: l.Purpose, Name: l.Name}, nil
}

// Accept implements AccountLinks.
func (uc *accountLinks) Accept(ctx context.Context, req AcceptLinkRequest) (LinkInfo, error) {
	l, err := uc.find(ctx, req.Token, req.ClientKey)
	if err != nil {
		return LinkInfo{}, err
	}
	// Validate the password before the link is consumed, so a weak
	// password does not burn the link.
	hash, err := uc.guests.hash(ctx, req.Password)
	if err != nil {
		return LinkInfo{}, err
	}
	client := domain.NewClient(req.UserAgent, req.ClientKey)
	switch l.Purpose {
	case domain.LinkInvite:
		err = uc.acceptInvite(ctx, l, hash, client)
	case domain.LinkReset:
		err = uc.acceptReset(ctx, l, hash, client)
	default:
		err = domain.ErrInvalidAccountLink
	}
	if err != nil {
		return LinkInfo{}, err
	}
	uc.deps.Limiter.Reset(req.ClientKey)
	return LinkInfo{Purpose: l.Purpose, Name: l.Name}, nil
}

// consume deletes the link; false means another request used it first.
func (uc *accountLinks) consume(ctx context.Context, l domain.AccountLink) error {
	ok, err := uc.repo.Delete(ctx, l.ID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrInvalidAccountLink
	}
	return nil
}

func (uc *accountLinks) acceptInvite(ctx context.Context, l domain.AccountLink, hash string, client domain.Client) error {
	a, err := uc.inviteAccount(ctx, l, hash)
	if err != nil {
		return err
	}
	uc.deps.Events.record(ctx, domain.EventInviteAccepted, a.ID(), client, fmt.Sprintf("%s %s", a.Role(), a.Name()))
	return nil
}

// inviteAccount creates the invited account with the role, name and tabs of
// the invitation and consumes it; shared by the password and the provider
// way in (#528).
func (uc *accountLinks) inviteAccount(ctx context.Context, l domain.AccountLink, hash string) (domain.Account, error) {
	now := uc.deps.Clock.Now()
	var a domain.Account
	var err error
	switch l.Role {
	case domain.RoleMember:
		if !uc.deps.MembersEnabled {
			return domain.Account{}, domain.ErrMembersUnavailable
		}
		a, err = domain.NewMember(uc.deps.IDs.NewID(), l.Name, hash, now)
	case domain.RoleGuest:
		a, err = domain.NewGuest(uc.deps.IDs.NewID(), l.Name, hash, l.Tabs, now)
	default:
		return domain.Account{}, domain.ErrInvalidAccountLink
	}
	if err != nil {
		return domain.Account{}, err
	}
	if err := uc.withinLimits(ctx, l.Role); err != nil {
		return domain.Account{}, err
	}
	if err := uc.consume(ctx, l); err != nil {
		return domain.Account{}, err
	}
	if err := uc.deps.Accounts.Create(ctx, a); err != nil {
		return domain.Account{}, err
	}
	return a, nil
}

// withinLimits applies the same account limits as creating by hand.
func (uc *accountLinks) withinLimits(ctx context.Context, role domain.Role) error {
	all, err := uc.deps.Accounts.List(ctx)
	if err != nil {
		return fmt.Errorf("list accounts: %w", err)
	}
	n := 0
	for _, a := range all {
		if a.Role() == role {
			n++
		}
	}
	switch {
	case role == domain.RoleGuest && n >= domain.MaxGuests:
		return domain.ErrTooManyGuests
	case role == domain.RoleMember && n >= domain.MaxMembers:
		return domain.ErrTooManyMembers
	}
	return nil
}

func (uc *accountLinks) acceptReset(ctx context.Context, l domain.AccountLink, hash string, client domain.Client) error {
	a, err := uc.deps.Accounts.Get(ctx, l.AccountID)
	if errors.Is(err, domain.ErrGuestNotFound) || errors.Is(err, domain.ErrAccountNotFound) {
		_, _ = uc.repo.Delete(ctx, l.ID)
		return domain.ErrInvalidAccountLink // the account was deleted meanwhile
	}
	if err != nil {
		return err
	}
	if a.Role() != domain.RoleMember && a.Role() != domain.RoleGuest {
		return domain.ErrInvalidAccountLink
	}
	if err := uc.consume(ctx, l); err != nil {
		return err
	}
	if err := a.ChangePasswordHash(hash, uc.deps.Clock.Now()); err != nil {
		return err
	}
	if err := uc.deps.Accounts.Update(ctx, a); err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	uc.deps.passwordSet(ctx, a.ID())
	// Whoever knew the old password is signed out everywhere.
	if err := uc.deps.Sessions.DeleteAllFor(ctx, a.ID()); err != nil {
		return fmt.Errorf("end sessions: %w", err)
	}
	uc.deps.Events.record(ctx, domain.EventPasswordReset, a.ID(), client, "reset link")
	return nil
}
