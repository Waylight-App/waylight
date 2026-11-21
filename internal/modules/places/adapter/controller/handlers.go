// Package controller contains the places HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"waylight/internal/modules/places/adapter/presenter"
	"waylight/internal/modules/places/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the places endpoints.
type Handlers struct{ uc usecase.Places }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Places) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/places.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Places(list))
}

// Add handles POST /api/places.
func (h *Handlers) Add(w http.ResponseWriter, r *http.Request) {
	var dto PlaceInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	p, err := h.uc.Add(r.Context(), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Place(p))
}

// Update handles PUT /api/places/{id}: rename or move a place.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto PlaceInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	p, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Place(p))
}

// Delete handles DELETE /api/places/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func input(d PlaceInputDTO) usecase.Input {
	return usecase.Input{Name: d.Name, Label: d.Label, Lat: d.Lat, Lon: d.Lon}
}
