// Package usecase holds the birthday use cases (#287).
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"waylight/internal/modules/birthdays/domain"
	"waylight/internal/shared/kernel"
)

// Settings are the reminder and the optional CardDAV address book; the
// password is a secret.
type Settings struct {
	Reminder     domain.Reminder
	LastReminded string // date (YYYY-MM-DD) of the last reminder
	CardDAVURL   string
	Username     string
	Password     string
}

// Repository stores birthdays and settings (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Birthday, error)
	Save(ctx context.Context, b domain.Birthday) error
	Delete(ctx context.Context, id string) error
	// ReplaceSource replaces all entries of a source (CardDAV sync).
	ReplaceSource(ctx context.Context, source string, list []domain.Birthday) error
	Settings(ctx context.Context) (Settings, error)
	SaveSettings(ctx context.Context, s Settings) error
	// SyncAccounts lists the accounts that have an address book.
	SyncAccounts(ctx context.Context) ([]string, error)
}

// AddressBook reads the vCards of a CardDAV address book (output port).
type AddressBook interface {
	Cards(ctx context.Context, url, username, password string) ([]domain.Card, error)
}

// SettingsView never has the password.
type SettingsView struct {
	Reminder    domain.Reminder
	CardDAVURL  string
	Username    string
	HasPassword bool
}

// SettingsInput changes the settings; an empty password keeps the stored
// one, an empty URL disconnects the address book.
type SettingsInput struct {
	RemindEnabled bool
	RemindAt      string
	CardDAVURL    string
	Username      string
	Password      string
}

// ImportReport says what a vCard import did.
type ImportReport struct{ Added, Duplicates int }

// Birthdays is the input port.
type Birthdays interface {
	Upcoming(ctx context.Context, limit int) ([]domain.Upcoming, error)
	All(ctx context.Context) ([]domain.Upcoming, error)
	Add(ctx context.Context, name string, month, day, year int) (domain.Birthday, error)
	Delete(ctx context.Context, id string) error
	ImportVCard(ctx context.Context, data []byte) (ImportReport, error)
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	// Sync reads the CardDAV address book (job).
	Sync(ctx context.Context) error
	// Remind pushes today's birthdays once a day (job).
	Remind(ctx context.Context) (int, error)
	// Today lists the birthdays of today and tomorrow (for the briefing).
	Soon(ctx context.Context, days int) ([]domain.Upcoming, error)
}

// Deps bundles the output ports.
type Deps struct {
	Repo     Repository
	Book     AddressBook
	IDs      kernel.IDGenerator
	Clock    kernel.Clock
	Location *time.Location
	Alerts   kernel.AlertSink // nil = no reminders
}

type birthdays struct{ d Deps }

// New returns the use cases.
func New(d Deps) Birthdays {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &birthdays{d: d}
}

// MaxImportBytes limits a vCard file.
const MaxImportBytes = 5 << 20

func (uc *birthdays) All(ctx context.Context) ([]domain.Upcoming, error) {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list birthdays: %w", err)
	}
	return domain.Next(list, uc.d.Clock.Now(), uc.d.Location), nil
}

func (uc *birthdays) Upcoming(ctx context.Context, limit int) ([]domain.Upcoming, error) {
	all, err := uc.All(ctx)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (uc *birthdays) Soon(ctx context.Context, days int) ([]domain.Upcoming, error) {
	all, err := uc.All(ctx)
	if err != nil {
		return nil, err
	}
	var out []domain.Upcoming
	for _, u := range all {
		if u.InDays < days {
			out = append(out, u)
		}
	}
	return out, nil
}

func (uc *birthdays) count(ctx context.Context) (int, error) {
	list, err := uc.d.Repo.List(ctx)
	return len(list), err
}

func (uc *birthdays) Add(ctx context.Context, name string, month, day, year int) (domain.Birthday, error) {
	n, err := uc.count(ctx)
	if err != nil {
		return domain.Birthday{}, err
	}
	if n >= domain.MaxEntries {
		return domain.Birthday{}, domain.ErrTooMany
	}
	b, err := domain.NewBirthday(uc.d.IDs.NewID(), name, month, day, year, domain.SourceManual)
	if err != nil {
		return domain.Birthday{}, err
	}
	return b, uc.d.Repo.Save(ctx, b)
}

// errSyncedEntry: the next sync would bring a deleted address book entry
// back, so it has to be removed at its source.
var errSyncedEntry = kernel.NewValidationError("id", "This birthday comes from the connected address book; change or remove it there.")

func (uc *birthdays) Delete(ctx context.Context, id string) error {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list birthdays: %w", err)
	}
	for _, b := range list {
		if b.ID == id && b.Source == domain.SourceCardDAV {
			return errSyncedEntry
		}
	}
	return uc.d.Repo.Delete(ctx, id)
}

func key(name string, m time.Month, d int) string {
	return strings.ToLower(name) + "|" + fmt.Sprint(int(m), "-", d)
}

func (uc *birthdays) ImportVCard(ctx context.Context, data []byte) (ImportReport, error) {
	if len(data) > MaxImportBytes {
		return ImportReport{}, kernel.NewValidationError("file", "The vCard file must be at most 5 MB.")
	}
	cards := domain.ParseVCards(string(data))
	if len(cards) == 0 {
		return ImportReport{}, kernel.NewValidationError("file", "The file contains no contacts with a birthday (vCard, .vcf).")
	}
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return ImportReport{}, err
	}
	known := map[string]bool{}
	for _, b := range list {
		known[key(b.Name, b.Month, b.Day)] = true
	}
	var r ImportReport
	for _, c := range cards {
		b, err := domain.NewBirthday(uc.d.IDs.NewID(), c.Name, c.Month, c.Day, c.Year, domain.SourceVCard)
		if err != nil || known[key(b.Name, b.Month, b.Day)] {
			r.Duplicates++
			continue
		}
		if len(list)+r.Added >= domain.MaxEntries {
			return r, domain.ErrTooMany
		}
		if err := uc.d.Repo.Save(ctx, b); err != nil {
			return r, err
		}
		known[key(b.Name, b.Month, b.Day)] = true
		r.Added++
	}
	return r, nil
}

