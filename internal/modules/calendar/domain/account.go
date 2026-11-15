package domain

import (
	"regexp"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Provider is a calendar service connected by signing in (#282).
type Provider string

// Providers that can be connected.
const (
	ProviderGoogle    Provider = "google"
	ProviderMicrosoft Provider = "microsoft"
)

// Valid reports whether p is a known provider.
func (p Provider) Valid() bool { return p == ProviderGoogle || p == ProviderMicrosoft }

// Label is the provider's name as shown to people.
func (p Provider) Label() string {
	if p == ProviderMicrosoft {
		return "Microsoft"
	}
	return "Google"
}

// MaxAccounts limits connected accounts; each one costs requests on every refresh.
const MaxAccounts = 5

// Errors of connected accounts.
var (
	ErrUnknownProvider = kernel.NewValidationError("provider", "Choose Google or Microsoft.")
	ErrAccountNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The account is not connected (any more).")
	ErrTooManyAccounts = kernel.NewPublicError(kernel.ErrConflict, "At most 5 accounts can be connected.")
	ErrSignInAgain     = kernel.NewPublicError(kernel.ErrUnavailable, "The account needs to be signed in again.")
)

// Grant holds the tokens of an account; ClientSecret belongs to the user's
// app registration and is kept with them because refreshing needs it.
type Grant struct {
	ClientSecret string
	Access       string
	Refresh      string
	Expiry       time.Time
}

// AccountCalendar is one calendar of a connected account.
type AccountCalendar struct {
	ID       string
	Name     string
	Color    string // #rrggbb or ""
	Writable bool
	Enabled  bool
}

// Account is a Google or Microsoft account connected with the user's own
// app registration (ClientID).
type Account struct {
	ID        string
	Provider  Provider
	ClientID  string
	Name      string // e-mail address or display name of the account
	Grant     Grant
	Calendars []AccountCalendar
}

// Enabled returns the calendars shown in the agenda.
func (a Account) Enabled() []AccountCalendar {
	var out []AccountCalendar
	for _, c := range a.Calendars {
		if c.Enabled {
			out = append(out, c)
		}
	}
	return out
}

// WithEnabled switches the calendars on whose IDs are listed and all
// others off; unknown IDs are ignored.
func (a Account) WithEnabled(ids []string) Account {
	on := map[string]bool{}
	for _, id := range ids {
		on[id] = true
	}
	cals := make([]AccountCalendar, len(a.Calendars))
	for i, c := range a.Calendars {
		c.Enabled = on[c.ID]
		cals[i] = c
	}
	a.Calendars = cals
	return a
}

// MergeCalendars takes the calendars the provider reports now and keeps the
// on/off choice of those that existed before; new ones start switched on.
func (a Account) MergeCalendars(fresh []AccountCalendar) Account {
	was := map[string]bool{}
	known := map[string]bool{}
	for _, c := range a.Calendars {
		known[c.ID] = true
		was[c.ID] = c.Enabled
	}
	out := make([]AccountCalendar, 0, len(fresh))
	for _, c := range fresh {
		c.Enabled = !known[c.ID] || was[c.ID]
		out = append(out, c)
	}
	a.Calendars = out
	return a
}

// accountCalendarRef is how a calendar of an account is chosen when an event
// is created, next to plain CalDAV collection URLs.
const accountCalendarRef = "account:"

// CalendarRef is the value that selects an account calendar for new events.
func CalendarRef(accountID, calendarID string) string {
	return accountCalendarRef + accountID + ":" + calendarID
}

// ParseCalendarRef splits a CalendarRef; ok is false for CalDAV URLs.
func ParseCalendarRef(ref string) (accountID, calendarID string, ok bool) {
	rest, found := strings.CutPrefix(ref, accountCalendarRef)
	if !found {
		return "", "", false
	}
	accountID, calendarID, ok = strings.Cut(rest, ":")
	return accountID, calendarID, ok && accountID != "" && calendarID != ""
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// SafeColor keeps a provider colour only when it is a plain #rrggbb value,
// since it ends up in a style attribute.
func SafeColor(c string) string {
	if hexColor.MatchString(c) {
		return strings.ToLower(c)
	}
	return ""
}
