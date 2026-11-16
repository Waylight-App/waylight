package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits of lists kept on trisse-home (#471).
const (
	maxListName = 60
	// MaxLists bounds the lists one account owns.
	MaxLists = 20
	// MaxItemsPerList bounds a list; a shopping list is not an archive.
	MaxItemsPerList = 500
	// CurrentCalDAV selects the connected CalDAV list in the widget.
	CurrentCalDAV = "caldav"
)

// Errors of lists on trisse-home.
var (
	ErrListNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The shopping list does not exist or is not shared with you.")
	ErrNotYourList  = kernel.NewPublicError(kernel.ErrForbidden, "Only the person who created the list can rename, share or delete it.")
	ErrTooManyLists = kernel.NewPublicError(kernel.ErrConflict, "You already have 20 shopping lists; delete one first.")
	ErrListFull     = kernel.NewPublicError(kernel.ErrConflict, "The list has 500 items; clear the ticked ones first.")
)

// List is a shopping list kept on trisse-home (#471): an alternative to a
// CalDAV task list that needs no server and can be shared like a note.
type List struct {
	ID         string
	Name       string
	Access     kernel.Access
	Created    time.Time
	Open, Done int // item counts, filled when listed
}

// NormaliseListName validates a list name.
func NormaliseListName(s string) (string, error) {
	s = strings.Join(strings.Fields(s), " ")
	if s == "" || utf8.RuneCountInString(s) > maxListName {
		return "", kernel.NewValidationError("name", "The name must be between 1 and 60 characters long.")
	}
	return s, nil
}

// ExportText writes the items one per line, open first, ticked ones marked.
func ExportText(name string, items []Item) string {
	var b strings.Builder
	b.WriteString(name + "\n\n")
	for _, it := range ForExport(items) {
		if it.Done {
			b.WriteString("[x] ")
		} else {
			b.WriteString("[ ] ")
		}
		b.WriteString(it.Summary + "\n")
	}
	return b.String()
}

// ForExport orders all items like the widget, without the limit on ticked
// ones.
func ForExport(items []Item) []Item {
	var open, done []Item
	for _, it := range items {
		if it.Done {
			done = append(done, it)
		} else {
			open = append(open, it)
		}
	}
	return append(ForDisplay(open), done...)
}

// icsText escapes a value for an iCalendar TEXT property (RFC 5545 3.3.11).
func icsText(s string) string {
	return strings.NewReplacer(`\`, `\\`, ";", `\;`, ",", `\,`, "\r\n", `\n`, "\n", `\n`, "\r", `\n`).Replace(s)
}

// fold breaks a content line after 75 octets (RFC 5545 3.1), never inside a
// UTF-8 character; continuation lines start with a space.
func fold(s string) string {
	var b strings.Builder
	limit := 75
	for len(s) > limit {
		cut := limit
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		b.WriteString(s[:cut] + "\r\n ")
		s = s[cut:]
		limit = 74 // the leading space counts
	}
	b.WriteString(s + "\r\n")
	return b.String()
}

func icsTime(t time.Time) string { return t.UTC().Format("20060102T150405Z") }

// ExportICS writes the items as VTODOs, so a task app can import the list.
// host makes the UIDs unique across instances.
func ExportICS(list List, items []Item, host string, now time.Time) string {
	var b strings.Builder
	line := func(s string) { b.WriteString(fold(s)) }
	line("BEGIN:VCALENDAR")
	line("VERSION:2.0")
	line("PRODID:-//trisse-home//shopping//EN")
	line("X-WR-CALNAME:" + icsText(list.Name))
	for _, it := range ForExport(items) {
		line("BEGIN:VTODO")
		line(fmt.Sprintf("UID:%s@%s", it.ID, host))
		line("DTSTAMP:" + icsTime(now))
		if !it.Created.IsZero() {
			line("CREATED:" + icsTime(it.Created))
		}
		line("SUMMARY:" + icsText(it.Summary))
		if it.Done {
			line("STATUS:COMPLETED")
			if !it.Completed.IsZero() {
				line("COMPLETED:" + icsTime(it.Completed))
			}
		} else {
			line("STATUS:NEEDS-ACTION")
		}
		line("END:VTODO")
	}
	line("END:VCALENDAR")
	return b.String()
}
