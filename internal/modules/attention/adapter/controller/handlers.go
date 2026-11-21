// Package controller serves the attention list over HTTP.
package controller

import (
	"net/http"
	"strconv"

	"waylight/internal/modules/attention/adapter/presenter"
	"waylight/internal/modules/attention/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serve the attention list.
type Handlers struct{ uc usecase.Attention }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Attention) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/attention; ?refresh=true asks the services again
// instead of answering from the kept statuses.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	fresh, _ := strconv.ParseBool(r.URL.Query().Get("refresh"))
	items, err := h.uc.List(r.Context(), fresh)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.List(items))
}
