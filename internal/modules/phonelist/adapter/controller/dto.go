package controller

// EntryInputDTO is the body of POST /api/phonelist/entries and PUT /api/phonelist/entries/{id}.
type EntryInputDTO struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Team  string `json:"team"`
}
