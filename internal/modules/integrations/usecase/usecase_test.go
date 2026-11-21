package usecase

import (
	"context"
	"errors"
	"maps"
	"slices"
	"testing"
	"time"

	"waylight/internal/modules/integrations/domain"
	"waylight/internal/shared/kernel"
)

type memRepo struct{ m map[string]domain.Connection }

func (r *memRepo) List(context.Context) ([]domain.Connection, error) {
	keys := slices.Sorted(maps.Keys(r.m))
	out := make([]domain.Connection, 0, len(keys))
	for _, k := range keys {
		out = append(out, r.m[k])
	}
	return out, nil
}

func (r *memRepo) Get(_ context.Context, id string) (domain.Connection, error) {
	c, ok := r.m[id]
	if !ok {
		return domain.Connection{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *memRepo) Save(_ context.Context, c domain.Connection) error { r.m[c.ID] = c; return nil }
func (r *memRepo) Delete(_ context.Context, id string) error         { delete(r.m, id); return nil }

func (r *memRepo) ReplaceAll(_ context.Context, list []domain.Connection) error {
	r.m = map[string]domain.Connection{}
	for _, c := range list {
		r.m[c.ID] = c
	}
	return nil
}

type clock struct{}

func (clock) Now() time.Time { return time.Unix(1_800_000_000, 0) }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return string(rune('a' + i.n - 1)) }

func admin() context.Context {
	return kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "a", Role: "admin"})
}

func newUC(widgets []Widget) (Connections, *memRepo) {
	repo := &memRepo{m: map[string]domain.Connection{}}
	return New(Deps{
		Repo: repo, Clock: clock{}, IDs: &ids{},
		Widgets: func(context.Context) ([]Widget, error) { return widgets, nil },
	}), repo
}

func TestOnlyAdmins(t *testing.T) {
	uc, _ := newUC(nil)
	member := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "m", Role: "member"})
	if _, err := uc.List(member, ""); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("list: %v", err)
	}
	if _, err := uc.Add(member, Input{Service: "immich", Name: "Home", URL: "http://h"}); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("add: %v", err)
	}
}

func TestAddListAndNames(t *testing.T) {
	uc, repo := newUC([]Widget{{ID: "w1", TabID: "t1", Kind: "immich", ConnectionID: "a"}, {ID: "w2", TabID: "t1", Kind: "notes"}})
	ctx := admin()
	home, err := uc.Add(ctx, Input{Service: "immich", Name: "Home", URL: "http://192.168.178.20:2283", Credential: map[string]string{"apiKey": "secret"}})
	if err != nil || home.ID != "a" || !slices.Equal(home.Fields, []string{"apiKey"}) {
		t.Fatalf("%+v %v", home, err)
	}
	if repo.m["a"].Credential["apiKey"] != "secret" {
		t.Fatal("the credential must be stored")
	}
	if _, err := uc.Add(ctx, Input{Service: "immich", Name: "home", URL: "http://h"}); !errors.Is(err, domain.ErrNameTaken) {
		t.Fatalf("same name: %v", err)
	}
	if _, err := uc.Add(ctx, Input{Service: "proxmox", Name: "Home", URL: "https://pve:8006",
		Credential: map[string]string{"tokenId": "w@pve!d", "secret": "s"}}); err != nil {
		t.Fatalf("the name may repeat across services: %v", err)
	}
	list, err := uc.List(ctx, "immich")
	if err != nil || len(list) != 1 || len(list[0].UsedBy) != 1 || list[0].UsedBy[0].WidgetID != "w1" {
		t.Fatalf("%+v %v", list, err)
	}
	all, _ := uc.List(ctx, "")
	if len(all) != 2 {
		t.Fatalf("all: %+v", all)
	}
}

