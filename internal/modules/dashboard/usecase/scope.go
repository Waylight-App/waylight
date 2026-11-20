package usecase

import (
	"context"
	"fmt"
	"slices"

	"waylight/internal/shared/kernel"
)

// TabScope answers which widget kinds are placed on a set of tabs; the
// auth module uses it to decide what a guest may read (input port).
type TabScope interface {
	WidgetKindsOnTabs(ctx context.Context, tabIDs []string) ([]string, error)
	// TabIDs lists every tab in display order (the display of an appliance
	// shows them all, #140).
	TabIDs(ctx context.Context) ([]string, error)
	// FirstWidgetConfig returns the settings (JSON) of the first widget of
	// kind on the account's tabs, in tab order and then top to bottom.
	FirstWidgetConfig(ctx context.Context, kind string) (string, bool, error)
}

type tabScope struct{ deps Deps }

// NewTabScope returns the TabScope use case.
func NewTabScope(deps Deps) TabScope { return &tabScope{deps: deps} }

// WidgetKindsOnTabs implements TabScope; the result is sorted and unique.
func (uc *tabScope) WidgetKindsOnTabs(ctx context.Context, tabIDs []string) ([]string, error) {
	widgets, err := uc.deps.Widgets.ListWidgets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list widgets: %w", err)
	}
	var kinds []string
	for _, w := range widgets {
		if slices.Contains(tabIDs, w.TabID()) && !slices.Contains(kinds, w.Kind()) {
			kinds = append(kinds, w.Kind())
		}
	}
	slices.Sort(kinds)
	return kinds, nil
}

// TabIDs implements TabScope: the admin's tabs, which the display of an
// appliance shows (#140); members' personal tabs are not part of it.
func (uc *tabScope) TabIDs(ctx context.Context) ([]string, error) {
	tabs, err := uc.deps.Tabs.ListTabs(kernel.WithPrincipal(ctx, kernel.AdminPrincipal()))
	if err != nil {
		return nil, fmt.Errorf("list tabs: %w", err)
	}
	ids := make([]string, 0, len(tabs))
	for _, t := range tabs {
		ids = append(ids, t.ID())
	}
	return ids, nil
}

// FirstWidgetConfig implements TabScope.
func (uc *tabScope) FirstWidgetConfig(ctx context.Context, kind string) (string, bool, error) {
	tabs, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return "", false, fmt.Errorf("list tabs: %w", err)
	}
	widgets, err := uc.deps.Widgets.ListWidgets(ctx)
	if err != nil {
		return "", false, fmt.Errorf("list widgets: %w", err)
	}
	for _, t := range tabs {
		for _, w := range widgets {
			if w.TabID() == t.ID() && w.Kind() == kind {
				return w.Config(), true, nil
			}
		}
	}
	return "", false, nil
}
