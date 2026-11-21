package usecase

import (
	"context"
	"net/http"
	"slices"

	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/shared/kernel"
)

// Compact implements SmartHome (#614): the name and state of one entity,
// with a toggle or press as quick action while actions are enabled for it.
// A state is neither good nor bad, so the level stays none.
func (uc *smartHome) Compact(ctx context.Context, id string) (kernel.CompactState, error) {
	if id == "" {
		return kernel.CompactState{}, kernel.NewValidationError("entity", "Choose one entity for the compact view.")
	}
	if !domain.ValidEntityID(id) {
		return kernel.CompactState{}, domain.ErrEntityNotFound
	}
	states, err := uc.current(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	i := slices.IndexFunc(states, func(e domain.Entity) bool { return e.ID == id })
	if i < 0 {
		return kernel.CompactState{}, domain.ErrEntityNotFound
	}
	e := uc.view(states[i])
	state := kernel.CompactState{
		Title:     kernel.NewTextf("{name}", "name", e.Name),
		Detail:    stateText(e.Entity),
		Level:     kernel.CompactNone,
		UpdatedAt: e.LastChanged,
	}
	if state.UpdatedAt.IsZero() {
		state.UpdatedAt = uc.d.Clock.Now()
	}
	if label, ok := actionLabel(e); ok {
		state.Action = &kernel.CompactAction{
			ID:     "smarthome." + string(e.Action),
			Label:  label,
			Method: http.MethodPost,
			// Entity IDs match [a-z0-9_.], so they need no escaping in a path.
			Path: "/api/smarthome/entities/" + e.ID + "/action",
		}
	}
	return state, nil
}

// stateText names the common switch states so clients can translate them;
// sensor values keep their own text and unit.
func stateText(e domain.Entity) kernel.Text {
	switch e.State {
	case "on":
		return kernel.NewText("On")
	case "off":
		return kernel.NewText("Off")
	case "unavailable":
		return kernel.NewText("Unavailable")
	}
	if e.Unit != "" {
		return kernel.NewTextf("{state} {unit}", "state", e.State, "unit", e.Unit)
	}
	return kernel.NewTextf("{state}", "state", e.State)
}

// actionLabel says what the quick action does with the entity's current
// state; false when the entity has no action for this instance.
func actionLabel(e EntityView) (kernel.Text, bool) {
	switch e.Action {
	case domain.ActionToggle:
		if e.State == "on" {
			return kernel.NewText("Turn off"), true
		}
		return kernel.NewText("Turn on"), true
	case domain.ActionPress:
		return kernel.NewText("Run"), true
	}
	return kernel.Text{}, false
}
