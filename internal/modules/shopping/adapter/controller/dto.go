package controller

// SettingsInputDTO is the body of PUT /api/shopping/settings; an empty
// password keeps the stored one.
type SettingsInputDTO struct {
	ListURL  string `json:"listUrl"`
	Username string `json:"username"`
	Password string `json:"password"`
	// SignIn is a finished Nextcloud sign-in of the setup assistant
	// (#654); it replaces username and password.
	SignIn string `json:"signIn,omitempty"`
}

// NewItemDTO is the body of POST /api/shopping/items.
type NewItemDTO struct {
	Summary string `json:"summary"`
}

// ItemStateDTO is the body of PUT /api/shopping/items/{id}.
type ItemStateDTO struct {
	Done bool `json:"done"`
}

// ListRequestDTO creates or changes a list on Waylight (#471).
type ListRequestDTO struct {
	Name string `json:"name"`
	// Visibility is "private", "everyone" or "group:<id>"; empty on create
	// uses the default for new records.
	Visibility string `json:"visibility,omitempty"`
}

// CurrentRequestDTO chooses what the widget shows.
type CurrentRequestDTO struct {
	// List is a list ID or "caldav".
	List string `json:"list"`
}
