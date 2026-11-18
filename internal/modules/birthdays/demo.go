package birthdays

import (
	"context"
	"time"

	"waylight/internal/shared/kernel"
)

// SeedDemo implements modules.DemoSeeder: birthdays spread over the next
// weeks, so the widget always shows who is next.
func (m *Module) SeedDemo(ctx context.Context) error {
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	existing, err := m.uc.All(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	now := time.Now()
	people := []struct {
		name string
		in   int // days from today
		born int // year of birth, 0 = unknown
	}{
		{"Anna Schröder", 9, 1991},
		{"Grandma Hilde", 2, 1946},
		{"Jonas Meyer", 17, 1988},
		{"Lea (neighbour)", 31, 0},
		{"Paul", 54, 2019},
		{"Dad", 76, 1962},
	}
	for _, p := range people {
		d := now.AddDate(0, 0, p.in)
		if _, err := m.uc.Add(ctx, p.name, int(d.Month()), d.Day(), p.born); err != nil {
			return err
		}
	}
	return nil
}
