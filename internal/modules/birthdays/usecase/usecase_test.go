package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"trisse-home/internal/modules/birthdays/domain"
	"trisse-home/internal/shared/kernel"
)

type repo struct {
	m map[string]domain.Birthday
	s Settings
}

func (r *repo) List(context.Context) ([]domain.Birthday, error) {
	var out []domain.Birthday
	for _, b := range r.m {
		out = append(out, b)
	}
	return out, nil
}
func (r *repo) Save(_ context.Context, b domain.Birthday) error { r.m[b.ID] = b; return nil }
func (r *repo) Delete(_ context.Context, id string) error {
	if _, ok := r.m[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.m, id)
	return nil
}
func (r *repo) ReplaceSource(_ context.Context, source string, list []domain.Birthday) error {
	for id, b := range r.m {
		if b.Source == source {
			delete(r.m, id)
		}
	}
	for _, b := range list {
		r.m[b.ID] = b
	}
	return nil
}
func (r *repo) Settings(context.Context) (Settings, error) { return r.s, nil }

// SyncAccounts: the fake has one address book, the owner's, when configured.
func (r *repo) SyncAccounts(context.Context) ([]string, error) {
	if r.s.CardDAVURL == "" {
		return nil, nil
	}
	return []string{"owner"}, nil
}

func (r *repo) SaveSettings(_ context.Context, s Settings) error { r.s = s; return nil }

type book struct {
	cards []domain.Card
	fail  bool
}

func (b *book) Cards(_ context.Context, _, _, pw string) ([]domain.Card, error) {
	if b.fail || pw != "pw" {
		return nil, errors.New("401")
	}
	return b.cards, nil
}

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return fmt.Sprint("id", i.n) }

type alerts struct{ got []kernel.Alert }

func (a *alerts) Alert(_ context.Context, al kernel.Alert) { a.got = append(a.got, al) }

func TestBirthdays(t *testing.T) {
	ctx := context.Background()
	r := &repo{m: map[string]domain.Birthday{}}
	bk := &book{cards: []domain.Card{{UID: "u1", Name: "Carla", Month: 11, Day: 14, Year: 1980}}}
	c := &clk{t: time.Date(2026, 11, 14, 7, 0, 0, 0, time.UTC)}
	al := &alerts{}
	uc := New(Deps{Repo: r, Book: bk, IDs: &ids{}, Clock: c, Location: time.UTC, Alerts: al})

	if _, err := uc.Add(ctx, "Anna", 2, 30, 0); err == nil {
		t.Fatal("invalid date")
	}
	a, _ := uc.Add(ctx, "Anna", 11, 20, 1990)
	vcf := "BEGIN:VCARD\nFN:Anna\nBDAY:1990-11-20\nEND:VCARD\nBEGIN:VCARD\nFN:Bernd\nBDAY:--12-01\nEND:VCARD\n"
	if rep, err := uc.ImportVCard(ctx, []byte(vcf)); err != nil || rep.Added != 1 || rep.Duplicates != 1 {
		t.Fatalf("Anna is already there: %+v %v", rep, err)
	}
	if _, err := uc.ImportVCard(ctx, []byte("no vcard")); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
	// CardDAV: tested and read when saved.
	if _, err := uc.SaveSettings(ctx, SettingsInput{CardDAVURL: "https://cloud.example/dav/contacts", Username: "u", Password: "x"}); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatal(err)
	}
	v, err := uc.SaveSettings(ctx, SettingsInput{RemindEnabled: true, RemindAt: "08:00", CardDAVURL: "https://cloud.example/dav/contacts", Username: "u", Password: "pw"})
	if err != nil || !v.HasPassword || v.CardDAVURL != "https://cloud.example/dav/contacts/" {
		t.Fatalf("%+v %v", v, err)
	}
	up, _ := uc.Upcoming(ctx, 2)
	if len(up) != 2 || up[0].Name != "Carla" || up[0].InDays != 0 || up[0].Age != 46 || up[1].Name != "Anna" {
		t.Fatalf("%+v", up)
	}
	// Sync replaces the CardDAV entries only; a failing book keeps them.
	bk.cards = []domain.Card{{UID: "u2", Name: "Dieter", Month: 1, Day: 2}}
	_ = uc.Sync(ctx)
	bk.fail = true
	if err := uc.Sync(ctx); err == nil {
		t.Fatal("error expected")
	}
	if all, _ := uc.All(ctx); len(all) != 3 {
		t.Fatalf("Anna, Bernd, Dieter: %+v", all)
	}
	// Reminder: once a day after 08:00, for today's birthdays.
	bk.fail = false
	bk.cards = []domain.Card{{UID: "u1", Name: "Carla", Month: 11, Day: 14, Year: 1980}}
	_ = uc.Sync(ctx)
	if n, _ := uc.Remind(ctx); n != 0 {
		t.Fatal("before 08:00")
	}
	c.t = c.t.Add(2 * time.Hour)
	if n, _ := uc.Remind(ctx); n != 1 || al.got[0].Title != "🎂 Carla" || al.got[0].Body != "Turns 46 today." {
		t.Fatalf("%v %+v", n, al.got)
	}
	if n, _ := uc.Remind(ctx); n != 0 {
		t.Fatal("only once a day")
	}
	if soon, _ := uc.Soon(ctx, 2); len(soon) != 1 {
		t.Fatalf("%+v", soon)
	}
	// Address book entries are removed at their source, not here.
	if err := uc.Delete(ctx, "carddav-u1"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("synced entry deleted: %v", err)
	}
	// Disconnecting the address book removes its entries.
	if _, err := uc.SaveSettings(ctx, SettingsInput{}); err != nil {
		t.Fatal(err)
	}
	if all, _ := uc.All(ctx); len(all) != 2 {
		t.Fatalf("%+v", all)
	}
	if err := uc.Delete(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
}
