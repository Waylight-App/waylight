package domain

import (
	"testing"
	"time"
)

func TestTodo(t *testing.T) {
	now := time.Date(2026, 11, 10, 15, 0, 0, 0, time.UTC)
	if _, err := NewTodo("1", "  ", time.Time{}, 0, now); err == nil {
		t.Fatal("expected error for blank title")
	}
	due := time.Date(2026, 11, 9, 23, 30, 0, 0, time.UTC)
	todo, err := NewTodo("1", "Pay rent", due, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if todo.Due().Hour() != 0 {
		t.Fatal("due date not truncated to the day")
	}
	if !todo.IsOverdue(now) {
		t.Fatal("expected overdue")
	}
	todo.SetDone(true, now)
	if !todo.Done() || todo.IsOverdue(now) {
		t.Fatal("done to-dos are never overdue")
	}
	todo.SetDone(true, now.Add(time.Hour))
	if !todo.CompletedAt().Equal(now) {
		t.Fatal("completion time must not change when marking done twice")
	}
	todo.SetDone(false, now)
	if todo.Done() {
		t.Fatal("reopen failed")
	}
}

func TestSources(t *testing.T) {
	s, err := NewSource("s1", "  Nextcloud   Aufgaben ", KindCalDAV, "https://cloud.example/remote.php/dav/calendars/anna/tasks", " anna ", "pw")
	if err != nil || s.Name != "Nextcloud Aufgaben" || s.URL != "https://cloud.example/remote.php/dav/calendars/anna/tasks/" || s.Username != "anna" {
		t.Fatalf("%+v %v", s, err)
	}
	for _, bad := range [][5]string{
		{"", "caldav", "https://a.example/x", "u", "p"},
		{"n", "todoist", "https://a.example/x", "u", "p"},
		{"n", "caldav", "ftp://a.example/x", "u", "p"},
		{"n", "caldav", "https://a.example/x", "", "p"},
		{"n", "caldav", "https://a.example/x", "u", ""},
	} {
		if _, err := NewSource("s", bad[0], SourceKind(bad[1]), bad[2], bad[3], bad[4]); err == nil {
			t.Errorf("accepted %v", bad)
		}
	}
	if !ValidSourceID("abc_1-2") || ValidSourceID(LocalSource) || ValidSourceID("a/b") || ValidSourceID("") {
		t.Fatal("source IDs")
	}
	if !ValidExternalID("a b@c.ics") || ValidExternalID("../x") || ValidExternalID("..") || ValidExternalID("a/b.ics") {
		t.Fatal("external IDs")
	}
	if c := CapabilitiesOf(""); !c.Repeat || !c.Ordering || !c.Reminders {
		t.Fatal(c)
	}
	if c := CapabilitiesOf(KindCalDAV); c.Repeat || c.Ordering {
		t.Fatal(c)
	}
}
