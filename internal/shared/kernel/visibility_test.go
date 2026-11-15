package kernel

import (
	"context"
	"testing"
)

func TestParseVisibility(t *testing.T) {
	for in, ok := range map[string]bool{
		"private": true, "everyone": true, "group:office": true,
		"": false, "public": false, "group:": false, "group:a b": false, "group:a,b": false, "group:x'": false,
	} {
		if _, err := ParseVisibility(in); (err == nil) != ok {
			t.Errorf("%q: err %v, want ok=%v", in, err, ok)
		}
	}
	if id, ok := GroupVisibility("office").Group(); !ok || id != "office" {
		t.Fatal(id, ok)
	}
	if _, ok := Everyone.Group(); ok {
		t.Fatal("everyone is no group")
	}
}

func TestVisibleTo(t *testing.T) {
	anna := Principal{AccountID: "anna", Role: "member", Groups: []string{"office"}}
	ben := Principal{AccountID: "ben", Role: "member"}
	admin := Principal{AccountID: "owner", Role: "owner"}
	guest := Principal{AccountID: "g1", Role: "guest", Groups: []string{"office"}}
	system := PrincipalFrom(context.Background())
	cases := []struct {
		name  string
		v     Visibility
		owner string
		p     Principal
		want  bool
	}{
		{"owner sees private", Private, "anna", anna, true},
		{"other member does not see private", Private, "anna", ben, false},
		{"admin has no extra right", Private, "anna", admin, false},
		{"system sees private", Private, "anna", system, true},
		{"everyone is everyone", Everyone, "anna", ben, true},
		{"guests see everyone", Everyone, "anna", guest, true},
		{"group member sees group", GroupVisibility("office"), "ben", anna, true},
		{"guest in group sees group", GroupVisibility("office"), "ben", guest, true},
		{"outsider does not see group", GroupVisibility("office"), "anna", ben, false},
		{"empty owner never matches", Private, "", Principal{Role: "member"}, false},
	}
	for _, c := range cases {
		if got := c.v.VisibleTo(c.p, c.owner); got != c.want {
			t.Errorf("%s: got %v", c.name, got)
		}
	}
}

func TestPrincipalWithoutRequestIsSystem(t *testing.T) {
	p := PrincipalFrom(context.Background())
	if !p.IsSystem() || p.AccountID != SystemAccountID || p.IsGuest() || p.TabScoped() {
		t.Fatalf("%+v", p)
	}
}

func TestDefaultVisibility(t *testing.T) {
	if DefaultVisibility(Principal{Role: "member"}) != Private {
		t.Fatal("members start private")
	}
	if DefaultVisibility(AdminPrincipal()) != Everyone {
		t.Fatal("the admin keeps everyone by default")
	}
	admin := AdminPrincipal()
	admin.PrivateByDefault = true
	if DefaultVisibility(admin) != Private {
		t.Fatal("the admin's choice of private is respected")
	}
}
