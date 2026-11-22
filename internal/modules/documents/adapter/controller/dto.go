package controller

// DocumentInputDTO is the body of POST /api/documents and
// PUT /api/documents/{id}.
type DocumentInputDTO struct {
	Title string `json:"title"`
	// PurchasedOn and WarrantyEnd are YYYY-MM-DD or empty.
	PurchasedOn string `json:"purchasedOn"`
	WarrantyEnd string `json:"warrantyEnd"`
	Note        string `json:"note"`
	// Visibility is "private", "everyone" or "group:<id>". Empty uses the
	// default for new records and keeps the visibility of an existing one.
	Visibility string `json:"visibility,omitempty"`
}
