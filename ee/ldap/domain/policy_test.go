package domain

import (
	"slices"
	"testing"
)

func TestGroupName(t *testing.T) {
	for dn, want := range map[string]string{"cn=Office,ou=groups,dc=example,dc=org": "Office", "CN=Sales": "Sales", "plain": "plain"} {
		if got := GroupName(dn); got != want {
			t.Errorf("%q: %q", dn, got)
		}
	}
}

func TestPolicy(t *testing.T) {
	u := User{Groups: []string{"cn=Office,ou=groups,dc=x", "cn=VPN,ou=groups,dc=x"}}
	if !(Policy{}).Allows(u) || !(Policy{MemberGroups: []string{"office"}}).Allows(u) || (Policy{MemberGroups: []string{"Sales"}}).Allows(u) {
		t.Error("member groups")
	}
	if (Policy{}).Allows(User{Disabled: true}) {
		t.Error("disabled user allowed")
	}
	if got := (Policy{SyncGroups: []string{"cn=Office,ou=groups,dc=x"}}).GroupsToSync(u); !slices.Equal(got, []string{"Office"}) {
		t.Errorf("sync: %v", got)
	}
}

func TestADDisabled(t *testing.T) {
	if !ADDisabled("514") || ADDisabled("512") || ADDisabled("") || ADDisabled("x") {
		t.Error("userAccountControl")
	}
}
