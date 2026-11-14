package gateway

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/caldav/caldavtest"
)

// A to-do as Nextcloud Tasks stores it (anonymised recording), with
// properties trisse-home does not know.
const nextcloudTodo = "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Nextcloud Tasks v0.16.1\r\nBEGIN:VTODO\r\n" +
	"UID:b6f1c2d0-1111-4a5b-9c1d-2e3f4a5b6c7d\r\nCREATED:20261101T080000Z\r\nLAST-MODIFIED:20261101T080000Z\r\nDTSTAMP:20261101T080000Z\r\n" +
	"SUMMARY:Steuererklärung\r\nPRIORITY:5\r\nX-APPLE-SORT-ORDER:12\r\nDUE;VALUE=DATE:20261130\r\nSTATUS:NEEDS-ACTION\r\n" +
	"BEGIN:VALARM\r\nACTION:DISPLAY\r\nTRIGGER:-PT1H\r\nEND:VALARM\r\nEND:VTODO\r\nEND:VCALENDAR\r\n"

func setup(t *testing.T) (*caldavtest.Server, *CalDAVProviders, domain.Source) {
	srv := caldavtest.New(t)
	p := NewCalDAVProviders(http.DefaultClient)
	p.now = func() time.Time { return time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC) }
	return srv, p, domain.Source{ID: "nc", Kind: domain.KindCalDAV, URL: srv.ListURL(), Username: srv.User, Password: srv.Password}
}

func TestCalDAVListAndCache(t *testing.T) {
	ctx := context.Background()
	srv, p, s := setup(t)
	srv.Put("steuer.ics", nextcloudTodo)
	srv.Put("later.ics", strings.NewReplacer("Steuererklärung", "Später", "DUE;VALUE=DATE:20261130\r\n", "", "b6f1", "c7a2").Replace(nextcloudTodo))
	srv.Put("event.ics", "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:e\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n")

	todos, err := p.Open(s).List(ctx)
	if err != nil || len(todos) != 2 {
		t.Fatalf("%+v %v", todos, err)
	}
	first := todos[0]
	if first.ID() != "steuer.ics" || first.Title() != "Steuererklärung" || !first.Due().Equal(time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)) ||
		first.Position() != 0 || first.Version() == 0 || todos[1].Title() != "Später" {
		t.Fatalf("dated first: %+v", todos)
	}
	// Unchanged CTag: nothing but the CTag is asked for.
	before := srv.Count("REPORT etags") + srv.Count("REPORT multiget")
	if _, err := p.Open(s).List(ctx); err != nil || srv.Count("REPORT etags")+srv.Count("REPORT multiget") != before {
		t.Fatal("refetched an unchanged list", err)
	}
	// A change in another app: only that object is downloaded again.
	srv.Put("later.ics", strings.Replace(srv.Data("later.ics"), "SUMMARY:Später", "SUMMARY:Gleich", 1))
	multigets := srv.Count("REPORT multiget")
	todos, _ = p.Open(s).List(ctx)
	if srv.Count("REPORT multiget") != multigets+1 || todos[1].Title() != "Gleich" {
		t.Fatalf("change not picked up: %+v", todos)
	}
	// Servers without CTag compare ETags on every refresh.
	srv.NoCTag = true
	srv.Put("steuer.ics", strings.Replace(nextcloudTodo, "STATUS:NEEDS-ACTION", "STATUS:COMPLETED", 1))
	if todos, _ = p.Open(s).List(ctx); !todos[0].Done() && !todos[1].Done() {
		t.Fatalf("done in the other app: %+v", todos)
	}
}

