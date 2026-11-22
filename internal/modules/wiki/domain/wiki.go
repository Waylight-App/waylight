// Package domain holds the vendor-neutral wiki search types (#321).
//
// Ring: Entities.
package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// Kind is the wiki software behind a connection.
type Kind string

// Supported wikis; each is a gateway adapter behind the PageSearcher port.
const (
	KindBookStack Kind = "bookstack"
	KindOutline   Kind = "outline"
	KindWikiJS    Kind = "wikijs"
)

// Valid reports whether k is supported.
func (k Kind) Valid() bool { return k == KindBookStack || k == KindOutline || k == KindWikiJS }

// Limits of the wiki search.
const (
	MaxWikis = 5
	// MinQueryLength keeps one typed letter from asking every wiki.
	MinQueryLength = 2
	MaxQueryLength = 200
	// MaxResultsPerWiki keeps the search panel readable.
	MaxResultsPerWiki = 5
	maxNameLength     = 60
	maxExcerptLength  = 240
)

// Errors of the wiki search.
var (
	ErrNotFound    = kernel.NewPublicError(kernel.ErrNotFound, "The wiki is not connected (any more).")
	ErrKind        = kernel.NewValidationError("kind", "Choose BookStack, Outline or Wiki.js.")
	ErrName        = kernel.NewValidationError("name", "The name must have between 1 and 60 characters.")
	ErrTooMany     = kernel.NewPublicError(kernel.ErrConflict, "At most 5 wikis can be connected.")
	ErrTokenNeeded = kernel.NewValidationError("token", "Enter the access token of the wiki.")
)

// Wiki is one connected wiki. Token is the API credential; for BookStack it
// is "<token id>:<token secret>". Everyone who uses this instance sees what
// that token may read.
type Wiki struct {
	ID    string
	Name  string
	Kind  Kind
	URL   string
	Token string
}

// Validate checks a wiki and normalises its fields.
func (w *Wiki) Validate() error {
	w.Name = strings.TrimSpace(w.Name)
	if w.Name == "" || utf8.RuneCountInString(w.Name) > maxNameLength {
		return ErrName
	}
	if !w.Kind.Valid() {
		return ErrKind
	}
	u, err := kernel.ParseTarget("url", w.URL, kernel.WithTargetPath())
	if err != nil {
		return err
	}
	w.URL = strings.TrimRight(u.String(), "/")
	w.Token = strings.TrimSpace(w.Token)
	if w.Token == "" {
		return ErrTokenNeeded
	}
	return nil
}

// Page is a search result, the same for every wiki software.
type Page struct {
	Title   string
	Excerpt string
	URL     string
	// Space is the book, collection or locale the page belongs to.
	Space   string
	Updated time.Time
	// Wiki is the name of the connection it came from.
	Wiki string
}

// NormaliseQuery trims the query; ok is false when it is too short to send.
func NormaliseQuery(q string) (string, bool) {
	q = strings.Join(strings.Fields(q), " ")
	if utf8.RuneCountInString(q) < MinQueryLength {
		return "", false
	}
	if r := []rune(q); len(r) > MaxQueryLength {
		q = string(r[:MaxQueryLength])
	}
	return q, true
}

// Excerpt turns a highlighted HTML snippet from a wiki into short plain
// text: tags are dropped, the common entities decoded, whitespace folded.
// The UI never renders it as HTML.
func Excerpt(html string) string {
	var b strings.Builder
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>' && inTag:
			inTag = false
			b.WriteRune(' ')
		case !inTag:
			b.WriteRune(r)
		}
	}
	text := strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'", "&#039;", "'", "&nbsp;", " ").Replace(b.String())
	text = strings.Join(strings.Fields(text), " ")
	if r := []rune(text); len(r) > maxExcerptLength {
		text = string(r[:maxExcerptLength-1]) + "…"
	}
	return text
}
