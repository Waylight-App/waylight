package domain

import (
	"testing"
	"time"
)

const cards = "BEGIN:VCARD\r\nVERSION:4.0\r\nUID:urn:uuid:1\r\nFN:Anna Weber\r\nTITLE:Reception\r\nORG:Example Ltd;Front desk\r\n" +
	"TEL;TYPE=work:+49 30 111\r\nTEL;TYPE=cell;PREF=1:+49 170 222\r\nEMAIL:anna@example.org\r\nEND:VCARD\r\n" +
	"BEGIN:VCARD\r\nVERSION:3.0\r\nFN:Ben\r\n  Ott\r\nitem1.EMAIL;TYPE=INTERNET:ben@example.org\r\nEND:VCARD\r\n" +
	"BEGIN:VCARD\r\nFN:Nobody to reach\r\nEND:VCARD\r\n" +
	"BEGIN:VCARD\r\nTEL:+49 1\r\nEND:VCARD\r\n"

func TestParseVCards(t *testing.T) {
	got := ParseVCards(cards)
	if len(got) != 2 {
		t.Fatalf("cards without a name or a way to reach them are skipped: %+v", got)
	}
	anna, ben := got[0], got[1]
	if anna.Name != "Anna Weber" || anna.Role != "Reception" || anna.Team != "Front desk" || anna.Phone != "+49 170 222" || anna.Email != "anna@example.org" {
		t.Fatalf("anna: %+v", anna)
	}
	if ben.Name != "Ben Ott" || ben.Email != "ben@example.org" || ben.Phone != "" {
		t.Fatalf("ben (folded name, grouped e-mail): %+v", ben)
	}
	if again := ParseVCards(cards); again[0].ID != anna.ID || again[1].ID != ben.ID {
		t.Fatal("IDs must stay the same between syncs")
	}
	e, ok := FromCard(anna, time.Unix(0, 0))
	if !ok || e.Source != SourceCardDAV || e.ID != anna.ID {
		t.Fatalf("%+v %v", e, ok)
	}
	if _, ok := FromCard(Card{ID: "x", EntryInput: EntryInput{Name: "A", Phone: "call me"}}, time.Unix(0, 0)); ok {
		t.Fatal("a card that breaks the entry rules is left out")
	}
}
