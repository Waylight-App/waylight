// Package domain holds the birthday rules (#287).
//
// Ring: Entities.
package domain

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// Sources of an entry.
const (
	SourceManual  = "manual"
	SourceVCard   = "vcard"
	SourceCardDAV = "carddav"
)

// MaxEntries limits the list.
const MaxEntries = 2000

// Errors.
var (
	ErrNotFound    = kernel.NewPublicError(kernel.ErrNotFound, "The birthday does not exist (any more).")
	ErrInvalidDate = kernel.NewValidationError("date", "Enter a valid date (day and month; the year is optional).")
	ErrName        = kernel.NewValidationError("name", "The name must have between 1 and 100 characters.")
	ErrTooMany     = kernel.NewValidationError("name", "At most 2000 birthdays can be stored.")
)

// Birthday is a person's day of birth; Year is 0 when unknown.
type Birthday struct {
	ID     string
	Name   string
	Month  time.Month
	Day    int
	Year   int
	Source string
}

// NewBirthday validates an entry (29 February is allowed).
func NewBirthday(id, name string, month, day, year int, source string) (Birthday, error) {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" || utf8.RuneCountInString(name) > 100 {
		return Birthday{}, ErrName
	}
	if month < 1 || month > 12 || day < 1 || year < 0 || year > 9999 {
		return Birthday{}, ErrInvalidDate
	}
	ref := year
	if ref == 0 {
		ref = 2000 // a leap year, so 29 February is possible
	}
	if time.Date(ref, time.Month(month), day, 0, 0, 0, 0, time.UTC).Day() != day {
		return Birthday{}, ErrInvalidDate
	}
	return Birthday{ID: id, Name: name, Month: time.Month(month), Day: day, Year: year, Source: source}, nil
}

// Upcoming is the next occurrence of a birthday.
type Upcoming struct {
	Birthday
	Date   time.Time // next birthday (today included), midnight in loc
	InDays int
	Age    int // age reached on Date; 0 when the year is unknown
}

// on returns the birthday in year y; 29 February falls on 28 February in
// common years.
func (b Birthday) on(y int, loc *time.Location) time.Time {
	d := b.Day
	if b.Month == time.February && d == 29 && time.Date(y, 2, 29, 0, 0, 0, 0, loc).Day() != 29 {
		d = 28
	}
	return time.Date(y, b.Month, d, 0, 0, 0, 0, loc)
}

// Next returns the coming birthdays from today, soonest first.
func Next(list []Birthday, now time.Time, loc *time.Location) []Upcoming {
	today := time.Date(now.In(loc).Year(), now.In(loc).Month(), now.In(loc).Day(), 0, 0, 0, 0, loc)
	out := make([]Upcoming, 0, len(list))
	for _, b := range list {
		d := b.on(today.Year(), loc)
		if d.Before(today) {
			d = b.on(today.Year()+1, loc)
		}
		u := Upcoming{Birthday: b, Date: d, InDays: int(d.Sub(today).Hours()/24 + 0.5)}
		if b.Year > 0 && d.Year() > b.Year {
			u.Age = d.Year() - b.Year
		}
		out = append(out, u)
	}
	slices.SortStableFunc(out, func(a, b Upcoming) int {
		if a.InDays != b.InDays {
			return a.InDays - b.InDays
		}
		return strings.Compare(a.Name, b.Name)
	})
	return out
}

var (
	bdayFull    = regexp.MustCompile(`^(\d{4})-?(\d{2})-?(\d{2})`)
	bdayNoYear  = regexp.MustCompile(`^--(\d{2})-?(\d{2})`)
	unfoldLines = strings.NewReplacer("\r\n ", "", "\r\n\t", "", "\n ", "", "\n\t", "")
)

// ParseBDAY reads a vCard BDAY value: 1990-05-17, 19900517, --05-17,
// --0517, also with a time part. ok is false for text values.
func ParseBDAY(v string) (month, day, year int, ok bool) {
	v = strings.TrimSpace(v)
	if m := bdayFull.FindStringSubmatch(v); m != nil {
		year, _ = strconv.Atoi(m[1])
		month, _ = strconv.Atoi(m[2])
		day, _ = strconv.Atoi(m[3])
		if year == 1604 { // Apple's marker for "year unknown"
			year = 0
		}
		return month, day, year, true
	}
	if m := bdayNoYear.FindStringSubmatch(v); m != nil {
		month, _ = strconv.Atoi(m[1])
		day, _ = strconv.Atoi(m[2])
		return month, day, 0, true
	}
	return 0, 0, 0, false
}

// Card is a contact with a birthday found in vCard data.
type Card struct {
	UID   string
	Name  string
	Month int
	Day   int
	Year  int
}

// ParseVCards finds the contacts with a birthday in vCard text (one or
// many cards); cards without a usable BDAY or name are skipped.
func ParseVCards(data string) []Card {
	var out []Card
	var cur *Card
	for _, line := range strings.Split(unfoldLines.Replace(data), "\n") {
		line = strings.TrimRight(line, "\r")
		name, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		prop := strings.ToUpper(name)
		if i := strings.IndexByte(prop, ';'); i >= 0 {
			prop = prop[:i]
		}
		if j := strings.IndexByte(prop, '.'); j >= 0 { // grouped property, e.g. item1.BDAY
			prop = prop[j+1:]
		}
		switch {
		case prop == "BEGIN" && strings.EqualFold(value, "VCARD"):
			cur = &Card{}
		case prop == "END" && strings.EqualFold(value, "VCARD") && cur != nil:
			if cur.Name != "" && cur.Month > 0 {
				out = append(out, *cur)
			}
			cur = nil
		case cur == nil:
		case prop == "FN":
			cur.Name = unescape(value)
		case prop == "UID":
			cur.UID = value
		case prop == "BDAY" && !strings.Contains(strings.ToUpper(name), "VALUE=TEXT"):
			if m, d, y, ok := ParseBDAY(value); ok {
				cur.Month, cur.Day, cur.Year = m, d, y
			}
		}
	}
	return out
}

func unescape(s string) string {
	return strings.TrimSpace(strings.NewReplacer(`\,`, ",", `\;`, ";", `\n`, " ", `\N`, " ", `\\`, `\`).Replace(s))
}

// Reminder says whether and when a reminder is pushed on the day.
type Reminder struct {
	Enabled bool
	At      string // "HH:MM"
}

var hhmm = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// NewReminder validates the reminder time; empty means 08:00.
func NewReminder(enabled bool, at string) (Reminder, error) {
	if at == "" {
		at = "08:00"
	}
	if !hhmm.MatchString(at) {
		return Reminder{}, kernel.NewValidationError("remindAt", "Enter the time as HH:MM.")
	}
	return Reminder{Enabled: enabled, At: at}, nil
}

// Due reports whether the reminder of the day should go out now: enabled,
// the time has come, and it was not sent today yet.
func (r Reminder) Due(now time.Time, lastSent string, loc *time.Location) bool {
	if !r.Enabled {
		return false
	}
	local := now.In(loc)
	return local.Format("15:04") >= r.At && lastSent != local.Format(time.DateOnly)
}
