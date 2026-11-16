// Package domain holds the rules of live presence (#346): who is in the
// office, at home or away today, set by each person, never guessed.
//
// Ring: Entities.
package domain

import (
	"slices"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Status is where someone works today.
type Status string

// Statuses. StatusNone means "not set": nothing is ever inferred.
const (
	StatusOffice Status = "office"
	StatusHome   Status = "home"
	StatusAway   Status = "away"
	StatusNone   Status = ""
)

// Settable lists the statuses a person can set.
var Settable = []Status{StatusOffice, StatusHome, StatusAway}

// DateLayout is a calendar day.
const DateLayout = "2006-01-02"

// Errors.
var (
	ErrStatus   = kernel.NewValidationError("status", "Choose office, home or away.")
	ErrReturnOn = kernel.NewValidationError("returnOn", "The return day must be a day after today, as YYYY-MM-DD.")
	ErrNotSelf  = kernel.NewPublicError(kernel.ErrForbidden, "Guests, kiosk displays and API tokens cannot set a status.")
	ErrDisabled = kernel.NewPublicError(kernel.ErrNotFound, "Presence is switched off for this organisation.")
	ErrAdmin    = kernel.NewPublicError(kernel.ErrForbidden, "Only admins can switch presence on or off.")
	ErrNoGroup  = kernel.NewValidationError("group", "This group does not exist.")
)

// Entry is a status someone set for one day. Entries of earlier days are
// meaningless and deleted: no history is kept.
type Entry struct {
	AccountID string
	Status    Status
	// Day the status was set for (DateLayout); it lapses after that day.
	Day string
	// ReturnOn is an optional day someone who is away is back (DateLayout).
	ReturnOn string
}

// NewEntry validates a status set by a person on day.
func NewEntry(accountID string, status Status, returnOn, day string) (Entry, error) {
	if !slices.Contains(Settable, status) {
		return Entry{}, ErrStatus
	}
	returnOn = strings.TrimSpace(returnOn)
	if returnOn != "" {
		if status != StatusAway {
			returnOn = "" // a return day only means something while away
		} else if r, err := time.Parse(DateLayout, returnOn); err != nil || r.Format(DateLayout) <= day {
			return Entry{}, ErrReturnOn
		}
	}
	return Entry{AccountID: accountID, Status: status, Day: day, ReturnOn: returnOn}, nil
}

// MaySet reports whether p may set its own status: people, not guests,
// kiosk displays or API tokens.
func MaySet(p kernel.Principal) bool {
	return p.AccountID != "" && !p.TabScoped() && !p.IsToken() && !p.IsSystem()
}

// Source says where a shown status comes from.
type Source string

// Sources.
const (
	SourceSelf    Source = "self"
	SourceAbsence Source = "absence"
	SourceNone    Source = "none"
)

// Absent is a planned absence from #299 on the day.
type Absent struct {
	Until  string
	Remote bool
}

// Person is one member as the widget shows them.
type Person struct {
	AccountID string
	Name      string
	Status    Status
	Source    Source
	ReturnOn  string
}

// Resolve decides what to show for one person on day: a planned absence
// wins over a manual status (a remote day reads "home"), a status of an
// earlier day has lapsed, and without either the status is not set.
func Resolve(accountID, name string, e *Entry, a *Absent, day string) Person {
	p := Person{AccountID: accountID, Name: name, Status: StatusNone, Source: SourceNone}
	switch {
	case a != nil && a.Remote:
		p.Status, p.Source = StatusHome, SourceAbsence
	case a != nil:
		p.Status, p.Source = StatusAway, SourceAbsence
		if a.Until > day {
			p.ReturnOn = nextDay(a.Until)
		}
	case e != nil && e.Day == day:
		p.Status, p.Source, p.ReturnOn = e.Status, SourceSelf, e.ReturnOn
	}
	return p
}

func nextDay(day string) string {
	t, err := time.Parse(DateLayout, day)
	if err != nil {
		return ""
	}
	return t.AddDate(0, 0, 1).Format(DateLayout)
}

// Counts sums a group up per status; StatusNone counts the not set.
func Counts(people []Person) map[Status]int {
	out := map[Status]int{StatusOffice: 0, StatusHome: 0, StatusAway: 0, StatusNone: 0}
	for _, p := range people {
		out[p.Status]++
	}
	return out
}

// SortPeople orders by status (office, home, away, not set), then name.
func SortPeople(people []Person) {
	rank := map[Status]int{StatusOffice: 0, StatusHome: 1, StatusAway: 2, StatusNone: 3}
	slices.SortStableFunc(people, func(a, b Person) int {
		if d := rank[a.Status] - rank[b.Status]; d != 0 {
			return d
		}
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
}
