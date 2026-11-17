// Package controller serves the photo frame endpoints (#286).
//
// Ring: Interface Adapters.
package controller

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"waylight/internal/modules/photos/adapter/presenter"
	"waylight/internal/modules/photos/domain"
	"waylight/internal/modules/photos/usecase"
	"waylight/internal/shared/httpapi"
)

// OrderInputDTO is the body of PUT /api/photos/order.
type OrderInputDTO struct {
	Album string   `json:"album"`
	IDs   []string `json:"ids"`
}

// uploadOverhead leaves room for the multipart boundaries and headers.
const uploadOverhead = 64 << 10

// Handlers serves the endpoints.
type Handlers struct{ uc usecase.Photos }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Photos) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/photos?album=… in display order.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context(), r.URL.Query().Get("album"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Photos(list))
}

// Albums handles GET /api/photos/albums.
func (h *Handlers) Albums(w http.ResponseWriter, r *http.Request) {
	o, err := h.uc.Overview(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Overview(o))
}

// Upload handles POST /api/photos?album=… with the picture as the
// multipart field "file"; one picture per request keeps memory bounded.
func (h *Handlers) Upload(w http.ResponseWriter, r *http.Request) {
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "multipart/form-data" {
		httpapi.WriteError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Send the picture as multipart/form-data.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxUploadBytes+uploadOverhead)
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
		if err != nil {
			writeUploadError(w, err)
			return
		}
		if part.FormName() != "file" {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(part, domain.MaxUploadBytes+1))
		if err != nil {
			writeUploadError(w, err)
			return
		}
		p, err := h.uc.Upload(r.Context(), r.URL.Query().Get("album"), data)
		if err != nil {
			httpapi.WriteDomainError(w, r, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, presenter.Photo(p))
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "The form has no file.")
}

func writeUploadError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "too_large", "A picture must be at most 25 MB.")
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
}

// Reorder handles PUT /api/photos/order.
func (h *Handlers) Reorder(w http.ResponseWriter, r *http.Request) {
	var dto OrderInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.Reorder(r.Context(), dto.Album, dto.IDs); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /api/photos/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Image handles GET /api/photos/{id}/image (?thumb=1 for the thumbnail).
// A picture never changes under its id, so the browser may keep it.
func (h *Handlers) Image(w http.ResponseWriter, r *http.Request) {
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
