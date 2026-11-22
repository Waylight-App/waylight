package domain

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// Limits of the store.
const (
	// MaxConnections bounds the connections of an instance; a choice with
	// more is no longer a quick pick in a widget.
	MaxConnections = 100
	// MaxNameLength limits the name, e.g. "Photos at home".
	MaxNameLength = 60
	// MaxFields bounds the credential fields (user name, password, token
	// ID and secret are the most any service needs).
	MaxFields = 4
	// MaxValueLength bounds one credential value; tokens are far shorter.
	MaxValueLength = 4096
)

var (
	// servicePattern matches the module names that own a service, e.g.
	// "immich" or "smarthome".
	servicePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,31}$`)
	// fieldPattern matches credential field names such as "apiKey".
	fieldPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]{0,31}$`)
)

// Errors of the store.
var (
	ErrNotFound  = kernel.NewPublicError(kernel.ErrNotFound, "The connection does not exist.")
	ErrTooMany   = kernel.NewPublicErrorf(kernel.ErrInvalid, "At most {max} connections can be saved.", "max", MaxConnections)
	ErrNameTaken = kernel.NewPublicError(kernel.ErrConflict, "A connection of this service with this name exists already.")
)

// Credential holds the secret fields of a connection by name, e.g.
// {"apiKey": "…"} or {"username": "…", "password": "…"}. It is stored
// encrypted and never leaves the server.
type Credential map[string]string

// Fields lists the names of the fields that are set, sorted; the API shows
// these instead of the values.
func (c Credential) Fields() []string {
	out := make([]string, 0, len(c))
	for k, v := range c {
		if v != "" {
			out = append(out, k)
		}
	}
	slices.Sort(out)
	return out
}

// NewCredential validates the fields; empty values are dropped.
func NewCredential(fields map[string]string) (Credential, error) {
	out := Credential{}
	for k, v := range fields {
		if !fieldPattern.MatchString(k) {
			return nil, kernel.NewValidationError("credential", "A credential field has an invalid name.")
		}
		if utf8.RuneCountInString(v) > MaxValueLength {
			return nil, kernel.NewValidationErrorf("credential", "A credential value must be at most {max} characters long.", "max", MaxValueLength)
		}
		if v != "" {
			out[k] = v
		}
	}
	if len(out) > MaxFields {
		return nil, kernel.NewValidationErrorf("credential", "A credential has at most {max} fields.", "max", MaxFields)
	}
	return out, nil
}

// Connection is a named connection to one service (#610): an address and
// the credential to use there. Several widgets choose the same connection,
// and a service can have several connections, e.g. two Proxmox clusters.
type Connection struct {
	ID string
	// Service is the module that talks to the service, e.g. "immich".
	Service string
	// Name is unique per service, e.g. "Home" or "Office".
	Name string
	// URL is the address of the service; a path is allowed for reverse
	// proxies and sub paths.
	URL        string
	Credential Credential
	CreatedAt  time.Time
}

// NewConnection validates and creates a connection.
func NewConnection(id, service, name, rawURL string, cred Credential, now time.Time) (Connection, error) {
	service = strings.TrimSpace(service)
	if !servicePattern.MatchString(service) {
		return Connection{}, kernel.NewValidationError("service", "The service is not valid.")
	}
	c := Connection{ID: id, Service: service, Credential: maps.Clone(cred), CreatedAt: now}
	if c.Credential == nil {
		c.Credential = Credential{}
	}
	if err := c.Change(name, rawURL); err != nil {
		return Connection{}, err
	}
	return c, nil
}

// Change renames the connection or moves it to another address after the
// same checks as NewConnection.
func (c *Connection) Change(name, rawURL string) error {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > MaxNameLength {
		return kernel.NewValidationErrorf("name", "The name must be between 1 and {max} characters long.", "max", MaxNameLength)
	}
	u, err := kernel.ParseTarget("url", strings.TrimRight(strings.TrimSpace(rawURL), "/"), kernel.WithTargetPath())
	if err != nil {
		return err
	}
	c.Name, c.URL = name, u.String()
	return nil
}

// SameName reports whether two names clash: names are compared without
// case, as people read them.
func SameName(a, b string) bool { return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b)) }

// Use is a widget whose settings name a connection ("connectionId", or
// one of "connectionIds").
type Use struct {
	WidgetID, TabID, Kind string
	// Mine is true when the widget is on a tab of the signed-in account;
	// other accounts' tabs are named without their title.
	Mine bool
}
