package domain

import (
	"errors"
	"testing"

	"waylight/internal/shared/kernel"
)

const today = "2026-11-16"

func TestNewEntry(t *testing.T) {
	if e, err := NewEntry("a", StatusAway, "2026-11-18", today); err != nil || e.ReturnOn != "2026-11-18" {
		t.Fatalf("%+v %v", e, err)
	}
	if e, _ := NewEntry("a", StatusHome, "2026-11-18", today); e.ReturnOn != "" {
		t.Fatal("a return day only belongs to away")
	}
	for _, c := range []struct {
		s Status
		r string
	}{{"beach", ""}, {StatusNone, ""}, {StatusAway, today}, {StatusAway, "18.11."}} {
		if _, err := NewEntry("a", c.s, c.r, today); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%+v accepted", c)
		}
	}
}

func TestMaySet(t *testing.T) {
	if !MaySet(kernel.Principal{AccountID: "a", Role: "member"}) || MaySet(kernel.Principal{AccountID: "g", Role: "guest"}) ||
		MaySet(kernel.Principal{AccountID: "k", Role: "kiosk"}) || MaySet(kernel.Principal{AccountID: "t", Role: "token"}) {
		t.Fatal("only people set their own status")
	}
}

func TestResolve(t *testing.T) {
	set := &Entry{Status: StatusOffice, Day: today}
	old := &Entry{Status: StatusOffice, Day: "2026-11-13"}
	cases := []struct {
		name   string
		e      *Entry
		a      *Absent
		status Status
		source Source
		ret    string
	}{
		{"own status", set, nil, StatusOffice, SourceSelf, ""},
		{"lapsed", old, nil, StatusNone, SourceNone, ""},
		{"nothing", nil, nil, StatusNone, SourceNone, ""},
		{"absence wins", set, &Absent{Until: "2026-11-20"}, StatusAway, SourceAbsence, "2026-11-21"},
		{"last day of absence", nil, &Absent{Until: today}, StatusAway, SourceAbsence, ""},
		{"remote day", set, &Absent{Until: today, Remote: true}, StatusHome, SourceAbsence, ""},
	}
	for _, c := range cases {
		p := Resolve("a", "Anna", c.e, c.a, today)
		if p.Status != c.status || p.Source != c.source || p.ReturnOn != c.ret {
			t.Errorf("%s: %+v", c.name, p)
		}
	}
}

func TestCountsAndSort(t *testing.T) {
	people := []Person{{Name: "zoe", Status: StatusNone}, {Name: "Ben", Status: StatusHome}, {Name: "anna", Status: StatusHome}, {Name: "x", Status: StatusOffice}}
	SortPeople(people)
	if people[0].Name != "x" || people[1].Name != "anna" || people[3].Name != "zoe" {
		t.Fatalf("order: %+v", people)
	}
	c := Counts(people)
	if c[StatusHome] != 2 || c[StatusNone] != 1 || c[StatusAway] != 0 {
		t.Fatalf("counts: %v", c)
	}
}