func TestUpdateKeepsTheCredentialOnlyAtTheSameAddress(t *testing.T) {
	uc, repo := newUC(nil)
	ctx := admin()
	c, _ := uc.Add(ctx, Input{Service: "immich", Name: "Home", URL: "http://h:2283", Credential: map[string]string{"apiKey": "k"}})
	if _, err := uc.Update(ctx, c.ID, Change{Name: "Photos", URL: "http://h:2283"}); err != nil || repo.m[c.ID].Credential["apiKey"] != "k" {
		t.Fatalf("rename: %v %+v", err, repo.m[c.ID])
	}
	var v *kernel.ValidationError
	if _, err := uc.Update(ctx, c.ID, Change{Name: "Photos", URL: "http://elsewhere:2283"}); !errors.As(err, &v) {
		t.Fatalf("a new address needs the credential again: %v", err)
	}
	if _, err := uc.Update(ctx, c.ID, Change{Name: "Photos", URL: "http://elsewhere:2283", Credential: map[string]string{"apiKey": "new"}}); err != nil ||
		repo.m[c.ID].Credential["apiKey"] != "new" {
		t.Fatalf("move with credential: %v", err)
	}
	if _, err := uc.ReplaceCredential(ctx, c.ID, map[string]string{"apiKey": "newer"}); err != nil || repo.m[c.ID].Credential["apiKey"] != "newer" {
		t.Fatalf("replace: %v", err)
	}
}

func TestRemoveInUseAsksFirst(t *testing.T) {
	uc, repo := newUC([]Widget{{ID: "w1", TabID: "t1", Kind: "immich", ConnectionID: "a"}})
	ctx := admin()
	_, _ = uc.Add(ctx, Input{Service: "immich", Name: "Home", URL: "http://h"})
	_, _ = uc.Add(ctx, Input{Service: "immich", Name: "Spare", URL: "http://s"})
	if err := uc.Remove(ctx, "a", false); !errors.Is(err, ErrInUse) {
		t.Fatalf("in use: %v", err)
	}
	if err := uc.Remove(ctx, "b", false); err != nil {
		t.Fatalf("unused: %v", err)
	}
	if err := uc.Remove(ctx, "a", true); err != nil || len(repo.m) != 0 {
		t.Fatalf("confirmed: %v %v", err, repo.m)
	}
	if err := uc.Remove(ctx, "a", true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("gone: %v", err)
	}
}

func TestConfigRoundTripKeepsIDsAndCredentials(t *testing.T) {
	repo := &memRepo{m: map[string]domain.Connection{}}
	created := time.Unix(1_700_000_000, 0)
	repo.m["x"] = domain.Connection{ID: "x", Service: "immich", Name: "Home", URL: "http://h:2283", Credential: domain.Credential{"apiKey": "k"}, CreatedAt: created}
	repo.m["y"] = domain.Connection{ID: "y", Service: "grafana", Name: "Ops", URL: "http://g:3000", Credential: domain.Credential{"token": "t"}}
	repo.m["z"] = domain.Connection{ID: "z", Service: "proxmox", Name: "Old", URL: "https://p:8006"}
	cfg := NewConfig(repo, clock{}, &ids{})
	ctx := context.Background()
	exported, err := cfg.Export(ctx)
	if err != nil || len(exported) != 3 || exported[0].Service != "grafana" {
		t.Fatalf("%+v %v", exported, err)
	}
	err = cfg.Apply(ctx, []ConfigConnection{
		{Service: "immich", Name: "home", URL: "http://h:2283"},  // same, other case
		{Service: "grafana", Name: "Ops", URL: "http://g2:3000"}, // moved
		{Service: "nextcloud", Name: "Cloud", URL: "https://cloud.example.org"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if c := repo.m["x"]; c.Credential["apiKey"] != "k" || c.Name != "home" || !c.CreatedAt.Equal(created) {
		t.Fatalf("kept: %+v", c)
	}
	if c := repo.m["y"]; len(c.Credential) != 0 || c.URL != "http://g2:3000" {
		t.Fatalf("a moved connection must drop its credential: %+v", c)
	}
	if _, ok := repo.m["z"]; ok || len(repo.m) != 3 {
		t.Fatalf("missing entries are removed: %+v", repo.m)
	}
	before := maps.Clone(repo.m)
	if err := cfg.Apply(ctx, []ConfigConnection{{Service: "immich", Name: "A", URL: "http://a"}, {Service: "immich", Name: "a", URL: "http://b"}}); !errors.Is(err, domain.ErrNameTaken) {
		t.Fatalf("duplicates: %v", err)
	}
	if err := cfg.Apply(ctx, []ConfigConnection{{Service: "immich", Name: "A", URL: "ftp://a"}}); err == nil {
		t.Fatal("an invalid entry must be refused")
	}
	if len(repo.m) != len(before) {
		t.Fatal("a refused section must change nothing")
	}
}
