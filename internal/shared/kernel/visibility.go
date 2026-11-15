package kernel

import (
	"errors"
	"slices"
	"strings"
)

// Visibility says who besides its owner may see a record of personal content
// (ADR 0018). It is stored as text: "private", "everyone" or "group:<id>".
type Visibility string

const (
	// Private records are seen by their owner only. Default for new records
	// of members.
	Private Visibility = "private"
	// Everyone means every account of the instance, guests included when the
	// widget is on a tab they may see. Existing records get it on migration,
	// so guests keep the view they had before.
	Everyone Visibility = "everyone"
)

const groupPrefix = "group:"

// ErrInvalidVisibility is returned for a value that is none of the above.
var ErrInvalidVisibility = errors.New("invalid visibility")

// GroupVisibility shares a record with the members of one group.
func GroupVisibility(groupID string) Visibility { return Visibility(groupPrefix + groupID) }

// ParseVisibility validates a stored or submitted value.
func ParseVisibility(s string) (Visibility, error) {
	switch v := Visibility(s); {
	case v == Private, v == Everyone:
		return v, nil
	case strings.HasPrefix(s, groupPrefix) && len(s) > len(groupPrefix) && !strings.ContainsAny(s[len(groupPrefix):], " ,'\""):
		return v, nil
	}
	return "", ErrInvalidVisibility
}

// Group returns the group a record is shared with, if any.
func (v Visibility) Group() (string, bool) {
	id, ok := strings.CutPrefix(string(v), groupPrefix)
	return id, ok && id != ""
}

// VisibleTo is the one rule for "may p see this record" (ADR 0018). The
// owner always may; admins have no extra right to other people's private
// records. Requests without a principal (background jobs) act as the
// instance and see everything, as they do today.
//
// Repositories do not call this per row: sqlutil.VisibleTo expresses the
// same rule as a WHERE clause, and a test keeps the two in step.
func (v Visibility) VisibleTo(p Principal, ownerID string) bool {
	if p.IsSystem() || (ownerID != "" && ownerID == p.AccountID) {
		return true
	}
	if v == Everyone {
		return true
	}
	if id, ok := v.Group(); ok {
		return slices.Contains(p.Groups, id)
	}
	return false
}

// ErrNotOwner is returned when someone tries to change or delete a record
// they can see but do not own (ADR 0018: only the owner changes a record).
var ErrNotOwner = NewPublicError(ErrForbidden, "Only the person who created this can change it.")

// DefaultVisibility is the visibility of a new record created by p. Members
// start private. The admin keeps today's behaviour, everyone, so a household
// without members notices no change (ADR 0018, migration step 3); asking the
// admin to switch to private comes with the first member.
func DefaultVisibility(p Principal) Visibility {
	if p.Role == "member" {
		return Private
	}
	return Everyone
}
