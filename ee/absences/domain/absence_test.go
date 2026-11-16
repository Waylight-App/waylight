package domain

import (
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestValidate(t *testing.T) {
	ok := Absence{From: "2026-12-22", To: "2027-01-02", Kind: KindHoliday}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, a := range map[string]Absence{
		"reversed":  {From: "2026-12-22", To: "2026-12-21", Kind: KindHoliday},
		"too long":  {From: "2026-01-01", To: "2027-01-02", Kind: KindHoliday},
		"bad day":   {From: "22.12.2026", To: "2026-12-23", Kind: KindHoliday},
		"bad kind":  {From: "2026-12-22", To: "2026-12-23", Kind: "away"},
		"long note": {From: "2026-12-22", To: "2026-12-23", Kind: KindOther, Note: string(make([]rune, 201))},
	} {
		if a.Validate() == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestPrivateReasonOnlyForThePersonAndAdmins(t *testing.T) {
	a := Absence{AccountID: "m1", Kind: KindSick, Note: "doctor", Private: true}
	for who, p := range map[string]kernel.Principal{
		"self":  {AccountID: "m1", Role: "member"},
		"admin": {AccountID: "a1", Role: "admin"},
		"owner": {AccountID: "owner", Role: "owner"},
	} {
		if got := a.SeenBy(p); got.Kind != KindSick || got.Note != "doctor" {
			t.Errorf("%s sees %+v", who, got)
		}
	}
	for who, p := range map[string]kernel.Principal{
		"colleague": {AccountID: "m2", Role: "member"},
		"kiosk":     {AccountID: "m1", Role: "kiosk"},
		"token":     {AccountID: "owner", Role: "token"},
	} {
		if got := a.SeenBy(p); got.Kind != KindAway || got.Note != "" {
			t.Errorf("%s sees %+v", who, got)
		}
	}
	open := Absence{AccountID: "m1", Kind: KindRemote, Note: "home office"}
	if got := open.SeenBy(kernel.Principal{AccountID: "m2", Role: "member"}); got.Kind != KindRemote {
		t.Errorf("a non-private absence is hidden: %+v", got)
	}
}

func TestMayChangeAndOverlaps(t *testing.T) {
	if !MayChange(kernel.Principal{AccountID: "m1", Role: "member"}, "m1") || MayChange(kernel.Principal{AccountID: "m1", Role: "member"}, "m2") ||
		!MayChange(kernel.Principal{AccountID: "a1", Role: "admin"}, "m2") || MayChange(kernel.Principal{AccountID: "g1", Role: "guest"}, "g1") {
		t.Fatal("rights")
	}
	a := Absence{From: "2026-12-22", To: "2026-12-24"}
	if !a.Overlaps("2026-12-24", "2026-12-30") || a.Overlaps("2026-12-25", "2026-12-30") {
		t.Fatal("overlap")
	}
}
