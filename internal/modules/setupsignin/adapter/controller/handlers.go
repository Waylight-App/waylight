package controller

import (
	"fmt"
	"net/http"

	"waylight/internal/modules/setupsignin/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serve the sign-ins of the setup assistant (#654).
type Handlers struct{ uc usecase.SignIns }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.SignIns) *Handlers { return &Handlers{uc: uc} }

// Start handles POST /api/setup/signin.
func (h *Handlers) Start(w http.ResponseWriter, r *http.Request) {
	var dto SetupSignInInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	s, err := h.uc.Start(r.Context(), dto.Service, dto.URL)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, SetupSignInDTO{ID: s.ID, LoginURL: s.LoginURL})
}

// Status handles GET /api/setup/signin/{id}.
func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.Status(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpapi.WriteJSON(w, http.StatusOK, SetupSignInStatusDTO{Status: string(s)})
}

// Cancel handles DELETE /api/setup/signin/{id}.
func (h *Handlers) Cancel(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Cancel(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// callbackPage is shown in the sign-in window. The assistant notices the
// result by asking the server, so the page needs no script and works under
// the strict content security policy.
const callbackPage = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width">
<title>Waylight</title><style>body{font:16px system-ui,sans-serif;margin:3rem auto;max-width:32rem;padding:0 1rem}</style></head>
<body><main><h1>%s</h1><p>%s</p><p lang="de">%s</p></main></body></html>`

// Callback handles Home Assistant's redirect with a code. It is public
// because the session cookie is not sent on a cross-site redirect and the
// redirect goes to PUBLIC_URL, which may differ from the address the
// assistant runs on; the state, known only to this server and usable once,
// ties it to the sign-in.
func (h *Handlers) Callback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	err := h.uc.Callback(r.Context(), q.Get("state"), q.Get("code"), q.Get("error"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	// The URL carried the code; it must not leak to the next page.
	w.Header().Set("Referrer-Policy", "no-referrer")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, callbackPage, "Sign-in failed",
			"Close this window. Waylight shows the reason and lets you paste a token instead.",
			"Schließe dieses Fenster. Waylight zeigt den Grund und lässt dich stattdessen ein Token einfügen.")
		return
	}
	_, _ = fmt.Fprintf(w, callbackPage, "Signed in",
		"You can close this window and return to Waylight.",
		"Du kannst dieses Fenster schließen und zu Waylight zurückkehren.")
}
