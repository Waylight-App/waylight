// Package presenter maps the watchlist to JSON (#284).
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/markets/usecase"
)

// PointDTO is one daily close.
type PointDTO struct {
	Day   string  `json:"day"`
	Value float64 `json:"value"`
}

// RowDTO is one watchlist entry; value is null when there is no quote.
type RowDTO struct {
	Symbol    string     `json:"symbol"`
	Value     *float64   `json:"value"`
	ChangePct float64    `json:"changePct"`
	AsOf      string     `json:"asOf"`
	History   []PointDTO `json:"history"`
	Stale     bool       `json:"stale"`
	// Error is "no_data" or "unavailable" when there is no quote.
	Error string `json:"error,omitempty"`
}

// Rows maps the use case answer.
func Rows(rows []usecase.Row) []RowDTO {
	out := make([]RowDTO, 0, len(rows))
	for _, r := range rows {
		dto := RowDTO{Symbol: r.Symbol.String(), Stale: r.Stale, Error: r.Error, History: []PointDTO{}}
		if q := r.Quote; q != nil {
			v := q.Value
			dto.Value, dto.ChangePct, dto.AsOf = &v, q.ChangePct, q.AsOf.Format(time.DateOnly)
			for _, p := range q.History {
				dto.History = append(dto.History, PointDTO{Day: p.Day.Format(time.DateOnly), Value: p.Value})
			}
		}
		out = append(out, dto)
	}
	return out
}
