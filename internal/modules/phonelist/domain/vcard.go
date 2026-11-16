package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Source tells where an entry comes from.
type Source string

// Entry sources.
const (
	SourceManual  Source = "manual"
	SourceCardDAV Source = "carddav"
)

// ErrReadOnlyEntry is returned when an entry from the address book is changed here.
var ErrReadOnlyEntry = kernel.NewPublicError(kernel.ErrConflict, "This entry comes from the address book; change it there.")

// Card is a person read from an address book, with a stable ID.
type Card struct {
	ID string
	EntryInput
}

// FromCard turns a card into an entry of the address book source. Cards
// that break the rules for entries are left out (ok is false).
func FromCard(c Card, now time.Time) (Entry, bool) {
	e, err := NewEntry(c.ID, c.EntryInput, now)
	if err != nil {
		return Entry{}, false
	}
	e.Source = SourceCardDAV
	return e, true
}

// unfold joins folded vCard lines (RFC 6350 §3.2): a line starting with a
// space or tab continues the previous one.
func unfold(data string) []string {
	var out []string
	for _, l := range strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n") {
		if (strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t")) && len(out) > 0 {
			out[len(out)-1] += l[1:]
			continue
		}
		out = append(out, l)
	}
	return out
}

// unescape decodes vCard text values (\, \; \n).
func unescape(s string) string {
	r := strings.NewReplacer(`\n`, " ", `\N`, " ", `\,`, ",", `\;`, ";", `\\`, `\`)
	return strings.TrimSpace(r.Replace(s))
}

// preferred reports whether a TEL or EMAIL line is marked as preferred,
// with PREF=1 (vCard 4) or TYPE=pref (vCard 3).
func preferred(params string) bool {
	p := strings.ToLower(params)
	return strings.Contains(p, "pref=1") || strings.Contains(p, "type=pref") || strings.Contains(p, ",pref")
}

// ParseVCards reads the people of one or more vCards: the formatted name,
// the title or role, the organisation unit as team, and the preferred (or
// first) phone number and e-mail address. Cards without a name, or without
// a number and an address, are skipped, like entries typed in by hand.
func ParseVCards(data string) []Card {
	var out []Card
	var cur *EntryInput
	var uid string
	var telPref, mailPref bool
	ids := map[string]bool{}
	for _, line := range unfold(data) {
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key, params, _ := strings.Cut(name, ";")
		// Grouped properties such as "item1.TEL" count like plain ones.
		if i := strings.LastIndexByte(key, '.'); i >= 0 {
			key = key[i+1:]
		}
		switch strings.ToUpper(key) {
		case "BEGIN":
			if strings.EqualFold(value, "VCARD") {
				cur, uid, telPref, mailPref = &EntryInput{}, "", false, false
			}
		case "END":
			if cur != nil && strings.EqualFold(value, "VCARD") {
				if cur.Name != "" && (cur.Phone != "" || cur.Email != "") {
					out = append(out, Card{ID: cardID(uid, *cur, ids), EntryInput: *cur})
				}
				cur = nil
			}
		}
		if cur == nil {
			continue
		}
		v := unescape(value)
		switch strings.ToUpper(key) {
		case "FN":
			cur.Name = v
		case "UID":
			uid = strings.TrimSpace(value)
		case "TITLE":
			cur.Role = v
		case "ROLE":
			if cur.Role == "" {
				cur.Role = v
			}
		case "ORG":
			// ORG is "Company;Unit"; the unit is the team, the company alone
			// says nothing in an internal list.
			if _, unit, ok := strings.Cut(value, ";"); ok {
				cur.Team = unescape(strings.Split(unit, ";")[0])
			}
		case "TEL":
			if p := preferred(params); cur.Phone == "" || (p && !telPref) {
				cur.Phone, telPref = strings.TrimPrefix(v, "tel:"), p
			}
		case "EMAIL":
			if p := preferred(params); cur.Email == "" || (p && !mailPref) {
				cur.Email, mailPref = v, p
			}
		}
	}
	return out
}

// cardID derives a stable entry ID from the card's UID (or its contents),
// so a sync keeps the same IDs and nothing jumps around in the widget.
func cardID(uid string, in EntryInput, seen map[string]bool) string {
	src := uid
	if src == "" {
		src = in.Name + "\x00" + in.Phone + "\x00" + in.Email
	}
	sum := sha256.Sum256([]byte("carddav\x00" + src))
	base := "cd-" + hex.EncodeToString(sum[:12])
	id := base
	// Two cards with the same UID (a broken export) still get two entries.
	for n := 2; seen[id]; n++ {
		id = fmt.Sprintf("%s-%d", base, n)
	}
	seen[id] = true
	return id
}
