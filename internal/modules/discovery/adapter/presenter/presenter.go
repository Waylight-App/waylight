// Package presenter maps found services to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import "waylight/internal/modules/discovery/domain"

// FoundDTO is a service found in the home network.
type FoundDTO struct {
	URL string `json:"url"`
	// Hint says where it was found, e.g. "Docker container immich_server".
	Hint string `json:"hint"`
}

// Found maps the list; empty stays an empty array.
func Found(list []domain.Found) []FoundDTO {
	out := make([]FoundDTO, 0, len(list))
	for _, f := range list {
		out = append(out, FoundDTO{URL: f.URL, Hint: f.Hint})
	}
	return out
}
