// Package controller is the HTTP API of the team calendar (#299).
//
// Ring: Interface adapters (controller).
package controller

import (
	"net/http"

	"trisse-home/ee/teamcalendar/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serve the API.
type Handlers struct{ s *usecase.Service }

// NewHandlers returns the handlers.
func NewHandlers(s *usecase.Service) *Handlers { return &Handlers{s: s} }

// SubscriptionDTO is an own subscription and who sees it.
type SubscriptionDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
}

// List handles GET /api/teamcalendar/subscriptions.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.s.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	out := make([]SubscriptionDTO, 0, len(list))
	for _, s := range list {
		out = append(out, SubscriptionDTO{ID: s.ID, Name: s.Name, Visibility: string(s.Visibility)})
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}

// Share handles PUT /api/teamcalendar/subscriptions/{id} {visibility}.
func (h *Handlers) Share(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Visibility string `json:"visibility"`
	}
	if err := httpapi.DecodeJSON(w, r, &in); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.s.Share(r.Context(), r.PathValue("id"), in.Visibility); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
