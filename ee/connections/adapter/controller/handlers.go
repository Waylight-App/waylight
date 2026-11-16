// Package controller serves connected accounts (#326).
//
// Ring: Interface adapters (controller).
package controller

import (
	"net/http"
	"net/url"
	"time"

	"trisse-home/ee/connections/domain"
	"trisse-home/ee/connections/usecase"
	"trisse-home/internal/shared/httpapi"
)

// ConnectionDTO is one service in GET /api/connections.
type ConnectionDTO struct {
	Service string `json:"service"`
	Name    string `json:"name"`
	// Configured is false until an admin registers trisse-home at the service.
	Configured bool `json:"configured"`
	// Status is "", "connected" or "expired" (connect again).
	Status string     `json:"status"`
	Since  *time.Time `json:"since,omitempty"`
}

// StartDTO is the answer of POST /api/connections/{service}.
type StartDTO struct {
	URL string `json:"url"`
}

// ClientDTO is one service in GET /api/connections/clients (admins).
type ClientDTO struct {
	Service   string `json:"service"`
	Name      string `json:"name"`
	ClientID  string `json:"clientId"`
	Tenant    string `json:"tenant"`
	HasSecret bool   `json:"hasSecret"`
}

// ClientsDTO lists the registrations and the redirect URL to register.
type ClientsDTO struct {
	RedirectURL string      `json:"redirectUrl"`
	Clients     []ClientDTO `json:"clients"`
}

// SaveClientDTO is the body of PUT /api/connections/clients/{service}; an
// empty clientSecret keeps the stored one.
type SaveClientDTO struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Tenant       string `json:"tenant"`
}

// Handlers serves the endpoints.
type Handlers struct{ s *usecase.Service }

// NewHandlers returns the handlers.
func NewHandlers(s *usecase.Service) *Handlers { return &Handlers{s: s} }

// List handles GET /api/connections.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	views, err := h.s.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	out := make([]ConnectionDTO, 0, len(views))
	for _, v := range views {
		d := ConnectionDTO{Service: v.Service.ID, Name: v.Service.Name, Configured: v.Configured, Status: string(v.Status)}
		if !v.Since.IsZero() {
			since := v.Since
			d.Since = &since
		}
		out = append(out, d)
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}

// Start handles POST /api/connections/{service}.
func (h *Handlers) Start(w http.ResponseWriter, r *http.Request) {
	u, err := h.s.Start(r.Context(), r.PathValue("service"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, StartDTO{URL: u})
}

// Callback handles GET /api/connections/callback, where the service sends
// the browser back, and redirects to the settings with the outcome.
func (h *Handlers) Callback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	service, err := h.s.Finish(r.Context(), q.Get("state"), q.Get("code"), q.Get("error"))
	// The section lives in the ee/ plugin, whose ids carry "plugin-" (ADR 0021).
	back := url.Values{"settings": {"plugin-connections"}}
	if err != nil {
		back.Set("connection_error", service)
	} else {
		back.Set("connected", service)
	}
	http.Redirect(w, r, "/?"+back.Encode(), http.StatusSeeOther)
}

// Disconnect handles DELETE /api/connections/{service}.
func (h *Handlers) Disconnect(w http.ResponseWriter, r *http.Request) {
	if err := h.s.Disconnect(r.Context(), r.PathValue("service")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Clients handles GET /api/connections/clients (admins).
func (h *Handlers) Clients(w http.ResponseWriter, r *http.Request) {
	views, redirect, err := h.s.ClientsList(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	out := ClientsDTO{RedirectURL: redirect, Clients: []ClientDTO{}}
	for _, v := range views {
		out.Clients = append(out.Clients, ClientDTO{Service: v.Service.ID, Name: v.Service.Name, ClientID: v.ClientID, Tenant: v.Tenant, HasSecret: v.HasSecret})
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}

// SaveClient handles PUT /api/connections/clients/{service} (admins).
func (h *Handlers) SaveClient(w http.ResponseWriter, r *http.Request) {
	var in SaveClientDTO
	if err := httpapi.DecodeJSON(w, r, &in); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	err := h.s.SaveClient(r.Context(), domain.Client{Service: r.PathValue("service"), ClientID: in.ClientID, Secret: in.ClientSecret, Tenant: in.Tenant})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
