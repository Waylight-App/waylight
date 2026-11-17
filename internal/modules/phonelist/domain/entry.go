package domain

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// Limits of an entry.
const (
	MaxNameLength  = 100
	MaxRoleLength  = 100
	MaxTeamLength  = 60
	MaxPhoneLength = 40
	MaxEntries     = 1000
)

// ErrEntryNotFound is returned for unknown entries.
var ErrEntryNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The phone list entry does not exist.")

// ErrTooManyEntries is returned at the limit.
var ErrTooManyEntries = kernel.NewPublicError(kernel.ErrConflict, "The phone list holds at most 1000 entries.")

// phonePattern allows what people write in phone numbers: digits, a leading
// plus, spaces and the usual separators.
var phonePattern = regexp.MustCompile(`^\+?[0-9 ()./-]+$`)

// Entry is one person in the phone list (#300).
type Entry struct {
	ID    string
	Name  string
	Role  string
	Phone string
	Email string
	// Team is the department or group shown next to the name, such as
	// "Accounting"; free text, not an account group.
	Team      string
	CreatedAt time.Time
	// Source is manual (typed in by an admin) or carddav (read-only, from
	// the address book).
	Source Source
}

// EntryInput carries the editable fields.
type EntryInput struct {
	Name, Role, Phone, Email, Team string
}

// NewEntry validates the input and creates an entry.
func NewEntry(id string, in EntryInput, now time.Time) (Entry, error) {
	e := Entry{ID: id, CreatedAt: now, Source: SourceManual}
	if err := e.Update(in); err != nil {
		return Entry{}, err
	}
	return e, nil
}

// Update replaces the editable fields after validation. A name and at
// least a phone number or an e-mail address are required.
func (e *Entry) Update(in EntryInput) error {
	name := strings.TrimSpace(in.Name)
	if name == "" || utf8.RuneCountInString(name) > MaxNameLength {
		return kernel.NewValidationError("name", "The name must be between 1 and 100 characters long.")
	}
	role := strings.TrimSpace(in.Role)
	if utf8.RuneCountInString(role) > MaxRoleLength {
		return kernel.NewValidationError("role", "The role must be at most 100 characters long.")
	}
	team := strings.TrimSpace(in.Team)
	if utf8.RuneCountInString(team) > MaxTeamLength {
		return kernel.NewValidationError("team", "The team must be at most 60 characters long.")
	}
	phone := strings.TrimSpace(in.Phone)
	if phone != "" && (len(phone) > MaxPhoneLength || !phonePattern.MatchString(phone) || DialNumber(phone) == "") {
		return kernel.NewValidationError("phone", "Please enter a phone number with digits, spaces and + ( ) / . - only.")
	}
	email := strings.TrimSpace(in.Email)
	if email != "" {
		a, err := mail.ParseAddress(email)
		if err != nil || a.Name != "" || a.Address != email {
			return kernel.NewValidationError("email", "Please enter a valid e-mail address.")
		}
	}
	if phone == "" && email == "" {
		return kernel.NewValidationError("phone", "Enter a phone number or an e-mail address.")
	}
	e.Name, e.Role, e.Team, e.Phone, e.Email = name, role, team, phone, email
	return nil
}

// DialNumber is the number for a tel: link: digits and a leading plus.
func DialNumber(phone string) string {
	var b strings.Builder
	for i, r := range strings.TrimSpace(phone) {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '+' && i == 0:
			b.WriteRune(r)
		}
	}
	if s := b.String(); s != "" && s != "+" {
		return s
	}
	return ""
}
