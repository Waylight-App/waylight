package modules

// InfoResponse is the DTO returned by GET /api/modules.
type InfoResponse struct {
	Modules     []string `json:"modules"`
	WidgetKinds []string `json:"widgetKinds"`
	// DataKinds maps module names to who owns their data (ADR 0018), so the
	// interface can offer members only the settings they may change.
	DataKinds map[string]string `json:"dataKinds"`
}
