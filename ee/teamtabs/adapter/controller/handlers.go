// Package controller serves the team tab endpoints (#295).
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/ee/teamtabs/usecase"
	"trisse-home/internal/shared/httpapi"
)

// TabDTO is an own tab and who else sees it.
type TabDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
}

// ShareDTO is the body of PUT /api/team-tabs/{id}: "private", "everyone" or "group:<id>".
type ShareDTO struct {
	Visibility string `json:"visibility"`
}

// Handlers serves the endpoints.
type Handlers struct{ svc *usecase.Service }

// NewHandlers returns the handlers.
func NewHandlers(svc *usecase.Service) *Handlers { return &Handlers{svc: svc} }

// List handles GET /api/team-tabs (admins): their own tabs and visibility.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	tabs, err := h.svc.Tabs(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	out := make([]TabDTO, 0, len(tabs))
	for _, t := range tabs {
		out = append(out, TabDTO{ID: t.ID, Name: t.Name, Visibility: string(t.Visibility)})
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}

// Share handles PUT /api/team-tabs/{id} (admins).
func (h *Handlers) Share(w http.ResponseWriter, r *http.Request) {
	var dto ShareDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.svc.Share(r.Context(), r.PathValue("id"), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, ShareDTO{Visibility: string(v)})
}
