// Package smartsearch answers questions about containers in the smart
// search (ADR 0026): the overview and the status of one container.
package smartsearch

import (
	"context"
	"fmt"
	"strings"

	"waylight/internal/modules/docker/usecase"
	"waylight/internal/shared/answers"
)

var intents = []answers.Intent{
	{ID: "docker.overview", Source: "docker", Answer: answers.List, WidgetKind: "docker",
		Terms: append(answers.Terms(3, "docker", "container", "containers"),
			answers.Terms(1, "laufen", "laufende", "running", "gestoppt", "stopped", "down", "kaputt")...)},
	{ID: "docker.status", Source: "docker", Prefixes: []string{"docker"}, Answer: answers.Status, WidgetKind: "docker",
		Slots: []answers.Slot{{Name: "container", Type: answers.SlotEntity, EntityKind: "container", Required: true}},
		Terms: answers.Terms(1, "status", "lauft", "laeuft", "running", "image", "update")},
}

var messages = answers.Messages{
	answers.English: {
		"containers": "Containers: {running} of {total} running", "no_containers": "All containers are running.",
		"gone": "This container no longer exists.",
	},
	answers.German: {
		"containers": "Container: {running} von {total} laufen", "no_containers": "Alle Container laufen.",
		"gone": "Diesen Container gibt es nicht mehr.",
	},
}

// Source answers from the cached container list.
type Source struct{ containers usecase.Containers }

// New returns the source on containers.
func New(containers usecase.Containers) *Source { return &Source{containers: containers} }

// ID implements answers.Source.
func (*Source) ID() string { return "docker" }

// Intents implements answers.Source.
func (*Source) Intents() []answers.Intent { return intents }

// Entities implements answers.Source: the containers by name.
func (s *Source) Entities(ctx context.Context) ([]answers.Entity, error) {
	list, err := s.containers.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("containers: %w", err)
	}
	out := make([]answers.Entity, len(list.Containers))
	for i, c := range list.Containers {
		out[i] = answers.Entity{Kind: "container", ID: c.ID, Name: c.Name}
	}
	return out, nil
}

// indicator maps state and health to a traffic light, as the widget does.
func indicator(c usecase.ContainerView) answers.State {
	switch {
	case c.State == "restarting" || c.State == "dead":
		return answers.StateError
	case c.State != "running":
		return answers.StateOff
	case c.Health == "unhealthy":
		return answers.StateError
	case c.Health == "starting":
		return answers.StateWarning
	default:
		return answers.StateOK
	}
}

// Resolve implements answers.Source.
func (s *Source) Resolve(ctx context.Context, r answers.Request) (answers.Answer, error) {
	list, err := s.containers.List(ctx)
	if err != nil {
		return answers.Answer{}, fmt.Errorf("containers: %w", err)
	}
	l := r.Lang
	if r.Match.Intent.ID == "docker.overview" {
		a := answers.Answer{
			Kind: answers.List, At: list.UpdatedAt, Empty: messages.T(l, "no_containers"),
			Title: messages.T(l, "containers", "running", list.Running, "total", list.Total),
		}
		for _, c := range list.Containers {
			if st := indicator(c); st != answers.StateOK {
				a.Items = append(a.Items, answers.Item{Label: c.Name, Value: c.Status, State: st})
			}
		}
		return a, nil
	}
	target, _ := r.Match.EntitySlot("container")
	for _, c := range list.Containers {
		if c.ID == target.ID || c.Name == target.Name {
			return answers.Answer{
				Kind: answers.Status, At: list.UpdatedAt, Title: c.Name, State: indicator(c), Text: c.Status,
				Detail: strings.TrimSpace(c.Image),
			}, nil
		}
	}
	return answers.Answer{Kind: answers.Text, Title: target.Name, Text: messages.T(l, "gone")}, nil
}
