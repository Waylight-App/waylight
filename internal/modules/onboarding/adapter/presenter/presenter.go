// Package presenter maps the onboarding state to its JSON DTO.
//
// Ring: Interface Adapters.
package presenter

import "waylight/internal/modules/onboarding/usecase"

// StatusDTO is returned by the onboarding endpoints.
type StatusDTO struct {
	Completed bool `json:"completed"`
	// Step is the wizard step the account stopped at; empty before it started.
	Step string `json:"step"`
}

// Status maps a StatusResponse.
func Status(r usecase.StatusResponse) StatusDTO {
	return StatusDTO{Completed: r.Completed, Step: r.Step}
}
