// Package controller serves the wiki search endpoints (#321).
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"waylight/internal/modules/wiki/adapter/presenter"
	"waylight/internal/modules/wiki/domain"
	"waylight/internal/modules/wiki/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the endpoints.
type Handlers struct{ uc usecase.Wikis }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Wikis) *Handlers { return &Handlers{uc: uc} }

// Search handles GET /api/wiki/search?q=…
func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Search(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Search(res))
}

// List handles GET /api/wiki/wikis.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	out := make([]presenter.WikiDTO, 0, len(list))
	for _, v := range list {
		out = append(out, presenter.Wiki(v))
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}

func input(dto WikiInputDTO) usecase.WikiInput {
	return usecase.WikiInput{Name: dto.Name, Kind: domain.Kind(dto.Kind), URL: dto.URL, Token: dto.Token}
}

// Create handles POST /api/wiki/wikis.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto WikiInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Create(r.Context(), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Wiki(v))
}

// Update handles PUT /api/wiki/wikis/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto WikiInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Wiki(v))
}

// Delete handles DELETE /api/wiki/wikis/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Test handles POST /api/wiki/wikis/{id}/test.
func (h *Handlers) Test(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Test(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
