package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"waylight/internal/modules/phonelist/domain"
	"waylight/internal/shared/kernel"
)

type memRepo struct {
	m        map[string]domain.Entry
	settings Settings
}

func (r *memRepo) ReplaceSource(_ context.Context, src domain.Source, list []domain.Entry) error {
	for id, e := range r.m {
		if e.Source == src {
			delete(r.m, id)
		}
	}
	for _, e := range list {
		e.Source = src
		r.m[e.ID] = e
	}
	return nil
}
func (r *memRepo) Settings(context.Context) (Settings, error)       { return r.settings, nil }
func (r *memRepo) SaveSettings(_ context.Context, s Settings) error { r.settings = s; return nil }

type fakeBook struct {
	cards []domain.Card
	err   error
}

func (b *fakeBook) Cards(context.Context, string, string, string) ([]domain.Card, error) {
	return b.cards, b.err
}

func (r *memRepo) List(context.Context) ([]domain.Entry, error) {
	out := make([]domain.Entry, 0, len(r.m))
	for _, e := range r.m {
		out = append(out, e)
	}
	return out, nil
}

func (r *memRepo) Get(_ context.Context, id string) (domain.Entry, error) {
	e, ok := r.m[id]
	if !ok {
		return domain.Entry{}, domain.ErrEntryNotFound
	}
	return e, nil
}

func (r *memRepo) Save(_ context.Context, e domain.Entry) error { r.m[e.ID] = e; return nil }
func (r *memRepo) Delete(_ context.Context, id string) error    { delete(r.m, id); return nil }

type clk struct{}

func (clk) Now() time.Time { return time.Unix(0, 0) }

type seq struct{ n int }

func (s *seq) NewID() string { s.n++; return fmt.Sprintf("e%d", s.n) }

func TestPhonelist(t *testing.T) {
	ctx := context.Background()
	admin := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: "owner", Role: "owner"})
	member := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: "m1", Role: "member"})
	p := NewPhonelist(Deps{Repo: &memRepo{m: map[string]domain.Entry{}}, Clock: clk{}, IDs: &seq{}})

	if _, err := p.Add(member, domain.EntryInput{Name: "Ben", Phone: "1"}); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatal("members only read the list", err)
	}
	anna, err := p.Add(admin, domain.EntryInput{Name: "anna Weber", Role: "Reception", Phone: "+49 30 1", Team: "Front desk"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Add(admin, domain.EntryInput{Name: "Ben Ott", Role: "Accounting", Email: "ben@example.org"}); err != nil {
		t.Fatal(err)
	}
	list, _ := p.List(member)
	if len(list) != 2 || list[0].Name != "anna Weber" {
		t.Fatalf("sorted by name ignoring case: %+v", list)
	}
	if hits, _ := p.Search(member, "front ANNA"); len(hits) != 1 || hits[0].ID != anna.ID {
		t.Fatalf("search: %+v", hits)
	}
	if hits, _ := p.Search(member, "  "); len(hits) != 0 {
		t.Fatal("an empty query finds nothing")
	}
	if _, err := p.Update(member, anna.ID, domain.EntryInput{Name: "x", Phone: "1"}); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatal(err)
	}
	if e, err := p.Update(admin, anna.ID, domain.EntryInput{Name: "Anna Weber", Phone: "2"}); err != nil || e.Phone != "2" {
		t.Fatal(e, err)
	}
	if err := p.Delete(member, anna.ID); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatal(err)
	}
	if err := p.Delete(admin, anna.ID); err != nil {
		t.Fatal(err)
	}
	if err := p.Delete(admin, anna.ID); !errors.Is(err, domain.ErrEntryNotFound) {
		t.Fatal(err)
	}
}

func TestAddressBook(t *testing.T) {
	ctx := context.Background()
	admin := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: "owner", Role: "owner"})
	member := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: "m1", Role: "member"})
	book := &fakeBook{cards: []domain.Card{{ID: "cd-1", EntryInput: domain.EntryInput{Name: "Anna", Phone: "1"}}}}
	repo := &memRepo{m: map[string]domain.Entry{}}
	p := NewPhonelist(Deps{Repo: repo, Book: book, Clock: clk{}, IDs: &seq{}})

	if _, err := p.SaveSettings(member, SettingsInput{CardDAVURL: "https://dav.example/book"}); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatal(err)
	}
	if _, err := p.SaveSettings(admin, SettingsInput{CardDAVURL: "https://dav.example/book", Username: "u"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("a password is needed", err)
	}
	v, err := p.SaveSettings(admin, SettingsInput{CardDAVURL: "https://dav.example/book", Username: "u", Password: "pw"})
	if err != nil || v.CardDAVURL != "https://dav.example/book/" || !v.HasPassword {
		t.Fatal(v, err)
	}
	list, _ := p.List(member)
	if len(list) != 1 || list[0].Source != domain.SourceCardDAV {
		t.Fatalf("%+v", list)
	}
	if _, err := p.Update(admin, "cd-1", domain.EntryInput{Name: "x", Phone: "1"}); !errors.Is(err, domain.ErrReadOnlyEntry) {
		t.Fatal("address book entries are changed in the book", err)
	}
	if err := p.Delete(admin, "cd-1"); !errors.Is(err, domain.ErrReadOnlyEntry) {
		t.Fatal(err)
	}

	// A failing sync keeps the entries and records the error.
	book.err = errors.New("401 Unauthorized")
	if err := p.Sync(ctx); err == nil {
		t.Fatal("sync error must be reported")
	}
	if list, _ := p.List(member); len(list) != 1 || repo.settings.LastError == "" {
		t.Fatalf("%+v %q", list, repo.settings.LastError)
	}

	// Switching the book off removes its people; manual entries stay.
	_, _ = p.Add(admin, domain.EntryInput{Name: "Desk", Phone: "2"})
	if _, err := p.SaveSettings(admin, SettingsInput{}); err != nil {
		t.Fatal(err)
	}
	if list, _ := p.List(member); len(list) != 1 || list[0].Name != "Desk" {
		t.Fatalf("%+v", list)
	}
}
