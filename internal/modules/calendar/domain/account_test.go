package domain

import "testing"

func TestCalendarRef(t *testing.T) {
	ref := CalendarRef("acc1", "primary:with:colons@example.org")
	a, c, ok := ParseCalendarRef(ref)
	if !ok || a != "acc1" || c != "primary:with:colons@example.org" {
		t.Fatalf("%q %q %v", a, c, ok)
	}
	for _, s := range []string{"https://dav.example/cal/", "account:", "account:x", "account::y"} {
		if _, _, ok := ParseCalendarRef(s); ok {
			t.Errorf("%q parsed", s)
		}
	}
}

func TestMergeCalendarsKeepsChoices(t *testing.T) {
	a := Account{Calendars: []AccountCalendar{{ID: "work", Enabled: false}, {ID: "home", Enabled: true}, {ID: "gone", Enabled: true}}}
	a = a.MergeCalendars([]AccountCalendar{{ID: "home"}, {ID: "work"}, {ID: "new"}})
	got := map[string]bool{}
	for _, c := range a.Calendars {
		got[c.ID] = c.Enabled
	}
	if len(got) != 3 || got["work"] || !got["home"] || !got["new"] {
		t.Fatalf("%v", got)
	}
	if e := a.WithEnabled([]string{"work"}).Enabled(); len(e) != 1 || e[0].ID != "work" {
		t.Fatalf("%+v", e)
	}
}

func TestSafeColor(t *testing.T) {
	if SafeColor("#A1B2C3") != "#a1b2c3" || SafeColor("red") != "" || SafeColor("#123;x") != "" {
		t.Fatal("colour filter")
	}
}
