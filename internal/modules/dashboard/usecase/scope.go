package usecase

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"waylight/internal/modules/dashboard/domain"
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
	// OwnWidgets lists the widgets on the tabs the account owns, in tab
	// order and then top to bottom; tabs others share with it are left out
	// because only their owner can change them (#599).
	OwnWidgets(ctx context.Context) ([]OwnWidget, error)
	// InstanceWidgets lists the widgets on every tab of the instance, of
	// every account, for instance-wide lookups such as which widgets use a
	// connection (#610). Mine marks the tabs of the account of ctx.
	InstanceWidgets(ctx context.Context) ([]InstanceWidget, error)
}

// OwnWidget is a widget as OwnWidgets returns it.
type OwnWidget struct {
	ID, TabID, Kind, Config string
}

// InstanceWidget is a widget as InstanceWidgets returns it.
type InstanceWidget struct {
	OwnWidget
	Mine bool
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

// OwnWidgets implements TabScope.
func (uc *tabScope) OwnWidgets(ctx context.Context) ([]OwnWidget, error) {
	p := kernel.PrincipalFrom(ctx)
	tabs, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tabs: %w", err)
	}
	widgets, err := uc.deps.Widgets.ListWidgets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list widgets: %w", err)
	}
	slices.SortStableFunc(widgets, func(a, b domain.Widget) int {
		return cmp.Or(cmp.Compare(a.Rect().Y, b.Rect().Y), cmp.Compare(a.Rect().X, b.Rect().X))
	})
	var out []OwnWidget
	for _, t := range tabs {
		if !t.Access().Mine(p) {
			continue
		}
		for _, w := range widgets {
			if w.TabID() == t.ID() {
				out = append(out, OwnWidget{ID: w.ID(), TabID: w.TabID(), Kind: w.Kind(), Config: w.Config()})
			}
		}
	}
	return out, nil
}

// InstanceWidgets implements TabScope. It reads as the system, which sees
// every account's tabs; callers only pass on IDs, kinds and settings
// fields they need, never tab titles of other accounts.
func (uc *tabScope) InstanceWidgets(ctx context.Context) ([]InstanceWidget, error) {
	p := kernel.PrincipalFrom(ctx)
	sys := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: kernel.SystemAccountID, Role: "system"})
	tabs, err := uc.deps.Tabs.ListTabs(sys)
	if err != nil {
		return nil, fmt.Errorf("list tabs: %w", err)
	}
	widgets, err := uc.deps.Widgets.ListWidgets(sys)
	if err != nil {
		return nil, fmt.Errorf("list widgets: %w", err)
	}
	mine := map[string]bool{}
	for _, t := range tabs {
		mine[t.ID()] = t.Access().Mine(p)
	}
	out := make([]InstanceWidget, 0, len(widgets))
	for _, w := range widgets {
		out = append(out, InstanceWidget{OwnWidget: OwnWidget{ID: w.ID(), TabID: w.TabID(), Kind: w.Kind(), Config: w.Config()}, Mine: mine[w.TabID()]})
	}
	return out, nil
}
