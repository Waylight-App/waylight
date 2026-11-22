// Package controller contains the meter HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"waylight/internal/modules/meters/adapter/presenter"
	"waylight/internal/modules/meters/domain"
	"waylight/internal/modules/meters/usecase"
	"waylight/internal/shared/httpapi"
	"waylight/internal/shared/kernel"
)

// Handlers serves the meter endpoints.
type Handlers struct{ uc usecase.Meters }

// NewHandlers returns meter handlers.
func NewHandlers(uc usecase.Meters) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/meters.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Meters(list))
}

func input(dto MeterInputDTO) domain.Input {
	return domain.Input{Name: dto.Name, Kind: dto.Kind, Unit: dto.Unit}
}

// Create handles POST /api/meters.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto MeterInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Create(r.Context(), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Meter(v))
}

// Update handles PUT /api/meters/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto MeterInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Meter(v))
}

// Delete handles DELETE /api/meters/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Detail handles GET /api/meters/{id}?year=2026.
func (h *Handlers) Detail(w http.ResponseWriter, r *http.Request) {
	year := 0
	if raw := r.URL.Query().Get("year"); raw != "" {
		var err error
		if year, err = strconv.Atoi(raw); err != nil || year < domain.MinYear || year > 9999 {
			httpapi.WriteDomainError(w, r, kernel.NewValidationError("year", "Enter a year like 2026."))
			return
		}
	}
	d, err := h.uc.Detail(r.Context(), r.PathValue("id"), year)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Detail(d))
}

// AddReading handles POST /api/meters/{id}/readings.
func (h *Handlers) AddReading(w http.ResponseWriter, r *http.Request) {
	var dto ReadingInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	rd, err := h.uc.AddReading(r.Context(), r.PathValue("id"), domain.ReadingInput{Day: dto.Day, Value: dto.Value})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Reading(rd))
}

// DeleteReading handles DELETE /api/meters/{id}/readings/{day}.
func (h *Handlers) DeleteReading(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteReading(r.Context(), r.PathValue("id"), r.PathValue("day")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
