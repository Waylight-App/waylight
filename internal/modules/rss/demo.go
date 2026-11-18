package rss

import "context"

// SeedDemo implements modules.DemoSeeder: the public demo loads its example
// feeds right away, so the news widget has items after the start and after
// every reset (the sample data only starts this fetch in the background).
// The demo answers the feed requests itself; failures are stored on the
// feeds and retried by the refresh job.
func (m *Module) SeedDemo(ctx context.Context) error {
	_ = m.uc.RefreshAll(ctx)
	return nil
}
