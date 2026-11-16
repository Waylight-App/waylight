package usecase

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"trisse-home/internal/modules/phonelist/domain"
	"trisse-home/internal/shared/kernel"
)

// Phonelist is the input port of the module (#300). Everyone who can see
// the widget reads the list; only admins change it.
type Phonelist interface {
	List(ctx context.Context) ([]domain.Entry, error)
	// Search finds entries whose name, role or team contain every word.
	Search(ctx context.Context, query string) ([]domain.Entry, error)
	Add(ctx context.Context, in domain.EntryInput) (domain.Entry, error)
	Update(ctx context.Context, id string, in domain.EntryInput) (domain.Entry, error)
	Delete(ctx context.Context, id string) error
	// Settings and SaveSettings are the address book of the instance
	// (admins only); saving reads the book once to check it.
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	// Sync reads the address book again (job).
	Sync(ctx context.Context) error
}

// Repository persists entries (output port, implemented in adapter/gateway).
type Repository interface {
	List(ctx context.Context) ([]domain.Entry, error)
	Get(ctx context.Context, id string) (domain.Entry, error)
	Save(ctx context.Context, e domain.Entry) error
	Delete(ctx context.Context, id string) error
	// ReplaceSource swaps all entries of a source in one transaction.
	ReplaceSource(ctx context.Context, source domain.Source, entries []domain.Entry) error
	Settings(ctx context.Context) (Settings, error)
	SaveSettings(ctx context.Context, s Settings) error
}

// AddressBook reads the people of a CardDAV address book (output port).
type AddressBook interface {
	Cards(ctx context.Context, url, username, password string) ([]domain.Card, error)
}

// Settings is the stored address book configuration.
type Settings struct {
	CardDAVURL, Username, Password string
	LastSync                       time.Time
	LastError                      string
}

// SettingsInput changes the address book; an empty password keeps the
// stored one, an empty URL switches the address book off.
type SettingsInput struct {
	CardDAVURL, Username, Password string
}

// SettingsView is the address book configuration without the password.
type SettingsView struct {
	CardDAVURL, Username string
	HasPassword          bool
	LastSync             time.Time
	LastError            string
}

// Deps bundles the output ports.
type Deps struct {
	Repo  Repository
	Book  AddressBook
	Clock kernel.Clock
	IDs   kernel.IDGenerator
}

type uc struct{ d Deps }

// NewPhonelist returns the phone list use cases.
func NewPhonelist(d Deps) Phonelist { return &uc{d: d} }

// adminOnly lets admins and the system (config, import) change the list.
func adminOnly(ctx context.Context) error {
	if p := kernel.PrincipalFrom(ctx); !p.IsAdmin() && !p.IsSystem() {
		return kernel.ErrAdminOnly
	}
	return nil
}

// List implements Phonelist, sorted by name.
func (u *uc) List(ctx context.Context) ([]domain.Entry, error) {
	entries, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	slices.SortFunc(entries, func(a, b domain.Entry) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return entries, nil
}

// Search implements Phonelist. The list is small (at most MaxEntries), so
// it is filtered in memory.
func (u *uc) Search(ctx context.Context, query string) ([]domain.Entry, error) {
	words := strings.Fields(strings.ToLower(query))
	all, err := u.List(ctx)
	if err != nil || len(words) == 0 {
		return []domain.Entry{}, err
	}
	out := []domain.Entry{}
	for _, e := range all {
		text := strings.ToLower(e.Name + " " + e.Role + " " + e.Team)
		if !slices.ContainsFunc(words, func(w string) bool { return !strings.Contains(text, w) }) {
			out = append(out, e)
		}
	}
	return out, nil
}

// Add implements Phonelist.
func (u *uc) Add(ctx context.Context, in domain.EntryInput) (domain.Entry, error) {
	if err := adminOnly(ctx); err != nil {
		return domain.Entry{}, err
	}
	existing, err := u.d.Repo.List(ctx)
	if err != nil {
		return domain.Entry{}, fmt.Errorf("list entries: %w", err)
	}
	if len(existing) >= domain.MaxEntries {
		return domain.Entry{}, domain.ErrTooManyEntries
	}
	e, err := domain.NewEntry(u.d.IDs.NewID(), in, u.d.Clock.Now())
	if err != nil {
		return domain.Entry{}, err
	}
	if err := u.d.Repo.Save(ctx, e); err != nil {
		return domain.Entry{}, fmt.Errorf("save entry: %w", err)
	}
	return e, nil
}

// Update implements Phonelist.
func (u *uc) Update(ctx context.Context, id string, in domain.EntryInput) (domain.Entry, error) {
	if err := adminOnly(ctx); err != nil {
		return domain.Entry{}, err
	}
	e, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return domain.Entry{}, err
	}
	if e.Source == domain.SourceCardDAV {
		return domain.Entry{}, domain.ErrReadOnlyEntry
	}
	if err := e.Update(in); err != nil {
		return domain.Entry{}, err
	}
	if err := u.d.Repo.Save(ctx, e); err != nil {
		return domain.Entry{}, fmt.Errorf("save entry: %w", err)
	}
	return e, nil
}

