package controller

// MeterInputDTO is the body of POST /api/meters and PUT /api/meters/{id}.
type MeterInputDTO struct {
	Name string `json:"name"`
	// Kind is "electricity", "gas", "water", "heat" or "other".
	Kind string `json:"kind"`
	Unit string `json:"unit"`
	// Visibility is "private", "everyone" or "group:<id>". Empty shares a new
	// meter with everyone and keeps the visibility of an existing one.
	Visibility string `json:"visibility,omitempty"`
}

// ReadingInputDTO is the body of POST /api/meters/{id}/readings.
type ReadingInputDTO struct {
	// Day is YYYY-MM-DD.
	Day   string  `json:"day"`
	Value float64 `json:"value"`
}
