package controller

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"trisse-home/internal/modules/appearance/adapter/presenter"
	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/modules/appearance/usecase"
	"trisse-home/internal/shared/httpapi"
)

// BackgroundInputDTO is the body of PUT /api/appearance/background.
type BackgroundInputDTO struct {
	BackgroundImage string `json:"backgroundImage"`
}

// uploadOverhead leaves room for multipart boundaries and headers.
const uploadOverhead = 64 << 10

// BackgroundHandlers serve uploaded backgrounds (#277).
type BackgroundHandlers struct{ uc usecase.Backgrounds }

// NewBackgroundHandlers returns the handlers.
func NewBackgroundHandlers(uc usecase.Backgrounds) *BackgroundHandlers {
	return &BackgroundHandlers{uc: uc}
}

// List handles GET /api/appearance/backgrounds.
func (h *BackgroundHandlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Backgrounds(list))
}

// Upload handles POST /api/appearance/backgrounds with the picture as the
// multipart field "file".
func (h *BackgroundHandlers) Upload(w http.ResponseWriter, r *http.Request) {
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "multipart/form-data" {
		httpapi.WriteError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Send the picture as multipart/form-data.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxBackgroundUploadBytes+uploadOverhead)
	mr, err := r.MultipartReader()
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
		return
	}
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			httpapi.WriteDomainError(w, r, domain.ErrBackgroundTooLarge)
			return
		}
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
			return
		}
		if part.FormName() != "file" {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(part, domain.MaxBackgroundUploadBytes+1))
		if err != nil {
			httpapi.WriteDomainError(w, r, domain.ErrBackgroundTooLarge)
			return
		}
		b, err := h.uc.Upload(r.Context(), data)
		if err != nil {
			httpapi.WriteDomainError(w, r, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, presenter.Background(b))
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "The form has no file.")
}

// Image handles GET /api/appearance/backgrounds/{id} (?thumb=1 for the
// preview). A picture never changes under its id, so it may be cached.
func (h *BackgroundHandlers) Image(w http.ResponseWriter, r *http.Request) {
	data, err := h.uc.Image(r.Context(), r.PathValue("id"), r.URL.Query().Get("thumb") == "1")
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	// The bytes are a JPEG this server encoded itself, never the upload.
	_, _ = w.Write(data) //nolint:gosec // see above; nosniff and the sandbox CSP add defence in depth
}

// Delete handles DELETE /api/appearance/backgrounds/{id}.
func (h *BackgroundHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Set handles PUT /api/appearance/background: only the global background.
func (h *BackgroundHandlers) Set(w http.ResponseWriter, r *http.Request) {
	var dto BackgroundInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Set(r.Context(), dto.BackgroundImage)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Appearance(res))
}
