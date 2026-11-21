// Package controller contains the connections HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"waylight/internal/modules/connections/adapter/presenter"
	"waylight/internal/modules/connections/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the connections endpoints.
type Handlers struct{ uc usecase.Connections }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Connections) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/connections[?service=immich].
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context(), r.URL.Query().Get("service"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Connections(list))
}

// Add handles POST /api/connections.
func (h *Handlers) Add(w http.ResponseWriter, r *http.Request) {
	var dto ConnectionInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Add(r.Context(), usecase.Input{Service: dto.Service, Name: dto.Name, URL: dto.URL, Credential: dto.Credential})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Connection(v))
}

// Update handles PUT /api/connections/{id}: rename or move a connection.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto ConnectionChangeDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Update(r.Context(), r.PathValue("id"), usecase.Change{Name: dto.Name, URL: dto.URL, Credential: dto.Credential})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Connection(v))
}

// ReplaceCredential handles PUT /api/connections/{id}/credential.
func (h *Handlers) ReplaceCredential(w http.ResponseWriter, r *http.Request) {
	var dto CredentialDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.ReplaceCredential(r.Context(), r.PathValue("id"), dto.Credential)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Connection(v))
}

// Remove handles DELETE /api/connections/{id}[?confirm=true]; a connection
// in use is only removed with confirm=true.
func (h *Handlers) Remove(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Remove(r.Context(), r.PathValue("id"), r.URL.Query().Get("confirm") == "true"); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
