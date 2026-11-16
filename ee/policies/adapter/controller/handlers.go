// Package controller serves the admin policies (#304).
//
// Ring: Interface adapters (controller).
package controller

import (
	"net/http"

	"trisse-home/ee/policies/domain"
	"trisse-home/ee/policies/usecase"
	"trisse-home/internal/shared/httpapi"
)

// PoliciesDTO is GET and PUT /api/policies.
type PoliciesDTO struct {
	RequireSecondFactor bool `json:"requireSecondFactor"`
	// MinLength of new passwords is 0 (the built-in 12) or 12 to 64.
	MinLength int `json:"minPasswordLength"`
	// MaxSessionHours is 0 (no limit) or up to 2160.
	MaxSessionHours int `json:"maxSessionHours"`
	// MemberModules lists the modules members may use; null allows all.
	MemberModules []string `json:"memberModules"`
}

// Handlers serves the endpoints.
type Handlers struct{ s *usecase.Service }

// NewHandlers returns the handlers.
func NewHandlers(s *usecase.Service) *Handlers { return &Handlers{s: s} }

// Get handles GET /api/policies (admins).
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	set, err := h.s.Get(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, PoliciesDTO(set))
}

// Save handles PUT /api/policies (admins).
func (h *Handlers) Save(w http.ResponseWriter, r *http.Request) {
	var in PoliciesDTO
	if err := httpapi.DecodeJSON(w, r, &in); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	set, err := h.s.Save(r.Context(), domain.Settings(in))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, PoliciesDTO(set))
}
