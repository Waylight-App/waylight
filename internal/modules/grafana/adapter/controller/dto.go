package controller

// SettingsInputDTO is the body of PUT /api/grafana/settings; an empty token
// keeps the stored one.
type SettingsInputDTO struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}
