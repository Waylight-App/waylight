// Package smartsearch answers "next bus" and "departures <stop>" in the
// smart search (ADR 0026). A named stop is looked up; without one the stop
// of the account's first transit widget is used.
package smartsearch

import (
	"context"
	"fmt"
	"strings"

	"waylight/internal/modules/transit/usecase"
	"waylight/internal/shared/answers"
)

// maxItems is the longest list in an answer.
const maxItems = 6

var intents = []answers.Intent{
	{ID: "transit.departures", Source: "transit", Prefixes: []string{"abfahrt", "bus", "departures"}, Answer: answers.List,
		WidgetKind: "transit",
		Slots:      []answers.Slot{{Name: "place", Type: answers.SlotPlace}},
		Terms: append(append(
			answers.Terms(3, "abfahrt", "abfahrten", "departure", "departures", "bus", "tram", "strassenbahn", "ubahn", "sbahn"),
			answers.Terms(2, "bahn", "zug", "train", "haltestelle", "stop")...),
			answers.Terms(1, "nachste", "next", "fahrt")...)},
}

var messages = answers.Messages{
	answers.English: {
		"departures": "Departures", "no_stops": "No stops found.", "empty": "No departures in the near future.",
		"cancelled":  "cancelled",
		"which_stop": `Name a stop, e.g. "departures Oldenburg Hbf", or choose one in the transit widget.`,
	},
	answers.German: {
		"departures": "Abfahrten", "no_stops": "Keine Haltestellen gefunden.", "empty": "In nächster Zeit keine Abfahrten.",
		"cancelled":  "fällt aus",
		"which_stop": "Nenne eine Haltestelle, z. B. „Abfahrt Oldenburg Hbf“, oder wähle eine im ÖPNV-Widget.",
	},
}

// Source answers from the journey planner.
type Source struct{ transit usecase.Transit }

// New returns the source on transit.
func New(transit usecase.Transit) *Source { return &Source{transit: transit} }

// ID implements answers.Source.
func (*Source) ID() string { return "transit" }

// Intents implements answers.Source.
func (*Source) Intents() []answers.Intent { return intents }

// Entities implements answers.Source.
func (*Source) Entities(context.Context) ([]answers.Entity, error) { return nil, nil }

// Resolve implements answers.Source.
func (s *Source) Resolve(ctx context.Context, r answers.Request) (answers.Answer, error) {
	l := r.Lang
	var stopID, stopName string
	if named := r.Match.Slots["place"].Text; named != "" {
		stops, err := s.transit.SearchStops(ctx, named)
		if err != nil {
			return answers.Answer{}, fmt.Errorf("search stops: %w", err)
		}
		if len(stops) == 0 {
			return answers.Answer{Kind: answers.Text, Title: named, Text: messages.T(l, "no_stops")}, nil
		}
		stopID, stopName = stops[0].ID, stops[0].Name
	} else {
		cfg := r.WidgetConfig(ctx, "transit")
		stopID, _ = cfg["stopId"].(string)
		stopName, _ = cfg["stopName"].(string)
	}
	if stopID == "" {
		return answers.Answer{Kind: answers.Text, Title: messages.T(l, "departures"), Text: messages.T(l, "which_stop")}, nil
	}
	deps, err := s.transit.Departures(ctx, stopID, maxItems)
	if err != nil {
		return answers.Answer{}, fmt.Errorf("departures: %w", err)
	}
	if stopName == "" {
		stopName = messages.T(l, "departures")
	}
	a := answers.Answer{Kind: answers.List, Title: stopName, Empty: messages.T(l, "empty"), At: r.Now}
	for _, d := range deps {
		item := answers.Item{Label: d.Line + " → " + d.Headsign}
		switch {
		case d.Cancelled:
			item.Value, item.State = messages.T(l, "cancelled"), answers.StateError
		default:
			parts := []string{answers.MinutesUntil(l, d.When(), r.Now)}
			if delay := d.DelayMinutes(); delay > 0 {
				parts = append(parts, fmt.Sprintf("+%d", delay))
				if delay >= 5 {
					item.State = answers.StateWarning
				}
			}
			if d.Track != "" {
				parts = append(parts, d.Track)
			}
			item.Value = strings.Join(parts, " · ")
		}
		a.Items = append(a.Items, item)
	}
	return a, nil
}