func TestCalDAVWrites(t *testing.T) {
	ctx := context.Background()
	srv, p, s := setup(t)
	srv.Put("steuer.ics", nextcloudTodo)
	l := p.Open(s)

	// Edit: only title, due date and completion change.
	got, err := l.Get(ctx, "steuer.ics")
	if err != nil {
		t.Fatal(err)
	}
	if err := got.Edit("Steuer erklären", got.Due().AddDate(0, 0, 1)); err != nil {
		t.Fatal(err)
	}
	got.SetDone(true, p.now())
	if err := l.Save(ctx, got); err != nil {
		t.Fatal(err)
	}
	data := srv.Data("steuer.ics")
	for _, want := range []string{"SUMMARY:Steuer erklären", "DUE;VALUE=DATE:20261201", "STATUS:COMPLETED", "PRIORITY:5", "X-APPLE-SORT-ORDER:12", "BEGIN:VALARM"} {
		if !strings.Contains(data, want) {
			t.Fatalf("missing %q:\n%s", want, data)
		}
	}
	// Server wins: a write based on an old read fails with a conflict.
	stale := p.Open(s)
	old, _ := stale.Get(ctx, "steuer.ics")
	srv.Put("steuer.ics", strings.Replace(data, "PRIORITY:5", "PRIORITY:1", 1))
	_ = old.Edit("Mein Titel", old.Due())
	if err := stale.Save(ctx, old); !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatal("overwrote a change of the other app", err)
	}
	if fresh, _ := p.Open(s).Get(ctx, "steuer.ics"); fresh.Version() == old.Version() {
		t.Fatal("the version must follow the ETag")
	}

	// Create with a client-generated ID; Get finds it with or without ".ics".
	n, _ := domain.NewTodo("0b7e9d2c-4f1a-4c8e-9a51-2d3c4b5a6f70", "Brief", time.Time{}, 0, p.now())
	if err := l.Save(ctx, n); err != nil {
		t.Fatal(err)
	}
	created, err := l.Get(ctx, "0b7e9d2c-4f1a-4c8e-9a51-2d3c4b5a6f70")
	if err != nil || created.ID() != "0b7e9d2c-4f1a-4c8e-9a51-2d3c4b5a6f70.ics" || created.Title() != "Brief" {
		t.Fatalf("%+v %v", created, err)
	}
	if err := p.Open(s).Save(ctx, n); !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatal("a create must never overwrite", err)
	}
	if err := l.Delete(ctx, created.ID()); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Get(ctx, created.ID()); !errors.Is(err, domain.ErrTodoNotFound) {
		t.Fatal(err)
	}
	if todos, _ := p.Open(s).List(ctx); len(todos) != 1 {
		t.Fatalf("list after writes: %+v", todos)
	}
}

func TestCalDAVDiscoverAndLogin(t *testing.T) {
	ctx := context.Background()
	srv, p, s := setup(t)
	lists, err := p.Discover(ctx, domain.KindCalDAV, srv.URL+caldavtest.Root, srv.User, srv.Password)
	if err != nil || len(lists) != 1 || lists[0].URL != srv.ListURL() || lists[0].Name != "Aufgaben" {
		t.Fatalf("%+v %v", lists, err)
	}
	s.Password = "wrong"
	if _, err := p.Open(s).List(ctx); err == nil {
		t.Fatal("wrong password accepted")
	}
}

func TestSourceRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewSourceRepository(dbtest.Open(t, "todos", Migrations()), prefixCipher{})
	s := domain.Source{ID: "nc", Name: "Nextcloud", Kind: domain.KindCalDAV, URL: "https://cloud.example/tasks/", Username: "anna", Password: "geheim"}
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	var enc string
	_ = repo.db.QueryRowContext(ctx, `SELECT password_encrypted FROM todos_sources`).Scan(&enc)
	if enc != "enc:geheim" {
		t.Fatal("password stored in plain text:", enc)
	}
	s.Name = "Cloud"
	_ = repo.Save(ctx, s)
	if got, err := repo.Get(ctx, "nc"); err != nil || got != s {
		t.Fatalf("%+v %v", got, err)
	}
	if all, _ := repo.List(ctx); len(all) != 1 {
		t.Fatal(all)
	}
	_ = repo.Delete(ctx, "nc")
	if _, err := repo.Get(ctx, "nc"); !errors.Is(err, domain.ErrSourceNotFound) {
		t.Fatal(err)
	}
}

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }
