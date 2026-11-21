package controller

// NoteInputDTO is the body of POST /api/board and PUT /api/board/{id}.
type NoteInputDTO struct {
	Text   string `json:"text"`
	Pinned bool   `json:"pinned"`
	// ExpiresOn is the last day the note is shown, YYYY-MM-DD, or empty
	// to keep it.
	ExpiresOn string `json:"expiresOn"`
	// Visibility is "private", "everyone" or "group:<id>". Empty shares a
	// new note with everyone and keeps the visibility of an existing one.
	Visibility string `json:"visibility,omitempty"`
}
