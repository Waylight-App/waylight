package usecase

import (
	"context"
	"fmt"
	"slices"
	"time"

	"waylight/internal/modules/board/domain"
	"waylight/internal/shared/kernel"
)

// Board is the input port of the module; every method is one use case.
type Board interface {
	// List returns the visible notes that have not expired, pinned first.
	List(ctx context.Context) ([]NoteView, error)
	// Create adds a note. An empty visibility shares it with everyone.
	Create(ctx context.Context, in domain.Input, visibility string) (NoteView, error)
	// Update changes a note of the principal; an empty visibility keeps it.
	Update(ctx context.Context, id string, in domain.Input, visibility string) (NoteView, error)
	Delete(ctx context.Context, id string) error
	// Cleanup removes the notes of every account that expired before
	// today (job).
	Cleanup(ctx context.Context) (int, error)
}

// NoteView is a note as the principal sees it.
type NoteView struct {
	domain.Note
	// Mine says whether the principal wrote the note and may change it.
	Mine bool
}

// Repository persists notes (output port, implemented in adapter/gateway).
// Reads return only what the principal in ctx may see; Save and Delete
// change only the principal's own notes (ADR 0018).
type Repository interface {
	List(ctx context.Context) ([]domain.Note, error)
	Get(ctx context.Context, id string) (domain.Note, error)
	Save(ctx context.Context, n domain.Note) error
	Delete(ctx context.Context, id string) error
	// DeleteExpired removes notes of all accounts whose last day is
	// before day.
	DeleteExpired(ctx context.Context, day time.Time) (int, error)
}

// Deps bundles the output ports. Location decides what "today" is
// (nil = time.Local).
type Deps struct {
	Repo     Repository
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
	Location *time.Location
}

type uc struct{ d Deps }

// NewBoard returns the family board use cases.
func NewBoard(d Deps) Board {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &uc{d: d}
}

func (u *uc) today() time.Time { return u.d.Clock.Now().In(u.d.Location) }

func (u *uc) view(ctx context.Context, n domain.Note) NoteView {
	return NoteView{Note: n, Mine: n.Access.Mine(kernel.PrincipalFrom(ctx))}
}

// List implements Board. Expired notes are filtered here as well, so they
// disappear at midnight even before the cleanup job runs.
func (u *uc) List(ctx context.Context) ([]NoteView, error) {
	list, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	today := u.today()
	list = slices.DeleteFunc(list, func(n domain.Note) bool { return n.ExpiredOn(today) })
	slices.SortStableFunc(list, func(a, b domain.Note) int {
		switch {
		case domain.Less(a, b):
			return -1
		case domain.Less(b, a):
			return 1
		}
		return 0
	})
	out := make([]NoteView, 0, len(list))
	for _, n := range list {
		out = append(out, u.view(ctx, n))
	}
	return out, nil
}

// shareFor is the visibility of a new note: everyone unless the writer
// chose otherwise, because the board is for the whole household.
func shareFor(p kernel.Principal, raw string) (kernel.Visibility, error) {
	if raw == "" {
		return kernel.Everyone, nil
	}
	return kernel.CheckShare(p, raw)
}

// Create implements Board.
func (u *uc) Create(ctx context.Context, in domain.Input, visibility string) (NoteView, error) {
	p := kernel.PrincipalFrom(ctx)
	v, err := shareFor(p, visibility)
	if err != nil {
		return NoteView{}, err
	}
	n, err := domain.New(u.d.IDs.NewID(), in, kernel.Access{Owner: p.AccountID, Visibility: v}, u.d.Clock.Now(), u.today())
	if err != nil {
		return NoteView{}, err
	}
	if err := u.d.Repo.Save(ctx, n); err != nil {
		return NoteView{}, fmt.Errorf("save note: %w", err)
	}
	return u.view(ctx, n), nil
}

// Update implements Board.
func (u *uc) Update(ctx context.Context, id string, in domain.Input, visibility string) (NoteView, error) {
	p := kernel.PrincipalFrom(ctx)
	n, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return NoteView{}, err
	}
	if !n.Access.Mine(p) {
		return NoteView{}, kernel.ErrNotOwner
	}
	if visibility != "" {
		if n.Access.Visibility, err = kernel.CheckShare(p, visibility); err != nil {
			return NoteView{}, err
		}
	}
	if n, err = n.Update(in); err != nil {
		return NoteView{}, err
	}
	if err := u.d.Repo.Save(ctx, n); err != nil {
		return NoteView{}, err
	}
	return u.view(ctx, n), nil
}

// Delete implements Board.
func (u *uc) Delete(ctx context.Context, id string) error { return u.d.Repo.Delete(ctx, id) }

// Cleanup implements Board.
func (u *uc) Cleanup(ctx context.Context) (int, error) {
	t := u.today()
	return u.d.Repo.DeleteExpired(ctx, time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC))
}
