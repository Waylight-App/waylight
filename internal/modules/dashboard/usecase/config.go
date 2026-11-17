package usecase

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"waylight/internal/modules/dashboard/domain"
	"waylight/internal/shared/kernel"
)

// ConfigPhone places a widget in the phone layout of its tab (#92).
type ConfigPhone struct {
	Order  int
	Rows   int
	Hidden bool
}

// ConfigWidget is a widget in waylight.yaml (#88). Config is the opaque
// JSON document owned by the widget's module.
type ConfigWidget struct {
	Kind   string
	Rect   domain.Rect
	Phone  *ConfigPhone
	Config string
}

// ConfigTab is a tab with its widgets in reading order.
type ConfigTab struct {
	Name       string
	Background string
	Rules      []domain.RuleInput
	Widgets    []ConfigWidget
}

// ConfigRepository adds the atomic replacement of all tabs and widgets.
type ConfigRepository interface {
	TabRepository
	WidgetRepository
	// ReplaceAll makes tabs the complete set of tabs and widgets the
	// complete set of widgets; tabs whose IDs are missing are deleted.
	ReplaceAll(ctx context.Context, tabs []domain.Tab, widgets []domain.Widget) error
}

// Config exports and applies the dashboard section (input port).
type Config interface {
	Export(ctx context.Context) ([]ConfigTab, error)
	Apply(ctx context.Context, tabs []ConfigTab) error
}

type config struct {
	repo  ConfigRepository
	kinds KindCatalog
	ids   kernel.IDGenerator
}

// NewConfig returns the Config use case.
func NewConfig(repo ConfigRepository, kinds KindCatalog, ids kernel.IDGenerator) Config {
	return &config{repo: repo, kinds: kinds, ids: ids}
}

// readingOrder sorts widgets top to bottom, then left to right, so the
// file does not depend on the order in which they were created.
func readingOrder(ws []domain.Widget) {
	slices.SortStableFunc(ws, func(a, b domain.Widget) int {
		return cmp.Or(cmp.Compare(a.Rect().Y, b.Rect().Y), cmp.Compare(a.Rect().X, b.Rect().X))
	})
}

// Export and Apply work on the admin's tabs: config as code describes the
// instance, and members' personal tabs are not part of it (ADR 0018). They
// also run at start without a request, as the system, which would see
// every account's tabs.
func (uc *config) Export(ctx context.Context) ([]ConfigTab, error) {
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	tabs, err := uc.repo.ListTabs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tabs: %w", err)
	}
	out := make([]ConfigTab, 0, len(tabs))
	for _, t := range tabs {
		ws, err := uc.repo.ListWidgetsByTab(ctx, t.ID())
		if err != nil {
			return nil, fmt.Errorf("list widgets: %w", err)
		}
		readingOrder(ws)
		phone := map[string]*ConfigPhone{}
		if t.Mobile() != nil {
			for i, e := range domain.ResolveMobileLayout(t.Mobile(), ws).Entries() {
				phone[e.WidgetID] = &ConfigPhone{Order: i + 1, Rows: e.Rows, Hidden: e.Hidden}
			}
		}
		ct := ConfigTab{Name: t.Name(), Background: t.Background()}
		for _, r := range t.Rules() {
			from, to := r.Range()
			ct.Rules = append(ct.Rules, domain.RuleInput{Days: r.Days(), From: from, To: to, Networks: r.Networks()})
		}
		for _, w := range ws {
			ct.Widgets = append(ct.Widgets, ConfigWidget{Kind: w.Kind(), Rect: w.Rect(), Phone: phone[w.ID()], Config: w.Config()})
		}
		out = append(out, ct)
	}
	return out, nil
}

var errDuplicateTab = kernel.NewValidationError("tabs", "Each tab name can be used only once.")

// Apply keeps the IDs of tabs with the same name, so guest access to a tab
// survives a re-import; widgets always get new IDs. An empty list is fine:
// the dashboard creates its "Home" tab on the next visit, as after setup.
func (uc *config) Apply(ctx context.Context, in []ConfigTab) error {
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	old, err := uc.repo.ListTabs(ctx)
	if err != nil {
		return fmt.Errorf("list tabs: %w", err)
	}
	ids := map[string]string{}
	for _, t := range old {
		ids[strings.ToLower(t.Name())] = t.ID()
	}
	seen := map[string]bool{}
	var tabs []domain.Tab
	var widgets []domain.Widget
	for pos, ct := range in {
		key := strings.ToLower(strings.TrimSpace(ct.Name))
		if seen[key] {
			return errDuplicateTab
		}
		seen[key] = true
		id, ok := ids[key]
		if !ok {
			id = uc.ids.NewID()
		}
		tab, tabWidgets, err := uc.buildTab(id, pos, ct)
		if err != nil {
			return fmt.Errorf("tab %q: %w", ct.Name, err)
		}
		tabs = append(tabs, tab)
		widgets = append(widgets, tabWidgets...)
	}
	return uc.repo.ReplaceAll(ctx, tabs, widgets)
}

func (uc *config) buildTab(id string, pos int, ct ConfigTab) (domain.Tab, []domain.Widget, error) {
	tab, err := domain.NewTab(id, ct.Name, pos)
	if err != nil {
		return domain.Tab{}, nil, err
	}
	if err := tab.SetBackground(ct.Background); err != nil {
		return domain.Tab{}, nil, err
	}
	rules, err := domain.NewTabRules(ct.Rules)
	if err != nil {
		return domain.Tab{}, nil, err
	}
	tab = tab.WithRules(rules)
	type placed struct {
		order int
		entry domain.MobileEntry
	}
	var phone []placed
	widgets := make([]domain.Widget, 0, len(ct.Widgets))
	for _, cw := range ct.Widgets {
		// A widget of a disabled module would silently vanish from the
		// page, so it is an error rather than something to skip.
		if !uc.kinds.Available(cw.Kind) {
			return domain.Tab{}, nil, fmt.Errorf("%w: %s", domain.ErrUnknownWidgetKind, cw.Kind)
		}
		rect, err := domain.NewRect(cw.Rect.X, cw.Rect.Y, cw.Rect.W, cw.Rect.H)
		if err != nil {
			return domain.Tab{}, nil, err
		}
		w, err := domain.NewWidget(uc.ids.NewID(), id, cw.Kind, rect, cw.Config)
		if err != nil {
			return domain.Tab{}, nil, err
		}
		widgets = append(widgets, w)
		if cw.Phone != nil {
			phone = append(phone, placed{cw.Phone.Order, domain.MobileEntry{WidgetID: w.ID(), Rows: cw.Phone.Rows, Hidden: cw.Phone.Hidden}})
		}
	}
	// Without any phone entry the tab keeps the layout derived from the grid.
	if len(phone) > 0 {
		slices.SortStableFunc(phone, func(a, b placed) int { return cmp.Compare(a.order, b.order) })
		entries := make([]domain.MobileEntry, 0, len(phone))
		for _, p := range phone {
			entries = append(entries, p.entry)
		}
		m, err := domain.NewMobileLayout(entries, widgets)
		if err != nil {
			return domain.Tab{}, nil, err
		}
		tab = tab.WithMobile(m)
	}
	return tab, widgets, nil
}
