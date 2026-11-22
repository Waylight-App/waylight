// Package presenter maps meters to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"math"

	"waylight/internal/modules/meters/domain"
	"waylight/internal/modules/meters/usecase"
)

// ReadingDTO is one reading.
type ReadingDTO struct {
	// Day is YYYY-MM-DD.
	Day   string  `json:"day"`
	Value float64 `json:"value"`
}

// MeterDTO is one meter in API responses.
type MeterDTO struct {
	ID string `json:"id"`
	// Name is the meter's own name, e.g. "Electricity, hallway".
	Name string `json:"name"`
	// Kind is "electricity", "gas", "water", "heat" or "other".
	Kind string `json:"kind"`
	// Unit is what the counter counts, e.g. "kWh" or "m³".
	Unit string `json:"unit"`
	// Last is the newest reading; null while there is none.
	Last *ReadingDTO `json:"last"`
	// Visibility is "private", "everyone" or "group:<id>" (ADR 0018).
	Visibility string `json:"visibility"`
	// Mine is true when the viewer created the meter and may change it.
	Mine bool `json:"mine"`
}

// ConsumptionDTO is the consumption of one year.
type ConsumptionDTO struct {
	Year int `json:"year"`
	// Months has 12 entries, January first; null where the readings do not
	// cover the month.
	Months []*float64 `json:"months"`
	// Total is the consumption between the first and the last reading of the
	// year; null when the readings do not reach into the year.
	Total *float64 `json:"total"`
	// Complete is true when Total covers the whole year.
	Complete bool `json:"complete"`
}

// DetailDTO is a meter with its readings and the consumption of one year.
type DetailDTO struct {
	Meter MeterDTO `json:"meter"`
	// Readings are newest first.
	Readings    []ReadingDTO   `json:"readings"`
	Consumption ConsumptionDTO `json:"consumption"`
	// Years are the calendar years with readings, the latest first.
	Years []int `json:"years"`
}

// amount rounds to three decimals, which is finer than any meter shows.
func amount(a domain.Amount) *float64 {
	if !a.Known {
		return nil
	}
	v := math.Round(a.Value*1000) / 1000
	return &v
}

// Reading maps one reading.
func Reading(r domain.Reading) ReadingDTO {
	return ReadingDTO{Day: domain.Day(r.Day), Value: r.Value}
}

// Meter maps one meter.
func Meter(v usecase.MeterView) MeterDTO {
	dto := MeterDTO{ID: v.ID, Name: v.Name, Kind: string(v.Kind), Unit: v.Unit, Visibility: string(v.Access.Visibility), Mine: v.Mine}
	if v.Last != nil {
		last := Reading(*v.Last)
		dto.Last = &last
	}
	return dto
}

// Meters maps a list of meters.
func Meters(list []usecase.MeterView) []MeterDTO {
	out := make([]MeterDTO, 0, len(list))
	for _, v := range list {
		out = append(out, Meter(v))
	}
	return out
}

// Detail maps a meter with its readings and consumption.
func Detail(d usecase.Detail) DetailDTO {
	c := ConsumptionDTO{Year: d.Consumption.Year, Months: make([]*float64, 0, len(d.Consumption.Months)),
		Total: amount(d.Consumption.Total), Complete: d.Consumption.Complete}
	for _, m := range d.Consumption.Months {
		c.Months = append(c.Months, amount(m))
	}
	out := DetailDTO{Meter: Meter(d.MeterView), Readings: make([]ReadingDTO, 0, len(d.Readings)), Consumption: c, Years: d.Years}
	if out.Years == nil {
		out.Years = []int{}
	}
	for _, r := range d.Readings {
		out.Readings = append(out.Readings, Reading(r))
	}
	return out
}
