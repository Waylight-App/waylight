package usecase

import (
	"context"
	"net/http"

	"waylight/ee/presence/domain"
	"waylight/internal/shared/kernel"
)

// Compact is the short state of a presence widget (#632): how many people
// of the group are in the office and at home, the number away as detail,
// and the viewer's own status as chips when they may set it. Who is where
// is neither good nor bad, so the level stays none. changes is false while
// the licence only allows reading, which hides the chips.
func (s *Service) Compact(ctx context.Context, groupID string, changes bool) (kernel.CompactState, error) {
	b, err := s.Board(ctx, groupID)
	if err != nil {
		return kernel.CompactState{}, err
	}
	c := b.Counts
	state := kernel.CompactState{
		Title: kernel.NewTextf("{office} in the office · {home} at home",
			"office", c[domain.StatusOffice], "home", c[domain.StatusHome]),
		Level:     kernel.CompactNone,
		UpdatedAt: s.d.Clock.Now(),
	}
	if n := c[domain.StatusAway]; n > 0 {
		state.Detail = kernel.NewTextf("{count} away", "count", n)
	}
	if changes && b.Mine.AccountID != "" {
		state.Action = statusAction(b.Mine.Status)
	}
	return state, nil
}

// statusAction offers office, home and away as chips, the current status
// selected; each chip sends the body PUT /api/presence/me expects.
func statusAction(current domain.Status) *kernel.CompactAction {
	labels := map[domain.Status]kernel.Text{
		domain.StatusOffice: kernel.NewText("Office"),
		domain.StatusHome:   kernel.NewText("Home"),
		domain.StatusAway:   kernel.NewText("Away"),
	}
	a := &kernel.CompactAction{ID: "presence.set", Label: kernel.NewText("My status today"), Method: http.MethodPut, Path: "/api/presence/me"}
	for _, st := range domain.Settable {
		a.Choices = append(a.Choices, kernel.CompactChoice{ID: string(st), Label: labels[st],
			Body: map[string]any{"status": string(st)}, Selected: st == current})
	}
	return a
}
