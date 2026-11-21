// Package controller contains the focus timer HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"time"

	"waylight/internal/modules/focus/domain"

	"waylight/internal/modules/focus/adapter/presenter"
	"waylight/internal/modules/focus/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the focus timer endpoints.
type Handlers struct {
	uc usecase.Focus
	// now is the clock the time left is measured against.
	now func() time.Time
}

// NewHandlers returns focus timer handlers.
func NewHandlers(uc usecase.Focus, now func() time.Time) *Handlers {
	return &Handlers{uc: uc, now: now}
}

// Record handles POST /api/focus/sessions and answers with the updated
// statistics.
func (h *Handlers) Record(w http.ResponseWriter, r *http.Request) {
	var dto SessionInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if _, err := h.uc.RecordSession(r.Context(), dto.Minutes, dto.EndedAt); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	h.write(w, r, http.StatusCreated)
}

// Stats handles GET /api/focus/stats.
func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) { h.write(w, r, http.StatusOK) }

func (h *Handlers) write(w http.ResponseWriter, r *http.Request, status int) {
	days, err := h.uc.WeeklyStats(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, status, presenter.Week(days))
}

// Timer handles GET /api/focus/timer.
func (h *Handlers) Timer(w http.ResponseWriter, r *http.Request) {
	t, err := h.uc.Timer(r.Context())
	h.writeTimer(w, r, t, err)
}

// ControlTimer handles POST /api/focus/timer/{action}: start, pause, reset
// or skip.
func (h *Handlers) ControlTimer(w http.ResponseWriter, r *http.Request) {
	t, err := h.uc.ControlTimer(r.Context(), usecase.TimerAction(r.PathValue("action")))
	h.writeTimer(w, r, t, err)
}

// ConfigureTimer handles PUT /api/focus/timer/config.
func (h *Handlers) ConfigureTimer(w http.ResponseWriter, r *http.Request) {
	var dto TimerConfigInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	cfg, err := domain.NewTimerConfig(dto.Work, dto.Short, dto.Long, dto.LongEvery)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	t, err := h.uc.ConfigureTimer(r.Context(), cfg)
	h.writeTimer(w, r, t, err)
}

func (h *Handlers) writeTimer(w http.ResponseWriter, r *http.Request, t domain.Timer, err error) {
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Timer(t, h.now()))
}
