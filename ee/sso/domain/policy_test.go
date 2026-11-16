package domain

import (
	"slices"
	"testing"
)

func TestPolicyAllows(t *testing.T) {
	p := Policy{Emails: []string{"@Example.org", "freelancer@mail.test"}, Groups: []string{"trisse-users"}}
	for name, tc := range map[string]struct {
		id   Identity
		want bool
	}{
		"verified address in the domain": {Identity{Subject: "s", Email: "anna@example.org", EmailVerified: true}, true},
		"listed address":                 {Identity{Subject: "s", Email: "Freelancer@mail.test", EmailVerified: true}, true},
		"unverified address":             {Identity{Subject: "s", Email: "anna@example.org"}, false},
		"other domain":                   {Identity{Subject: "s", Email: "anna@example.org.evil", EmailVerified: true}, false},
		"allowed group":                  {Identity{Subject: "s", Groups: []string{"TRISSE-users"}}, true},
		"no subject":                     {Identity{Email: "anna@example.org", EmailVerified: true}, false},
		"nothing matches":                {Identity{Subject: "s", Groups: []string{"sales"}}, false},
	} {
		if got := p.Allows(tc.id); got != tc.want {
			t.Errorf("%s: got %v", name, got)
		}
	}
}

func TestGroupsToSyncAndLeave(t *testing.T) {
	id := Identity{Groups: []string{"Office", "office", " ", "Sales", "VPN-Users"}}
	if got := (Policy{}).GroupsToSync(id); !slices.Equal(got, []string{"Office", "Sales", "VPN-Users"}) {
		t.Errorf("all: %v", got)
	}
	if got := (Policy{SyncGroups: []string{"office", "sales"}}).GroupsToSync(id); !slices.Equal(got, []string{"Office", "Sales"}) {
		t.Errorf("filtered: %v", got)
	}
	if got := Leave([]string{"Office", "Sales"}, []string{"office"}); !slices.Equal(got, []string{"Sales"}) {
		t.Errorf("leave: %v", got)
	}
}

func TestAccountName(t *testing.T) {
	if n := (Identity{Username: "anna", Email: "a@x.org"}).AccountName(); n != "anna" {
		t.Errorf("username: %q", n)
	}
	if n := (Identity{Email: "a@x.org", Name: "Anna"}).AccountName(); n != "a@x.org" {
		t.Errorf("email: %q", n)
	}
}
