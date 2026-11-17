package shopping

import (
	"context"

	"waylight/internal/shared/kernel"
)

// SeedDemo implements modules.DemoSeeder: a local shopping list with a few
// items, two of them already in the basket.
func (m *Module) SeedDemo(ctx context.Context) error {
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	existing, err := m.listsUC.All(ctx)
	if err != nil || len(existing.Lists) > 0 {
		return err
	}
	list, err := m.listsUC.Create(ctx, "Weekly shop", "")
	if err != nil {
		return err
	}
	for i, summary := range []string{"Oat milk", "Bananas", "Coffee beans", "Sourdough bread", "Tomatoes", "Dish soap", "Cheese", "Basil"} {
		item, err := m.listsUC.Add(ctx, list.ID, summary)
		if err != nil {
			return err
		}
		if i == 1 || i == 5 {
			if err := m.listsUC.SetDone(ctx, list.ID, item.ID, true); err != nil {
				return err
			}
		}
	}
	return m.listsUC.SetCurrent(ctx, list.ID)
}
