package controller

// WikiInputDTO creates or changes a wiki; an empty token keeps the stored one.
type WikiInputDTO struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	URL   string `json:"url"`
	Token string `json:"token"`
}
