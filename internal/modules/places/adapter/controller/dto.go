package controller

// PlaceInputDTO is the body of POST /api/places and PUT /api/places/{id}.
type PlaceInputDTO struct {
	Name  string  `json:"name"`
	Label string  `json:"label"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
}
