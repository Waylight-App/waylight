package domain

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// LocalSource is the ID of the built-in list stored in trisse-home.
const LocalSource = "local"

// SourceKind is the type of an external to-do provider (#169).
type SourceKind string

// Provider kinds.
const (
	KindCalDAV  SourceKind = "caldav"
	KindTodoist SourceKind = "todoist" // #170
)

// MaxSources limits the configured external lists.
const MaxSources = 20

// Errors of the sources.
var (
	ErrSourceNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The to-do list does not exist (any more).")
	ErrTooManySources = kernel.NewValidationError("name", "At most 20 to-do lists can be connected.")
	// ErrRateLimited is returned while a service asks to slow down.
	ErrRateLimited = kernel.NewPublicError(kernel.ErrUnavailable, "The service asks to slow down. The list is loaded again in a minute.")
	// ErrUnsupported rejects actions the provider of a list cannot store.
	ErrUnsupported = kernel.NewValidationError("repeat", "This list does not support repeating to-dos or reminders.")
)

// Capabilities say what a provider can store; the UI hides the rest.
type Capabilities struct {
	Ordering  bool
	Repeat    bool
	Reminders bool
}

// CapabilitiesOf returns what a list of the given kind supports.
func CapabilitiesOf(kind SourceKind) Capabilities {
	if kind == "" {
		return Capabilities{Ordering: true, Repeat: true, Reminders: true}
	}
	return Capabilities{}
}

// Source is a connected external to-do list. Password is a secret: it is
// stored encrypted and never returned by the API.
type Source struct {
	ID       string
	Name     string
	Kind     SourceKind
	URL      string // CalDAV: the list collection, ending with "/"
	Username string
	Password string // CalDAV: app password; Todoist: API token
	ListID   string // Todoist: the project ID
}

var sourceIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// ValidSourceID reports whether id can name a source.
func ValidSourceID(id string) bool { return sourceIDPattern.MatchString(id) && id != LocalSource }

var listIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// NewSource validates a list: a CalDAV collection with user and app
// password, or a Todoist project (listID) with an API token as password.
func NewSource(id, name string, kind SourceKind, listURL, username, password, listID string) (Source, error) {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" || utf8.RuneCountInString(name) > 80 {
		return Source{}, kernel.NewValidationError("name", "The name must have between 1 and 80 characters.")
	}
	switch kind {
	case KindTodoist:
		token := strings.TrimSpace(password)
		if token == "" || len(token) > 200 || strings.ContainsAny(token, " \r\n") {
			return Source{}, kernel.NewValidationError("password", "Enter the API token from Todoist's settings (Integrations → Developer).")
		}
		if !listIDPattern.MatchString(listID) {
			return Source{}, kernel.NewValidationError("listId", "Choose a Todoist project.")
		}
		return Source{ID: id, Name: name, Kind: kind, Password: token, ListID: listID}, nil
	case KindCalDAV:
	default:
		return Source{}, kernel.NewValidationError("kind", "Unknown kind of to-do list.")
	}
	u, err := kernel.ParseTarget("url", listURL, kernel.WithTargetPath())
	if err != nil {
		return Source{}, err
	}
	if strings.TrimSpace(username) == "" || password == "" {
		return Source{}, kernel.NewValidationError("username", "User name and app password are required.")
	}
	return Source{ID: id, Name: name, Kind: kind, URL: u.String() + "/", Username: strings.TrimSpace(username), Password: password}, nil
}

var objectNamePattern = regexp.MustCompile(`^[^/\x00-\x1f]{1,200}$`)

// ValidExternalID reports whether id can name a to-do of an external list
// (a CalDAV object name such as "abc.ics").
func ValidExternalID(id string) bool {
	return objectNamePattern.MatchString(id) && id != "." && id != ".." && !strings.Contains(id, "\\")
}
