package answers

import (
	"cmp"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Kind is the shape of an answer. A new kind is added only when at least
// three sources need it.
type Kind string

// The kinds of answers.
const (
	Metric  Kind = "metric"
	Status  Kind = "status"
	List    Kind = "list"
	Text    Kind = "text"
	Actions Kind = "actions"
)

// SlotType is what a slot of an intent holds.
type SlotType int

// The slot types.
const (
	SlotEntity SlotType = iota
	SlotPlace
	SlotDay
	SlotNumber
	SlotAction
)

// Slot is a typed part of a query, e.g. the container in "restart jellyfin".
type Slot struct {
	Name     string
	Type     SlotType
	Required bool
	// EntityKind is the kind of entity slots, e.g. "container".
	EntityKind string
	// Alone adds points when the filled slot (with the intent's terms) makes
	// up the whole query: "morgen" alone asks for tomorrow's agenda.
	Alone float64
}

// Term is a folded word with its weight for an intent.
type Term struct {
	Word   string
	Weight float64
}

// Terms is a helper for weighted terms: Terms(3, "muell", "abfall").
func Terms(weight float64, words ...string) []Term {
	out := make([]Term, len(words))
	for i, w := range words {
		out[i] = Term{Word: w, Weight: weight}
	}
	return out
}

// Intent is one kind of question as data.
type Intent struct {
	ID string
	// Source is the ID of the Source that answers the intent.
	Source string
	Terms  []Term
	Slots  []Slot
	Answer Kind
	// WidgetKind is the widget that shows the details ("open widget").
	WidgetKind string
	// Prefixes select the intent directly, e.g. "cpu" for "/cpu".
	Prefixes []string
}

// Entity is something with a name the user can refer to (container, service).
type Entity struct {
	Kind, ID, Name string
}

// Action is a state change a query asks for.
type Action string

// The actions.
const (
	Start   Action = "start"
	Stop    Action = "stop"
	Restart Action = "restart"
)

// SlotValue is a filled slot; only the field of its type is set.
type SlotValue struct {
	Type   SlotType
	Entity Entity
	Text   string
	Offset int
	Number float64
	Action Action
}

// Match is an interpreted query: intent, score and filled slots.
type Match struct {
	Intent Intent
	Score  float64
	Slots  map[string]SlotValue
	// Missing are required slots that could not be filled (asks back).
	Missing []string
	// Suggestions are entities of the missing kind, for asking back.
	Suggestions []Entity
}

// EntitySlot returns the entity in slot name, if filled.
func (m Match) EntitySlot(name string) (Entity, bool) {
	v, ok := m.Slots[name]
	return v.Entity, ok && v.Type == SlotEntity
}

// Threshold is the minimum score of an answer; MaxAnswers the number shown.
const (
	Threshold  = 1.0
	MaxAnswers = 3
)

var actionWords = map[string]Action{
	"restart": Restart, "neustart": Restart, "neustarten": Restart, "reboot": Restart,
	"start": Start, "starte": Start, "starten": Start,
	"stop": Stop, "stopp": Stop, "stoppe": Stop, "stoppen": Stop, "anhalten": Stop,
}

var dayWords = map[string]int{
	"heute": 0, "today": 0, "jetzt": 0, "now": 0, "morgen": 1, "tomorrow": 1, "ubermorgen": 2,
}

// weekdays by time.Weekday; only the full names count, two-letter
// abbreviations collide with words.
var weekdays = [7][2]string{
	{"sonntag", "sunday"}, {"montag", "monday"}, {"dienstag", "tuesday"}, {"mittwoch", "wednesday"},
	{"donnerstag", "thursday"}, {"freitag", "friday"}, {"samstag", "saturday"},
}

// DayOffset returns the days from today to a day word ("morgen", "freitag").
func DayOffset(token string, today time.Time) (int, bool) {
	if d, ok := dayWords[token]; ok {
		return d, true
	}
	for wd, names := range weekdays {
		if token == names[0] || token == names[1] {
			return (wd - int(today.Weekday()) + 7) % 7, true
		}
	}
	return 0, false
}

// joinPhrases joins "neu starten" to one action word.
func joinPhrases(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "neu" && i+1 < len(tokens) && strings.HasPrefix(tokens[i+1], "start") {
			out = append(out, "neustarten") // the verb; "Neustart" alone is the noun (uptime)
			i++
			continue
		}
		out = append(out, tokens[i])
	}
	return out
}

