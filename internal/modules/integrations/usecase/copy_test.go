package usecase

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"waylight/internal/modules/integrations/domain"
)

// copyRepo adds the copy records to memRepo; corrupt changes what is read
// back, like a cipher that does not round-trip.
type copyRepo struct {
	*memRepo
	done    map[string]bool
	corrupt bool
}

func (r *copyRepo) Copied(context.Context) (map[string]bool, error) { return maps.Clone(r.done), nil }

func (r *copyRepo) SaveCopy(_ context.Context, source string, c domain.Connection, verify func(domain.Connection) error) error {
	if r.done[source] {
		return nil
	}
	back := c
	back.Credential = maps.Clone(c.Credential)
	if r.corrupt {
		for k := range back.Credential {
			back.Credential[k] = "garbled"
		}
	}
	if err := verify(back); err != nil {
		return err
	}
	r.m[c.ID] = c
	r.done[source] = true
	return nil
}

func newCopier() (Copier, *copyRepo) {
	repo := &copyRepo{memRepo: &memRepo{m: map[string]domain.Connection{}}, done: map[string]bool{}}
	return NewCopier(repo, services(&[]string{}), clock{}, &ids{}), repo
}

// fixtures are module settings as the modules return them, including the
// empty optional fields a module stores.
func fixtures() []Saved {
	return []Saved{
		{Source: "immich", Service: "immich", URL: "http://192.168.178.20:2283/", Credential: map[string]string{"apiKey": "imm-Key-1"}},
		{Source: "proxmox", Service: "proxmox", URL: "https://pve.lan:8006",
			Credential: map[string]string{"tokenId": "waylight@pve!dash", "secret": "0b1c-uuid", "fingerprint": ""}},
		{Source: "fuel", Service: "tankerkoenig", URL: "https://creativecommons.tankerkoenig.de/json",
			Credential: map[string]string{"apiKey": "00000000-0000-0000-0000-000000000002"}},
	}
}

func TestCopyOnceAndVerbatim(t *testing.T) {
	c, repo := newCopier()
	ctx := context.Background()
	res, err := c.Copy(ctx, fixtures())
	if err != nil || !slices.Equal(res.Copied, []string{"immich", "proxmox", "fuel"}) || len(res.Skipped) != 0 {
		t.Fatalf("%+v %v", res, err)
	}
	byService := map[string]domain.Connection{}
	for _, x := range repo.m {
		byService[x.Service] = x
	}
	if x := byService["immich"]; x.Name != "Immich" || x.URL != "http://192.168.178.20:2283" || x.Credential["apiKey"] != "imm-Key-1" {
		t.Fatalf("immich: %+v", x)
	}
	if x := byService["proxmox"]; !maps.Equal(x.Credential, domain.Credential{"tokenId": "waylight@pve!dash", "secret": "0b1c-uuid"}) {
		t.Fatalf("proxmox: %+v", x)
	}
	// A second run, e.g. the next start, copies nothing and changes nothing.
	before := maps.Clone(repo.m)
	res, err = c.Copy(ctx, fixtures())
	if err != nil || len(res.Copied) != 0 || len(res.Skipped) != 0 || !maps.EqualFunc(before, repo.m, func(a, b domain.Connection) bool { return a.ID == b.ID }) {
		t.Fatalf("second run: %+v %v", res, err)
	}
	// A connection the admin removed later is not brought back.
	for id, x := range repo.m {
		if x.Service == "immich" {
			delete(repo.m, id)
		}
	}
	if res, _ = c.Copy(ctx, fixtures()); len(res.Copied) != 0 {
		t.Fatalf("removed copy came back: %+v", res)
	}
}

func TestCopyNumbersATakenName(t *testing.T) {
	c, repo := newCopier()
	repo.m["own"] = domain.Connection{ID: "own", Service: "immich", Name: "Immich", URL: "http://other:2283"}
	saved := []Saved{
		{Source: "a", Service: "immich", URL: "http://h:2283", Credential: map[string]string{"apiKey": "k"}},
		{Source: "b", Service: "immich", Name: strings.Repeat("x", domain.MaxNameLength), URL: "http://i:2283", Credential: map[string]string{"apiKey": "k"}},
		{Source: "c", Service: "immich", Name: strings.Repeat("x", domain.MaxNameLength), URL: "http://j:2283", Credential: map[string]string{"apiKey": "k"}},
	}
	if res, err := c.Copy(context.Background(), saved); err != nil || len(res.Copied) != 3 {
		t.Fatalf("%+v %v", res, err)
	}
	var names []string
	for _, x := range repo.m {
		names = append(names, x.Name)
	}
	slices.Sort(names)
	want := []string{"Immich", "Immich 2", strings.Repeat("x", domain.MaxNameLength), strings.Repeat("x", domain.MaxNameLength-2) + " 2"}
	slices.Sort(want)
	if !slices.Equal(names, want) {
		t.Fatalf("%q", names)
	}
}

func TestCopySkipsWhatCannotBeCopiedSafely(t *testing.T) {
	c, repo := newCopier()
	saved := []Saved{
		{Source: "gone", Service: "gitlab", URL: "http://g", Credential: map[string]string{"token": "t"}},
		{Source: "partial", Service: "proxmox", URL: "https://p:8006", Credential: map[string]string{"tokenId": "u@pam!x", "secret": ""}},
		{Source: "query", Service: "immich", URL: "http://h:2283/?x=1", Credential: map[string]string{"apiKey": "k"}},
		{Source: "stray", Service: "immich", URL: "http://h:2283", Credential: map[string]string{"apiKey": "k", "password": "p"}},
	}
	res, err := c.Copy(context.Background(), saved)
	if err != nil || len(res.Copied) != 0 || len(res.Skipped) != 4 || len(repo.m) != 0 || len(repo.done) != 0 {
		t.Fatalf("%+v %v %v", res, err, repo.m)
	}
	for src, why := range res.Skipped {
		if strings.Contains(why, "u@pam!x") || why == "" {
			t.Errorf("%s: %q", src, why)
		}
	}
}

func TestCopyMismatchIsRetried(t *testing.T) {
	c, repo := newCopier()
	repo.corrupt = true
	res, err := c.Copy(context.Background(), fixtures()[:1])
	if err != nil || res.Skipped["immich"] != ErrCopyMismatch.Error() || len(repo.m) != 0 || repo.done["immich"] {
		t.Fatalf("%+v %v", res, err)
	}
	if strings.Contains(fmt.Sprint(res), "imm-Key-1") {
		t.Fatal("the result must not carry the credential")
	}
	repo.corrupt = false
	if res, _ = c.Copy(context.Background(), fixtures()[:1]); !slices.Equal(res.Copied, []string{"immich"}) {
		t.Fatalf("retry: %+v", res)
	}
}

func TestCopyStopsAtTheLimit(t *testing.T) {
	c, repo := newCopier()
	for i := range domain.MaxConnections {
		id := fmt.Sprint("x", i)
		repo.m[id] = domain.Connection{ID: id, Service: "immich", Name: id}
	}
	res, _ := c.Copy(context.Background(), fixtures()[:1])
	if res.Skipped["immich"] != domain.ErrTooMany.Error() || repo.done["immich"] {
		t.Fatalf("%+v", res)
	}
}
