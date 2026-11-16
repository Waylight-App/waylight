// Package domain holds the rules of directory sign-in (#396): who of the
// directory's users becomes a member and which of their groups count.
//
// Ring: Entities.
package domain

import (
	"slices"
	"strings"
)

// User is a directory entry found for a login name.
type User struct {
	DN string
	// Groups are the DNs of the groups the user is in (memberOf).
	Groups []string
	// Disabled is set for accounts the directory has switched off, such as an
	// Active Directory account with the ACCOUNTDISABLE flag.
	Disabled bool
}

// Policy decides who signs in as a member.
type Policy struct {
	// MemberGroups lets in only users in one of these groups (names or DNs);
	// empty lets in every user found below the base DN.
	MemberGroups []string
	// SyncGroups limits which directory groups become trisse-home groups;
	// empty takes all of them.
	SyncGroups []string
	// AdminGroups makes everyone in one of these groups (names or DNs) an
	// admin (#434); leaving the group takes the role back at the next sign-in.
	AdminGroups []string
}

// IsAdmin reports whether the user's groups make them an admin.
func (p Policy) IsAdmin(u User) bool {
	return slices.ContainsFunc(u.Groups, func(g string) bool { return matches(p.AdminGroups, g) })
}

// GroupName is the first value of a DN: "cn=Office,ou=groups,dc=x" → "Office".
func GroupName(dn string) string {
	first, _, _ := strings.Cut(dn, ",")
	_, value, ok := strings.Cut(first, "=")
	if !ok {
		return strings.TrimSpace(dn)
	}
	return strings.TrimSpace(value)
}

func matches(list []string, dn string) bool {
	name := GroupName(dn)
	return slices.ContainsFunc(list, func(x string) bool {
		x = strings.TrimSpace(x)
		return strings.EqualFold(x, dn) || strings.EqualFold(x, name)
	})
}

// Allows reports whether the user may sign in as a member.
func (p Policy) Allows(u User) bool {
	if u.Disabled {
		return false
	}
	if len(p.MemberGroups) == 0 {
		return true
	}
	return slices.ContainsFunc(u.Groups, func(g string) bool { return matches(p.MemberGroups, g) })
}

// GroupsToSync returns the names of the user's groups that become
// trisse-home groups, without duplicates.
func (p Policy) GroupsToSync(u User) []string {
	var out []string
	for _, dn := range u.Groups {
		name := GroupName(dn)
		if name == "" || slices.ContainsFunc(out, func(x string) bool { return strings.EqualFold(x, name) }) {
			continue
		}
		if len(p.SyncGroups) == 0 || matches(p.SyncGroups, dn) {
			out = append(out, name)
		}
	}
	return out
}

// ADDisabled reads the ACCOUNTDISABLE bit of Active Directory's
// userAccountControl attribute.
func ADDisabled(userAccountControl string) bool {
	n := 0
	for _, r := range strings.TrimSpace(userAccountControl) {
		if r < '0' || r > '9' {
			return false
		}
		n = n*10 + int(r-'0')
	}
	return n&0x2 != 0
}
