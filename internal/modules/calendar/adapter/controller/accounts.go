package controller

import (
	"fmt"
	"net/http"

	"waylight/internal/modules/calendar/domain"
	"waylight/internal/modules/calendar/usecase"
	"waylight/internal/shared/httpapi"
)

// AccountHandlers serve the connected Google and Microsoft accounts (#282).
type AccountHandlers struct{ uc usecase.Accounts }

// NewAccountHandlers returns the handlers.
func NewAccountHandlers(uc usecase.Accounts) *AccountHandlers { return &AccountHandlers{uc: uc} }

// List handles GET /api/calendar/accounts.
func (h *AccountHandlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.ListAccounts(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	out := AccountsDTO{RedirectURL: h.uc.RedirectURL(), Accounts: make([]AccountDTO, 0, len(list))}
	for _, a := range list {
		out.Accounts = append(out.Accounts, account(a))
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}

func account(a usecase.AccountView) AccountDTO {
	cals := make([]AccountCalendarDTO, 0, len(a.Calendars))
	for _, c := range a.Calendars {
		cals = append(cals, AccountCalendarDTO(c))
	}
	return AccountDTO{ID: a.ID, Provider: string(a.Provider), Name: a.Name, Calendars: cals}
}

// Start handles POST /api/calendar/accounts/oauth/start.
func (h *AccountHandlers) Start(w http.ResponseWriter, r *http.Request) {
	var dto SignInInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	s, err := h.uc.StartSignIn(r.Context(), domain.Provider(dto.Provider), dto.ClientID, dto.ClientSecret)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, SignInStartDTO{URL: s.URL, State: s.State})
}

// callbackPage is shown in the sign-in window; the settings dialog notices
// the result by itself, so the page needs no script.
const callbackPage = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width">
<title>Waylight</title><style>body{font:16px system-ui,sans-serif;margin:3rem auto;max-width:32rem;padding:0 1rem}</style></head>
<body><h1>%s</h1><p>%s</p></body></html>`

// Callback handles the provider's redirect with a code (or an error). It is
// public because the session cookie may not be sent on a cross-site
// redirect; the state, known only to this server, ties it to the sign-in.
func (h *AccountHandlers) Callback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	err := h.uc.FinishSignIn(r.Context(), q.Get("state"), q.Get("code"), q.Get("error"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = fmt.Fprintf(w, callbackPage, "Sign-in failed / Anmeldung fehlgeschlagen",
			"Close this window and check client ID, client secret and the redirect address. / Schließe dieses Fenster und prüfe Client-ID, Client-Secret und Weiterleitungsadresse.")
		return
	}
	_, _ = fmt.Fprintf(w, callbackPage, "Signed in / Angemeldet",
		"You can close this window; the calendars appear in Waylight. / Du kannst dieses Fenster schließen; die Kalender erscheinen in Waylight.")
}

// Status handles GET /api/calendar/accounts/oauth/state/{state}.
func (h *AccountHandlers) Status(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.SignInStatus(r.PathValue("state"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, SignInStatusDTO{Status: s})
}

// SetCalendars handles PUT /api/calendar/accounts/{id}.
func (h *AccountHandlers) SetCalendars(w http.ResponseWriter, r *http.Request) {
	var dto AccountCalendarsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	a, err := h.uc.SetCalendars(r.Context(), r.PathValue("id"), dto.Enabled)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, account(a))
}

// Disconnect handles DELETE /api/calendar/accounts/{id}.
func (h *AccountHandlers) Disconnect(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Disconnect(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Writable handles GET /api/calendar/writable: calendars for new events.
func (h *AccountHandlers) Writable(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.Writable(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	out := make([]CalendarChoiceDTO, 0, len(list))
	for _, c := range list {
		out = append(out, CalendarChoiceDTO(c))
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}
