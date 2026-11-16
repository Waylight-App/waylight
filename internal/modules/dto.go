package modules

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
}
