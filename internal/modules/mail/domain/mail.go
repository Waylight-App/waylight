// Package domain holds the mail widget rules (#283).
//
// Ring: Entities.
package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// Limits.
const (
	MaxAccounts = 10
	// Newest is how many unread messages are listed per account.
	Newest = 10
)

// Errors.
var (
	ErrAccountNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The mail account does not exist (any more).")
	ErrTooManyAccounts = kernel.NewValidationError("name", "At most 10 mail accounts can be added.")
)

// Account is an IMAP inbox; Password is a secret (stored encrypted, never
// returned).
type Account struct {
	ID       string
	Name     string
	Host     string
	Port     int
	Username string
	Password string
	Folder   string
	// WebmailURL is opened by a click on the widget (optional).
	WebmailURL string
}

var hostName = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.-]{0,251}[A-Za-z0-9])?$`)

// NewAccount validates an account; port 0 means 993, folder "" means INBOX.
func NewAccount(id, name, host string, port int, username, password, folder, webmail string) (Account, error) {
	a := Account{ID: id, Name: strings.Join(strings.Fields(name), " "), Host: strings.ToLower(strings.TrimSpace(host)), Port: port,
		Username: strings.TrimSpace(username), Password: password, Folder: strings.TrimSpace(folder), WebmailURL: strings.TrimSpace(webmail)}
	if a.Port == 0 {
		a.Port = 993
	}
	if a.Folder == "" {
		a.Folder = "INBOX"
	}
	switch {
	case a.Name == "" || utf8.RuneCountInString(a.Name) > 60:
		return Account{}, kernel.NewValidationError("name", "The name must have between 1 and 60 characters.")
	case !hostName.MatchString(a.Host):
		return Account{}, kernel.NewValidationError("host", "Enter the IMAP server name, e.g. imap.example.com.")
	case a.Port < 1 || a.Port > 65535:
		return Account{}, kernel.NewValidationError("port", "The port must be between 1 and 65535 (IMAP over TLS: 993).")
	case a.Username == "" || a.Password == "":
		return Account{}, kernel.NewValidationError("username", "User name and app password are required.")
	case strings.ContainsAny(a.Username+a.Password+a.Folder, "\r\n\x00") || len(a.Folder) > 200:
		return Account{}, kernel.NewValidationError("folder", "User name, password and folder must not contain line breaks.")
	}
	if a.WebmailURL != "" {
		u, err := kernel.ParseTarget("webmailUrl", a.WebmailURL, kernel.WithTargetQuery())
		if err != nil {
			return Account{}, err
		}
		a.WebmailURL = u.String()
	}
	return a, nil
}

// Message is the header of one unread message; the body is never read.
type Message struct {
	From    string
	Subject string
	Date    time.Time
}

// Inbox is what the widget shows of one account.
type Inbox struct {
	Unread   int
	Messages []Message
}

// SenderName returns the display name of a From header, or the address.
func SenderName(from string) string {
	from = strings.TrimSpace(from)
	if i := strings.Index(from, "<"); i > 0 {
		if name := strings.Trim(strings.TrimSpace(from[:i]), `"'`); name != "" {
			return name
		}
	}
	return strings.Trim(from, "<>")
}
