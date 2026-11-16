// Package controller is the admin API of the branding (#305). Reading the
// branding is public and lives in the core (internal/platform/branding).
//
// Ring: Interface adapters (controller).
package controller

import (
	"errors"
	"io"
	"net/http"
	"time"

	"trisse-home/ee/branding/domain"
	"trisse-home/ee/branding/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serve the API.
type Handlers struct{ s *usecase.Service }

// NewHandlers returns the handlers.
func NewHandlers(s *usecase.Service) *Handlers { return &Handlers{s: s} }

// SettingsDTO is the settings page.
type SettingsDTO struct {
	Name          string               `json:"name"`
	Accent        string               `json:"accent"`
	Secondary     string               `json:"secondary"`
	HidePoweredBy bool                 `json:"hidePoweredBy"`
	DefaultTheme  string               `json:"defaultTheme"`
	LockTheme     bool                 `json:"lockTheme"`
	Assets        map[string]time.Time `json:"assets,omitempty"`
	Applied       bool                 `json:"applied"`
	Warnings      []string             `json:"warnings"`
}

func dto(v usecase.View) SettingsDTO {
	w := v.Warnings
	if w == nil {
		w = []string{}
	}
	return SettingsDTO{Name: v.Name, Accent: v.Accent, Secondary: v.Secondary, HidePoweredBy: v.HidePoweredBy, DefaultTheme: v.DefaultTheme,
		LockTheme: v.LockTheme, Assets: v.Assets,
		Applied: v.Applied, Warnings: w}
}

func write(w http.ResponseWriter, r *http.Request, v usecase.View, err error) {
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, dto(v))
}

// Get handles GET /api/branding/settings.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	v, err := h.s.Get(r.Context())
	write(w, r, v, err)
}

// Save handles PUT /api/branding/settings.
func (h *Handlers) Save(w http.ResponseWriter, r *http.Request) {
	var in SettingsDTO
	if err := httpapi.DecodeJSON(w, r, &in); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.s.Save(r.Context(), domain.Settings{Name: in.Name, Accent: in.Accent, Secondary: in.Secondary, HidePoweredBy: in.HidePoweredBy,
		DefaultTheme: in.DefaultTheme, LockTheme: in.LockTheme})
	write(w, r, v, err)
}

// Upload handles PUT /api/branding/assets/{kind}; the body is the image.
func (h *Handlers) Upload(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, domain.MaxUpload+1))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		httpapi.WriteDomainError(w, r, domain.ErrImage)
		return
	}
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.s.Upload(r.Context(), r.PathValue("kind"), data)
	write(w, r, v, err)
}

// Delete handles DELETE /api/branding/assets/{kind}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	v, err := h.s.Delete(r.Context(), r.PathValue("kind"))
	write(w, r, v, err)
}
