package main

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/dashboard/usecase"
	"waylight/internal/modules/shopping"
	"waylight/internal/shared/kernel"
)

// householdSetup sets up the shared records of a household template such as
// Family (#871), so the dashboard imports no other module:
//
//   - shopping: the list named after the tab with this visibility becomes
//     the applier's current list; it is created once and reused by every
//     member who applies the template, so they all tick off the same list;
//   - board: new notes of the widget default to this visibility.
func householdSetup(registry *modules.Registry) usecase.SharedSetup {
	var shop *shopping.Module
	for _, m := range registry.All() {
		if s, ok := m.(*shopping.Module); ok {
			shop = s
		}
	}
	return func(ctx context.Context, kind, tabName string, v kernel.Visibility) (map[string]any, error) {
		switch kind {
		case "shopping":
			if shop == nil {
				return nil, nil
			}
			return nil, shop.UseSharedList(ctx, tabName, v)
		case "board":
			return map[string]any{"visibility": string(v)}, nil
		}
		return nil, nil
	}
}
