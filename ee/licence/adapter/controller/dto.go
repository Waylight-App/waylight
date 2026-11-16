package controller

// LicenceDTO is the licence as GET /api/licence returns it to admins (#307).
type LicenceDTO struct {
	// Status is missing, invalid, valid, expired or over-seats.
	Status string `json:"status"`
	// Source is "environment" (LICENCE_KEY, read-only here), "settings" or "".
	Source string `json:"source"`
	// Problem explains an invalid key.
	Problem  string `json:"problem,omitempty"`
	Licensee string `json:"licensee,omitempty"`
	// Issued and Until are days (YYYY-MM-DD); the key is valid through Until.
	Issued   string   `json:"issued,omitempty"`
	Until    string   `json:"until,omitempty"`
	Seats    int      `json:"seats"`
	Accounts int      `json:"accounts"`
	Features []string `json:"features"`
	// Unknown lists features of the key this version cannot unlock.
	Unknown []string `json:"unknown"`
}

// FeaturesDTO is the answer of GET /api/licence/features for every signed-in
// account: "full", "read" or "none" per known feature.
type FeaturesDTO struct {
	Features map[string]string `json:"features"`
}

// InstallDTO is the body of PUT /api/licence.
type InstallDTO struct {
	Key string `json:"key"`
}
