package controller

import (
	"fmt"
	"net/http"
	"time"

	"waylight/internal/modules/todos/adapter/presenter"
	"waylight/internal/modules/todos/usecase"
	"waylight/internal/shared/httpapi"
	"waylight/internal/shared/kernel"
)

// UseCases bundles the to-do input ports.
type UseCases struct {
	List           usecase.ListTodos
	Create         usecase.CreateTodo
	Update         usecase.UpdateTodo
	Delete         usecase.DeleteTodo
	ClearCompleted usecase.ClearCompleted
	// Sources manages connected external lists (#169); nil = not offered.
	Sources usecase.Sources
	// Share changes the visibility of a local to-do (ADR 0018, #301).
	Share usecase.ShareTodo
}

// Handlers serves the to-do endpoints.
type Handlers struct{ uc UseCases }

// NewHandlers returns to-do handlers.
func NewHandlers(uc UseCases) *Handlers { return &Handlers{uc: uc} }

func (d TodoInputDTO) toInput() (usecase.TodoInput, error) {
	in := usecase.TodoInput{Title: d.Title, ID: d.ID, Version: d.Version}
	if d.Due != "" {
		due, err := time.Parse(time.DateOnly, d.Due)
		if err != nil {
			return in, kernel.NewValidationError("due", "The due date must look like 2026-12-24.")
		}
		in.Due = due
	}
	if d.Repeat != nil {
		in.Repeat = &usecase.RepeatInput{Kind: d.Repeat.Kind, Every: d.Repeat.Every, Days: d.Repeat.Days}
	}
	if d.Reminder != nil {
		in.Reminder = &usecase.ReminderInput{At: d.Reminder.At, DaysBefore: d.Reminder.DaysBefore}
	}
	return in, nil
}

// source is the list a request is about: ?source=<id>, empty = local.
func source(r *http.Request) string { return r.URL.Query().Get("source") }

// List handles GET /api/todos?done=true&source=<id>.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.List.Execute(r.Context(), source(r), r.URL.Query().Get("done") == "true")
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Todos(res))
}

// Create handles POST /api/todos.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto TodoInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	in, err := dto.toInput()
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	in.Source = source(r)
	res, err := h.uc.Create.Execute(r.Context(), in)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Todo(res))
}

// Update handles PUT /api/todos/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto TodoInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	in, err := dto.toInput()
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	in.Source = source(r)
	res, err := h.uc.Update.Execute(r.Context(), r.PathValue("id"), in, dto.Done)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Todo(res))
}

// Share handles PUT /api/todos/{id}/visibility.
func (h *Handlers) Share(w http.ResponseWriter, r *http.Request) {
	var dto VisibilityDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Share.Execute(r.Context(), r.PathValue("id"), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Todo(v))
}

// Delete handles DELETE /api/todos/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete.Execute(r.Context(), source(r), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ClearCompleted handles POST /api/todos/clear-completed.
func (h *Handlers) ClearCompleted(w http.ResponseWriter, r *http.Request) {
	n, err := h.uc.ClearCompleted.Execute(r.Context(), source(r))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, ClearedDTO{Removed: n})
}

// Sources handles GET /api/todos/settings/sources.
func (h *Handlers) Sources(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Sources.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Sources(res))
}

func (h *Handlers) saveSource(w http.ResponseWriter, r *http.Request, id string, status int) {
	var dto SourceInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Sources.Save(r.Context(), dto.toInput(id))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, status, presenter.Source(res))
}

// CreateSource handles POST /api/todos/settings/sources (tests the connection).
func (h *Handlers) CreateSource(w http.ResponseWriter, r *http.Request) {
	h.saveSource(w, r, "", http.StatusCreated)
}

// UpdateSource handles PUT /api/todos/settings/sources/{id}; an empty
// password keeps the stored one.
func (h *Handlers) UpdateSource(w http.ResponseWriter, r *http.Request) {
	h.saveSource(w, r, r.PathValue("id"), http.StatusOK)
}

// DeleteSource handles DELETE /api/todos/settings/sources/{id}.
func (h *Handlers) DeleteSource(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Sources.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DiscoverSources handles POST /api/todos/settings/sources/discover: the
// task lists of an account (id names a stored list whose password is used
// when password is empty).
func (h *Handlers) DiscoverSources(w http.ResponseWriter, r *http.Request) {
	var dto SourceInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Sources.Discover(r.Context(), dto.toInput(dto.ID))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Discovered(res))
}

// OAuthInfo handles GET /api/todos/settings/oauth.
func (h *Handlers) OAuthInfo(w http.ResponseWriter, _ *http.Request) {
	httpapi.WriteJSON(w, http.StatusOK, OAuthInfoDTO{RedirectURL: h.uc.Sources.OAuthInfo().RedirectURL})
}

// StartOAuth handles POST /api/todos/settings/oauth/start: kind, client ID
// (username) and client secret (password) of the user's app registration.
func (h *Handlers) StartOAuth(w http.ResponseWriter, r *http.Request) {
	var dto SourceInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Sources.StartOAuth(r.Context(), dto.toInput(""))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, OAuthStartDTO{URL: res.URL, State: res.State})
}

// callbackPage is shown in the sign-in window; the settings dialog notices
// the result by itself, so the page needs no script.
const callbackPage = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width">
<title>waylight</title><style>body{font:16px system-ui,sans-serif;margin:3rem auto;max-width:32rem;padding:0 1rem}</style></head>
<body><h1>%s</h1><p>%s</p></body></html>`

// OAuthCallback handles GET /api/todos/settings/oauth/callback, where the
// provider sends the browser back with a code (or an error).
func (h *Handlers) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	err := h.uc.Sources.FinishOAuth(r.Context(), q.Get("state"), q.Get("code"), q.Get("error"))
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
		"You can close this window and choose the list in waylight. / Du kannst dieses Fenster schließen und in waylight die Liste wählen.")
}

// OAuthStatus handles GET /api/todos/settings/oauth/state/{state}.
func (h *Handlers) OAuthStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.uc.Sources.OAuthStatus(r.PathValue("state"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, OAuthStatusDTO{Status: status})
}
