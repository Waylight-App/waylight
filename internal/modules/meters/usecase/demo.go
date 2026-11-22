package usecase

import (
	"context"
	"time"

	"waylight/internal/modules/meters/domain"
)

// SeedDemo adds two meters with a year of monthly readings to the public
// demo, once. The readings end on the first of the current month, so the
// chart always shows the last twelve months.
func SeedDemo(ctx context.Context, d Deps) error {
	c := NewMeters(d)
	existing, err := c.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	now := d.Clock.Now()
	last := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	for _, s := range []struct {
		in    domain.Input
		start float64
		// monthly is the consumption of each month, January first.
		monthly [12]float64
	}{
		{domain.Input{Name: "Electricity", Kind: "electricity", Unit: "kWh"}, 14200,
			[12]float64{310, 285, 270, 240, 225, 210, 205, 215, 235, 265, 295, 330}},
		{domain.Input{Name: "Water", Kind: "water", Unit: "m³"}, 482.4,
			[12]float64{11, 10, 11, 12, 13, 15, 16, 16, 13, 12, 11, 11}},
	} {
		v, err := c.Create(ctx, s.in, "")
		if err != nil {
			return err
		}
		// 13 readings reach back to the first of the month a year ago.
		value := s.start
		for i := -12; i <= 0; i++ {
			day := last.AddDate(0, i, 0)
			if i > -12 {
				value += s.monthly[(int(day.Month())+10)%12] // consumption of the month before
			}
			if _, err := c.AddReading(ctx, v.ID, domain.ReadingInput{Day: domain.Day(day), Value: value}); err != nil {
				return err
			}
		}
	}
	return nil
}
