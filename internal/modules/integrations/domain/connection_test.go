package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

func TestNewConnection(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	c, err := NewConnection("1", "immich", " Home ", "http://192.168.178.20:2283/", Credential{"apiKey": "k"}, now)
	if err != nil || c.Name != "Home" || c.URL != "http://192.168.178.20:2283" || c.Credential["apiKey"] != "k" {
		t.Fatalf("%+v %v", c, err)
	}
	for name, args := range map[string][3]string{
		"service": {"Immich!", "Home", "http://h"},
		"name":    {"immich", " ", "http://h"},
		"long":    {"immich", strings.Repeat("x", MaxNameLength+1), "http://h"},
		"url":     {"immich", "Home", "ftp://h"},
	} {
		var v *kernel.ValidationError
		if _, err := NewConnection("1", args[0], args[1], args[2], nil, now); !errors.As(err, &v) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestCredential(t *testing.T) {
	c, err := NewCredential(map[string]string{"username": "anna", "password": "p", "tokenId": ""})
	if err != nil || !slices.Equal(c.Fields(), []string{"password", "username"}) {
		t.Fatalf("%v %v", c, err)
	}
	if _, err := NewCredential(map[string]string{"api key": "x"}); err == nil {
		t.Error("invalid field name accepted")
	}
	if _, err := NewCredential(map[string]string{"a": "1", "b": "2", "c": "3", "d": "4", "e": "5"}); err == nil {
		t.Error("too many fields accepted")
	}
	if _, err := NewCredential(map[string]string{"token": strings.Repeat("x", MaxValueLength+1)}); err == nil {
		t.Error("too long value accepted")
	}
}

func TestSameName(t *testing.T) {
	if !SameName("Home", " home") || SameName("Home", "Office") {
		t.Fatal("names are compared without case and spaces")
	}
}
