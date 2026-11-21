package domain

import (
	"errors"
	"maps"
	"testing"
)

func TestSpecCheckCredential(t *testing.T) {
	s := Spec{Service: "proxmox", Fields: []Field{{Name: "tokenId"}, {Name: "secret", Secret: true}, {Name: "fingerprint", Optional: true}}}
	for name, tc := range map[string]struct {
		cred     Credential
		complete bool
		want     error
	}{
		"complete":            {Credential{"tokenId": "u@pam!w", "secret": "s"}, true, nil},
		"optional field":      {Credential{"tokenId": "u@pam!w", "secret": "s", "fingerprint": "AA:BB"}, true, nil},
		"unknown field":       {Credential{"tokenId": "u@pam!w", "secret": "s", "apiKey": "k"}, true, ErrUnknownField},
		"missing field":       {Credential{"tokenId": "u@pam!w"}, true, ErrMissingField},
		"partial from config": {Credential{}, false, nil},
		"unknown, partial":    {Credential{"apiKey": "k"}, false, ErrUnknownField},
	} {
		if err := s.CheckCredential(tc.cred, tc.complete); !errors.Is(err, tc.want) {
			t.Errorf("%s: %v, want %v", name, err, tc.want)
		}
	}
}

func TestSpecVisibleHidesSecrets(t *testing.T) {
	s := Spec{Fields: []Field{{Name: "username"}, {Name: "password", Secret: true}}}
	got := s.Visible(Credential{"username": "admin", "password": "hunter2", "stray": "x"})
	if !maps.Equal(got, map[string]string{"username": "admin"}) {
		t.Fatalf("%v", got)
	}
}

func TestSpecsSorted(t *testing.T) {
	specs := NewSpecs([]Spec{{Service: "radarr"}, {Service: "immich"}})
	if got := specs.Sorted(); len(got) != 2 || got[0].Service != "immich" {
		t.Fatalf("%+v", got)
	}
}
