// Package controller serves the mail endpoints (#283).
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"waylight/internal/modules/mail/adapter/presenter"
	"waylight/internal/modules/mail/usecase"
	"waylight/internal/shared/httpapi"
)

// AccountInputDTO is the body of POST/PUT /api/mail/settings/accounts; an
// empty password keeps the stored one.
type AccountInputDTO struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Folder     string `json:"folder"`
	WebmailURL string `json:"webmailUrl"`
}

// Handlers serves the endpoints.
type Handlers struct{ uc usecase.Mail }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Mail) *Handlers { return &Handlers{uc: uc} }

// Accounts handles GET /api/mail/settings/accounts.
func (h *Handlers) Accounts(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.Accounts(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Accounts(list))
}

func (h *Handlers) save(w http.ResponseWriter, r *http.Request, id string, status int) {
	var dto AccountInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.SaveAccount(r.Context(), usecase.AccountInput{ID: id, Name: dto.Name, Host: dto.Host, Port: dto.Port,
		Username: dto.Username, Password: dto.Password, Folder: dto.Folder, WebmailURL: dto.WebmailURL})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, status, presenter.Account(v))
}

// CreateAccount handles POST /api/mail/settings/accounts (tests the login).
func (h *Handlers) CreateAccount(w http.ResponseWriter, r *http.Request) {
	h.save(w, r, "", http.StatusCreated)
}

// UpdateAccount handles PUT /api/mail/settings/accounts/{id}.
func (h *Handlers) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	h.save(w, r, r.PathValue("id"), http.StatusOK)
}

// DeleteAccount handles DELETE /api/mail/settings/accounts/{id}.
func (h *Handlers) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteAccount(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Inbox handles GET /api/mail/inbox?account=<id> ("" = all accounts).
func (h *Handlers) Inbox(w http.ResponseWriter, r *http.Request) {
	v, err := h.uc.Inbox(r.Context(), r.URL.Query().Get("account"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Inbox(v))
}
