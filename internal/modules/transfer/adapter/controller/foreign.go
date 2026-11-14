package controller

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"trisse-home/internal/modules/transfer/domain"
	"trisse-home/internal/modules/transfer/usecase"
	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
)

// ForeignHandlers serve the import from other dashboards (#77).
type ForeignHandlers struct{ uc usecase.Importer }

// NewForeignHandlers returns the handlers.
func NewForeignHandlers(uc usecase.Importer) *ForeignHandlers { return &ForeignHandlers{uc: uc} }

// Import handles POST /api/transfer/import/{format}: the files as
// multipart parts named "file"; ?apply=true applies the result, otherwise
// it is only a preview.
func (h *ForeignHandlers) Import(w http.ResponseWriter, r *http.Request) {
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "multipart/form-data" {
		httpapi.WriteError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Send the files as multipart/form-data.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxConfigBytes)
	mr, err := r.MultipartReader()
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
		return
	}
	files := map[string][]byte{}
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			httpapi.WriteDomainError(w, r, kernel.NewValidationError("file", "The files are larger than 2 MB together."))
			return
		}
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
			return
		}
		if part.FormName() != "file" {
			continue
		}
		data, err := io.ReadAll(part)
		if err != nil {
			httpapi.WriteDomainError(w, r, kernel.NewValidationError("file", "The files are larger than 2 MB together."))
			return
		}
		files[part.FileName()] = data
	}
	if len(files) == 0 {
		httpapi.WriteDomainError(w, r, kernel.NewValidationError("file", "Choose at least one file."))
		return
	}
	out, err := h.uc.Import(r.Context(), r.PathValue("format"), files, r.URL.Query().Get("apply") == "true")
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	s := out.Summary
	dto := ForeignImportDTO{YAML: string(out.YAML), Summary: ImportSummaryDTO{Tiles: s.Tiles, Bookmarks: s.Bookmarks, Widgets: s.Widgets,
		Skipped: s.Skipped, Tab: s.Tab, Warnings: append([]string{}, s.Warnings...)}}
	if out.Result != nil {
		res := ConfigResultDTO(*out.Result)
		dto.Result = &res
	}
	httpapi.WriteJSON(w, http.StatusOK, dto)
}
