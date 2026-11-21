package controller

// ConnectionInputDTO is the body of POST /api/connections.
type ConnectionInputDTO struct {
	Service    string            `json:"service"`
	Name       string            `json:"name"`
	URL        string            `json:"url"`
	Credential map[string]string `json:"credential"`
}

// ConnectionChangeDTO is the body of PUT /api/connections/{id}; a missing
// credential keeps the stored one while the address stays the same.
type ConnectionChangeDTO struct {
	Name       string            `json:"name"`
	URL        string            `json:"url"`
	Credential map[string]string `json:"credential,omitempty"`
}

// CredentialDTO is the body of PUT /api/connections/{id}/credential.
type CredentialDTO struct {
	Credential map[string]string `json:"credential"`
}
