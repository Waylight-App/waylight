package answers

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// foldChain removes diacritics after canonical decomposition.
var foldChain = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// Fold lower-cases s and removes diacritics, so "Büro" matches "buro".
func Fold(s string) string {
	out, _, err := transform.String(foldChain, s)
	if err != nil {
		out = s
	}
	return strings.ReplaceAll(strings.ToLower(out), "ß", "ss")
}

// stopWords carry no meaning for an intent (German and English).
var stopWords = func() map[string]bool {
	words := "der die das den dem des ein eine einen einem einer ist sind wie was wo wann wer welche welcher welches " +
		"mein meine meinen meiner mir mich ich du es im in am an auf zu zum zur mit von vom fur bei noch gerade " +
		"jetzt aktuell bitte mal denn doch ja und oder " +
		// Auxiliary verbs carry no topic; as two-letter prefixes ("do")
		// they would otherwise match terms like "docker".
		"habe hast hat haben muss musst soll sollte kann gibt schon " +
		// Verbs around a weather question would otherwise be read as a place.
		"brauche brauchst brauchen anziehen mitnehmen " +
		"the a an is are how what where when who which my me i you it in on at to of for with please now current " +
		"currently show tell and or do does did will would can could should need there be"
	m := map[string]bool{}
	for _, w := range strings.Fields(words) {
		m[w] = true
	}
	return m
}()

// Tokenize splits a query into folded words without punctuation and stop
// words.
func Tokenize(query string) []string {
	parts := strings.FieldsFunc(Fold(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '.' && r != '_' && r != '-'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(p, "._-")
		if p != "" && !stopWords[p] {
			out = append(out, p)
		}
	}
	return out
}

// Quality says how a word matched a term; exact > prefix > fuzzy.
type Quality float64

// The points per quality.
const (
	NoMatch Quality = 0
	Fuzzy   Quality = 0.6
	Prefix  Quality = 0.8
	Exact   Quality = 1
)

// minPrefix is the shortest word that matches as a prefix while typing.
const minPrefix = 2

// maxDistance is the allowed edit distance for a term length: 1 from 4,
// 2 from 8 characters.
func maxDistance(length int) int {
	switch {
	case length >= 8:
		return 2
	case length >= 4:
		return 1
	default:
		return 0
	}
}

// boundedDistance is the Damerau-Levenshtein distance (optimal string
// alignment). It returns limit+1 as soon as the distance must exceed the
// limit, so obvious mismatches never fill the whole matrix.
func boundedDistance(a, b []rune, limit int) int {
	if abs(len(a)-len(b)) > limit {
		return limit + 1
	}
	prev2 := make([]int, len(b)+1)
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		rowMin := i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			v := min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				v = min(v, prev2[j-2]+1)
			}
			cur[j] = v
			rowMin = min(rowMin, v)
		}
		if rowMin > limit {
			return limit + 1
		}
		prev2, prev = prev, cur
	}
	return prev[len(b)]
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// MatchToken matches a folded word against a folded term: exactly, as a
// prefix while typing (from two characters), or with a few typos. Words of
// up to three characters only match exactly or by prefix, to avoid noise.
func MatchToken(token, term string) Quality {
	if token == term {
		return Exact
	}
	t, w := []rune(token), []rune(term)
	if len(t) >= minPrefix && strings.HasPrefix(term, token) {
		return Prefix
	}
	if len(t) <= 3 || len(w) == 0 {
		return NoMatch
	}
	// Typos rarely hit the first letter; without this "serien" would be "ferien".
	if t[0] != w[0] {
		return NoMatch
	}
	limit := maxDistance(len(w))
	if limit == 0 || boundedDistance(t, w, limit) > limit {
		return NoMatch
	}
	return Fuzzy
}
