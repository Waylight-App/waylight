// Package smartsearch answers "when is the bin collected" in the smart
// search (ADR 0026).
package smartsearch

import (
	"context"
	"fmt"
	"strings"

	"waylight/internal/modules/waste/usecase"
	"waylight/internal/shared/answers"
)

// maxItems is the longest list in an answer.
const maxItems = 6

var intents = []answers.Intent{
	{ID: "waste.next", Source: "waste", Prefixes: []string{"muell", "mull", "abfall", "waste"}, Answer: answers.List, WidgetKind: "waste",
		Terms: append(append(answers.Terms(3, "mull", "muell", "abfall", "mullabfuhr", "tonne", "waste", "garbage", "trash", "bin"),
			answers.Terms(2, "gelbe", "gelber", "sack", "biotonne", "restmull", "papiertonne", "recycling")...),
			answers.Terms(1, "abholung", "abfuhr", "collection", "raus", "out")...)},
}

var messages = answers.Messages{
	answers.English: {"title": "Waste collection", "empty": "No upcoming collections."},
	answers.German:  {"title": "Müllabfuhr", "empty": "Keine anstehenden Abholungen."},
}

// Source answers from the next collections.
type Source struct{ waste usecase.Waste }

// New returns the source on waste.
func New(waste usecase.Waste) *Source { return &Source{waste: waste} }

// ID implements answers.Source.
func (*Source) ID() string { return "waste" }

// Intents implements answers.Source.
func (*Source) Intents() []answers.Intent { return intents }

// Entities implements answers.Source.
func (*Source) Entities(context.Context) ([]answers.Entity, error) { return nil, nil }

// Resolve implements answers.Source. The types come ordered by the next
// date; tomorrow and today need attention.
func (s *Source) Resolve(ctx context.Context, r answers.Request) (answers.Answer, error) {
	o, err := s.waste.Upcoming(ctx)
	if err != nil {
		return answers.Answer{}, fmt.Errorf("waste: %w", err)
	}
	l := r.Lang
	a := answers.Answer{Kind: answers.List, Title: messages.T(l, "title"), Empty: o.Error, At: o.UpdatedAt}
	if a.Empty == "" {
		a.Empty = messages.T(l, "empty")
	}
	for i, w := range o.Types {
		if i >= maxItems {
			break
		}
		item := answers.Item{Label: strings.TrimSpace(w.Style.Icon + " " + w.Style.Label), Value: answers.InDays(l, w.DaysUntil)}
		if w.DaysUntil <= 1 {
			item.State = answers.StateWarning
		}
		a.Items = append(a.Items, item)
	}
	return a, nil
}
