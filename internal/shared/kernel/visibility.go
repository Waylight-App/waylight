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
// and admins (people with a space of their own, ADR 0022) start private. The admin keeps today's behaviour, everyone, so a household
// without members notices no change, until they choose private for their
// new records once members exist (ADR 0018, migration step 3).
func DefaultVisibility(p Principal) Visibility {
	if p.Role == "member" || p.Role == "admin" || p.PrivateByDefault {
		return Private
	}
	return Everyone
}

// Access is who owns a record and who else may see it (ADR 0018).
type Access struct {
	Owner      string
	Visibility Visibility
}

// NewAccess is the access of a record p creates now.
func NewAccess(p Principal) Access {
	return Access{Owner: p.AccountID, Visibility: DefaultVisibility(p)}
}

// Mine reports whether p owns the record and may therefore change it and
// its visibility.
func (a Access) Mine(p Principal) bool {
	return p.IsSystem() || (a.Owner != "" && a.Owner == p.AccountID)
}

// ErrShareNotAllowed refuses a visibility the principal may not give.
var ErrShareNotAllowed = NewValidationError("visibility", "Share with only you, everyone or one of your groups.")

// CheckShare parses a visibility p wants to give one of its records:
// private, everyone, or a group p belongs to. Admins may name any group;
// they manage groups without being in them.
func CheckShare(p Principal, raw string) (Visibility, error) {
	v, err := ParseVisibility(raw)
	if err != nil {
		return "", ErrShareNotAllowed
	}
	if id, ok := v.Group(); ok && !p.IsAdmin() && !slices.Contains(p.Groups, id) {
		return "", ErrShareNotAllowed
	}
	return v, nil
}
