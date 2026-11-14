package domain

import (
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestExpandEnv(t *testing.T) {
	env := map[string]string{"HOME_DOMAIN": "example.org", "EMPTY": ""}
	lookup := func(k string) (string, bool) { v, ok := env[k]; return v, ok }
	for in, want := range map[string]string{
		"https://jf.${HOME_DOMAIN}/": "https://jf.example.org/",
		"a${EMPTY}b":                 "ab",
		"literal $${HOME_DOMAIN}":    "literal ${HOME_DOMAIN}",
		"no refs, $5 and {x}":        "no refs, $5 and {x}",
	} {
		if got, err := ExpandEnv(in, lookup); err != nil || got != want {
			t.Errorf("%q: got %q, %v", in, got, err)
		}
	}
	if _, err := ExpandEnv("${NOPE}", lookup); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("unset variable: %v", err)
	}
}

func TestCheckConfigVersion(t *testing.T) {
	if CheckConfigVersion(1) != nil || CheckConfigVersion(2) == nil || CheckConfigVersion(0) == nil {
		t.Fatal("only version 1 is supported")
	}
}
