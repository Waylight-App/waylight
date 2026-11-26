package usecase

import (
	"context"
	"fmt"

	"waylight/internal/modules/dashboard/domain"
	"waylight/internal/shared/kernel"
)

// SeedSampleData fills an empty dashboard with example tabs and widgets
// (input port).
type SeedSampleData interface {
	Execute(ctx context.Context) error
}

// SampleWidget is a widget of the sample layout.
type SampleWidget struct {
	Kind       string
	X, Y, W, H int
	Config     string
	// Fallback replaces the widget, without its config, when its kind is
	// unavailable (Tonight needs the weather module; the clock does not).
	Fallback string
}

// SampleTab is a tab of the sample layout.
type SampleTab struct {
	Name    string
	Widgets []SampleWidget
}

// EverydayLayout is the sample layout of the "everyday" start page (#281):
// only the Home tab.
func EverydayLayout(weatherConfig string) []SampleTab {
	return SampleLayout(weatherConfig)[:1]
}

// sampleRowHeight is the height, in grid rows, of the widgets of the sample
// layouts: with the card padding of the design guide it leaves room for a
// list of four items.
const sampleRowHeight = 6

// SampleLayout returns the layout shown after the first setup; weatherConfig
// is the JSON configuration of the Tonight card (location chosen by the
// operator). It has the composition of the design mockups, and every row adds
// up to the 12 columns (#793): Home opens with the Tonight card (5 columns),
// the to-dos (4) and the notes (3), then three widgets of four columns twice;
// Server puts the speed chart (8) next to the backups (4), then three lists
// of four columns.
func SampleLayout(weatherConfig string) []SampleTab {
	row := func(y int, widgets ...SampleWidget) []SampleWidget {
		x := 0
		for i := range widgets {
			widgets[i].X, widgets[i].Y, widgets[i].H = x, y, sampleRowHeight
			x += widgets[i].W
		}
		return widgets
	}
	var home, server []SampleWidget
	home = append(home, row(0,
		SampleWidget{Kind: "tonight", W: 5, Config: weatherConfig, Fallback: "clock"},
		SampleWidget{Kind: "todos", W: 4},
		SampleWidget{Kind: "notes", W: 3, Config: `{"scope":"pinned"}`},
	)...)
	home = append(home, row(sampleRowHeight,
		SampleWidget{Kind: "calendar", W: 4, Config: `{"days":7}`},
		SampleWidget{Kind: "bookmarks", W: 4},
		SampleWidget{Kind: "rss", W: 4, Config: `{"unreadOnly":true,"limit":15}`},
	)...)
	home = append(home, row(2*sampleRowHeight,
		SampleWidget{Kind: "habits", W: 4},
		SampleWidget{Kind: "birthdays", W: 4},
		SampleWidget{Kind: "focus", W: 4},
	)...)
	server = append(server, row(0,
		SampleWidget{Kind: "speedtest", W: 8},
		SampleWidget{Kind: "backup-status", W: 4},
	)...)
	server = append(server, row(sampleRowHeight,
		SampleWidget{Kind: "system", W: 4},
		SampleWidget{Kind: "docker", W: 4},
		SampleWidget{Kind: "diskhealth", W: 4},
	)...)
	return []SampleTab{
		{Name: "Home", Widgets: home},
		{Name: "Server", Widgets: server},
	}
}

type seedSampleData struct {
	deps Deps
	tabs []SampleTab
	// mark flags the tabs as example tabs, which the template choice on
	// first start may replace (#670).
	mark bool
}

// NewSeedSampleData returns the SeedSampleData use case for the given
// layout. Its tabs are marked as example tabs.
func NewSeedSampleData(deps Deps, tabs []SampleTab) SeedSampleData {
	return &seedSampleData{deps: deps, tabs: tabs, mark: true}
}

// NewSeedShowcase returns the SeedSampleData use case for the demo
// showcase, whose tabs are the demo itself and are never replaced.
func NewSeedShowcase(deps Deps, tabs []SampleTab) SeedSampleData {
	return &seedSampleData{deps: deps, tabs: tabs}
}

