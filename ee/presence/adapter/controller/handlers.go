// Package controller is the HTTP API of live presence (#346).
//
// Ring: Interface adapters (controller).
package controller

import (
	"net/http"

	"trisse-home/ee/presence/domain"
	"trisse-home/ee/presence/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serve the API.
type Handlers struct{ s *usecase.Service }

// NewHandlers returns the handlers.
func NewHandlers(s *usecase.Service) *Handlers { return &Handlers{s: s} }

// PersonDTO is one person; status "" is not set, source tells whether the
// person set it ("self") or it comes from a planned absence ("absence").
type PersonDTO struct {
	AccountID string `json:"accountId"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Source    string `json:"source"`
	ReturnOn  string `json:"returnOn,omitempty"`
}

// CountsDTO sums the board up.
type CountsDTO struct {
	Office int `json:"office"`
	Home   int `json:"home"`
	Away   int `json:"away"`
	None   int `json:"none"`
}

// BoardDTO is a group's presence today; mine is null for viewers who
// cannot set a status.
type BoardDTO struct {
	Day    string      `json:"day"`
	People []PersonDTO `json:"people"`
	Counts CountsDTO   `json:"counts"`
	Mine   *PersonDTO  `json:"mine"`
}

// StatusDTO sets the viewer's status; "" clears it.
type StatusDTO struct {
	Status   string `json:"status"`
	ReturnOn string `json:"returnOn"`
}

// SettingsDTO is the organisation-wide switch.
type SettingsDTO struct {
	Enabled bool `json:"enabled"`
}

func person(p domain.Person) PersonDTO {
	return PersonDTO{AccountID: p.AccountID, Name: p.Name, Status: string(p.Status), Source: string(p.Source), ReturnOn: p.ReturnOn}
}

// Board handles GET /api/presence?group=.
func (h *Handlers) Board(w http.ResponseWriter, r *http.Request) {
	b, err := h.s.Board(r.Context(), r.URL.Query().Get("group"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	out := BoardDTO{Day: b.Day, People: make([]PersonDTO, 0, len(b.People)), Counts: CountsDTO{
		Office: b.Counts[domain.StatusOffice], Home: b.Counts[domain.StatusHome], Away: b.Counts[domain.StatusAway],
		None: b.Counts[domain.StatusNone]}}
	for _, p := range b.People {
		out.People = append(out.People, person(p))
	}
	if b.Mine.AccountID != "" {
		mine := person(b.Mine)
		out.Mine = &mine
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}

// SetMine handles PUT /api/presence/me.
func (h *Handlers) SetMine(w http.ResponseWriter, r *http.Request) {
	var in StatusDTO
	if err := httpapi.DecodeJSON(w, r, &in); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.s.Set(r.Context(), domain.Status(in.Status), in.ReturnOn); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Settings handles GET /api/presence/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	on, err := h.s.Enabled(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, SettingsDTO{Enabled: on})
}

// SaveSettings handles PUT /api/presence/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var in SettingsDTO
	if err := httpapi.DecodeJSON(w, r, &in); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.s.SetEnabled(r.Context(), in.Enabled); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, in)
}
