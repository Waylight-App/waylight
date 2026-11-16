package domain

import (
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestSortPinnedFirstThenNewest(t *testing.T) {
	t0 := time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)
	list := []Announcement{{ID: "old", CreatedAt: t0}, {ID: "new", CreatedAt: t0.Add(time.Hour)}, {ID: "pin", Pinned: true, CreatedAt: t0.Add(-time.Hour)}}
	Sort(list)
	if list[0].ID != "pin" || list[1].ID != "new" || list[2].ID != "old" {
		t.Fatalf("order %v %v %v", list[0].ID, list[1].ID, list[2].ID)
	}
}

func TestWhoMayWriteAndChange(t *testing.T) {
	set := Settings{WriterGroup: "g1"}
	admin := kernel.Principal{AccountID: "a1", Role: "admin"}
	writer := kernel.Principal{AccountID: "m1", Role: "member", Groups: []string{"g1"}}
	reader := kernel.Principal{AccountID: "m2", Role: "member"}
	guest := kernel.Principal{AccountID: "g", Role: "guest", Groups: []string{"g1"}}
	mine := Announcement{AuthorID: "m1"}
	if !set.MayWrite(admin) || !set.MayWrite(writer) || set.MayWrite(reader) {
		t.Fatal("write rights")
	}
	if !set.MayChange(writer, mine) || set.MayChange(writer, Announcement{AuthorID: "a1"}) || !set.MayChange(admin, mine) {
		t.Fatal("change rights")
	}
	if (Settings{}).MayWrite(writer) {
		t.Fatal("no writer group, yet a member may write")
	}
	if set.MayWrite(guest) {
		t.Fatal("a guest in the writer group may write")
	}
}

func TestActiveAndValidate(t *testing.T) {
	now := time.Now()
	if !(Announcement{}).Active(now) || (Announcement{ExpiresAt: now.Add(-time.Second)}).Active(now) {
		t.Fatal("active")
	}
	a := Announcement{Title: "  "}
	if a.Validate() == nil {
		t.Fatal("empty title accepted")
	}
}
