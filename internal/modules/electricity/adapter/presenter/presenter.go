// Package presenter maps electricity prices to JSON (#288).
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/electricity/usecase"
)

// HourDTO is one hour; start is RFC 3339 in the server's time zone.
type HourDTO struct {
	Start      string  `json:"start"`
	CentPerKWh float64 `json:"centPerKWh"`
	Cheapest   bool    `json:"cheapest"`
}

// PricesDTO is returned by GET /api/electricity/prices.
type PricesDTO struct {
	Zone      string    `json:"zone"`
	Hours     []HourDTO `json:"hours"`
	FetchedAt string    `json:"fetchedAt"`
	Stale     bool      `json:"stale"`
}

// Prices maps the use case answer.
func Prices(p usecase.Prices) PricesDTO {
	out := PricesDTO{Zone: string(p.Zone), Hours: make([]HourDTO, 0, len(p.Hours)), FetchedAt: p.FetchedAt.Format(time.RFC3339), Stale: p.Stale}
	for _, h := range p.Hours {
		out.Hours = append(out.Hours, HourDTO{Start: h.Start.Format(time.RFC3339), CentPerKWh: h.CentPerKWh, Cheapest: h.Cheapest})
	}
	return out
}
