package controller

import (
	"net/http"

	"waylight/internal/modules/dashboard/adapter/presenter"
	"waylight/internal/modules/dashboard/usecase"
	"waylight/internal/shared/httpapi"
)

// BuiltinHandlers serve the built-in tab templates (#589).
type BuiltinHandlers struct{ uc usecase.BuiltinTemplates }

// NewBuiltinHandlers returns the handlers.
func NewBuiltinHandlers(uc usecase.BuiltinTemplates) *BuiltinHandlers {
	return &BuiltinHandlers{uc: uc}
}

// List handles GET /api/dashboard/templates/builtin.
func (h *BuiltinHandlers) List(w http.ResponseWriter, _ *http.Request) {
	httpapi.WriteJSON(w, http.StatusOK, presenter.Builtins(h.uc.List()))
}

// CreateTab handles POST /api/dashboard/templates/builtin/{id}/tabs.
func (h *BuiltinHandlers) CreateTab(w http.ResponseWriter, r *http.Request) {
	var dto BuiltinTabDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	var loc *usecase.Location
	if dto.Location != nil {
		loc = &usecase.Location{Lat: dto.Location.Lat, Lon: dto.Location.Lon, Label: dto.Location.Label, PlaceID: dto.Location.PlaceID}
	}
	res, err := h.uc.CreateTab(r.Context(), r.PathValue("id"), dto.Name, loc, dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Tab(res))
}
