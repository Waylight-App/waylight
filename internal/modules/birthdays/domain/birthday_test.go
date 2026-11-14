package domain

import (
	"testing"
	"time"
)

func TestNewBirthday(t *testing.T) {
	if _, err := NewBirthday("a", "Anna", 2, 29, 0, SourceManual); err != nil {
		t.Fatal("29 February without year", err)
	}
	for _, bad := range [][3]int{{2, 30, 0}, {13, 1, 0}, {2, 29, 2023}, {4, 31, 1990}, {0, 1, 0}} {
		if _, err := NewBirthday("a", "Anna", bad[0], bad[1], bad[2], SourceManual); err == nil {
			t.Errorf("accepted %v", bad)
		}
	}
	if _, err := NewBirthday("a", "  ", 1, 1, 0, SourceManual); err == nil {
		t.Error("empty name")
	}
}

func TestNext(t *testing.T) {
	loc := time.FixedZone("CET", 3600)
	now := time.Date(2026, 11, 14, 23, 30, 0, 0, loc)
	list := []Birthday{
		{Name: "Today", Month: 11, Day: 14, Year: 1990},
		{Name: "Yesterday", Month: 11, Day: 13},
		{Name: "Leap", Month: 2, Day: 29, Year: 2000},
		{Name: "Soon", Month: 12, Day: 1, Year: 2010},
	}
	got := Next(list, now, loc)
	if got[0].Name != "Today" || got[0].InDays != 0 || got[0].Age != 36 {
		t.Fatalf("today: %+v", got[0])
	}
	if got[1].Name != "Soon" || got[1].InDays != 17 || got[1].Age != 16 {
		t.Fatalf("soon: %+v", got[1])
	}
	if got[2].Name != "Leap" || got[2].Date.Day() != 28 || got[2].Age != 27 {
		t.Fatalf("29 February in 2027 falls on the 28th: %+v", got[2])
	}
	if got[3].Name != "Yesterday" || got[3].InDays != 364 || got[3].Age != 0 {
		t.Fatalf("next year, age unknown: %+v", got[3])
	}
}

func TestVCards(t *testing.T) {
	data := "BEGIN:VCARD\r\nVERSION:3.0\r\nFN:Anna\r\n  Muster\r\nUID:u1\r\nBDAY:1990-05-17\r\nEND:VCARD\r\n" +
		"BEGIN:VCARD\r\nVERSION:4.0\r\nFN:Bernd\\, Jr.\r\nBDAY:--0317\r\nEND:VCARD\r\n" +
		"BEGIN:VCARD\r\nFN:Carla\r\nitem1.BDAY;VALUE=date:1604-08-09\r\nEND:VCARD\r\n" +
		"BEGIN:VCARD\r\nFN:Dora\r\nBDAY;VALUE=text:circa 1800\r\nEND:VCARD\r\n" +
		"BEGIN:VCARD\r\nFN:Emil\r\nEND:VCARD\r\n"
	cards := ParseVCards(data)
	if len(cards) != 3 {
		t.Fatalf("%+v", cards)
	}
	if cards[0].Name != "Anna Muster" || cards[0].Year != 1990 || cards[0].Month != 5 || cards[0].UID != "u1" {
		t.Fatalf("%+v", cards[0])
	}
	if cards[1].Name != "Bernd, Jr." || cards[1].Year != 0 || cards[1].Day != 17 {
		t.Fatalf("%+v", cards[1])
	}
	if cards[2].Year != 0 || cards[2].Month != 8 {
		t.Fatalf("Apple's 1604 means no year: %+v", cards[2])
	}
	if m, d, y, ok := ParseBDAY("19850102T000000Z"); !ok || m != 1 || d != 2 || y != 1985 {
		t.Fatal(m, d, y, ok)
	}
}

func TestReminder(t *testing.T) {
	loc := time.UTC
	r, err := NewReminder(true, "")
	if err != nil || r.At != "08:00" {
		t.Fatal(r, err)
	}
	if _, err := NewReminder(true, "8 Uhr"); err == nil {
		t.Fatal("bad time")
	}
	morning := time.Date(2026, 11, 14, 7, 59, 0, 0, loc)
	if r.Due(morning, "", loc) || !r.Due(morning.Add(time.Minute), "", loc) || r.Due(morning.Add(time.Hour), "2026-11-14", loc) {
		t.Fatal("due once per day after the time")
	}
}
