package controller

import (
	"net/http"

	"waylight/internal/modules/auth/adapter/presenter"
	"waylight/internal/modules/auth/domain"
	"waylight/internal/shared/httpapi"
	"waylight/internal/shared/kernel"
)

// linkingAccount is the account that may link a sign-in: a person signed in
// with a session, never a kiosk display or an API token.
func linkingAccount(r *http.Request) (string, bool) {
	p := kernel.PrincipalFrom(r.Context())
	return p.AccountID, p.AccountID != "" && !p.IsKiosk() && !p.IsToken() && !p.IsSystem()
}

// ListIdentities handles GET /api/auth/identities (#527).
func (h *Handlers) ListIdentities(w http.ResponseWriter, r *http.Request) {
	account, ok := linkingAccount(r)
	if h.uc.SSO == nil || !ok {
		httpapi.WriteJSON(w, http.StatusOK, presenter.Identities(nil))
		return
	}
	list, err := h.uc.SSO.Links(r.Context(), account)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Identities(list))
}

// BeginIdentityLink handles POST /api/auth/identities/link: it binds the
// state to this browser like OIDCStart and returns the provider URL, which
// the page then opens; the callback stores the link (#527).
func (h *Handlers) BeginIdentityLink(w http.ResponseWriter, r *http.Request) {
	account, ok := linkingAccount(r)
	if h.uc.SSO == nil {
		httpapi.WriteDomainError(w, r, domain.ErrSSOUnavailable)
		return
	}
	if !ok {
		httpapi.WriteDomainError(w, r, kernel.ErrForbidden)
		return
	}
	url, state, err := h.uc.SSO.BeginLink(r.Context(), account)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: oidcStateCookie, Value: state, Path: "/api/auth/oidc/", HttpOnly: true, //nolint:gosec // Secure follows COOKIE_SECURE, as in OIDCStart
		Secure: h.cookies.Secure, SameSite: http.SameSiteLaxMode, MaxAge: 600})
	httpapi.WriteJSON(w, http.StatusOK, presenter.LinkStartDTO{Redirect: url})
}

// DeleteIdentity handles DELETE /api/auth/identities/{subject}.
func (h *Handlers) DeleteIdentity(w http.ResponseWriter, r *http.Request) {
	account, ok := linkingAccount(r)
	if h.uc.SSO == nil || !ok {
		httpapi.WriteDomainError(w, r, domain.ErrIdentityNotFound)
		return
	}
	client := domain.NewClient(r.UserAgent(), clientKey(r))
	if err := h.uc.SSO.Unlink(r.Context(), account, r.PathValue("subject"), client); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SignUpWithSSO handles POST /api/auth/link/sso: accept an invitation
// through the provider instead of a password (#528). Like BeginIdentityLink
// it binds the state to this browser and returns the provider URL; the
// callback creates the account and signs in.
func (h *Handlers) SignUpWithSSO(w http.ResponseWriter, r *http.Request) {
	var dto LinkTokenDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if h.uc.SSO == nil {
		httpapi.WriteDomainError(w, r, domain.ErrSSOUnavailable)
		return
	}
	url, state, err := h.uc.SSO.BeginSignUp(r.Context(), dto.Token, clientKey(r))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: oidcStateCookie, Value: state, Path: "/api/auth/oidc/", HttpOnly: true, //nolint:gosec // Secure follows COOKIE_SECURE, as in OIDCStart
		Secure: h.cookies.Secure, SameSite: http.SameSiteLaxMode, MaxAge: 600})
	httpapi.WriteJSON(w, http.StatusOK, presenter.LinkStartDTO{Redirect: url})
}

// DeviceFlowRequestDTO is the body of POST /api/auth/microsoft/poll.
type DeviceFlowRequestDTO struct {
	Flow string `json:"flow"`
}

// StartMicrosoftLink handles POST /api/auth/microsoft/link: link a Microsoft
// account to the signed-in account by code (#535).
func (h *Handlers) StartMicrosoftLink(w http.ResponseWriter, r *http.Request) {
	account, ok := linkingAccount(r)
	if !ok {
		httpapi.WriteDomainError(w, r, kernel.ErrForbidden)
		return
	}
	h.startDevice(w, r, account)
}

// StartMicrosoftSignIn handles POST /api/auth/microsoft/start (public):
// sign in with a linked Microsoft account by code (#535).
func (h *Handlers) StartMicrosoftSignIn(w http.ResponseWriter, r *http.Request) {
	h.startDevice(w, r, "")
}

func (h *Handlers) startDevice(w http.ResponseWriter, r *http.Request, account string) {
	if h.uc.SSO == nil {
		httpapi.WriteDomainError(w, r, domain.ErrSSOUnavailable)
		return
	}
	f, err := h.uc.SSO.BeginDevice(r.Context(), account)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.DeviceFlow(f))
}

// PollMicrosoft handles POST /api/auth/microsoft/poll (public): 202 while
// the code is not confirmed, then the result: linked, a session (sets the
// cookie, as the password login) or a ticket for the local second factor.
func (h *Handlers) PollMicrosoft(w http.ResponseWriter, r *http.Request) {
	var dto DeviceFlowRequestDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if h.uc.SSO == nil {
		httpapi.WriteDomainError(w, r, domain.ErrSSOUnavailable)
		return
	}
	client := domain.NewClient(r.UserAgent(), clientKey(r))
	res, err := h.uc.SSO.PollDevice(r.Context(), dto.Flow, client)
	switch {
	case err != nil:
		httpapi.WriteDomainError(w, r, err)
	case res.Pending:
		httpapi.WriteJSON(w, http.StatusAccepted, presenter.DevicePollDTO{Pending: true})
	case res.Linked:
		httpapi.WriteJSON(w, http.StatusOK, presenter.DevicePollDTO{Linked: true})
	case res.Ticket != "":
		httpapi.WriteJSON(w, http.StatusOK, presenter.DevicePollDTO{Ticket: res.Ticket})
	default:
		h.respondWithSession(w, r, res.Session, nil)
	}
}