// Delete implements Phonelist.
func (u *uc) Delete(ctx context.Context, id string) error {
	if err := adminOnly(ctx); err != nil {
		return err
	}
	e, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if e.Source == domain.SourceCardDAV {
		return domain.ErrReadOnlyEntry
	}
	return u.d.Repo.Delete(ctx, id)
}

func view(s Settings) SettingsView {
	return SettingsView{CardDAVURL: s.CardDAVURL, Username: s.Username, HasPassword: s.Password != "", LastSync: s.LastSync, LastError: s.LastError}
}

// Settings implements Phonelist.
func (u *uc) Settings(ctx context.Context) (SettingsView, error) {
	if err := adminOnly(ctx); err != nil {
		return SettingsView{}, err
	}
	s, err := u.d.Repo.Settings(ctx)
	return view(s), err
}

// SaveSettings implements Phonelist. The URL is checked like every address
// an admin configures (kernel.ParseTarget), and the book is read once, so a
// wrong password shows up right away instead of at the next sync.
func (u *uc) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	if err := adminOnly(ctx); err != nil {
		return SettingsView{}, err
	}
	old, err := u.d.Repo.Settings(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	if strings.TrimSpace(in.CardDAVURL) == "" {
		// Off: the people from the book leave the list with it.
		if err := u.d.Repo.ReplaceSource(ctx, domain.SourceCardDAV, nil); err != nil {
			return SettingsView{}, err
		}
		s := Settings{}
		return view(s), u.d.Repo.SaveSettings(ctx, s)
	}
	target, err := kernel.ParseTarget("cardDavUrl", in.CardDAVURL, kernel.WithTargetPath())
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
	s := Settings{CardDAVURL: strings.TrimSuffix(target.String(), "/") + "/", Username: strings.TrimSpace(in.Username), Password: pw}
	if err := u.syncWith(ctx, &s); err != nil {
		return SettingsView{}, kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	return view(s), u.d.Repo.SaveSettings(ctx, s)
}

// Sync implements Phonelist. A failure keeps the previous entries and is
// shown in the settings.
func (u *uc) Sync(ctx context.Context) error {
	s, err := u.d.Repo.Settings(ctx)
	if err != nil || s.CardDAVURL == "" {
		return err
	}
	syncErr := u.syncWith(ctx, &s)
	if err := u.d.Repo.SaveSettings(ctx, s); err != nil {
		return err
	}
	return syncErr
}

// syncWith reads the book and replaces the address book entries; it
// records the outcome in s.
func (u *uc) syncWith(ctx context.Context, s *Settings) error {
	cards, err := u.d.Book.Cards(ctx, s.CardDAVURL, s.Username, s.Password)
	if err != nil {
		s.LastError = err.Error()
		return err
	}
	now := u.d.Clock.Now()
	entries := make([]domain.Entry, 0, len(cards))
	for _, c := range cards {
		if e, ok := domain.FromCard(c, now); ok && len(entries) < domain.MaxEntries {
			entries = append(entries, e)
		}
	}
	if err := u.d.Repo.ReplaceSource(ctx, domain.SourceCardDAV, entries); err != nil {
		return err
	}
	s.LastSync, s.LastError = now, ""
	return nil
}
