// Package controller contains the document HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"

	"waylight/internal/modules/documents/adapter/presenter"
	"waylight/internal/modules/documents/domain"
	"waylight/internal/modules/documents/usecase"
	"waylight/internal/shared/httpapi"
)

// uploadOverhead leaves room for the multipart boundaries and headers.
const uploadOverhead = 64 << 10

// Handlers serves the document endpoints.
type Handlers struct{ uc usecase.Documents }

// NewHandlers returns document handlers.
func NewHandlers(uc usecase.Documents) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/documents.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Documents(list))
}

func input(dto DocumentInputDTO) domain.Input {
	return domain.Input{Title: dto.Title, PurchasedOn: dto.PurchasedOn, WarrantyEnd: dto.WarrantyEnd, Note: dto.Note}
}

// Create handles POST /api/documents.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto DocumentInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Create(r.Context(), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Document(v))
}

// Update handles PUT /api/documents/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto DocumentInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Document(v))
}

// Delete handles DELETE /api/documents/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UploadFile handles POST /api/documents/{id}/file with the file as the
// multipart field "file"; the body is capped before it is read.
func (h *Handlers) UploadFile(w http.ResponseWriter, r *http.Request) {
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "multipart/form-data" {
		httpapi.WriteError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Send the file as multipart/form-data.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxFileBytes+uploadOverhead)
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
		data, err := io.ReadAll(io.LimitReader(part, domain.MaxFileBytes+1))
		if err != nil {
			writeUploadError(w, err)
			return
		}
		v, err := h.uc.AttachFile(r.Context(), r.PathValue("id"), part.FileName(), data)
		if err != nil {
			httpapi.WriteDomainError(w, r, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, presenter.Document(v))
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "The form has no file.")
}

func writeUploadError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "too_large", "The file must be at most 10 MB.")
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
}

// File handles GET /api/documents/{id}/file: always a download, served with
// the type found in the bytes at upload (a photo or a PDF).
func (h *Handlers) File(w http.ResponseWriter, r *http.Request) {
	d, data, err := h.uc.File(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", d.File.ContentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": d.File.Name})+
		"; filename*=UTF-8''"+url.PathEscape(d.File.Name))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = w.Write(data) //nolint:gosec // an attachment of a type checked at upload; nosniff and the sandbox CSP add defence in depth
}
