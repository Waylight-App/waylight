// Package domain holds the rules of absences (#299): who is away on which
// days, and who may see why.
//
// Ring: Entities.
package domain

import (
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Kind is why someone is away.
type Kind string

// Kinds. KindAway is what others see of a private absence.
const (
	KindHoliday Kind = "holiday"
	KindSick    Kind = "sick"
	KindRemote  Kind = "remote"
	KindOther   Kind = "other"
	KindAway    Kind = "away"
)

var kinds = []Kind{KindHoliday, KindSick, KindRemote, KindOther}

// Limits.
const (
	MaxDays  = 366 // one absence spans at most a year
	MaxNote  = 200
	MaxRange = 62 // days per list request
	// DateLayout is a calendar day, without time or zone: absences are days.
	DateLayout = "2006-01-02"
)

// Absence is one person away from From to To, both days included.
type Absence struct {
	ID        string
	AccountID string
	From, To  string // DateLayout
	Kind      Kind
	Note      string
	// Private hides kind and note from everyone but the person and admins;
	// a reason can be health data.
	Private   bool
	CreatedBy string
	CreatedAt time.Time
}

// Errors.
var (
	ErrNotFound   = kernel.NewPublicError(kernel.ErrNotFound, "The absence does not exist.")
	ErrNotAllowed = kernel.NewPublicError(kernel.ErrForbidden, "You can only enter and change your own absences.")
	ErrNoAccount  = kernel.NewValidationError("accountId", "This account does not exist.")
)

// ParseDay reads a calendar day.
func ParseDay(s string) (time.Time, error) { return time.Parse(DateLayout, strings.TrimSpace(s)) }

// Validate checks the fields.
func (a *Absence) Validate() error {
	from, err := ParseDay(a.From)
	if err != nil {
		return kernel.NewValidationError("from", "Enter the first day as YYYY-MM-DD.")
	}
	to, err := ParseDay(a.To)
	if err != nil {
		return kernel.NewValidationError("to", "Enter the last day as YYYY-MM-DD.")
	}
	switch {
	case to.Before(from):
		return kernel.NewValidationError("to", "The last day must not be before the first.")
	case to.Sub(from) >= MaxDays*24*time.Hour:
		return kernel.NewValidationError("to", "An absence can span at most a year.")
	case !slices.Contains(kinds, a.Kind):
		return kernel.NewValidationError("kind", "Choose holiday, sick, remote or other.")
	}
	a.Note = strings.TrimSpace(a.Note)
	if utf8.RuneCountInString(a.Note) > MaxNote {
		return kernel.NewValidationError("note", "The note must be at most 200 characters.")
	}
	a.From, a.To = from.Format(DateLayout), to.Format(DateLayout)
	return nil
}

// Overlaps reports whether the absence touches the days from..to.
func (a Absence) Overlaps(from, to string) bool { return a.From <= to && a.To >= from }

// MayChange reports whether p may enter, edit or delete absences of account.
func MayChange(p kernel.Principal, account string) bool {
	if p.TabScoped() || p.IsToken() {
		return false
	}
	return p.IsAdmin() || (account != "" && account == p.AccountID)
}

// SeenBy returns the absence as p may see it: the person and admins see
// everything; others see a private absence only as "away", without a note.
func (a Absence) SeenBy(p kernel.Principal) Absence {
	if (p.IsAdmin() && !p.IsToken()) || (!p.TabScoped() && p.AccountID == a.AccountID) {
		return a
	}
	if a.Private {
		a.Kind, a.Note = KindAway, ""
	}
	return a
}
