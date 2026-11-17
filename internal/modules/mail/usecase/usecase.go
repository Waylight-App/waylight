// Package usecase holds the mail widget use cases (#283).
//
// Ring: Use Cases.
package usecase

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"waylight/internal/modules/mail/domain"
	"waylight/internal/shared/kernel"
)

// cacheTTL: a widget asks every few minutes; this keeps logins rare.
const cacheTTL = 2 * time.Minute

// AccountRepository stores the accounts with encrypted passwords (output port).
type AccountRepository interface {
	List(ctx context.Context) ([]domain.Account, error)
	Get(ctx context.Context, id string) (domain.Account, error)
	Save(ctx context.Context, a domain.Account) error
	Delete(ctx context.Context, id string) error
}

// Reader reads an inbox read-only (output port); IMAP is the adapter.
type Reader interface {
	// Test logs in and opens the folder.
	Test(ctx context.Context, a domain.Account) error
	// Unread returns the unread count and the newest unread headers.
	Unread(ctx context.Context, a domain.Account, newest int) (domain.Inbox, error)
}

// AccountInput is an account to add or change; an empty password keeps the
// stored one.
type AccountInput struct {
	ID, Name, Host     string
	Port               int
	Username, Password string
	Folder, WebmailURL string
}

// AccountView never has the password.
type AccountView struct {
	ID, Name, Host string
	Port           int
	Username       string
	Folder         string
	WebmailURL     string
}

// MessageView is one unread message of the widget.
type MessageView struct {
	Account string
	From    string
	Subject string
	Date    time.Time
}

// InboxView is the widget's answer: the sum over the chosen accounts.
type InboxView struct {
	Unread   int
	Messages []MessageView
	// WebmailURL is the address to open (the single account's, if set).
	WebmailURL string
	// Failed names accounts that could not be read.
	Failed []string
}

// Mail is the input port.
type Mail interface {
	Accounts(ctx context.Context) ([]AccountView, error)
	SaveAccount(ctx context.Context, in AccountInput) (AccountView, error)
	DeleteAccount(ctx context.Context, id string) error
	// Inbox of one account, or of all accounts when id is "".
	Inbox(ctx context.Context, id string) (InboxView, error)
}

type cached struct {
	at    time.Time
	inbox domain.Inbox
}

type mail struct {
	repo   AccountRepository
	reader Reader
	ids    kernel.IDGenerator
	clock  kernel.Clock
	mu     sync.Mutex
	cache  map[string]cached
}

// New returns the use cases.
func New(repo AccountRepository, reader Reader, ids kernel.IDGenerator, clock kernel.Clock) Mail {
	return &mail{repo: repo, reader: reader, ids: ids, clock: clock, cache: map[string]cached{}}
}

func view(a domain.Account) AccountView {
	return AccountView{ID: a.ID, Name: a.Name, Host: a.Host, Port: a.Port, Username: a.Username, Folder: a.Folder, WebmailURL: a.WebmailURL}
}

func unavailable(err error) error {
	var pe *kernel.PublicError
	if errors.As(err, &pe) {
		return err
	}
	return kernel.TargetFailure(kernel.ErrUnavailable, err)
}

func (uc *mail) Accounts(ctx context.Context) ([]AccountView, error) {
	all, err := uc.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	out := make([]AccountView, 0, len(all))
	for _, a := range all {
		out = append(out, view(a))
	}
	return out, nil
}

func (uc *mail) SaveAccount(ctx context.Context, in AccountInput) (AccountView, error) {
	id, pw := in.ID, in.Password
	if id == "" {
		all, err := uc.repo.List(ctx)
		if err != nil {
			return AccountView{}, fmt.Errorf("list accounts: %w", err)
		}
		if len(all) >= domain.MaxAccounts {
			return AccountView{}, domain.ErrTooManyAccounts
		}
		id = uc.ids.NewID()
	} else {
		old, err := uc.repo.Get(ctx, id)
		if err != nil {
			return AccountView{}, err
		}
		if pw == "" {
			pw = old.Password
		}
	}
	a, err := domain.NewAccount(id, in.Name, in.Host, in.Port, in.Username, pw, in.Folder, in.WebmailURL)
	if err != nil {
		return AccountView{}, err
	}
	if err := uc.reader.Test(ctx, a); err != nil {
		return AccountView{}, unavailable(err)
	}
	if err := uc.repo.Save(ctx, a); err != nil {
		return AccountView{}, fmt.Errorf("save account: %w", err)
	}
	uc.mu.Lock()
	delete(uc.cache, a.ID)
	uc.mu.Unlock()
	return view(a), nil
}

func (uc *mail) DeleteAccount(ctx context.Context, id string) error {
	if _, err := uc.repo.Get(ctx, id); err != nil {
		return err
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *mail) inbox(ctx context.Context, a domain.Account) (domain.Inbox, error) {
	now := uc.clock.Now()
	uc.mu.Lock()
	c, ok := uc.cache[a.ID]
	uc.mu.Unlock()
	if ok && now.Sub(c.at) < cacheTTL {
		return c.inbox, nil
	}
	in, err := uc.reader.Unread(ctx, a, domain.Newest)
	if err != nil {
		if ok {
			return c.inbox, nil // last known state
		}
		return domain.Inbox{}, err
	}
	uc.mu.Lock()
	uc.cache[a.ID] = cached{at: now, inbox: in}
	uc.mu.Unlock()
	return in, nil
}

func (uc *mail) Inbox(ctx context.Context, id string) (InboxView, error) {
	var accounts []domain.Account
	if id != "" {
		a, err := uc.repo.Get(ctx, id)
		if err != nil {
			return InboxView{}, err
		}
		accounts = []domain.Account{a}
	} else {
		all, err := uc.repo.List(ctx)
		if err != nil {
			return InboxView{}, fmt.Errorf("list accounts: %w", err)
		}
		accounts = all
	}
	if len(accounts) == 0 {
		return InboxView{}, kernel.NewPublicError(kernel.ErrNotFound, "No mail account is set up yet.")
	}
	results := make([]domain.Inbox, len(accounts))
	errs := make([]error, len(accounts))
	var wg sync.WaitGroup
	for i, a := range accounts {
		wg.Go(func() { results[i], errs[i] = uc.inbox(ctx, a) })
	}
	wg.Wait()
	out := InboxView{Messages: []MessageView{}}
	if len(accounts) == 1 {
		out.WebmailURL = accounts[0].WebmailURL
	}
	for i, a := range accounts {
		if errs[i] != nil {
			out.Failed = append(out.Failed, a.Name)
			continue
		}
		out.Unread += results[i].Unread
		for _, m := range results[i].Messages {
			out.Messages = append(out.Messages, MessageView{Account: a.Name, From: domain.SenderName(m.From), Subject: m.Subject, Date: m.Date})
		}
	}
	if len(out.Failed) == len(accounts) {
		return InboxView{}, unavailable(errs[0])
	}
	slices.SortStableFunc(out.Messages, func(a, b MessageView) int { return cmp.Compare(b.Date.Unix(), a.Date.Unix()) })
	if len(out.Messages) > domain.Newest {
		out.Messages = out.Messages[:domain.Newest]
	}
	return out, nil
}
