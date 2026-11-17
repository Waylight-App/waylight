// Package controller serves the electricity price endpoint (#288).
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"waylight/internal/modules/electricity/adapter/presenter"
	"waylight/internal/modules/electricity/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the endpoint.
type Handlers struct{ uc usecase.Electricity }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Electricity) *Handlers { return &Handlers{uc: uc} }

// Prices handles GET /api/electricity/prices?zone=DE-LU.
func (h *Handlers) Prices(w http.ResponseWriter, r *http.Request) {
	p, err := h.uc.Prices(r.Context(), r.URL.Query().Get("zone"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Prices(p))
}
