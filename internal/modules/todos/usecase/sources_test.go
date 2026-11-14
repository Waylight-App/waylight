package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/shared/kernel"
)

type memSources struct{ m map[string]domain.Source }

func (r *memSources) List(context.Context) ([]domain.Source, error) {
	var out []domain.Source
	for _, s := range r.m {
		out = append(out, s)
	}
	return out, nil
}
func (r *memSources) Get(_ context.Context, id string) (domain.Source, error) {
	s, ok := r.m[id]
	if !ok {
		return s, domain.ErrSourceNotFound
	}
	return s, nil
}
func (r *memSources) Save(_ context.Context, s domain.Source) error { r.m[s.ID] = s; return nil }
func (r *memSources) Delete(_ context.Context, id string) error   { delete(r.m, id); return nil }

// fakeProviders serves one external list per source ID; a to-do read back
// after a write gets a new version, like an ETag.
type fakeProviders struct {
	lists  map[string]*memRepo
	opened []domain.Source
	fail   error
}

type fakeList struct {
	*memRepo
	fail error
}

func (l fakeList) List(ctx context.Context) ([]domain.Todo, error) {
	if l.fail != nil {
		return nil, l.fail
	}
	return l.memRepo.List(ctx)
}
func (l fakeList) Save(_ context.Context, t domain.Todo) error {
	if old, ok := l.m[t.ID()]; ok {
		t = t.WithVersion(old.Version() + 10)
	} else {
		t = t.WithVersion(100)
	}
	l.m[t.ID()] = t
	return nil
}

func (p *fakeProviders) Open(s domain.Source) Repository {
	p.opened = append(p.opened, s)
	if p.lists[s.ID] == nil {
		p.lists[s.ID] = &memRepo{m: map[string]domain.Todo{}}
	}
	return fakeList{p.lists[s.ID], p.fail}
}

func (p *fakeProviders) Discover(_ context.Context, _ domain.SourceKind, url, user, pw string) ([]DiscoveredList, error) {
	if pw != "pw" {
		return nil, errors.New("401")
	}
	return []DiscoveredList{{URL: url + "tasks/", Name: "Tasks"}}, nil
}

func externalDeps() (Deps, *fakeProviders, *memSources) {
	p := &fakeProviders{lists: map[string]*memRepo{}}
	src := &memSources{m: map[string]domain.Source{}}
	return Deps{Repo: &memRepo{m: map[string]domain.Todo{}}, Clock: &clk{t: time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC)},
		IDs: &ids{}, Sources: src, Providers: p}, p, src
}

func TestSources(t *testing.T) {
	ctx := context.Background()
	d, p, src := externalDeps()
	uc := NewSources(d)
	in := SourceInput{Name: "Nextcloud", Kind: "caldav", URL: "https://cloud.example/dav/tasks", Username: "anna", Password: "pw"}
	v, err := uc.Save(ctx, in)
	if err != nil || v.ID == "" || v.URL != "https://cloud.example/dav/tasks/" || v.Capabilities.Repeat || len(p.opened) != 1 {
		t.Fatalf("%+v %v", v, err)
	}
	// Editing without a password keeps the stored one.
	if _, err := uc.Save(ctx, SourceInput{ID: v.ID, Name: "Cloud", Kind: "caldav", URL: in.URL, Username: "anna"}); err != nil || src.m[v.ID].Password != "pw" || src.m[v.ID].Name != "Cloud" {
		t.Fatalf("%+v %v", src.m[v.ID], err)
	}
	// The connection is tested before saving.
	p.fail = errors.New("connection refused")
	if _, err := uc.Save(ctx, SourceInput{Name: "x", Kind: "caldav", URL: in.URL, Username: "a", Password: "b"}); !errors.Is(err, kernel.ErrUnavailable) || len(src.m) != 1 {
		t.Fatal("untested list stored", err)
	}
	p.fail = nil
	if got, err := uc.Discover(ctx, SourceInput{ID: v.ID, Kind: "caldav", URL: "https://cloud.example/dav/", Username: "anna"}); err != nil || len(got) != 1 {
		t.Fatalf("discover with the stored password: %+v %v", got, err)
	}
	if _, err := uc.Discover(ctx, SourceInput{Kind: "caldav", URL: "https://cloud.example/dav/", Username: "anna", Password: "x"}); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatal(err)
	}
	for range domain.MaxSources - 1 {
		_, _ = uc.Save(ctx, in)
	}
	if _, err := uc.Save(ctx, in); !errors.Is(err, domain.ErrTooManySources) {
		t.Fatal("limit", err)
	}
	if err := uc.Delete(ctx, v.ID); err != nil || src.m[v.ID].ID != "" {
		t.Fatal(err)
	}
	if err := uc.Delete(ctx, v.ID); !errors.Is(err, domain.ErrSourceNotFound) {
		t.Fatal(err)
	}
	if list, _ := uc.List(ctx); len(list) != domain.MaxSources-1 {
		t.Fatal(len(list))
	}
}

func TestExternalList(t *testing.T) {
	ctx := context.Background()
	d, p, src := externalDeps()
	src.m["nc"] = domain.Source{ID: "nc", Name: "NC", Kind: domain.KindCalDAV}
	in := func(title string) TodoInput { return TodoInput{Title: title, Source: "nc"} }

	a, err := NewCreateTodo(d).Execute(ctx, in("Steuer"))
	if err != nil || a.Version != 100 {
		t.Fatalf("created (read back with the provider's version): %+v %v", a, err)
	}
	if local, _ := NewListTodos(d).Execute(ctx, "", true); len(local) != 0 {
		t.Fatal("went to the local list")
	}
	if list, _ := NewListTodos(d).Execute(ctx, "nc", true); len(list) != 1 || list[0].Title != "Steuer" {
		t.Fatalf("%+v", list)
	}
	// Repeat and reminders are refused for lists that cannot store them.
	rep := in("x")
	rep.Due, rep.Repeat = time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), &RepeatInput{Kind: "weekly", Every: 1, Days: []int{1}}
	if _, err := NewCreateTodo(d).Execute(ctx, rep); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatal(err)
	}
	// Server wins: an edit made from an older version is refused.
	up := in("Steuer erklären")
	up.Version = a.Version
	b, err := NewUpdateTodo(d).Execute(ctx, a.ID, up, true)
	if err != nil || !b.Done || b.Version != 110 {
		t.Fatalf("%+v %v", b, err)
	}
	if _, err := NewUpdateTodo(d).Execute(ctx, a.ID, up, false); !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatal("stale version accepted", err)
	}
	if n, err := NewClearCompleted(d).Execute(ctx, "nc"); err != nil || n != 1 || len(p.lists["nc"].m) != 0 {
		t.Fatal(n, err)
	}
	c, _ := NewCreateTodo(d).Execute(ctx, in("Brief"))
	if err := NewDeleteTodo(d).Execute(ctx, "nc", c.ID); err != nil {
		t.Fatal(err)
	}
	// Unknown lists and provider failures.
	for _, s := range []string{"gone", "../x"} {
		if _, err := NewListTodos(d).Execute(ctx, s, false); !errors.Is(err, domain.ErrSourceNotFound) {
			t.Fatal(s, err)
		}
	}
	p.fail = errors.New("dial tcp: connection refused")
	if _, err := NewListTodos(d).Execute(ctx, "nc", false); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatal(err)
	}
	if _, err := NewListTodos(Deps{Repo: d.Repo, Clock: d.Clock}).Execute(ctx, "nc", false); !errors.Is(err, domain.ErrSourceNotFound) {
		t.Fatal("no providers wired", err)
	}
}
