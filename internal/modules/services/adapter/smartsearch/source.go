// Package smartsearch answers questions about the monitored services in the
// smart search (ADR 0026): which are down and the status of one.
package smartsearch

import (
	"context"
	"fmt"
	"time"

	"waylight/internal/modules/services/usecase"
	"waylight/internal/shared/answers"
)

var intents = []answers.Intent{
	{ID: "services.overview", Source: "services", Answer: answers.List, WidgetKind: "services",
		Terms: append(append(answers.Terms(3, "dienste", "services", "dienst"),
			answers.Terms(2, "offline", "down", "erreichbar", "reachable", "ausgefallen", "alles", "everything")...),
			answers.Terms(1, "online", "up", "ok")...)},
	{ID: "services.status", Source: "services", Prefixes: []string{"service"}, Answer: answers.Status, WidgetKind: "services",
		Slots: []answers.Slot{{Name: "service", Type: answers.SlotEntity, EntityKind: "service", Required: true}},
		Terms: answers.Terms(1, "status", "online", "erreichbar", "antwortzeit", "latency", "ping")},
}

var messages = answers.Messages{
	answers.English: {
		"services": "Services: {online} of {total} online", "all_online": "All services are online.", "none": "No services yet.",
		"gone": "This service no longer exists.", "online": "online", "offline": "offline", "unknown": "not checked yet",
	},
	answers.German: {
		"services": "Dienste: {online} von {total} online", "all_online": "Alle Dienste sind online.", "none": "Noch keine Dienste.",
		"gone": "Diesen Dienst gibt es nicht mehr.", "online": "online", "offline": "offline", "unknown": "noch nicht geprüft",
	},
}

var states = map[string]answers.State{"online": answers.StateOK, "offline": answers.StateError}

func state(s string) answers.State {
	if st, ok := states[s]; ok {
		return st
	}
	return answers.StateOff
}

// Source answers from the latest checks of the service tiles.
type Source struct{ services usecase.Services }

// New returns the source on services.
func New(services usecase.Services) *Source { return &Source{services: services} }

// ID implements answers.Source.
func (*Source) ID() string { return "services" }

// Intents implements answers.Source.
func (*Source) Intents() []answers.Intent { return intents }

// Entities implements answers.Source: the visible services by name.
func (s *Source) Entities(ctx context.Context) ([]answers.Entity, error) {
	list, err := s.visible(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]answers.Entity, len(list))
	for i, v := range list {
		out[i] = answers.Entity{Kind: "service", ID: v.ID, Name: v.Name}
	}
	return out, nil
}

func (s *Source) visible(ctx context.Context) ([]usecase.ServiceView, error) {
	list, err := s.services.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("services: %w", err)
	}
	out := list[:0:0]
	for _, v := range list {
		if !v.Hidden {
			out = append(out, v)
		}
	}
	return out, nil
}

// ownServices leaves out the example tiles of the sample data, which do not
// count towards the overview (#854).
func ownServices(list []usecase.ServiceView) []usecase.ServiceView {
	out := list[:0:0]
	for _, v := range list {
		if !v.Example {
			out = append(out, v)
		}
	}
	return out
}

// Resolve implements answers.Source.
func (s *Source) Resolve(ctx context.Context, r answers.Request) (answers.Answer, error) {
	list, err := s.visible(ctx)
	if err != nil {
		return answers.Answer{}, err
	}
	l := r.Lang
	if r.Match.Intent.ID == "services.overview" {
		list = ownServices(list)
		var latest time.Time
		a := answers.Answer{Kind: answers.List, Empty: messages.T(l, "all_online")}
		for _, v := range list {
			if v.CheckedAt.After(latest) {
				latest = v.CheckedAt
			}
			if v.State == "online" {
				continue
			}
			value := v.Error
			if value == "" {
				value = messages.T(l, stateKey(v.State))
			}
			a.Items = append(a.Items, answers.Item{Label: v.Name, Value: value, State: state(v.State)})
		}
		a.At = latest
		a.Title = messages.T(l, "services", "online", len(list)-len(a.Items), "total", len(list))
		if len(list) == 0 {
			a.Empty = messages.T(l, "none")
		}
		return a, nil
	}
	target, _ := r.Match.EntitySlot("service")
	for _, v := range list {
		if v.ID != target.ID {
			continue
		}
		detail := v.Error
		if v.State == "online" {
			detail = fmt.Sprintf("%d ms", v.LatencyMs)
		}
		if v.URL != "" {
			if detail != "" {
				detail += " · "
			}
			detail += v.URL
		}
		return answers.Answer{
			Kind: answers.Status, At: v.CheckedAt, Title: v.Name, State: state(v.State),
			Text: messages.T(l, stateKey(v.State)), Detail: detail,
		}, nil
	}
	return answers.Answer{Kind: answers.Text, Title: target.Name, Text: messages.T(l, "gone")}, nil
}

func stateKey(s string) string {
	if s == "online" || s == "offline" {
		return s
	}
	return "unknown"
}
