package answers

import "time"

// Severity of a metric; set by the source, clients only show it.
type Severity string

// The severities.
const (
	OK       Severity = "ok"
	Warning  Severity = "warning"
	Critical Severity = "critical"
)

// State of a status or a list row.
type State string

// The states.
const (
	StateOK      State = "ok"
	StateWarning State = "warning"
	StateError   State = "error"
	StateOff     State = "off"
)

// Item is one row of a list answer.
type Item struct {
	Label   string
	Value   string
	State   State
	Percent *float64
}

// Answer is one answer of the smart search, already in the client's
// language. Which fields are set depends on Kind.
type Answer struct {
	Kind   Kind
	Intent string
	Title  string
	// WidgetKind is the widget with the details ("open widget").
	WidgetKind string
	// At is when the data was measured; zero when unknown.
	At time.Time

	// Metric
	Value    string
	Detail   string
	Percent  *float64
	Severity Severity

	// Status (with Text and Detail)
	State State

	// Text and Status
	Text string
	// Suggestions are names to pick when the search asks back.
	Suggestions []string

	// List
	Items []Item
	Empty string
}

// Percent returns a pointer for the optional percentage fields.
func Percent(p float64) *float64 { return &p }
