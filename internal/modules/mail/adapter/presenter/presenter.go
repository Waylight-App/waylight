// Package presenter maps mail accounts and inboxes to JSON (#283).
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/mail/usecase"
)

// AccountDTO never contains the password.
type AccountDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Folder     string `json:"folder"`
	WebmailURL string `json:"webmailUrl"`
}

// MessageDTO is one unread message (headers only).
type MessageDTO struct {
	Account string `json:"account"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
}

// InboxDTO is returned by GET /api/mail/inbox.
type InboxDTO struct {
	Unread     int          `json:"unread"`
	Messages   []MessageDTO `json:"messages"`
	WebmailURL string       `json:"webmailUrl"`
	Failed     []string     `json:"failed"`
}

// Accounts maps account views.
func Accounts(vs []usecase.AccountView) []AccountDTO {
	out := make([]AccountDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, Account(v))
	}
	return out
}

// Account maps one account view.
func Account(v usecase.AccountView) AccountDTO { return AccountDTO(v) }

// Inbox maps the inbox view.
func Inbox(v usecase.InboxView) InboxDTO {
	out := InboxDTO{Unread: v.Unread, Messages: make([]MessageDTO, 0, len(v.Messages)), WebmailURL: v.WebmailURL, Failed: append([]string{}, v.Failed...)}
	for _, m := range v.Messages {
		d := ""
		if !m.Date.IsZero() {
			d = m.Date.Format(time.RFC3339)
		}
		out.Messages = append(out.Messages, MessageDTO{Account: m.Account, From: m.From, Subject: m.Subject, Date: d})
	}
	return out
}
