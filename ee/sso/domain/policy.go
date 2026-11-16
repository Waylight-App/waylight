// Package domain holds the rules of single sign-on for members (#302): who
// the identity provider may let in and which of their groups count.
//
// Ring: Entities.
package domain

import (
	"slices"
	"strings"
)

// Identity is a person as the identity provider confirmed them.
type Identity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Username      string
	Groups        []string
}

// Policy decides who becomes a member through single sign-on.
type Policy struct {
	// Emails lists addresses ("anna@example.org") or domains ("@example.org")
	// whose people may sign in. Only verified addresses count.
	Emails []string
	// Groups lets in everyone in one of these provider groups.
	Groups []string
	// SyncGroups limits which provider groups become trisse-home groups;
	// empty takes all of them.
	SyncGroups []string
}

// Allows reports whether the person may sign in as a member.
func (p Policy) Allows(id Identity) bool {
	if id.Subject == "" {
		return false
	}
	if id.EmailVerified && id.Email != "" {
		email := strings.ToLower(id.Email)
		_, domain, _ := strings.Cut(email, "@")
		for _, e := range p.Emails {
			e = strings.ToLower(strings.TrimSpace(e))
			if e == email || (strings.HasPrefix(e, "@") && e == "@"+domain) {
				return true
			}
		}
	}
	for _, g := range id.Groups {
		if containsFold(p.Groups, g) {
			return true
		}
	}
	return false
}

// GroupsToSync returns the person's provider groups that become trisse-home
// groups, without duplicates.
func (p Policy) GroupsToSync(id Identity) []string {
	var out []string
	for _, g := range id.Groups {
		g = strings.TrimSpace(g)
		if g == "" || containsFold(out, g) {
			continue
		}
		if len(p.SyncGroups) == 0 || containsFold(p.SyncGroups, g) {
			out = append(out, g)
		}
	}
	return out
}

// Leave lists the groups synced last time that the person is no longer in.
// Groups an admin added by hand were never synced, so they are kept.
func Leave(previous, now []string) []string {
	var out []string
	for _, g := range previous {
		if !containsFold(now, g) {
			out = append(out, g)
		}
	}
	return out
}

// AccountName is the name a new member starts with: the provider's short
// login name, else the e-mail address, else the full name. The auth module
// makes it valid and unique.
func (id Identity) AccountName() string {
	for _, n := range []string{id.Username, id.Email, id.Name} {
		if strings.TrimSpace(n) != "" {
			return n
		}
	}
	return "member"
}

func containsFold(list []string, s string) bool {
	return slices.ContainsFunc(list, func(x string) bool { return strings.EqualFold(strings.TrimSpace(x), s) })
}
