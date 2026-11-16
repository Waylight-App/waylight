package domain

import "testing"

func TestParseFilter(t *testing.T) {
	f, err := ParseFilter(`userName eq "Anna@Example.org"`)
	if err != nil || f.Attribute != "username" || !f.Matches("username", "anna@example.org") || f.Matches("externalid", "x") {
		t.Fatalf("%+v %v", f, err)
	}
	if f, err := ParseFilter(""); f != nil || err != nil {
		t.Fatal("empty filter")
	}
	for _, bad := range []string{`userName sw "a"`, `userName eq anna`, `userName`, `a eq "b" and c eq "d"`} {
		if _, err := ParseFilter(bad); err == nil {
			t.Errorf("accepted %q", bad)
		}
	}
}

func TestParseBool(t *testing.T) {
	for in, want := range map[any]bool{true: true, false: false, "False": false, "True": true} {
		if got, ok := ParseBool(in); !ok || got != want {
			t.Errorf("%v: %v %v", in, got, ok)
		}
	}
	if _, ok := ParseBool(1); ok {
		t.Error("number accepted")
	}
}

func TestMemberPath(t *testing.T) {
	if id, ok := MemberPath(`members[value eq "m1"]`); !ok || id != "m1" {
		t.Fatalf("%q %v", id, ok)
	}
	if _, ok := MemberPath("members"); ok {
		t.Fatal("plain path")
	}
}
