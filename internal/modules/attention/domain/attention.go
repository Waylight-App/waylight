package domain

import (
	"cmp"
	"slices"
	"time"
)

// Kind separates what stopped working from what was never set up.
type Kind string

const (
	// KindConnection is a connection that was set up and stopped working.
	KindConnection Kind = "connection"
	// KindSetup is a widget or integration that is not set up yet.
	KindSetup Kind = "setup"
)

// State is how a connection is doing, as a source reports it.
type State string

// The states a source reports.
const (
	// StateOK means the connection works.
	StateOK State = "ok"
	// StateSetup means it was never set up.
	StateSetup State = "setup"
	// StateFailing means it was set up and stopped working.
	StateFailing State = "failing"
)

// Action is the one thing a client offers for an entry: signIn, check, fix,
// choose or connect. The values are part of the public API.
type Action string

// The actions of the public API.
const (
	// ActionSignIn asks to sign in at the service again.
	ActionSignIn Action = "signIn"
	// ActionCheck asks to check again later; the service did not answer.
	ActionCheck Action = "check"
	// ActionFix asks to edit the refused credential.
	ActionFix Action = "fix"
	// ActionChoose asks to choose a missing widget setting.
	ActionChoose Action = "choose"
	// ActionConnect asks to set up the integration.
	ActionConnect Action = "connect"
)

// Message is a user-facing text as a template with {name} placeholders and
// its parameters; clients translate it by the template's message key.
type Message struct {
	// Text is the English message with the parameters filled in.
	Text     string
	Template string
	Params   map[string]string
}

// Status is what a source says about its connection.
type Status struct {
	State  State
	Reason Message
	Action Action
}

// Need is the missing piece of a widget's own settings.
type Need struct {
	Reason Message
	Field  string
}

// Item is one entry of the attention list.
type Item struct {
	// ID is stable while the entry exists, so clients can key and dismiss
	// animations on it.
	ID   string
	Kind Kind
	// Integration is the module the entry belongs to, e.g. "immich".
	Integration string
	// WidgetID and TabID point at the widget the entry belongs to; empty
	// for an instance integration that no widget of the account uses.
	WidgetID, TabID string
	// ConnectionID is the shared connection (#591) the entry belongs to;
	// empty until connections exist.
	ConnectionID string
	Title        string
	Reason       Message
	Since        time.Time
	Action       Action
	// Field is the widget setting to open for ActionChoose.
	Field string
}

// Sort puts what stopped working before what is not set up yet, and the
// oldest entry first within each group, so the list reads like a to-do
// list that does not reshuffle while it is open.
func Sort(items []Item) {
	rank := func(k Kind) int {
		if k == KindConnection {
			return 0
		}
		return 1
	}
	slices.SortStableFunc(items, func(a, b Item) int {
		return cmp.Or(
			cmp.Compare(rank(a.Kind), rank(b.Kind)),
			a.Since.Compare(b.Since),
			cmp.Compare(a.Title, b.Title),
			cmp.Compare(a.ID, b.ID),
		)
	})
}

// WidgetRef is a widget on one of the account's own tabs.
type WidgetRef struct {
	ID, TabID, Kind string
}

// Connection is how one integration's connection is doing, with the widgets
// of the account that use it, for Settings → Connections & services (#591).
type Connection struct {
	Integration string
	// Instance is true for a connection of the instance (admins set it up).
	Instance bool
	Status   Status
	Widgets  []WidgetRef
}
