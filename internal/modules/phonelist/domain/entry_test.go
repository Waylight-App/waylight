package domain

import (
	"strings"
	"testing"
	"time"
)

func TestNewEntry(t *testing.T) {
	e, err := NewEntry("1", EntryInput{Name: "  Anna Weber ", Role: "Reception", Phone: "+49 (30) 123-45 6", Team: "Front desk"}, time.Unix(0, 0))
	if err != nil || e.Name != "Anna Weber" || e.Phone != "+49 (30) 123-45 6" {
		t.Fatalf("%+v %v", e, err)
	}
	if got := DialNumber(e.Phone); got != "+4930123456" {
		t.Fatalf("dial %q", got)
	}
	for name, in := range map[string]EntryInput{
		"no name":          {Phone: "1"},
		"long name":        {Name: strings.Repeat("x", MaxNameLength+1), Phone: "1"},
		"letters in phone": {Name: "A", Phone: "call me"},
		"only plus":        {Name: "A", Phone: "+"},
		"bad e-mail":       {Name: "A", Email: "Anna <a@example.org>"},
		"no way to reach":  {Name: "A"},
		"long team":        {Name: "A", Phone: "1", Team: strings.Repeat("x", MaxTeamLength+1)},
	} {
		if _, err := NewEntry("1", in, time.Unix(0, 0)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := NewEntry("1", EntryInput{Name: "Desk", Email: "desk@example.org"}, time.Unix(0, 0)); err != nil {
		t.Fatal("an e-mail address alone is enough", err)
	}
}
