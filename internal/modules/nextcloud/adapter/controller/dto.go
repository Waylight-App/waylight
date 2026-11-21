package controller

// SettingsInputDTO is the body of PUT /api/nextcloud/settings; an empty app
// password keeps the stored one.
type SettingsInputDTO struct {
	URL         string `json:"url"`
	Username    string `json:"username"`
	AppPassword string `json:"appPassword"`
	// SignIn is a finished sign-in of the setup assistant (#654); it
	// replaces url, username and appPassword.
	SignIn string `json:"signIn,omitempty"`
}
