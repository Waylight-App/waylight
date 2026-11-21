package modules

import (
	"context"
	"net/http"

	"waylight/internal/shared/httpapi"
	"waylight/internal/shared/kernel"
)

// CompactProvider is implemented by modules whose widget has a compact view
// (#597). The platform serves the state at CompactRoute, so it passes the
// same permission checks as the rest of the module's API.
type CompactProvider interface {
	CompactState(ctx context.Context) (kernel.CompactState, error)
}

// CompactRoute is the pattern of a module's compact state endpoint.
func CompactRoute(module string) string { return "GET /api/" + module + "/compact" }

// CompactHandler answers GET /api/<module>/compact with the provider's state.
func CompactHandler(p CompactProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := p.CompactState(r.Context())
		if err != nil {
			httpapi.WriteDomainError(w, r, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, NewCompactResponse(state))
	}
}

// NewCompactResponse converts a compact state into its DTO.
func NewCompactResponse(s kernel.CompactState) CompactResponse {
	level := s.Level
	if level == "" {
		level = kernel.CompactNone
	}
	res := CompactResponse{Title: compactText(s.Title), Level: string(level), Count: s.Count, UpdatedAt: s.UpdatedAt}
	if s.Detail.Template != "" {
		d := compactText(s.Detail)
		res.Detail = &d
	}
	if a := s.Action; a != nil {
		action := CompactAction{ID: a.ID, Label: compactText(a.Label), Method: a.Method, Path: a.Path, Body: a.Body}
		for _, c := range a.Choices {
			action.Choices = append(action.Choices, CompactChoice{ID: c.ID, Label: compactText(c.Label), Body: c.Body, Selected: c.Selected})
		}
		res.Action = &action
	}
	return res
}

func compactText(t kernel.Text) CompactText {
	return CompactText{Key: kernel.MessageKey(t.Template), Text: t.String(), Params: t.Params}
}
