package modules

import "time"

// InfoResponse is the DTO returned by GET /api/modules.
type InfoResponse struct {
	Modules     []string `json:"modules"`
	WidgetKinds []string `json:"widgetKinds"`
	// DataKinds maps module names to who owns their data (ADR 0018), so the
	// interface can offer members only the settings they may change.
	DataKinds map[string]string `json:"dataKinds"`
	// ReadOnly names modules whose data stays readable but cannot be changed
	// right now, such as an ee/ feature whose licence key has expired (#307).
	ReadOnly []string `json:"readOnly"`
	// Compact lists the modules with a compact state (#634), so clients such
	// as the watch find them without a list of their own.
	Compact []CompactInfo `json:"compact"`
}

// CompactInfo names a module's compact endpoint and the query parameters
// it takes (empty for most modules; the smart home device needs "entity").
type CompactInfo struct {
	Module string   `json:"module"`
	Path   string   `json:"path"`
	Params []string `json:"params"`
}

// CompactText is a translatable text of a compact state (#597): clients
// look up Key in their message catalogue, fill in Params and fall back to
// the English Text when they do not know the key.
type CompactText struct {
	Key    string            `json:"key"`
	Text   string            `json:"text"`
	Params map[string]string `json:"params,omitempty"`
}

// CompactChoice is one chip of a compact action.
type CompactChoice struct {
	ID       string         `json:"id"`
	Label    CompactText    `json:"label"`
	Body     map[string]any `json:"body,omitempty"`
	Selected bool           `json:"selected"`
}

// CompactAction is the one primary action of a compact view: the request
// a client sends when it is used, or, with choices, one request per chip.
type CompactAction struct {
	ID      string          `json:"id"`
	Label   CompactText     `json:"label"`
	Method  string          `json:"method"`
	Path    string          `json:"path"`
	Body    map[string]any  `json:"body,omitempty"`
	Choices []CompactChoice `json:"choices,omitempty"`
}

// CompactResponse is the DTO of GET /api/<module>/compact: the short state
// of the module's widget. Level is one of none, ok, attention, problem.
type CompactResponse struct {
	Title     CompactText    `json:"title"`
	Detail    *CompactText   `json:"detail,omitempty"`
	Level     string         `json:"level"`
	Count     *int           `json:"count,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Action    *CompactAction `json:"action,omitempty"`
	// Empty is true when the widget has nothing to show (#619).
	Empty bool `json:"empty,omitempty"`
}

// HelpResponse is the DTO of GET /api/help/widgets (#678).
type HelpResponse struct {
	Widgets []WidgetHelpInfo `json:"widgets"`
}

// WidgetHelpInfo is the help of one widget kind; clients translate the
// texts like those of a compact state.
type WidgetHelpInfo struct {
	Kind    string      `json:"kind"`
	Shows   CompactText `json:"shows"`
	SetUp   CompactText `json:"setUp"`
	DocsURL string      `json:"docsUrl"`
}
