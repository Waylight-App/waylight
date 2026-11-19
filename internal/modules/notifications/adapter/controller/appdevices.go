package controller

import (
	"encoding/base64"
	"net/http"

	"waylight/internal/modules/notifications/adapter/presenter"

	"waylight/internal/modules/notifications/domain"
	"waylight/internal/modules/notifications/usecase"
	"waylight/internal/shared/httpapi"
	"waylight/internal/shared/kernel"
)

// AppDeviceHandlers serve the push devices of the mobile app (#564).
type AppDeviceHandlers struct{ uc usecase.AppDevices }

// NewAppDeviceHandlers returns the handlers.
func NewAppDeviceHandlers(uc usecase.AppDevices) *AppDeviceHandlers {
	return &AppDeviceHandlers{uc: uc}
}

// Register handles PUT /api/notifications/app-device.
func (h *AppDeviceHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var dto AppDeviceDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	key, err := base64.StdEncoding.DecodeString(dto.PublicKey)
	if err != nil {
		httpapi.WriteDomainError(w, r, kernel.NewValidationError("publicKey", "The public key must be base64."))
		return
	}
	d, err := h.uc.Register(r.Context(), usecase.AppDeviceInput{Name: dto.Name, Platform: domain.PushPlatform(dto.Platform), Token: dto.Token, PublicKey: key})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.AppDevice(d))
}

// Unregister handles DELETE /api/notifications/app-device.
func (h *AppDeviceHandlers) Unregister(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Unregister(r.Context()); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// List handles GET /api/notifications/app-devices.
func (h *AppDeviceHandlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.AppDevices(list))
}

// Remove handles DELETE /api/notifications/app-devices/{session}.
func (h *AppDeviceHandlers) Remove(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Remove(r.Context(), r.PathValue("session")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
