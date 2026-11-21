package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// Limits of a contract.
const (
	MaxNameLength     = 80
	MaxNoteLength     = 500
	MaxNoticeDays     = 730
	MaxRenewMonths    = 60
	MaxRemindDays     = 90
	MaxCostCents      = 100_000_000 // one million in the currency's main unit
	DefaultRemindDays = 14
)

// Period is how often the cost is due.
type Period string

// The periods a cost can have.
const (
	Monthly   Period = "monthly"
	Quarterly Period = "quarterly"
	Yearly    Period = "yearly"
)

// months is the length of the period in months; 0 for an unknown one.
func (p Period) months() int {
	switch p {
	case Monthly:
		return 1
	case Quarterly:
		return 3
	case Yearly:
		return 12
	}
	return 0
}

// ErrNotFound is returned for unknown contracts and for contracts the viewer
// may not see.
var ErrNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The contract does not exist.")

// Contract is a household contract with a cancellation deadline (#687).
type Contract struct {
	ID       string
	Name     string
	Provider string
	// CostCents is the cost per Period in the smallest currency unit.
	CostCents int
	Period    Period
	// TermEnd is the last day of the term as entered (midnight UTC).
	TermEnd time.Time
	// NoticeDays is the cancellation period before the end of a term.
	NoticeDays int
	// RenewMonths extends a term that was not cancelled; 0 means the
	// contract simply ends.
	RenewMonths int
	// RemindDays is how long before the deadline the reminder is sent;
	// 0 sends none.
	RemindDays int
	Note       string
	// RemindedFor is the deadline (YYYY-MM-DD) the last reminder was
	// sent for, so each deadline is reminded once.
	RemindedFor string
	Access      kernel.Access
	CreatedAt   time.Time
}

// Input is what a person enters for a contract.
type Input struct {
	Name        string
	Provider    string
	CostCents   int
	Period      string
	TermEnd     string // YYYY-MM-DD
	NoticeDays  int
	RenewMonths int
	RemindDays  int
	Note        string
}

// New validates the input and creates a contract.
func New(id string, in Input, access kernel.Access, now time.Time) (Contract, error) {
	c := Contract{ID: id, Access: access, CreatedAt: now}
	if err := c.apply(in); err != nil {
		return Contract{}, err
	}
	return c, nil
}

// Update changes everything but the owner. A changed term or notice period
// may move the deadline, which then gets a reminder of its own.
func (c Contract) Update(in Input) (Contract, error) {
	err := c.apply(in)
	return c, err
}

func (c *Contract) apply(in Input) error {
	name := strings.TrimSpace(in.Name)
	if name == "" || utf8.RuneCountInString(name) > MaxNameLength {
		return kernel.NewValidationError("name", "The name must be between 1 and 80 characters long.")
	}
	provider := strings.TrimSpace(in.Provider)
	if utf8.RuneCountInString(provider) > MaxNameLength {
		return kernel.NewValidationError("provider", "The provider may be at most 80 characters long.")
	}
	if in.CostCents < 0 || in.CostCents > MaxCostCents {
		return kernel.NewValidationError("costCents", "Enter a cost between 0 and 1,000,000.")
	}
	period := Period(in.Period)
	if period.months() == 0 {
		return kernel.NewValidationError("period", "The period must be monthly, quarterly or yearly.")
	}
	end, err := time.Parse(time.DateOnly, strings.TrimSpace(in.TermEnd))
	if err != nil {
		return kernel.NewValidationError("termEnd", "Enter a date like 2026-12-24.")
	}
	if in.NoticeDays < 0 || in.NoticeDays > MaxNoticeDays {
		return kernel.NewValidationError("noticeDays", "The notice period must be between 0 and 730 days.")
	}
	if in.RenewMonths < 0 || in.RenewMonths > MaxRenewMonths {
		return kernel.NewValidationError("renewMonths", "The renewal must be between 0 and 60 months.")
	}
	if in.RemindDays < 0 || in.RemindDays > MaxRemindDays {
		return kernel.NewValidationError("remindDays", "Remind between 0 and 90 days before the deadline.")
	}
	note := strings.TrimSpace(in.Note)
	if utf8.RuneCountInString(note) > MaxNoteLength {
		return kernel.NewValidationError("note", "The note may be at most 500 characters long.")
	}
	c.Name, c.Provider, c.CostCents, c.Period, c.TermEnd = name, provider, in.CostCents, period, end
	c.NoticeDays, c.RenewMonths, c.RemindDays, c.Note = in.NoticeDays, in.RenewMonths, in.RemindDays, note
	return nil
}

// MonthlyCents is the cost spread over one month, rounded to whole cents.
func (c Contract) MonthlyCents() int {
	m := c.Period.months()
	if m == 0 {
		return 0
	}
	return (c.CostCents + m/2) / m
}

// Term is a contract's state on one day.
type Term struct {
	// End is the end of the term the next deadline belongs to.
	End time.Time
	// Deadline is the last day to cancel for End.
	Deadline time.Time
	// Open is false when the contract ends without renewal and its last
	// deadline has passed: there is nothing left to cancel.
	Open bool
}

// TermOn returns the term whose deadline is next on day. A term whose
// deadline has passed renews by RenewMonths, as often as needed.
func (c Contract) TermOn(day time.Time) Term {
	today := dateOf(day)
	end := c.TermEnd
	deadline := end.AddDate(0, 0, -c.NoticeDays)
	for deadline.Before(today) {
		if c.RenewMonths == 0 {
			return Term{End: end, Deadline: deadline, Open: false}
		}
		end = end.AddDate(0, c.RenewMonths, 0)
		deadline = end.AddDate(0, 0, -c.NoticeDays)
	}
	return Term{End: end, Deadline: deadline, Open: true}
}

// RemindFrom is the first day of the reminder period of a term.
func (c Contract) RemindFrom(t Term) time.Time { return t.Deadline.AddDate(0, 0, -c.RemindDays) }

// Soon reports whether day is inside the reminder period of the next
// deadline.
func (c Contract) Soon(day time.Time) bool {
	t := c.TermOn(day)
	return t.Open && c.RemindDays > 0 && !dateOf(day).Before(c.RemindFrom(t))
}

// ReminderDue reports whether a reminder should go out on day: inside the
// reminder period and not yet sent for this deadline.
func (c Contract) ReminderDue(day time.Time) bool {
	return c.Soon(day) && c.RemindedFor != c.TermOn(day).Deadline.Format(time.DateOnly)
}

// TermEndDate is the entered end of term as YYYY-MM-DD.
func (c Contract) TermEndDate() string { return c.TermEnd.Format(time.DateOnly) }

func dateOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
