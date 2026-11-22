// Package controller contains the family board HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"waylight/internal/modules/board/adapter/presenter"
	"waylight/internal/modules/board/domain"
	"waylight/internal/modules/board/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the family board endpoints.
type Handlers struct{ uc usecase.Board }

// NewHandlers returns family board handlers.
func NewHandlers(uc usecase.Board) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/board.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Notes(list))
}

func input(dto NoteInputDTO) domain.Input {
	return domain.Input{Text: dto.Text, Pinned: dto.Pinned, ExpiresOn: dto.ExpiresOn}
}

// Create handles POST /api/board.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto NoteInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Create(r.Context(), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Note(v))
}

// Update handles PUT /api/board/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto NoteInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Note(v))
}

// Delete handles DELETE /api/board/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
