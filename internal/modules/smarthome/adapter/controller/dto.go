package controller

// SettingsInputDTO is the body of PUT /api/smarthome/settings. An empty
// token keeps the stored one.
type SettingsInputDTO struct {
	URL   string `json:"url"`
	Token string `json:"token"`
	// SignIn is a finished sign-in of the setup assistant (#654); it
	// replaces url and token.
	SignIn string `json:"signIn,omitempty"`
}
