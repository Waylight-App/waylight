package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

var now = time.Date(2026, 11, 23, 9, 0, 0, 0, time.UTC)

func day(s string) time.Time {
	d, _ := time.Parse(time.DateOnly, s)
	return d
}

func valid() Input {
	return Input{Name: "Mobile phone", CostCents: 1999, Period: "monthly", TermEnd: "2027-03-31", NoticeDays: 90, RenewMonths: 12, RemindDays: 14}
}

func TestValidation(t *testing.T) {
	c, err := New("c1", valid(), kernel.Access{}, now)
	if err != nil || c.Name != "Mobile phone" || c.MonthlyCents() != 1999 {
		t.Fatalf("%+v %v", c, err)
	}
	for name, change := range map[string]func(*Input){
		"no name":       func(i *Input) { i.Name = " " },
		"long provider": func(i *Input) { i.Provider = strings.Repeat("x", 81) },
		"negative cost": func(i *Input) { i.CostCents = -1 },
		"bad period":    func(i *Input) { i.Period = "weekly" },
		"no term end":   func(i *Input) { i.TermEnd = "" },
		"notice":        func(i *Input) { i.NoticeDays = 731 },
		"renew":         func(i *Input) { i.RenewMonths = 61 },
		"remind":        func(i *Input) { i.RemindDays = 91 },
		"note":          func(i *Input) { i.Note = strings.Repeat("x", 501) },
	} {
		in := valid()
		change(&in)
		if _, err := New("c", in, kernel.Access{}, now); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestMonthlyCents(t *testing.T) {
	for period, want := range map[string]int{"monthly": 12000, "quarterly": 4000, "yearly": 1000} {
		in := valid()
		in.CostCents, in.Period = 12000, period
		c, _ := New("c", in, kernel.Access{}, now)
		if got := c.MonthlyCents(); got != want {
			t.Errorf("%s: %d, want %d", period, got, want)
		}
	}
}

func TestTermOn(t *testing.T) {
	c, _ := New("c", valid(), kernel.Access{}, now) // ends 2027-03-31, 90 days notice
	term := c.TermOn(now)
	if !term.Open || term.Deadline.Format(time.DateOnly) != "2026-12-31" || term.End.Format(time.DateOnly) != "2027-03-31" {
		t.Fatalf("%+v", term)
	}
	// After the deadline has passed, the next term's deadline counts.
	term = c.TermOn(day("2027-01-01"))
	if !term.Open || term.End.Format(time.DateOnly) != "2028-03-31" || term.Deadline.Format(time.DateOnly) != "2028-01-01" {
		t.Fatalf("renewed: %+v", term)
	}
	// Several renewals at once, for a contract entered long after it began.
	old := valid()
	old.TermEnd = "2020-06-30"
	c2, _ := New("c", old, kernel.Access{}, now)
	if term := c2.TermOn(now); term.End.Format(time.DateOnly) != "2027-06-30" {
		t.Fatalf("old: %+v", term)
	}
	// Without renewal a passed deadline leaves nothing to cancel.
	fixed := valid()
	fixed.RenewMonths = 0
	c3, _ := New("c", fixed, kernel.Access{}, now)
	if term := c3.TermOn(day("2027-01-01")); term.Open {
		t.Fatalf("fixed: %+v", term)
	}
}

func TestReminder(t *testing.T) {
	c, _ := New("c", valid(), kernel.Access{}, now) // deadline 2026-12-31, remind 14 days
	if c.Soon(day("2026-12-16")) || c.ReminderDue(day("2026-12-16")) {
		t.Fatal("too early")
	}
	if !c.Soon(day("2026-12-17")) || !c.ReminderDue(day("2026-12-17")) {
		t.Fatal("reminder period starts 14 days before")
	}
	c.RemindedFor = "2026-12-31"
	if c.ReminderDue(day("2026-12-20")) || !c.Soon(day("2026-12-20")) {
		t.Fatal("reminded once per deadline")
	}
	// The next term's deadline gets its own reminder.
	if !c.ReminderDue(day("2027-12-25")) {
		t.Fatal("next deadline")
	}
	c.RemindDays = 0
	if c.Soon(day("2026-12-30")) || c.ReminderDue(day("2026-12-30")) {
		t.Fatal("no reminder when switched off")
	}
}
