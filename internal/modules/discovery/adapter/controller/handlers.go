// Package controller serves service discovery over HTTP.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"waylight/internal/modules/discovery/adapter/presenter"
	"waylight/internal/modules/discovery/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serve the discovery endpoint.
type Handlers struct{ uc usecase.Discovery }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Discovery) *Handlers { return &Handlers{uc: uc} }

// Find handles GET /api/discovery/{integration}.
func (h *Handlers) Find(w http.ResponseWriter, r *http.Request) {
	found, err := h.uc.Find(r.Context(), r.PathValue("integration"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Found(found))
}
