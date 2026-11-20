// Package smartsearch answers "whose birthday is next" in the smart search
// (ADR 0026).
package smartsearch

import (
	"context"
	"fmt"

	"waylight/internal/modules/birthdays/usecase"
	"waylight/internal/shared/answers"
)

// maxItems is the longest list in an answer.
const maxItems = 6

var intents = []answers.Intent{
	{ID: "birthdays.upcoming", Source: "birthdays", Prefixes: []string{"geburtstag", "birthday"}, Answer: answers.List,
		WidgetKind: "birthdays",
		Terms: append(answers.Terms(3, "geburtstag", "geburtstage", "birthday", "birthdays", "bday"),
			answers.Terms(1, "whose", "naechste", "nachste", "next", "gratulieren")...)},
}

var messages = answers.Messages{
	answers.English: {"title": "Birthdays", "empty": "No birthdays saved yet.", "turns": "turns {age}"},
	answers.German:  {"title": "Geburtstage", "empty": "Noch keine Geburtstage gespeichert.", "turns": "wird {age}"},
}

// Source answers from the account's birthdays.
type Source struct{ birthdays usecase.Birthdays }

// New returns the source on birthdays.
func New(birthdays usecase.Birthdays) *Source { return &Source{birthdays: birthdays} }

// ID implements answers.Source.
func (*Source) ID() string { return "birthdays" }

// Intents implements answers.Source.
func (*Source) Intents() []answers.Intent { return intents }

// Entities implements answers.Source.
func (*Source) Entities(context.Context) ([]answers.Entity, error) { return nil, nil }

// Resolve implements answers.Source.
func (s *Source) Resolve(ctx context.Context, r answers.Request) (answers.Answer, error) {
	list, err := s.birthdays.Upcoming(ctx, maxItems)
	if err != nil {
		return answers.Answer{}, fmt.Errorf("upcoming birthdays: %w", err)
	}
	l := r.Lang
	a := answers.Answer{Kind: answers.List, Title: messages.T(l, "title"), Empty: messages.T(l, "empty")}
	for _, b := range list {
		value := answers.InDays(l, b.InDays)
		if b.Age > 0 {
			value += " · " + messages.T(l, "turns", "age", b.Age)
		}
		item := answers.Item{Label: b.Name, Value: value}
		if b.InDays == 0 {
			item.State = answers.StateOK
		}
		a.Items = append(a.Items, item)
	}
	return a, nil
}
