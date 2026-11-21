package usecase

import (
	"context"
	"time"

	"waylight/internal/modules/contracts/domain"
)

// SeedDemo adds typical household contracts to the public demo, once. Term
// ends are relative to today, so one deadline is always close.
func SeedDemo(ctx context.Context, d Deps) error {
	c := NewContracts(d)
	existing, err := c.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	day := func(n int) string { return d.Clock.Now().AddDate(0, 0, n).Format(time.DateOnly) }
	for _, in := range []domain.Input{
		{Name: "Mobile phone", Provider: "Example Mobile", CostCents: 1999, Period: "monthly", TermEnd: day(40), NoticeDays: 30, RenewMonths: 1, RemindDays: 14},
		{Name: "Electricity", Provider: "Example Energy", CostCents: 8900, Period: "monthly", TermEnd: day(150), NoticeDays: 42, RenewMonths: 12, RemindDays: 21},
		{Name: "Gym", Provider: "City Fitness", CostCents: 3490, Period: "monthly", TermEnd: day(95), NoticeDays: 90, RenewMonths: 12, RemindDays: 14, Note: "Cancel in writing"},
		{Name: "Household insurance", Provider: "Example Insurance", CostCents: 12480, Period: "yearly", TermEnd: day(220), NoticeDays: 90, RenewMonths: 12, RemindDays: 30},
		{Name: "Video streaming", CostCents: 1399, Period: "monthly", TermEnd: day(12), NoticeDays: 0, RenewMonths: 1, RemindDays: 3},
	} {
		if _, err := c.Create(ctx, in, ""); err != nil {
			return err
		}
	}
	return nil
}
