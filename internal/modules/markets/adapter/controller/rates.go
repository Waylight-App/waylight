package controller

import (
	"net/http"

	"waylight/internal/modules/markets/adapter/presenter"
	"waylight/internal/modules/markets/usecase"
	"waylight/internal/shared/httpapi"
)

// RatesHandler serves the reference rates.
type RatesHandler struct{ rates *usecase.Rates }

// NewRatesHandler returns the handler.
func NewRatesHandler(r *usecase.Rates) *RatesHandler { return &RatesHandler{rates: r} }

// Rates handles GET /api/markets/rates.
func (h *RatesHandler) Rates(w http.ResponseWriter, r *http.Request) {
	v, err := h.rates.Current(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.RatesDTO{Base: v.Base, Date: v.Day.Format("2006-01-02"), Rates: v.Rates, Stale: v.Stale})
}
