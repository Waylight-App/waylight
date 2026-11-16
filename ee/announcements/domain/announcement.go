// Package domain holds the rules of announcements (#297): short notices
// written by admins or a chosen group and read by everyone.
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

// Limits keep an announcement a notice, not a document.
const (
	MaxTitle = 120
	MaxBody  = 10000
	// MaxActive bounds the announcements shown at once; older ones need to
	// expire or be deleted first.
	MaxActive = 50
)

// Announcement is one notice.
type Announcement struct {
	ID       string
	Title    string
	Body     string // Markdown, rendered by the client
	Pinned   bool
	AuthorID string
	// ExpiresAt is when the notice leaves the widget for the archive; zero
	// means it stays until deleted.
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Errors.
var (
	ErrNotFound   = kernel.NewPublicError(kernel.ErrNotFound, "The announcement does not exist.")
	ErrNotAllowed = kernel.NewPublicError(kernel.ErrForbidden, "You may not write announcements.")
	ErrTooMany    = kernel.NewPublicError(kernel.ErrConflict, "There are already 50 current announcements; let some expire or delete them first.")
	ErrNoReader   = kernel.NewPublicError(kernel.ErrForbidden, "Only members and admins can mark announcements as read.")
)

// Validate checks the fields a writer sets.
func (a *Announcement) Validate() error {
	a.Title = strings.TrimSpace(a.Title)
	a.Body = strings.TrimSpace(a.Body)
	switch {
	case a.Title == "" || utf8.RuneCountInString(a.Title) > MaxTitle:
		return kernel.NewValidationError("title", "The title must be 1 to 120 characters.")
	case utf8.RuneCountInString(a.Body) > MaxBody:
		return kernel.NewValidationError("body", "The text must be at most 10,000 characters.")
	}
	return nil
}

// Active reports whether the notice is shown in the widget at now.
func (a Announcement) Active(now time.Time) bool {
	return a.ExpiresAt.IsZero() || now.Before(a.ExpiresAt)
}

// Sort orders notices for the widget: pinned first, then the newest.
func Sort(list []Announcement) {
	slices.SortStableFunc(list, func(x, y Announcement) int {
		if x.Pinned != y.Pinned {
			if x.Pinned {
				return -1
			}
			return 1
		}
		return y.CreatedAt.Compare(x.CreatedAt)
	})
}

// Settings say who besides admins may write.
type Settings struct {
	// WriterGroup is a group ID whose members may write announcements and
	// change their own; empty leaves writing to admins.
	WriterGroup string
}

// MayWrite reports whether p may write announcements.
func (s Settings) MayWrite(p kernel.Principal) bool {
	if p.IsAdmin() {
		return true
	}
	// Groups can hold guests and kiosk displays too; they only read.
	if p.TabScoped() || p.IsToken() {
		return false
	}
	return s.WriterGroup != "" && slices.Contains(p.Groups, s.WriterGroup)
}

// MayChange reports whether p may edit or delete a: admins all, writers
// their own.
func (s Settings) MayChange(p kernel.Principal, a Announcement) bool {
	if p.IsAdmin() {
		return true
	}
	return s.MayWrite(p) && a.AuthorID == p.AccountID
}
