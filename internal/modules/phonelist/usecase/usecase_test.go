package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"trisse-home/internal/modules/phonelist/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ m map[string]domain.Entry }

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
