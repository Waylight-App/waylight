package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// Limits of a meter and its readings.
const (
	MaxNameLength = 60
	MaxUnitLength = 12
	// MaxValue keeps a reading far from float precision limits.
	MaxValue = 1e9
	// MinYear bounds the years a reading can fall in.
	MinYear = 2000
)

// Errors.
var (
	// ErrNotFound is returned for unknown meters and for meters the viewer
	// may not see.
	ErrNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The meter does not exist.")
	// ErrNoReading is returned when a meter has no reading on a day.
	ErrNoReading = kernel.NewPublicError(kernel.ErrNotFound, "The reading does not exist.")
)

// Kind is what a meter measures.
type Kind string

// The kinds of meter.
const (
	Electricity Kind = "electricity"
	Gas         Kind = "gas"
	Water       Kind = "water"
	Heat        Kind = "heat"
	Other       Kind = "other"
)

func parseKind(s string) (Kind, bool) {
	switch k := Kind(s); k {
	case Electricity, Gas, Water, Heat, Other:
		return k, true
	}
	return "", false
}

// Meter is a household meter with its unit (#689).
type Meter struct {
	ID        string
	Name      string
	Kind      Kind
	Unit      string // kWh, m³ ...
	Access    kernel.Access
	CreatedAt time.Time
}

// Input is what a person enters for a meter.
type Input struct {
	Name string
	Kind string
	Unit string
}

// New validates the input and creates a meter.
func New(id string, in Input, access kernel.Access, now time.Time) (Meter, error) {
	m := Meter{ID: id, Access: access, CreatedAt: now}
	if err := m.apply(in); err != nil {
		return Meter{}, err
	}
	return m, nil
}

// Update changes everything but the owner.
func (m Meter) Update(in Input) (Meter, error) {
	err := m.apply(in)
	return m, err
}

func (m *Meter) apply(in Input) error {
	name := strings.TrimSpace(in.Name)
	if name == "" || utf8.RuneCountInString(name) > MaxNameLength {
		return kernel.NewValidationError("name", "The name must be between 1 and 60 characters long.")
	}
	kind, ok := parseKind(in.Kind)
	if !ok {
		return kernel.NewValidationError("kind", "Choose electricity, gas, water, heat or other.")
	}
	unit := strings.TrimSpace(in.Unit)
	if unit == "" || utf8.RuneCountInString(unit) > MaxUnitLength {
		return kernel.NewValidationError("unit", "The unit must be between 1 and 12 characters long.")
	}
	m.Name, m.Kind, m.Unit = name, kind, unit
	return nil
}

// Day formats a stored day as YYYY-MM-DD.
func Day(t time.Time) string { return t.Format(time.DateOnly) }

// ParseDay reads a YYYY-MM-DD day as midnight UTC.
func ParseDay(field, text string) (time.Time, error) {
	day, err := time.Parse(time.DateOnly, strings.TrimSpace(text))
	if err != nil {
		return time.Time{}, kernel.NewValidationError(field, "Enter a date like 2026-12-24.")
	}
	return day, nil
}
