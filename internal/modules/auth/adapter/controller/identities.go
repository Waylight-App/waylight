package controller

import (
	"net/http"

	"trisse-home/internal/modules/auth/adapter/presenter"
	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
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
