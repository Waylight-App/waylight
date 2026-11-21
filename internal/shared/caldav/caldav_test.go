package caldav

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"waylight/internal/shared/caldav/caldavtest"
)

func setup(t *testing.T) (*caldavtest.Server, Client, Account) {
	srv := caldavtest.New(t)
	return srv, Client{HTTP: http.DefaultClient}, Account{URL: srv.ListURL(), Username: srv.User, Password: srv.Password}
}

const todo = "BEGIN:VCALENDAR\r\nBEGIN:VTODO\r\nUID:x\r\nSUMMARY:A\r\nEND:VTODO\r\nEND:VCALENDAR\r\n"

func TestDiscover(t *testing.T) {
	srv, c, _ := setup(t)
	for _, start := range []string{srv.URL + caldavtest.Root, srv.URL + caldavtest.Home, srv.ListURL()} {
		got, err := c.Discover(context.Background(), Account{URL: start, Username: srv.User, Password: srv.Password})
		if err != nil || len(got) != 1 || got[0].URL != srv.ListURL() || got[0].Name != "Aufgaben" {
			t.Fatalf("from %s: %+v %v", start, got, err)
		}
	}
	if _, err := c.Discover(context.Background(), Account{URL: srv.URL + caldavtest.Root, Username: "anna", Password: "wrong"}); !errors.Is(err, ErrUnauthorized) {
		t.Fatal(err)
	}
	// The same account holds an event calendar next to the task list.
	events, err := c.DiscoverEvents(context.Background(), Account{URL: srv.URL + caldavtest.Root, Username: srv.User, Password: srv.Password})
	if err != nil || len(events) != 1 || events[0].Name != "Persönlich" {
		t.Fatalf("events: %+v %v", events, err)
	}
}

func TestObjects(t *testing.T) {
	srv, c, a := setup(t)
	ctx := context.Background()
	ctag, err := c.CTag(ctx, a)
	if err != nil {
		t.Fatal(err)
	}
	etag, err := c.Put(ctx, a, "a b.ics", todo, "", true)
	if err != nil || etag == "" {
		t.Fatal(etag, err)
	}
	if _, err := c.Put(ctx, a, "a b.ics", todo, "", true); !errors.Is(err, ErrPreconditionFailed) {
		t.Fatal("If-None-Match must not overwrite", err)
	}
	if next, _ := c.CTag(ctx, a); next == ctag {
		t.Fatal("ctag did not change")
	}
	srv.Put("event.ics", "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n")
	list, err := c.ETags(ctx, a)
	if err != nil || len(list) != 1 || list[0].Name != "a b.ics" || list[0].ETag != etag || list[0].Data != "" {
		t.Fatalf("%+v %v", list, err)
	}
	got, err := c.Multiget(ctx, a, []string{"a b.ics"})
	if err != nil || len(got) != 1 || !strings.Contains(got[0].Data, "SUMMARY:A") {
		t.Fatalf("%+v %v", got, err)
	}
	full, err := c.Todos(ctx, a)
	if err != nil || len(full) != 1 || full[0].Data == "" {
		t.Fatalf("%+v %v", full, err)
	}
	o, err := c.Get(ctx, a, "a b.ics")
	if err != nil || o.ETag != etag {
		t.Fatal(o, err)
	}
	srv.Put("a b.ics", todo) // another app changes it
	if _, err := c.Put(ctx, a, "a b.ics", todo, etag, false); !errors.Is(err, ErrPreconditionFailed) {
		t.Fatal("If-Match must detect the change", err)
	}
	if err := c.Delete(ctx, a, "a b.ics"); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete(ctx, a, "a b.ics"); err != nil {
		t.Fatal("deleting a missing object is fine", err)
	}
	if _, err := c.Get(ctx, a, "a b.ics"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	srv.NoCTag = true
	if tag, err := c.CTag(ctx, a); err != nil || tag != "" {
		t.Fatal(tag, err)
	}
}
