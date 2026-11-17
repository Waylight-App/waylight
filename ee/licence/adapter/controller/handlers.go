// Package controller serves the licence endpoints (#307).
//
// Ring: Interface adapters (controller).
package controller

import (
	"net/http"
	"time"

	"waylight/ee/licence/domain"
	"waylight/ee/licence/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves /api/licence.
type Handlers struct{ s *usecase.Service }

// NewHandlers returns the handlers.
func NewHandlers(s *usecase.Service) *Handlers { return &Handlers{s: s} }

func present(v usecase.View) LicenceDTO {
	out := LicenceDTO{Status: string(v.Status), Source: string(v.Source), Problem: v.Problem, Accounts: v.Accounts,
		Features: []string{}, Unknown: []string{}}
	if l := v.Licence; l != nil {
		out.Licensee, out.Seats = l.Licensee, l.Seats
		out.Issued, out.Until = l.Issued.Format(time.DateOnly), l.Until.Format(time.DateOnly)
		out.Features = append(out.Features, l.Features...)
	}
	out.Unknown = append(out.Unknown, v.Unknown...)
	return out
}

var levelNames = map[domain.Level]string{domain.LevelNone: "none", domain.LevelRead: "read", domain.LevelFull: "full"}

// Get handles GET /api/licence (admins).
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	v, err := h.s.Status(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, present(v))
}

// Install handles PUT /api/licence (admins).
func (h *Handlers) Install(w http.ResponseWriter, r *http.Request) {
	var dto InstallDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.s.Install(r.Context(), dto.Key)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, present(v))
}

// Remove handles DELETE /api/licence (admins).
func (h *Handlers) Remove(w http.ResponseWriter, r *http.Request) {
	if err := h.s.Remove(r.Context()); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Features handles GET /api/licence/features (every signed-in account): the
// interface uses it for hints; access itself is enforced on the server.
func (h *Handlers) Features(w http.ResponseWriter, r *http.Request) {
	out := FeaturesDTO{Features: map[string]string{}}
	for f, l := range h.s.Levels(r.Context()) {
		out.Features[f] = levelNames[l]
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}
