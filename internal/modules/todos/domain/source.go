package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// LocalSource is the ID of the built-in list stored in waylight.
const LocalSource = "local"

// SourceKind is the type of an external to-do provider (#169).
type SourceKind string

// Provider kinds.
const (
	KindCalDAV  SourceKind = "caldav"
	KindTodoist SourceKind = "todoist" // #170
	// Signed in with OAuth and the user's own app registration (#171).
	KindMicrosoft SourceKind = "microsoft" // Microsoft To Do (Graph)
	KindGoogle    SourceKind = "google"    // Google Tasks
	KindTickTick  SourceKind = "ticktick"
)

// OAuth reports whether lists of this kind are connected by signing in.
func (k SourceKind) OAuth() bool {
	return k == KindMicrosoft || k == KindGoogle || k == KindTickTick
}

// OAuthGrant is the secret part of a signed-in list: the client secret of
// the user's app registration and the tokens. It is stored encrypted.
type OAuthGrant struct {
	ClientSecret string
	Access       string
	Refresh      string
	Expiry       time.Time // zero = unknown
}

// ErrSignInAgain is returned when the provider no longer accepts the grant.
var ErrSignInAgain = kernel.NewPublicError(kernel.ErrUnavailable, "The sign-in has expired or was revoked. Disconnect the list and connect it again.")

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
	ListID   string // Todoist, OAuth kinds: the project or list ID
	// Grant holds client secret and tokens of OAuth kinds; Username is
	// the client ID of the user's app registration.
	Grant OAuthGrant
}

var providerIDPattern = regexp.MustCompile(`^[^\x00-\x1f\x7f]{1,512}$`)

// NewOAuthSource validates a list of a kind that is connected by signing in.
func NewOAuthSource(id, name string, kind SourceKind, clientID, listID string, g OAuthGrant) (Source, error) {
	name = strings.Join(strings.Fields(name), " ")
	switch {
	case name == "" || utf8.RuneCountInString(name) > 80:
		return Source{}, kernel.NewValidationError("name", "The name must have between 1 and 80 characters.")
	case !kind.OAuth():
		return Source{}, kernel.NewValidationError("kind", "Unknown kind of to-do list.")
	case strings.TrimSpace(clientID) == "" || len(clientID) > 200:
		return Source{}, kernel.NewValidationError("username", "Enter the client ID of your app registration.")
	case !providerIDPattern.MatchString(listID) || listID == "." || listID == "..":
		return Source{}, kernel.NewValidationError("listId", "Choose a list.")
	case g.Access == "":
		return Source{}, ErrSignInAgain
	}
	return Source{ID: id, Name: name, Kind: kind, Username: strings.TrimSpace(clientID), ListID: listID, Grant: g}, nil
}

// ValidProviderID reports whether id can name a to-do of a list that is
// not CalDAV (IDs of APIs; they are always escaped in URLs).
func ValidProviderID(id string) bool {
	return providerIDPattern.MatchString(id) && id != "." && id != ".."
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
