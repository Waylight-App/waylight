// Package controller is the HTTP API of absences (#299).
//
// Ring: Interface adapters (controller).
package controller

import (
	"net/http"

	"trisse-home/ee/absences/domain"
	"trisse-home/ee/absences/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serve the API.
type Handlers struct{ s *usecase.Service }

// NewHandlers returns the handlers.
func NewHandlers(s *usecase.Service) *Handlers { return &Handlers{s: s} }

// AbsenceDTO is one absence as the viewer may see it; kind is "away" and
// note empty for someone else's private absence.
type AbsenceDTO struct {
	ID        string `json:"id"`
	AccountID string `json:"accountId"`
	Name      string `json:"name"`
	From      string `json:"from"`
	To        string `json:"to"`
	Kind      string `json:"kind"`
	Note      string `json:"note"`
	Private   bool   `json:"private"`
	Editable  bool   `json:"editable"`
}

// InputDTO enters or changes an absence; accountId only for admins.
type InputDTO struct {
	AccountID string `json:"accountId,omitempty"`
	From      string `json:"from"`
	To        string `json:"to"`
	Kind      string `json:"kind"`
	Note      string `json:"note"`
	Private   bool   `json:"private"`
}

func dto(v usecase.View) AbsenceDTO {
	return AbsenceDTO{ID: v.ID, AccountID: v.AccountID, Name: v.Name, From: v.From, To: v.To, Kind: string(v.Kind), Note: v.Note,
		Private: v.Private, Editable: v.Editable}
}

func (in InputDTO) input() usecase.Input {
	return usecase.Input{AccountID: in.AccountID, From: in.From, To: in.To, Kind: domain.Kind(in.Kind), Note: in.Note, Private: in.Private}
}

// List handles GET /api/absences?from=YYYY-MM-DD&to=YYYY-MM-DD.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.s.List(r.Context(), q.Get("from"), q.Get("to"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	out := make([]AbsenceDTO, 0, len(list))
	for _, v := range list {
		out = append(out, dto(v))
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}

// Create handles POST /api/absences.
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

// Update handles PUT /api/absences/{id}.
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

// Delete handles DELETE /api/absences/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.s.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
