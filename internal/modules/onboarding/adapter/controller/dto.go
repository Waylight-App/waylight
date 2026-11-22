package controller

// UpdateDTO is the body of PUT /api/onboarding: true completes the tour;
// false restarts it, from the given wizard step when one is sent and from
// the start of the wizard otherwise.
type UpdateDTO struct {
	Completed bool   `json:"completed"`
	Step      string `json:"step,omitempty"`
}
