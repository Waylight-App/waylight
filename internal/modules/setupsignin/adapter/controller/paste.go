package controller

import (
	"net/http"
	"time"

	"waylight/internal/modules/setupsignin/usecase"
	"waylight/internal/shared/httpapi"
)

// PasteHandlers serve Paste from phone in the setup assistant.
type PasteHandlers struct{ uc usecase.Pastes }

// NewPasteHandlers returns the handlers.
func NewPasteHandlers(uc usecase.Pastes) *PasteHandlers { return &PasteHandlers{uc: uc} }

// Create handles POST /api/setup/paste.
func (h *PasteHandlers) Create(w http.ResponseWriter, r *http.Request) {
	code, expires, err := h.uc.Create(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpapi.WriteJSON(w, http.StatusOK, SetupPasteDTO{Code: code, ExpiresAt: expires.UTC().Format(time.RFC3339)})
}

// Submit handles PUT /api/setup/paste/{code}.
func (h *PasteHandlers) Submit(w http.ResponseWriter, r *http.Request) {
	var dto SetupPasteInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.Submit(r.Context(), r.PathValue("code"), dto.Value); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Take handles GET /api/setup/paste/{code}.
func (h *PasteHandlers) Take(w http.ResponseWriter, r *http.Request) {
	value, ready, err := h.uc.Take(r.Context(), r.PathValue("code"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if !ready {
		httpapi.WriteJSON(w, http.StatusOK, SetupPasteStatusDTO{Status: "pending"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, SetupPasteStatusDTO{Status: "ready", Value: value})
}

// Cancel handles DELETE /api/setup/paste/{code}.
func (h *PasteHandlers) Cancel(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Cancel(r.Context(), r.PathValue("code")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
