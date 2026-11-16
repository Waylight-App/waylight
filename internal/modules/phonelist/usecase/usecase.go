package usecase

import (
	"context"
	"fmt"
	"slices"
	"strings"

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
}

// Repository persists entries (output port, implemented in adapter/gateway).
type Repository interface {
	List(ctx context.Context) ([]domain.Entry, error)
	Get(ctx context.Context, id string) (domain.Entry, error)
	Save(ctx context.Context, e domain.Entry) error
	Delete(ctx context.Context, id string) error
}

// Deps bundles the output ports.
type Deps struct {
	Repo  Repository
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
	if _, err := u.d.Repo.Get(ctx, id); err != nil {
		return err
	}
	return u.d.Repo.Delete(ctx, id)
}
