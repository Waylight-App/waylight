package kernel

import "time"

// CompactLevel says how much a compact state needs the viewer's attention.
// Clients colour only attention and problem, so a dashboard of healthy
// widgets stays quiet (#593).
type CompactLevel string

// The levels of a compact state, from calm to urgent.
const (
	// CompactNone has nothing to judge: no data yet, not set up, or a
	// widget whose state is neither good nor bad (the next event).
	CompactNone CompactLevel = "none"
	// CompactOK means everything the widget watches is fine.
	CompactOK CompactLevel = "ok"
	// CompactAttention asks for a look soon (an overdue to-do, a disk
	// filling up).
	CompactAttention CompactLevel = "attention"
	// CompactProblem means something is broken (a service down, a failed
	// backup).
	CompactProblem CompactLevel = "problem"
)

// Rank orders levels by urgency, so callers can pick the most urgent of
// several states: none and ok are 0 and 1, attention 2, problem 3.
func (l CompactLevel) Rank() int {
	switch l {
	case CompactOK:
		return 1
	case CompactAttention:
		return 2
	case CompactProblem:
		return 3
	default:
		return 0
	}
}

// Text is a user-facing text that clients translate: Template is the
// English message with {name} placeholders, Params their values, and
// MessageKey(Template) the key of its translations (#102).
type Text struct {
	Template string
	Params   map[string]string
}

// NewText returns a Text without placeholders. The template must be a
// string literal so scripts/servermessages can add it to the catalogue.
func NewText(template string) Text { return Text{Template: template} }

// NewTextf returns a Text with placeholders: template contains {name} and
// params are name/value pairs.
func NewTextf(template string, params ...any) Text {
	return Text{Template: template, Params: pairs(params)}
}

// String is the English text with its parameters filled in.
func (t Text) String() string { return fill(t.Template, t.Params) }

// CompactState is the short state of a widget (#597): what the compact web
// view, the app's small tiles, home screen widgets and the watch show
// instead of the full data, computed once on the server so that every
// client says the same.
type CompactState struct {
	// Title is the one line every client shows ("11 / 12 · 1 problem").
	Title Text
	// Detail is an optional second line; zero when there is none.
	Detail Text
	Level  CompactLevel
	// Count is an optional number for badges (open to-dos, services down);
	// nil when the widget has none.
	Count *int
	// UpdatedAt is when the underlying data was last read or changed.
	UpdatedAt time.Time
	// Action is the one primary action of the compact view (#598); nil when
	// the widget offers none.
	Action *CompactAction
	// Empty says the widget has nothing to show (no upcoming events, no
	// habits, nothing to do). The web shows such a widget as a quiet empty
	// state (#860) and the simple view leaves it out (#681); the title still
	// reads well on its own.
	Empty bool
}

// CompactAction is a request the client sends when the action is used, so
// a client can offer it without knowing the module. Choices turn it into a
// row of chips (presence: office, home, away); each choice brings its own
// body.
type CompactAction struct {
	// ID names the action for clients and tests ("focus.start").
	ID     string
	Label  Text
	Method string
	// Path is the API path the request goes to, e.g. /api/focus/start.
	Path string
	// Body is sent as JSON; nil sends no body.
	Body    map[string]any
	Choices []CompactChoice
}

// CompactChoice is one chip of a CompactAction.
type CompactChoice struct {
	ID       string
	Label    Text
	Body     map[string]any
	Selected bool
}

// IntPtr returns a pointer to n, for CompactState.Count.
func IntPtr(n int) *int { return &n }
