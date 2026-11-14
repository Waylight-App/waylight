// Package controller serves the sports endpoints (#285).
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/sports/adapter/presenter"
	"trisse-home/internal/modules/sports/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the endpoints.
type Handlers struct{ uc usecase.Sports }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Sports) *Handlers { return &Handlers{uc: uc} }

// Leagues handles GET /api/sports/leagues.
func (h *Handlers) Leagues(w http.ResponseWriter, _ *http.Request) {
	httpapi.WriteJSON(w, http.StatusOK, presenter.Leagues(h.uc.Leagues()))
}

// Teams handles GET /api/sports/teams?league=bl1.
func (h *Handlers) Teams(w http.ResponseWriter, r *http.Request) {
	ts, err := h.uc.Teams(r.Context(), r.URL.Query().Get("league"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Teams(ts))
}

// Overview handles GET /api/sports/overview?league=bl1&team=40.
func (h *Handlers) Overview(w http.ResponseWriter, r *http.Request) {
	team, _ := strconv.Atoi(r.URL.Query().Get("team"))
	o, err := h.uc.Overview(r.Context(), r.URL.Query().Get("league"), team)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Overview(o))
}
