package controller

import (
	"net/http"

	"waylight/internal/modules/waste/adapter/presenter"
	"waylight/internal/modules/waste/usecase"
	"waylight/internal/shared/httpapi"
)

// ReminderHandlers serves the reminder endpoints (#866).
type ReminderHandlers struct{ uc usecase.Reminders }

// NewReminderHandlers returns reminder handlers.
func NewReminderHandlers(uc usecase.Reminders) *ReminderHandlers { return &ReminderHandlers{uc: uc} }

// Reminder handles GET /api/waste/reminder.
func (h *ReminderHandlers) Reminder(w http.ResponseWriter, r *http.Request) {
	rem, err := h.uc.Reminder(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Reminder(rem))
}

// SaveReminder handles PUT /api/waste/reminder.
func (h *ReminderHandlers) SaveReminder(w http.ResponseWriter, r *http.Request) {
	var dto ReminderInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	rem, err := h.uc.SaveReminder(r.Context(), dto.Enabled, dto.Time)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Reminder(rem))
}
