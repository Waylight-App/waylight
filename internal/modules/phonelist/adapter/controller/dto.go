package controller

// EntryInputDTO is the body of POST /api/phonelist/entries and PUT /api/phonelist/entries/{id}.
type EntryInputDTO struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Team  string `json:"team"`
}

// SettingsInputDTO is the body of PUT /api/phonelist/settings; an empty
// password keeps the stored one, an empty URL switches the address book off.
type SettingsInputDTO struct {
	CardDAVURL string `json:"cardDavUrl"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}