type entityHit struct {
	entity  Entity
	quality float64
	used    int
}

// findEntity returns the best entity of a kind among the unused words; a
// word may match the whole name or one of its words ("home" for
// "Home Assistant").
func findEntity(tokens []string, used map[int]bool, entities []Entity, kind string) (entityHit, bool) {
	var best entityHit
	found := false
	for _, e := range entities {
		if e.Kind != kind {
			continue
		}
		name := Fold(e.Name)
		parts := Tokenize(e.Name)
		for i, token := range tokens {
			if used[i] {
				continue
			}
			q := MatchToken(token, name)
			for _, p := range parts {
				if q != NoMatch {
					break
				}
				q = MatchToken(token, p)
			}
			if q == NoMatch {
				continue
			}
			quality := float64(q)
			if name == token {
				quality *= 1.2
			}
			if !found || quality > best.quality {
				best, found = entityHit{entity: e, quality: quality, used: i}, true
			}
		}
	}
	return best, found
}

var numberRe = regexp.MustCompile(`^\d+([.,]\d+)?$`)

// interpretIntent interprets the words against one intent.
func interpretIntent(tokens []string, in Intent, entities []Entity, today time.Time, explicit bool) (Match, bool) {
	used := map[int]bool{}
	score := 0.0
	if explicit {
		score = 100
	}
	// Points from the intent's own words; a match that lacks a required
	// slot is only useful when the user asked for it.
	trigger := score
	if !explicit {
		// Every (term, word) pair that matches, best first: each word and
		// each term count once, so "containers" goes to the exact term.
		type pair struct {
			term, token int
			score       float64
		}
		var pairs []pair
		for ti, t := range in.Terms {
			for i, tok := range tokens {
				if q := MatchToken(tok, t.Word); q != NoMatch {
					pairs = append(pairs, pair{ti, i, t.Weight * float64(q)})
				}
			}
		}
		slices.SortStableFunc(pairs, func(a, b pair) int {
			return cmp.Or(cmp.Compare(b.score, a.score), cmp.Compare(a.term, b.term), cmp.Compare(a.token, b.token))
		})
		termsUsed := map[int]bool{}
		for _, p := range pairs {
			if termsUsed[p.term] || used[p.token] {
				continue
			}
			termsUsed[p.term], used[p.token] = true, true
			score += p.score
			trigger += p.score
		}
	}
	slots := map[string]SlotValue{}
	var missing []string
	var suggestions []Entity
	for _, slot := range in.Slots {
		var value *SlotValue
		switch slot.Type {
		case SlotEntity:
			if hit, ok := findEntity(tokens, used, entities, slot.EntityKind); ok {
				used[hit.used] = true
				value = &SlotValue{Type: SlotEntity, Entity: hit.entity}
				// A known name counts strongly: "jellyfin" alone selects its status.
				score += 2 * hit.quality
			} else {
				suggestions = nil
				for _, e := range entities {
					if e.Kind == slot.EntityKind && len(suggestions) < 5 {
						suggestions = append(suggestions, e)
					}
				}
			}
		case SlotAction:
			for i, t := range tokens {
				a, ok := actionWords[t]
				if used[i] || !ok {
					continue
				}
				used[i] = true
				// "starte pihole neu" is a restart.
				if a == Start && slices.Contains(tokens, "neu") {
					a = Restart
				}
				value = &SlotValue{Type: SlotAction, Action: a}
				score++
				trigger++
				break
			}
		case SlotDay:
			for i, t := range tokens {
				if d, ok := DayOffset(t, today); !used[i] && ok {
					used[i] = true
					value = &SlotValue{Type: SlotDay, Offset: d}
					break
				}
			}
		case SlotNumber:
			for i, t := range tokens {
				if !used[i] && numberRe.MatchString(t) {
					used[i] = true
					n, _ := strconv.ParseFloat(strings.Replace(t, ",", ".", 1), 64)
					value = &SlotValue{Type: SlotNumber, Number: n}
					break
				}
			}
		case SlotPlace:
			// Whatever is left after terms and the other slots, in the order typed.
			hasDay := slices.ContainsFunc(in.Slots, func(s Slot) bool { return s.Type == SlotDay })
			var words []string
			for i, t := range tokens {
				if used[i] {
					continue
				}
				if _, isAction := actionWords[t]; isAction {
					continue
				}
				if _, isDay := DayOffset(t, today); hasDay && isDay {
					continue
				}
				words = append(words, t)
			}
			if len(words) > 0 {
				value = &SlotValue{Type: SlotPlace, Text: strings.Join(words, " ")}
			}
		}
		if value != nil {
			slots[slot.Name] = *value
		} else if slot.Required {
			missing = append(missing, slot.Name)
		}
	}
	alone := 0.0
	for _, s := range in.Slots {
		if _, ok := slots[s.Name]; ok {
			alone += s.Alone
		}
	}
	if alone > 0 && len(used) == len(tokens) {
		score += alone
		trigger += alone
	}
	if score <= 0 || (len(missing) > 0 && trigger == 0) {
		return Match{}, false
	}
	if len(missing) == 0 {
		suggestions = nil
	}
	return Match{Intent: in, Score: score, Slots: slots, Missing: missing, Suggestions: suggestions}, true
}