// Execute implements SeedSampleData. An existing dashboard is never touched,
// and widgets of disabled modules are skipped.
func (uc *seedSampleData) Execute(ctx context.Context) error {
	// Sample tabs are the admin's; seeding runs right after setup as the
	// system, which would otherwise create and count tabs of every account.
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	existing, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return fmt.Errorf("list tabs: %w", err)
	}
	if len(existing) > 0 {
		return nil
	}
	for pos, st := range uc.tabs {
		tab, err := domain.NewTab(uc.deps.IDs.NewID(), st.Name, pos)
		if err != nil {
			return err
		}
		if err := uc.deps.Tabs.CreateTab(ctx, tab.WithSample(uc.mark)); err != nil {
			return fmt.Errorf("create tab: %w", err)
		}
		for _, sw := range st.Widgets {
			if !uc.deps.Kinds.Available(sw.Kind) {
				if sw.Fallback == "" || !uc.deps.Kinds.Available(sw.Fallback) {
					continue
				}
				sw.Kind, sw.Config = sw.Fallback, ""
			}
			rect, err := domain.NewRect(sw.X, sw.Y, sw.W, sw.H)
			if err != nil {
				return err
			}
			w, err := domain.NewWidget(uc.deps.IDs.NewID(), tab.ID(), sw.Kind, rect, sw.Config)
			if err != nil {
				return err
			}
			if err := uc.deps.Widgets.CreateWidget(ctx, w); err != nil {
				return fmt.Errorf("create widget: %w", err)
			}
		}
	}
	return nil
}

// SampleTabs settles the example tabs once the template choice on first
// start is made (#670): replaced by a template or "start empty", they are
// removed; kept, they become ordinary tabs (input port).
type SampleTabs interface {
	// Discard removes the account's example tabs and closes the gaps in the
	// tab order. Without other tabs the next load creates the default tab.
	Discard(ctx context.Context) error
	// Keep turns the account's example tabs into ordinary tabs.
	Keep(ctx context.Context) error
}

type sampleTabs struct{ deps Deps }

// NewSampleTabs returns the SampleTabs use case.
func NewSampleTabs(deps Deps) SampleTabs { return &sampleTabs{deps: deps} }

// ownSample splits the tabs the principal may change into example tabs and
// the rest. Team tabs others shared are left alone either way.
func (uc *sampleTabs) ownSample(ctx context.Context) (sample, rest []domain.Tab, err error) {
	tabs, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list tabs: %w", err)
	}
	p := kernel.PrincipalFrom(ctx)
	for _, t := range tabs {
		mine := t.Access().Owner == "" || t.Access().Mine(p)
		if mine && t.Sample() {
			sample = append(sample, t)
		} else {
			rest = append(rest, t)
		}
	}
	return sample, rest, nil
}

// Discard implements SampleTabs.
func (uc *sampleTabs) Discard(ctx context.Context) error {
	sample, rest, err := uc.ownSample(ctx)
	if err != nil || len(sample) == 0 {
		return err
	}
	for _, t := range sample {
		if err := uc.deps.Tabs.DeleteTab(ctx, t.ID()); err != nil {
			return fmt.Errorf("delete tab: %w", err)
		}
	}
	for i := range rest {
		rest[i].MoveTo(i)
	}
	if err := uc.deps.Tabs.SaveTabOrder(ctx, rest); err != nil {
		return fmt.Errorf("save tab order: %w", err)
	}
	return nil
}

// Keep implements SampleTabs.
func (uc *sampleTabs) Keep(ctx context.Context) error {
	sample, _, err := uc.ownSample(ctx)
	if err != nil {
		return err
	}
	for _, t := range sample {
		if err := uc.deps.Tabs.UpdateTab(ctx, t.WithSample(false)); err != nil {
			return fmt.Errorf("update tab: %w", err)
		}
	}
	return nil
}
