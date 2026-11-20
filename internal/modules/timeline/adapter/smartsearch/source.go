// Package smartsearch answers "what happened" in the smart search
// (ADR 0026) with the latest events of the timeline.
package smartsearch

import (
	"context"
	"fmt"

	"waylight/internal/modules/timeline/domain"
	"waylight/internal/modules/timeline/usecase"
	"waylight/internal/shared/answers"
)

// maxItems is the number of events in an answer.
const maxItems = 5

var intents = []answers.Intent{
	{ID: "timeline.recent", Source: "timeline", Prefixes: []string{"timeline"}, Answer: answers.List, WidgetKind: "timeline",
		Terms: append(append(answers.Terms(3, "timeline", "vorfalle", "vorfall", "ereignisse", "incidents", "events", "verlauf"),
			answers.Terms(2, "passiert", "happened", "los")...),
			answers.Terms(1, "zuletzt", "recent", "recently", "neues")...)},
}

var messages = answers.Messages{
	answers.English: {"title": "Latest events", "empty": "Nothing happened yet."},
	answers.German:  {"title": "Letzte Ereignisse", "empty": "Bisher ist nichts passiert."},
}

var states = map[domain.Severity]answers.State{
	domain.SeverityInfo: answers.StateOff, domain.SeverityWarning: answers.StateWarning, domain.SeverityCritical: answers.StateError,
}

// Source answers from the timeline.
type Source struct{ timeline usecase.Timeline }

// New returns the source on timeline.
func New(timeline usecase.Timeline) *Source { return &Source{timeline: timeline} }

// ID implements answers.Source.
func (*Source) ID() string { return "timeline" }

// Intents implements answers.Source.
func (*Source) Intents() []answers.Intent { return intents }

// Entities implements answers.Source.
func (*Source) Entities(context.Context) ([]answers.Entity, error) { return nil, nil }

// Resolve implements answers.Source.
func (s *Source) Resolve(ctx context.Context, r answers.Request) (answers.Answer, error) {
	events, err := s.timeline.List(ctx, usecase.Filter{Limit: maxItems})
	if err != nil {
		return answers.Answer{}, fmt.Errorf("timeline: %w", err)
	}
	l := r.Lang
	a := answers.Answer{Kind: answers.List, Title: messages.T(l, "title"), Empty: messages.T(l, "empty")}
	if len(events) > 0 {
		a.At = events[0].At
	}
	for _, e := range events {
		value := answers.Ago(l, e.At, r.Now)
		if e.Message != "" {
			value += " · " + e.Message
		}
		a.Items = append(a.Items, answers.Item{Label: e.Subject, Value: value, State: states[e.Severity]})
	}
	return a, nil
}
