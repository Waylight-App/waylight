package controller

// SetupSignInInputDTO is the body of POST /api/setup/signin: the service
// ("nextcloud" or "homeassistant") and the address entered in the
// assistant.
type SetupSignInInputDTO struct {
	Service string `json:"service"`
	URL     string `json:"url"`
}

// SetupSignInDTO is a started sign-in: its ID and the page the user signs
// in on. It carries no credential.
type SetupSignInDTO struct {
	ID       string `json:"id"`
	LoginURL string `json:"loginUrl"`
}

// SetupSignInStatusDTO answers GET /api/setup/signin/{id}: "pending" or
// "ready". A failed sign-in answers with an error instead.
type SetupSignInStatusDTO struct {
	Status string `json:"status"`
}

// SetupPasteDTO is a started code for Paste from phone: what the QR code
// carries and when it ends.
type SetupPasteDTO struct {
	Code string `json:"code"`
	// ExpiresAt is an RFC 3339 time.
	ExpiresAt string `json:"expiresAt"`
}

// SetupPasteInputDTO is the body of PUT /api/setup/paste/{code}: the key
// copied on the phone.
type SetupPasteInputDTO struct {
	Value string `json:"value"`
}

// SetupPasteStatusDTO answers GET /api/setup/paste/{code}. The value is
// set once, in the answer that hands it over.
type SetupPasteStatusDTO struct {
	Status string `json:"status"`
	Value  string `json:"value,omitempty"`
}
