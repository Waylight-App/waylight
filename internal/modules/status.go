package modules

import (
	"context"
	"errors"

	"waylight/internal/shared/kernel"
)

// State is how a module's connection to an outside service is doing (#599).
type State string

const (
	// StateOK means the connection works or the module is not in use.
	StateOK State = "ok"
	// StateSetup means the connection has not been set up yet.
	StateSetup State = "setup"
	// StateFailing means a connection that was set up stopped working.
	StateFailing State = "failing"
)

// Action is the one thing a client offers to resolve an attention entry.
// The values are part of the API of GET /api/attention.
type Action string

const (
	// ActionSignIn means the sign-in at the service expired; sign in again.
	ActionSignIn Action = "signIn"
	// ActionCheck means the service did not answer; check again later.
	ActionCheck Action = "check"
	// ActionFix means the service refused the stored credential; edit it.
	ActionFix Action = "fix"
	// ActionChoose means the widget misses a setting of its own (a stop, a place).
	ActionChoose Action = "choose"
	// ActionConnect means the integration is not set up yet.
	ActionConnect Action = "connect"
)

// ConnectionStatus is the answer of a StatusReporter.
type ConnectionStatus struct {
	State State
	// Reason explains a setup or failing state; its template is a message
	// key clients translate (#102). Nil for StateOK.
	Reason *kernel.PublicError
	Action Action
}

// StatusReporter is implemented by modules that connect to a service outside
// Waylight. The attention list (#599) asks it, so it must be cheap: answer
// from a cache where the module has one. For InstanceData modules the status
// is the instance's; for PersonalSettings modules it is the status of the
// account in ctx.
type StatusReporter interface {
	ConnectionStatus(ctx context.Context) ConnectionStatus
}

// SetupNeed is the one missing piece of a widget's own settings.
type SetupNeed struct {
	// Reason names what is missing, as a translatable message.
	Reason *kernel.PublicError
	// Field is the setting the widget misses, e.g. "stopId", so a client can
	// open the widget settings at it.
	Field string
}

// WidgetChecker is implemented by modules whose widgets need settings of
// their own before they can show anything, such as a stop or a place.
type WidgetChecker interface {
	// CheckWidget returns what the widget of kind with the given settings
	// (JSON) misses, or nil when it is complete.
	CheckWidget(kind, config string) *SetupNeed
}

// errNoAnswer explains a failure whose cause has no message of its own.
var errNoAnswer = kernel.NewPublicError(kernel.ErrUnavailable, "The service does not answer.")

// StatusOf turns the outcome of a cheap call to the service into a status.
// notConfigured is the module's own error for a connection that was never
// set up; refused credentials ask for a fix, everything else for a check.
func StatusOf(err, notConfigured error) ConnectionStatus {
	if err == nil {
		return ConnectionStatus{State: StateOK}
	}
	reason := errNoAnswer
	var pub *kernel.PublicError
	if errors.As(err, &pub) {
		reason = pub
	}
	switch {
	case notConfigured != nil && errors.Is(err, notConfigured):
		return ConnectionStatus{State: StateSetup, Reason: reason, Action: ActionConnect}
	case errors.Is(err, kernel.ErrUnauthorized), errors.Is(err, kernel.ErrForbidden):
		return ConnectionStatus{State: StateFailing, Reason: reason, Action: ActionFix}
	default:
		return ConnectionStatus{State: StateFailing, Reason: reason, Action: ActionCheck}
	}
}
