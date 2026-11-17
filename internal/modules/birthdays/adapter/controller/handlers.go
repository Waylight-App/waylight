// Package controller serves the birthday endpoints (#287).
//
// Ring: Interface Adapters.
package controller

import (
	"io"
	"net/http"
	"strconv"

	"waylight/internal/modules/birthdays/adapter/presenter"
	"waylight/internal/modules/birthdays/usecase"
	"waylight/internal/shared/httpapi"
)

// BirthdayInputDTO is the body of POST /api/birthdays; year 0 = unknown.
type BirthdayInputDTO struct {
	Name  string `json:"name"`
	Month int    `json:"month"`
	Day   int    `json:"day"`
	Year  int    `json:"year"`
}

// SettingsInputDTO is the body of PUT /api/birthdays/settings; an empty
// password keeps the stored one, an empty URL disconnects the address book.
type SettingsInputDTO struct {
	RemindEnabled bool   `json:"remindEnabled"`
	RemindAt      string `json:"remindAt"`
	CardDAVURL    string `json:"cardDavUrl"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

// Handlers serves the endpoints.
type Handlers struct{ uc usecase.Birthdays }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Birthdays) *Handlers { return &Handlers{uc: uc} }

// Upcoming handles GET /api/birthdays?limit=n (all when no limit).
func (h *Handlers) Upcoming(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	list, err := h.uc.Upcoming(r.Context(), max(limit, 0))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Upcoming(list))
}

// Create handles POST /api/birthdays.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto BirthdayInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if _, err := h.uc.Add(r.Context(), dto.Name, dto.Month, dto.Day, dto.Year); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// Delete handles DELETE /api/birthdays/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Import handles POST /api/birthdays/import with the .vcf file as body.
func (h *Handlers) Import(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, usecase.MaxImportBytes))
	if err != nil {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "too_large", "The vCard file must be at most 5 MB.")
		return
	}
	rep, err := h.uc.ImportVCard(r.Context(), data)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.ImportReportDTO(rep))
}

// Settings handles GET /api/birthdays/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	v, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(v))
}

// SaveSettings handles PUT /api/birthdays/settings (reads the address book first).
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.SaveSettings(r.Context(), usecase.SettingsInput(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(v))
}
