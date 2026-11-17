package controller

import (
	"errors"
	"io"
	"net/http"

	"waylight/internal/modules/transfer/domain"
	"waylight/internal/modules/transfer/usecase"
	"waylight/internal/shared/httpapi"
	"waylight/internal/shared/kernel"
)

// ConfigHandlers serves waylight.yaml (#88).
type ConfigHandlers struct{ uc usecase.Config }

// NewConfigHandlers returns the handlers.
func NewConfigHandlers(uc usecase.Config) *ConfigHandlers { return &ConfigHandlers{uc: uc} }

// Export handles GET /api/transfer/config: the settings as a YAML download.
func (h *ConfigHandlers) Export(w http.ResponseWriter, r *http.Request) {
	data, err := h.uc.Export(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="waylight.yaml"`)
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}

// Apply handles POST /api/transfer/config with the YAML file as body.
func (h *ConfigHandlers) Apply(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, domain.MaxConfigBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			httpapi.WriteDomainError(w, r, kernel.NewValidationError("file", "The file is larger than 2 MB."))
			return
		}
		httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "The file could not be read.")
		return
	}
	res, err := h.uc.Apply(r.Context(), data)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, ConfigResultDTO(res))
}

// Status handles GET /api/transfer/config/status: which sections
// CONFIG_FILE manages, so the UI can lock them.
func (h *ConfigHandlers) Status(w http.ResponseWriter, _ *http.Request) {
	httpapi.WriteJSON(w, http.StatusOK, ConfigStatusDTO{Managed: h.uc.Managed()})
}
