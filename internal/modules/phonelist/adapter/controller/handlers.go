// Package controller contains the phone list HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/phonelist/adapter/presenter"
	"trisse-home/internal/modules/phonelist/domain"
	"trisse-home/internal/modules/phonelist/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the phone list endpoints.
type Handlers struct{ uc usecase.Phonelist }

// NewHandlers returns phone list handlers.
func NewHandlers(uc usecase.Phonelist) *Handlers { return &Handlers{uc: uc} }

func (d EntryInputDTO) input() domain.EntryInput { return domain.EntryInput(d) }

// List handles GET /api/phonelist/entries; ?q= searches name, role and team.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	var entries []domain.Entry
	var err error
	if q := r.URL.Query().Get("q"); q != "" {
		entries, err = h.uc.Search(r.Context(), q)
	} else {
		entries, err = h.uc.List(r.Context())
	}
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Entries(entries))
}

// Add handles POST /api/phonelist/entries.
func (h *Handlers) Add(w http.ResponseWriter, r *http.Request) {
	var dto EntryInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	e, err := h.uc.Add(r.Context(), dto.input())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Entry(e))
}

// Update handles PUT /api/phonelist/entries/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto EntryInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	e, err := h.uc.Update(r.Context(), r.PathValue("id"), dto.input())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Entry(e))
}

// Delete handles DELETE /api/phonelist/entries/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
