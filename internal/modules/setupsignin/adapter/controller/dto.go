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