func (uc *birthdays) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.d.Repo.Settings(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	return SettingsView{Reminder: s.Reminder, CardDAVURL: s.CardDAVURL, Username: s.Username, HasPassword: s.Password != ""}, nil
}

func (uc *birthdays) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	old, err := uc.d.Repo.Settings(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	// The reminder goes to the instance's notification channels; until
	// there are channels per person only an admin switches it on (ADR 0018).
	if in.RemindEnabled && !kernel.PrincipalFrom(ctx).IsAdmin() {
		return SettingsView{}, kernel.ErrAdminOnly
	}
	r, err := domain.NewReminder(in.RemindEnabled, in.RemindAt)
	if err != nil {
		return SettingsView{}, err
	}
	s := Settings{Reminder: r, LastReminded: old.LastReminded}
	if strings.TrimSpace(in.CardDAVURL) != "" {
		u, err := kernel.ParseTarget("cardDavUrl", in.CardDAVURL, kernel.WithTargetPath())
		if err != nil {
			return SettingsView{}, err
		}
		pw := in.Password
		if pw == "" {
			pw = old.Password
		}
		if strings.TrimSpace(in.Username) == "" || pw == "" {
			return SettingsView{}, kernel.NewValidationError("username", "User name and app password are required.")
		}
		s.CardDAVURL, s.Username, s.Password = u.String()+"/", strings.TrimSpace(in.Username), pw
		cards, err := uc.d.Book.Cards(ctx, s.CardDAVURL, s.Username, s.Password)
		if err != nil {
			return SettingsView{}, kernel.TargetFailure(kernel.ErrUnavailable, err)
		}
		if err := uc.replace(ctx, cards); err != nil {
			return SettingsView{}, err
		}
	} else if old.CardDAVURL != "" {
		if err := uc.d.Repo.ReplaceSource(ctx, domain.SourceCardDAV, nil); err != nil {
			return SettingsView{}, err
		}
	}
	if err := uc.d.Repo.SaveSettings(ctx, s); err != nil {
		return SettingsView{}, err
	}
	return SettingsView{Reminder: s.Reminder, CardDAVURL: s.CardDAVURL, Username: s.Username, HasPassword: s.Password != ""}, nil
}

func (uc *birthdays) replace(ctx context.Context, cards []domain.Card) error {
	var list []domain.Birthday
	// Two accounts may sync the same contact; members' IDs carry the account
	// so they never collide (the owner keeps the IDs it always had).
	prefix := "carddav-"
	if a := kernel.PrincipalFrom(ctx).SettingsAccount(); a != kernel.SystemAccountID {
		prefix += a + "-"
	}
	for _, c := range cards {
		id := prefix + c.UID
		if c.UID == "" {
			id = uc.d.IDs.NewID()
		}
		if b, err := domain.NewBirthday(id, c.Name, c.Month, c.Day, c.Year, domain.SourceCardDAV); err == nil {
			list = append(list, b)
		}
	}
	if len(list) > domain.MaxEntries {
		list = list[:domain.MaxEntries]
	}
	return uc.d.Repo.ReplaceSource(ctx, domain.SourceCardDAV, list)
}

// Sync implements Birthdays: the job synchronises every account's address
// book with that account's credentials (ADR 0018).
func (uc *birthdays) Sync(ctx context.Context) error {
	if !kernel.PrincipalFrom(ctx).IsSystem() {
		return uc.syncOne(ctx)
	}
	accounts, err := uc.d.Repo.SyncAccounts(ctx)
	if err != nil {
		return err
	}
	var errs []error
	for _, a := range accounts {
		errs = append(errs, uc.syncOne(kernel.WithPrincipal(ctx, kernel.PrincipalFor(a))))
	}
	return errors.Join(errs...)
}

func (uc *birthdays) syncOne(ctx context.Context) error {
	s, err := uc.d.Repo.Settings(ctx)
	if err != nil || s.CardDAVURL == "" {
		return err
	}
	cards, err := uc.d.Book.Cards(ctx, s.CardDAVURL, s.Username, s.Password)
	if err != nil {
		return fmt.Errorf("read address book: %w", err) // keeps the last known entries
	}
	return uc.replace(ctx, cards)
}

func (uc *birthdays) Remind(ctx context.Context) (int, error) {
	if uc.d.Alerts == nil {
		return 0, nil
	}
	// The reminder is the admin's and lists the birthdays the admin sees,
	// never members' private ones, although the job runs as the system.
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	s, err := uc.d.Repo.Settings(ctx)
	if err != nil {
		return 0, err
	}
	now := uc.d.Clock.Now()
	if !s.Reminder.Due(now, s.LastReminded, uc.d.Location) {
		return 0, nil
	}
	today, err := uc.Soon(ctx, 1)
	if err != nil {
		return 0, err
	}
	s.LastReminded = now.In(uc.d.Location).Format(time.DateOnly)
	if err := uc.d.Repo.SaveSettings(ctx, s); err != nil { // first, so a failure never repeats pushes
		return 0, err
	}
	for _, u := range today {
		body := "Birthday today."
		if u.Age > 0 {
			body = fmt.Sprintf("Turns %d today.", u.Age)
		}
		uc.d.Alerts.Alert(ctx, kernel.Alert{Kind: "reminder", Title: "🎂 " + u.Name, Body: body})
	}
	return len(today), nil
}
