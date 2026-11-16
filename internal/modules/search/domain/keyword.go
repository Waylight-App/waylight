package domain

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// TargetKind says what an own keyword opens (#478).
type TargetKind string

// The targets of an own keyword: a service tile, a bookmark or a widget by
// its id, a web address, or a search that runs as if typed.
const (
	TargetService  TargetKind = "service"
	TargetBookmark TargetKind = "bookmark"
	TargetWidget   TargetKind = "widget"
	TargetURL      TargetKind = "url"
	TargetQuery    TargetKind = "query"
)

// Keywords are words, not symbols: "!" starts a bang and other symbols are
// left to the search field, so a keyword can never shadow those prefixes.
var keywordWordPattern = regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{N}_-]{0,23}$`)

const (
	maxTargetLength = 2000
	maxLabelLength  = 100
	maxIDLength     = 100
)

var (
	errKeywordWord   = kernel.NewValidationError("keyword", "A keyword is 1 to 24 letters, digits, - or _, starting with a letter or digit.")
	errKeywordKind   = kernel.NewValidationError("kind", "The target must be a service, bookmark, widget, web address or search.")
	errKeywordTarget = kernel.NewValidationError("target", "The target is missing or too long.")
	errKeywordURL    = kernel.NewValidationError("target", "The web address must start with http:// or https://.")
	errKeywordLabel  = kernel.NewValidationError("label", "The label can be at most 100 characters long.")
)

// Keyword is an own keyword of an account (#478): typing it in the search
// field and pressing Enter opens its target, e.g. "nas" → the NAS tile.
type Keyword struct {
	word   string
	kind   TargetKind
	target string
	label  string
}

// NewKeyword validates and creates a keyword. The word is trimmed and
// lower-cased; the label is what the search shows next to it (the name of
// the service or bookmark at the time it was chosen) and may be empty.
func NewKeyword(word string, kind TargetKind, target, label string) (Keyword, error) {
	word = strings.ToLower(strings.TrimSpace(word))
	if !keywordWordPattern.MatchString(word) {
		return Keyword{}, errKeywordWord
	}
	target = strings.TrimSpace(target)
	if target == "" || len(target) > maxTargetLength {
		return Keyword{}, errKeywordTarget
	}
	switch kind {
	case TargetURL:
		u, err := url.Parse(target)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return Keyword{}, errKeywordURL
		}
	case TargetService, TargetBookmark, TargetWidget:
		if len(target) > maxIDLength || strings.ContainsAny(target, " \t\n") {
			return Keyword{}, errKeywordTarget
		}
	case TargetQuery:
	default:
		return Keyword{}, errKeywordKind
	}
	label = strings.TrimSpace(label)
	if utf8.RuneCountInString(label) > maxLabelLength {
		return Keyword{}, errKeywordLabel
	}
	return Keyword{word: word, kind: kind, target: target, label: label}, nil
}

// Word returns the keyword as typed, lower-cased.
func (k Keyword) Word() string { return k.word }

// Kind returns what the keyword opens.
func (k Keyword) Kind() TargetKind { return k.kind }

// Target returns the id, web address or search text.
func (k Keyword) Target() string { return k.target }

// Label returns the display name of the target; may be empty.
func (k Keyword) Label() string { return k.label }

// CheckKeywordConflict rejects a keyword that is one of the quick capture
// prefixes: "t milk" must keep creating a to-do, whatever keywords exist.
func CheckKeywordConflict(word string, p CapturePrefixes) error {
	switch word {
	case p.Todo:
		return kernel.NewValidationErrorf("keyword", "“{word}” is the quick capture prefix for to-dos; choose another keyword or change the prefix.", "word", word)
	case p.Note:
		return kernel.NewValidationErrorf("keyword", "“{word}” is the quick capture prefix for notes; choose another keyword or change the prefix.", "word", word)
	case p.Bookmark:
		return kernel.NewValidationErrorf("keyword", "“{word}” is the quick capture prefix for bookmarks; choose another keyword or change the prefix.", "word", word)
	}
	return nil
}
