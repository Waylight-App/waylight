package domain

import (
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// MaxGroups bounds the number of groups.
const MaxGroups = 50

// Group is a set of accounts that records can be shared with (ADR 0018:
// visibility "group:<id>"). Members and guests can belong to groups; the
// admin sees everything shared anyway and needs no membership.
type Group struct {
	id        string
	name      string
	accounts  []string
	createdAt time.Time
}

// NewGroup creates a group; the name is shown in the sharing menu.
func NewGroup(id, name string, accounts []string, now time.Time) (Group, error) {
	g := Group{id: id, createdAt: now}
	if id == "" {
		return Group{}, kernel.NewValidationError("group", "id is required")
	}
	if err := g.Rename(name); err != nil {
		return Group{}, err
	}
	g.SetAccounts(accounts)
	return g, nil
}

// RestoreGroup rebuilds a group from persisted state.
func RestoreGroup(id, name string, accounts []string, createdAt time.Time) Group {
	return Group{id: id, name: name, accounts: accounts, createdAt: createdAt}
}

// Rename changes the display name: 1–40 characters, trimmed.
func (g *Group) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 40 {
		return kernel.NewValidationError("name", "The group name must be 1–40 characters.")
	}
	g.name = name
	return nil
}

// SetAccounts replaces the accounts in the group; the owner is never a
// member of a group, empty IDs and duplicates are dropped.
func (g *Group) SetAccounts(accounts []string) {
	out := []string{}
	for _, a := range accounts {
		if a != "" && a != OwnerID && !slices.Contains(out, a) {
			out = append(out, a)
		}
	}
	slices.Sort(out)
	g.accounts = out
}

// ID returns the group identifier.
func (g Group) ID() string { return g.id }

// Name returns the display name.
func (g Group) Name() string { return g.name }

// Accounts returns the IDs of the accounts in the group.
func (g Group) Accounts() []string { return slices.Clone(g.accounts) }

// CreatedAt returns when the group was created.
func (g Group) CreatedAt() time.Time { return g.createdAt }
