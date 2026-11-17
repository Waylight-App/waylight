// Package controller serves the watchlist endpoint (#284).
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strings"

	"waylight/internal/modules/markets/adapter/presenter"
	"waylight/internal/modules/markets/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the endpoint.
type Handlers struct{ uc usecase.Markets }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Markets) *Handlers { return &Handlers{uc: uc} }

// Watchlist handles GET /api/markets/watchlist?symbols=fx:USD,stooq:aapl.us.
func (h *Handlers) Watchlist(w http.ResponseWriter, r *http.Request) {
	var list []string
	if s := r.URL.Query().Get("symbols"); s != "" {
		list = strings.Split(s, ",")
	}
	rows, err := h.uc.Watchlist(r.Context(), list)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Rows(rows))
}
