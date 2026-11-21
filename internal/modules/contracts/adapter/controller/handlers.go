// Package controller contains the contract HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"waylight/internal/modules/contracts/adapter/presenter"
	"waylight/internal/modules/contracts/domain"
	"waylight/internal/modules/contracts/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the contract endpoints.
type Handlers struct{ uc usecase.Contracts }

// NewHandlers returns contract handlers.
func NewHandlers(uc usecase.Contracts) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/contracts.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Contracts(list))
}

func input(dto ContractInputDTO) domain.Input {
	return domain.Input{Name: dto.Name, Provider: dto.Provider, CostCents: dto.CostCents, Period: dto.Period, TermEnd: dto.TermEnd,
		NoticeDays: dto.NoticeDays, RenewMonths: dto.RenewMonths, RemindDays: dto.RemindDays, Note: dto.Note}
}

// Create handles POST /api/contracts.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto ContractInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Create(r.Context(), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Contract(v))
}

// Update handles PUT /api/contracts/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto ContractInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Contract(v))
}

// Delete handles DELETE /api/contracts/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
