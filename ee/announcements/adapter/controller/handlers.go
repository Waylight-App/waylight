// Package controller is the HTTP API of announcements (#297).
//
// Ring: Interface adapters (controller).
package controller

import (
	"net/http"
	"time"

	"trisse-home/ee/announcements/domain"
	"trisse-home/ee/announcements/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serve the API.
type Handlers struct{ s *usecase.Service }

// NewHandlers returns the handlers.
func NewHandlers(s *usecase.Service) *Handlers { return &Handlers{s: s} }

// AnnouncementDTO is one notice.
type AnnouncementDTO struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	Pinned     bool       `json:"pinned"`
	AuthorID   string     `json:"authorId"`
	AuthorName string     `json:"authorName"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	Read       bool       `json:"read"`
	Editable   bool       `json:"editable"`
}

// InputDTO creates or changes a notice.
type InputDTO struct {
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Pinned    bool       `json:"pinned"`
	ExpiresAt *time.Time `json:"expiresAt"`
	// Notify sends the notice to the notification channels.
	Notify bool `json:"notify"`
}

// SettingsDTO says who besides admins may write.
type SettingsDTO struct {
	WriterGroup string `json:"writerGroup"`
}

func dto(v usecase.View) AnnouncementDTO {
	out := AnnouncementDTO{ID: v.ID, Title: v.Title, Body: v.Body, Pinned: v.Pinned, AuthorID: v.AuthorID, AuthorName: v.AuthorName,
		CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, Read: v.Read, Editable: v.Editable}
	if !v.ExpiresAt.IsZero() {
		t := v.ExpiresAt
		out.ExpiresAt = &t
	}
	return out
}

func (in InputDTO) input() usecase.Input {
	u := usecase.Input{Title: in.Title, Body: in.Body, Pinned: in.Pinned, Notify: in.Notify}
	if in.ExpiresAt != nil {
		u.ExpiresAt = *in.ExpiresAt
	}
	return u
}

// List handles GET /api/announcements; ?archive=1 lists expired notices.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.s.List(r.Context(), r.URL.Query().Get("archive") == "1")
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	out := make([]AnnouncementDTO, 0, len(list))
	for _, v := range list {
		out = append(out, dto(v))
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}

// Create handles POST /api/announcements.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var in InputDTO
	if err := httpapi.DecodeJSON(w, r, &in); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.s.Create(r.Context(), in.input())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, dto(v))
}

// Update handles PUT /api/announcements/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var in InputDTO
	if err := httpapi.DecodeJSON(w, r, &in); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.s.Update(r.Context(), r.PathValue("id"), in.input())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, dto(v))
}

// Delete handles DELETE /api/announcements/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.s.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MarkRead handles POST /api/announcements/{id}/read.
func (h *Handlers) MarkRead(w http.ResponseWriter, r *http.Request) {
	if err := h.s.MarkRead(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Settings handles GET /api/announcements/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	set, err := h.s.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, SettingsDTO(set))
}

// SaveSettings handles PUT /api/announcements/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var in SettingsDTO
	if err := httpapi.DecodeJSON(w, r, &in); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.s.SaveSettings(r.Context(), domain.Settings(in)); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, in)
}
