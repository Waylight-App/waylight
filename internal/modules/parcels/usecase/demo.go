package usecase

import (
	"context"
	"fmt"
	"time"

	"waylight/internal/modules/parcels/domain"
)

// SeedDemo adds parcels in different states to the public demo. They use
// carriers without a tracking API or are already delivered, so the regular
// refresh never touches them.
func SeedDemo(ctx context.Context, d Deps) error {
	existing, err := d.Repo.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	now := d.Clock.Now()
	parcels := []struct {
		number, label string
		carrier       domain.Carrier
		status        domain.Status
		eta           time.Time
	}{
		{"H1004827365019", "Running shoes", domain.CarrierHermes,
			domain.Status{Code: domain.StatusTransit, Description: "The parcel is on its way to the delivery depot", Location: "Bremen", At: now.Add(-3 * time.Hour)},
			now.AddDate(0, 0, 1)},
		{"GL482019374652", "Raspberry Pi 5", domain.CarrierGLS,
			domain.Status{Code: domain.StatusPreTransit, Description: "Data received, parcel not yet handed over", At: now.Add(-20 * time.Hour)},
			now.AddDate(0, 0, 3)},
		{"01906234517788", "Birthday present for Anna", domain.CarrierDPD,
			domain.Status{Code: domain.StatusTransit, Description: "Out for delivery", Location: "Oldenburg", At: now.Add(-40 * time.Minute)},
			now},
		{"00340434161094042557", "Coffee beans", domain.CarrierDHL,
			domain.Status{Code: domain.StatusDelivered, Description: "Delivered to the neighbour", Location: "Oldenburg", At: now.Add(-26 * time.Hour)},
			time.Time{}},
	}
	for i, dp := range parcels {
		p, err := domain.NewParcel(d.IDs.NewID(), dp.number, dp.carrier, dp.label, now.Add(-time.Duration(i+1)*30*time.Hour))
		if err != nil {
			return err
		}
		p.Status, p.EstimatedDelivery, p.CheckedAt = dp.status, dp.eta, dp.status.At
		if err := d.Repo.Save(ctx, p); err != nil {
			return fmt.Errorf("save parcel: %w", err)
		}
	}
	return nil
}
