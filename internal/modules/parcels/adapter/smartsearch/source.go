// Package smartsearch answers "where is my parcel" in the smart search
// (ADR 0026).
package smartsearch

import (
	"context"
	"fmt"
	"time"

	"waylight/internal/modules/parcels/domain"
	"waylight/internal/modules/parcels/usecase"
	"waylight/internal/shared/answers"
)

// maxItems is the longest list in an answer.
const maxItems = 6

var intents = []answers.Intent{
	{ID: "parcels.status", Source: "parcels", Prefixes: []string{"paket", "parcel"}, Answer: answers.List, WidgetKind: "parcels",
		Terms: concat(
			answers.Terms(3, "paket", "pakete", "sendung", "sendungen", "parcel", "parcels", "package", "packages"),
			answers.Terms(3, "sendungsverfolgung", "paketverfolgung"),
			answers.Terms(2, "lieferung", "delivery", "dhl", "hermes", "dpd", "ups", "gls", "tracking", "track"),
			answers.Terms(1, "kommt", "coming", "zugestellt", "delivered"))},
}

func concat(parts ...[]answers.Term) []answers.Term {
	var out []answers.Term
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

var messages = answers.Messages{
	answers.English: {
		"title": "Parcels", "empty": "No tracked parcels", "eta": "expected {date}",
		"unknown": "Unknown", "pre-transit": "Announced", "transit": "On the way", "delivered": "Delivered", "failure": "Problem",
	},
	answers.German: {
		"title": "Pakete", "empty": "Keine verfolgten Pakete", "eta": "erwartet am {date}",
		"unknown": "Unbekannt", "pre-transit": "Angekündigt", "transit": "Unterwegs", "delivered": "Zugestellt", "failure": "Problem",
	},
}

var states = map[domain.StatusCode]answers.State{
	domain.StatusUnknown: answers.StateOff, domain.StatusDelivered: answers.StateOK, domain.StatusFailure: answers.StateError,
}

// Source answers from the tracked parcels.
type Source struct{ parcels usecase.Parcels }

// New returns the source on parcels.
func New(parcels usecase.Parcels) *Source { return &Source{parcels: parcels} }

// ID implements answers.Source.
func (*Source) ID() string { return "parcels" }

// Intents implements answers.Source.
func (*Source) Intents() []answers.Intent { return intents }

// Entities implements answers.Source.
func (*Source) Entities(context.Context) ([]answers.Entity, error) { return nil, nil }

// Resolve implements answers.Source.
func (s *Source) Resolve(ctx context.Context, r answers.Request) (answers.Answer, error) {
	list, err := s.parcels.List(ctx)
	if err != nil {
		return answers.Answer{}, fmt.Errorf("parcels: %w", err)
	}
	l := r.Lang
	a := answers.Answer{Kind: answers.List, Title: messages.T(l, "title"), Empty: messages.T(l, "empty")}
	var latest time.Time
	for i, p := range list {
		if p.Status.At.After(latest) {
			latest = p.Status.At
		}
		if i >= maxItems {
			continue
		}
		name := p.Label
		if name == "" {
			name = p.TrackingNumber
		}
		value := p.Error
		if value == "" {
			value = messages.T(l, string(p.Status.Code))
		}
		switch {
		case !p.EstimatedDelivery.IsZero() && !p.Delivered():
			value += " · " + messages.T(l, "eta", "date", answers.FormatDate(l, p.EstimatedDelivery))
		case p.Status.Location != "":
			value += " · " + p.Status.Location
		}
		item := answers.Item{Label: name, Value: value, State: states[p.Status.Code]}
		if p.Error != "" {
			item.State = answers.StateError
		}
		a.Items = append(a.Items, item)
	}
	a.At = latest
	return a, nil
}
