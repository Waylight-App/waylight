package usecase

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"trisse-home/internal/modules/wiki/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ list []domain.Wiki }

func (m *memRepo) List(context.Context) ([]domain.Wiki, error) { return m.list, nil }
func (m *memRepo) Get(_ context.Context, id string) (domain.Wiki, error) {
	for _, w := range m.list {
		if w.ID == id {
			return w, nil
		}
	}
	return domain.Wiki{}, domain.ErrNotFound
}
func (m *memRepo) Save(_ context.Context, w domain.Wiki) error {
	for i, x := range m.list {
		if x.ID == w.ID {
			m.list[i] = w
			return nil
		}
	}
	m.list = append(m.list, w)
	return nil
}
func (m *memRepo) Delete(_ context.Context, id string) error {
	for i, x := range m.list {
		if x.ID == id {
			m.list = append(m.list[:i], m.list[i+1:]...)
			return nil
		}
	}
	return domain.ErrNotFound
}

type fakeSearcher struct {
	calls atomic.Int32
	fail  bool
	slow  bool
}

func (f *fakeSearcher) Search(ctx context.Context, w domain.Wiki, q string, limit int) ([]domain.Page, error) {
	f.calls.Add(1)
	if f.slow {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if f.fail {
		return nil, errors.New("down")
	}
	var out []domain.Page
	for range limit + 2 { // more than asked for: the use case cuts
		out = append(out, domain.Page{Title: q + " in " + w.Name, URL: w.URL + "/p"})
	}
	return out, nil
}

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return string(rune('a' + i.n - 1)) }

func TestWikis(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{}
	bs, ol := &fakeSearcher{}, &fakeSearcher{fail: true}
	c := &clk{t: time.Unix(0, 0)}
	uc := New(Deps{Repo: repo, Searchers: map[domain.Kind]PageSearcher{domain.KindBookStack: bs, domain.KindOutline: ol}, IDs: &ids{}, Clock: c})

	a, err := uc.Create(ctx, WikiInput{Name: "Handbook", Kind: domain.KindBookStack, URL: "https://wiki.example.org", Token: "id:secret"})
	if err != nil || !a.HasToken {
		t.Fatalf("%+v %v", a, err)
	}
	_, _ = uc.Create(ctx, WikiInput{Name: "Office", Kind: domain.KindOutline, URL: "https://docs.example.org", Token: "k"})

	if r, _ := uc.Search(ctx, "x"); len(r.Pages) != 0 || bs.calls.Load() != 0 {
		t.Fatal("one letter must not reach a wiki")
	}
	r, err := uc.Search(ctx, "vpn")
	if err != nil || len(r.Pages) != domain.MaxResultsPerWiki || r.Pages[0].Wiki != "Handbook" || len(r.Failed) != 1 || r.Failed[0] != "Office" {
		t.Fatalf("%+v %v", r, err)
	}
	// A partial answer is not cached: the failing wiki is asked again.
	_, _ = uc.Search(ctx, "vpn")
	if ol.calls.Load() != 2 {
		t.Fatalf("outline calls %d", ol.calls.Load())
	}
	ol.fail = false
	_, _ = uc.Search(ctx, "VPN")
	_, _ = uc.Search(ctx, "vpn")
	if bs.calls.Load() != 3 {
		t.Fatalf("a complete answer is reused within a minute, got %d calls", bs.calls.Load())
	}
	c.t = c.t.Add(2 * time.Minute)
	_, _ = uc.Search(ctx, "vpn")
	if bs.calls.Load() != 4 {
		t.Fatalf("expired cache, got %d calls", bs.calls.Load())
	}

	// The stored token only follows the same wiki.
	if _, err := uc.Update(ctx, a.ID, WikiInput{Name: "Renamed", Kind: domain.KindBookStack, URL: "https://wiki.example.org/"}); err != nil {
		t.Fatalf("rename keeps the token: %v", err)
	}
	if _, err := uc.Update(ctx, a.ID, WikiInput{Name: "x", Kind: domain.KindBookStack, URL: "https://other.example"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("a new address needs the token again: %v", err)
	}
	if err := uc.Test(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	for range domain.MaxWikis {
		_, _ = uc.Create(ctx, WikiInput{Name: "w", Kind: domain.KindBookStack, URL: "https://w.example", Token: "t"})
	}
	if _, err := uc.Create(ctx, WikiInput{Name: "w", Kind: domain.KindBookStack, URL: "https://w.example", Token: "t"}); !errors.Is(err, kernel.ErrConflict) {
		t.Fatalf("limit: %v", err)
	}
}

func TestSlowWikiDoesNotBlock(t *testing.T) {
	repo := &memRepo{list: []domain.Wiki{{ID: "s", Name: "Slow", Kind: domain.KindWikiJS}, {ID: "f", Name: "Fast", Kind: domain.KindBookStack, URL: "https://f"}}}
	uc := New(Deps{Repo: repo, Searchers: map[domain.Kind]PageSearcher{domain.KindWikiJS: &fakeSearcher{slow: true}, domain.KindBookStack: &fakeSearcher{}},
		IDs: &ids{}, Clock: &clk{}})
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	r, err := uc.Search(ctx, "anything")
	if err != nil || len(r.Pages) == 0 || len(r.Failed) != 1 || time.Since(start) > time.Second {
		t.Fatalf("%+v %v after %v", r, err, time.Since(start))
	}
}
