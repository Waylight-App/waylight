package usecase

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"waylight/internal/modules/board/domain"
	"waylight/internal/shared/kernel"
)

// memRepo keeps notes in memory and applies the visibility rule like the
// SQLite gateway.
type memRepo struct{ notes map[string]domain.Note }

func (m *memRepo) List(ctx context.Context) ([]domain.Note, error) {
	var out []domain.Note
	for _, n := range m.notes {
		if n.Access.Visibility.VisibleTo(kernel.PrincipalFrom(ctx), n.Access.Owner) {
			out = append(out, n)
		}
	}
	return out, nil
}

func (m *memRepo) Get(ctx context.Context, id string) (domain.Note, error) {
	n, ok := m.notes[id]
	if !ok || !n.Access.Visibility.VisibleTo(kernel.PrincipalFrom(ctx), n.Access.Owner) {
		return domain.Note{}, domain.ErrNotFound
	}
	return n, nil
}

func (m *memRepo) Save(_ context.Context, n domain.Note) error { m.notes[n.ID] = n; return nil }
func (m *memRepo) Delete(_ context.Context, id string) error   { delete(m.notes, id); return nil }

func (m *memRepo) DeleteExpired(_ context.Context, day time.Time) (int, error) {
	k := 0
	for id, n := range m.notes {
		if n.ExpiredOn(day) {
			delete(m.notes, id)
			k++
		}
	}
	return k, nil
}

type clk struct{ now *time.Time }

func (c clk) Now() time.Time { return *c.now }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "n" + strconv.Itoa(i.n) }

var (
	anna = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben  = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
)

func TestBoard(t *testing.T) {
	now := time.Date(2026, 11, 23, 18, 0, 0, 0, time.UTC)
	repo := &memRepo{notes: map[string]domain.Note{}}
	b := NewBoard(Deps{Repo: repo, Clock: clk{&now}, IDs: &ids{}, Location: time.UTC})
	short, err := b.Create(anna, domain.Input{Text: "Parcel in the hallway", ExpiresOn: "2026-11-23"}, "")
	if err != nil || short.Access.Visibility != kernel.Everyone || !short.Mine {
		t.Fatalf("%+v %v", short, err)
	}
	now = now.Add(time.Minute)
	if _, err := b.Create(anna, domain.Input{Text: "Wi-Fi password on the fridge", Pinned: true}, ""); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	if _, err := b.Create(anna, domain.Input{Text: "Secret"}, "private"); err != nil {
		t.Fatal(err)
	}
	list, _ := b.List(ben)
	if len(list) != 2 || list[0].Text != "Wi-Fi password on the fridge" || list[1].Mine {
		t.Fatalf("ben sees %+v", list)
	}
	if _, err := b.Update(ben, short.ID, domain.Input{Text: "Mine"}, ""); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed anna's note: %v", err)
	}
	// The next day the expired note is gone from the list at once and
	// removed by the cleanup.
	now = now.Add(24 * time.Hour)
	if list, _ := b.List(anna); len(list) != 2 {
		t.Fatalf("after expiry: %+v", list)
	}
	if n, err := b.Cleanup(context.Background()); n != 1 || err != nil {
		t.Fatalf("cleanup: %d %v", n, err)
	}
}

func TestSeedDemoOnce(t *testing.T) {
	now := time.Date(2026, 11, 23, 8, 0, 0, 0, time.UTC)
	repo := &memRepo{notes: map[string]domain.Note{}}
	d := Deps{Repo: repo, Clock: clk{&now}, IDs: &ids{}, Location: time.UTC}
	ctx := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	if err := SeedDemo(ctx, d); err != nil || len(repo.notes) == 0 {
		t.Fatalf("%v %d", err, len(repo.notes))
	}
	n := len(repo.notes)
	if err := SeedDemo(ctx, d); err != nil || len(repo.notes) != n {
		t.Fatalf("second seed: %d → %d %v", n, len(repo.notes), err)
	}
}
