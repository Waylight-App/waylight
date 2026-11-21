package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// MaxTextLength limits a note; the board is for short messages.
const MaxTextLength = 500

// ErrNotFound is returned for unknown notes and for notes the viewer may
// not see.
var ErrNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The note does not exist.")

// Note is a short message on the family board (#685).
type Note struct {
	ID     string
	Text   string
	Pinned bool
	// ExpiresOn is the last day the note is shown (midnight UTC); zero
	// keeps it until it is removed.
	ExpiresOn time.Time
	Access    kernel.Access
	CreatedAt time.Time
}

// Input is what a person enters for a note.
type Input struct {
	Text   string
	Pinned bool
	// ExpiresOn is YYYY-MM-DD or empty.
	ExpiresOn string
}

// New validates the input and creates a note. A note cannot be created
// already expired.
func New(id string, in Input, access kernel.Access, now, today time.Time) (Note, error) {
	n := Note{ID: id, Access: access, CreatedAt: now}
	if err := n.apply(in); err != nil {
		return Note{}, err
	}
	if n.ExpiredOn(today) {
		return Note{}, kernel.NewValidationError("expiresOn", "Choose today or a later day.")
	}
	return n, nil
}

// Update changes text, pin and expiry.
func (n Note) Update(in Input) (Note, error) {
	err := n.apply(in)
	return n, err
}

func (n *Note) apply(in Input) error {
	text := strings.TrimSpace(in.Text)
	if text == "" || utf8.RuneCountInString(text) > MaxTextLength {
		return kernel.NewValidationError("text", "A note must be between 1 and 500 characters long.")
	}
	var expires time.Time
	if s := strings.TrimSpace(in.ExpiresOn); s != "" {
		d, err := time.Parse(time.DateOnly, s)
		if err != nil {
			return kernel.NewValidationError("expiresOn", "Enter a date like 2026-12-24.")
		}
		expires = d
	}
	n.Text, n.Pinned, n.ExpiresOn = text, in.Pinned, expires
	return nil
}

// ExpiredOn reports whether the note is past its last day on day.
func (n Note) ExpiredOn(day time.Time) bool {
	if n.ExpiresOn.IsZero() {
		return false
	}
	return n.ExpiresOn.Before(time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC))
}

// ExpiryDate is ExpiresOn as YYYY-MM-DD, or empty.
func (n Note) ExpiryDate() string {
	if n.ExpiresOn.IsZero() {
		return ""
	}
	return n.ExpiresOn.Format(time.DateOnly)
}

// Less puts pinned notes first, then the newest.
func Less(a, b Note) bool {
	if a.Pinned != b.Pinned {
		return a.Pinned
	}
	return a.CreatedAt.After(b.CreatedAt)
}
