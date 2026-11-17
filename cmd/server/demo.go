package main

import (
	"context"
	"fmt"
	"sync/atomic"

	"waylight/internal/platform/demo"
	"waylight/internal/platform/scheduler"
	"waylight/internal/shared/sqlutil"
)

// prepareDemo creates the demo account (which seeds the sample data), takes
// a snapshot of all exportable data and returns the job that restores it
// every DEMO_RESET_INTERVAL.
func prepareDemo(ctx context.Context, a *app) (scheduler.Task, error) {
	if err := a.auth.PrepareDemo(ctx); err != nil {
		return scheduler.Task{}, err
	}
	// Showcase data of the modules; the snapshot below includes it, so every
	// reset brings it back.
	if err := seedDemo(ctx, a); err != nil {
		return scheduler.Task{}, err
	}
	snapshot := map[string]map[string]sqlutil.Rows{}
	for name, p := range a.registry.Portables() {
		data, err := p.ExportData(ctx)
		if err != nil {
			return scheduler.Task{}, fmt.Errorf("snapshot %s: %w", name, err)
		}
		snapshot[name] = data
	}
	a.log.Warn("DEMO_MODE: public demo, visitors sign in with the password shown on the sign-in page; data is reset regularly", "reset_interval", a.cfg.DemoResetInterval)
	// The scheduler runs every task once at start; the data is fresh then,
	// and resetting would race with the first feed and weather refresh.
	var started atomic.Bool
	return scheduler.Task{Name: "demo.reset", Interval: a.cfg.DemoResetInterval, Run: func(ctx context.Context) error {
		if started.CompareAndSwap(false, true) {
			return nil
		}
		for name, p := range a.registry.Portables() {
			if err := p.ImportData(ctx, snapshot[name]); err != nil {
				return fmt.Errorf("reset %s: %w", name, err)
			}
		}
		// Modules without an export (drop zone, backup monitor) are not in the
		// snapshot; their seeders fill them again when a visitor emptied them
		// or the entries expired.
		if err := seedDemo(ctx, a); err != nil {
			return err
		}
		a.log.InfoContext(ctx, "demo data reset")
		return nil
	}}, nil
}

// seedDemo runs the demo seeders of all modules and sets up the example
// integrations. Each step only adds what is missing, so running it again is
// harmless.
func seedDemo(ctx context.Context, a *app) error {
	return demo.Seed(func() error {
		for name, s := range a.registry.DemoSeeders() {
			if err := s.SeedDemo(ctx); err != nil {
				return fmt.Errorf("demo data of %s: %w", name, err)
			}
		}
		return seedDemoIntegrations(ctx, newHandler(a))
	})
}
