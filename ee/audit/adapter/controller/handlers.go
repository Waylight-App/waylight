// Package controller serves the audit log to admins (#303).
//
// Ring: Interface adapters (controller).
package controller

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"waylight/ee/audit/domain"
	"waylight/ee/audit/usecase"
	"waylight/internal/shared/httpapi"
	"waylight/internal/shared/kernel"
)

// EntryDTO is one entry of GET /api/audit.
type EntryDTO struct {
	ID        string    `json:"id"`
	At        time.Time `json:"at"`
	AccountID string    `json:"accountId"`
	Role      string    `json:"role"`
	Module    string    `json:"module"`
	Route     string    `json:"route"`
	ObjectID  string    `json:"objectId"`
}

// SettingsDTO is GET and PUT /api/audit/settings.
type SettingsDTO struct {
	// RetentionDays is 0 (keep for good) or 30 to 3650.
	RetentionDays int `json:"retentionDays"`
}

// Handlers serves the endpoints.
type Handlers struct{ s *usecase.Service }

// NewHandlers returns the handlers.
func NewHandlers(s *usecase.Service) *Handlers { return &Handlers{s: s} }

var errBadFilter = kernel.NewPublicError(kernel.ErrInvalid, "from and to must be dates (YYYY-MM-DD) or times (RFC 3339), limit a number.")

func filter(r *http.Request) (domain.Filter, error) {
	q := r.URL.Query()
	f := domain.Filter{AccountID: q.Get("account"), Module: q.Get("module")}
	parse := func(v string) (time.Time, error) {
		if v == "" {
			return time.Time{}, nil
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, nil
		}
		return time.Parse(time.DateOnly, v)
	}
	var err1, err2, err3 error
	f.From, err1 = parse(q.Get("from"))
	f.To, err2 = parse(q.Get("to"))
	if v := q.Get("limit"); v != "" {
		f.Limit, err3 = strconv.Atoi(v)
	}
	if err1 != nil || err2 != nil || err3 != nil {
		return domain.Filter{}, errBadFilter
	}
	return f, nil
}

func dto(entries []domain.Entry) []EntryDTO {
	out := make([]EntryDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, EntryDTO(e))
	}
	return out
}

// List handles GET /api/audit.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	f, err := filter(r)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	entries, err := h.s.List(r.Context(), f)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, dto(entries))
}

// Export handles GET /api/audit/export?format=csv|json: all matching entries
// as a download.
func (h *Handlers) Export(w http.ResponseWriter, r *http.Request) {
	f, err := filter(r)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	entries, err := h.s.Export(r.Context(), f)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	name := "waylight-audit-" + time.Now().UTC().Format("20060102-150405")
	if r.URL.Query().Get("format") == "csv" {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="`+name+`.csv"`)
		cw := csv.NewWriter(w)
		_ = cw.Write([]string{"at", "account_id", "role", "module", "route", "object_id", "id"})
		for _, e := range entries {
			_ = cw.Write([]string{e.At.Format(time.RFC3339), e.AccountID, e.Role, e.Module, e.Route, e.ObjectID, e.ID})
		}
		cw.Flush()
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`.json"`)
	_ = json.NewEncoder(w).Encode(dto(entries))
}

// Settings handles GET /api/audit/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	days, err := h.s.Retention(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, SettingsDTO{RetentionDays: days})
}

// SaveSettings handles PUT /api/audit/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var in SettingsDTO
	if err := httpapi.DecodeJSON(w, r, &in); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.s.SetRetention(r.Context(), in.RetentionDays); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, in)
}