// Interpret reads a query: an explicit prefix ("/cpu", "/docker jellyfin")
// selects its intents directly; otherwise every intent collects points for
// matching words and entity names.
func Interpret(query string, intents []Intent, entities []Entity, today time.Time) []Match {
	trimmed := strings.TrimSpace(query)
	if rest, ok := strings.CutPrefix(trimmed, "/"); ok {
		head, tail, _ := strings.Cut(rest, " ")
		prefix := Fold(head)
		tokens := joinPhrases(Tokenize(tail))
		var best *Match
		for _, in := range intents {
			if !slices.Contains(in.Prefixes, prefix) {
				continue
			}
			m, ok := interpretIntent(tokens, in, entities, today, true)
			if !ok {
				continue
			}
			// "/docker jellyfin" is the status, "/docker restart jellyfin" the action.
			m.Score += float64(len(m.Slots)) - 2*float64(len(m.Missing))
			if best == nil || m.Score > best.Score {
				best = &m
			}
		}
		if best != nil {
			return []Match{*best}
		}
		if slices.ContainsFunc(intents, func(in Intent) bool { return slices.Contains(in.Prefixes, prefix) }) {
			return nil
		}
	}
	tokens := joinPhrases(Tokenize(trimmed))
	if len(tokens) == 0 {
		return nil
	}
	var out []Match
	for _, in := range intents {
		if m, ok := interpretIntent(tokens, in, entities, today, false); ok {
			out = append(out, m)
		}
	}
	return out
}

// Rank keeps matches above the threshold, best first, at most MaxAnswers.
func Rank(matches []Match) []Match {
	var out []Match
	for _, m := range matches {
		if m.Score >= Threshold {
			out = append(out, m)
		}
	}
	slices.SortStableFunc(out, func(a, b Match) int {
		return cmp.Or(cmp.Compare(b.Score, a.Score), cmp.Compare(a.Intent.ID, b.Intent.ID))
	})
	if len(out) > MaxAnswers {
		out = out[:MaxAnswers]
	}
	return out
}
