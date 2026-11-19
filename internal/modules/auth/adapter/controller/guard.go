package controller

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"waylight/internal/modules/auth/domain"
	"waylight/internal/modules/auth/usecase"
	"waylight/internal/shared/httpapi"
	"waylight/internal/shared/kernel"
)

// CSRFHeader carries the per-session CSRF token on state-changing requests.
const CSRFHeader = "X-CSRF-Token"

// Guard is middleware that only lets requests with a valid session through.
// State-changing methods additionally need the session's CSRF token, which a
// cross-site attacker cannot read (synchronizer token pattern). Bearer
// tokens are either app sessions (#559) or personal API tokens.
func (h *Handlers) Guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if raw, ok := bearerToken(r); ok {
			if domain.IsAppSessionToken(raw) {
				h.serveWithSession(w, r, raw, false, next)
				return
			}
			if h.uc.APITokens != nil {
				h.serveWithToken(w, r, raw, next)
				return
			}
		}
		h.serveWithSession(w, r, h.cookies.read(r), true, next)
	})
}

// serveWithSession authenticates a session token from the cookie or, for
// app sessions, from the Authorization header. Only cookies need the CSRF
// token: a browser attaches them to cross-site requests on its own, but
// never a bearer token.
func (h *Handlers) serveWithSession(w http.ResponseWriter, r *http.Request, raw string, viaCookie bool, next http.Handler) {
	res, err := h.uc.Authenticate.Execute(r.Context(), usecase.AuthenticateRequest{SessionToken: raw})
	if errors.Is(err, domain.ErrSessionInvalid) && viaCookie {
		h.cookies.clear(w)
	}
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if viaCookie && !isSafeMethod(r.Method) && !validCSRF(r.Header.Get(CSRFHeader), res.CSRFToken) {
		httpapi.WriteDomainError(w, r, kernel.NewPublicError(kernel.ErrForbidden, "Missing or invalid CSRF token."))
		return
	}
	if viaCookie && res.Refreshed {
		h.cookies.set(w, raw, res.ExpiresAt)
	}
	if err := h.uc.Authorizer.Authorize(r.Context(), res.Principal, r.Method, r.URL.Path); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	ctx := kernel.WithPrincipal(r.Context(), res.Principal)
	ctx = usecase.WithClient(ctx, domain.NewClient(r.UserAgent(), clientKey(r)))
	next.ServeHTTP(w, r.WithContext(ctx))
}

// sessionToken returns the caller's session token: the bearer token of an
// app session, or else the session cookie.
func (h *Handlers) sessionToken(r *http.Request) string {
	if raw, ok := bearerToken(r); ok && domain.IsAppSessionToken(raw) {
		return raw
	}
	return h.cookies.read(r)
}

func isSafeMethod(m string) bool {
	return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
}

func validCSRF(got, want string) bool {
	return got != "" && subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

// bearerToken returns the token of an "Authorization: Bearer" header.
func bearerToken(r *http.Request) (string, bool) {
	scheme, raw, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(raw) == "" {
		return "", false
	}
	return strings.TrimSpace(raw), true
}

// serveWithToken authenticates an API token. No cookie is involved, so a
// browser cannot be tricked into sending it: CSRF protection is not needed.
func (h *Handlers) serveWithToken(w http.ResponseWriter, r *http.Request, raw string, next http.Handler) {
	client := domain.NewClient(r.UserAgent(), clientKey(r))
	p, err := h.uc.APITokens.Authenticate(r.Context(), raw, client)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.Authorizer.Authorize(r.Context(), p, r.Method, r.URL.Path); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	ctx := usecase.WithClient(kernel.WithPrincipal(r.Context(), p), client)
	next.ServeHTTP(w, r.WithContext(ctx))
}
