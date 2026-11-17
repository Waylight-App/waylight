package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"waylight/internal/modules/readlater/domain"
	"waylight/internal/shared/kernel"
)

// demoArticle is a saved article of the public demo. The reader copy is a
// short text written for the demo, so nothing is fetched from the sites.
type demoArticle struct {
	url, title, site, excerpt string
	paragraphs                []string
	words                     int
	savedAgo                  time.Duration
	read                      bool
}

// SeedDemo saves a few articles with reader copies to the public demo.
func SeedDemo(ctx context.Context, repo Repository, clock kernel.Clock, ids kernel.IDGenerator) error {
	existing, err := repo.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	now := clock.Now()
	articles := []demoArticle{
		{url: "https://example.org/homelab/backups-3-2-1", title: "The 3-2-1 backup rule, explained for a homelab",
			site: "Homelab Notes", excerpt: "Three copies, two media, one offsite: what that means for a small server at home.",
			paragraphs: []string{
				"Three copies of your data, on two different media, with one copy offsite. The rule is old, but it still answers the question most homelab owners ask too late: what happens when the disk dies?",
				"Start with what you cannot replace: photos, documents, the configuration of your services. Everything else can be downloaded again.",
			}, words: 1840, savedAgo: 3 * time.Hour},
		{url: "https://example.org/cooking/sourdough-starter", title: "A sourdough starter in five days",
			site: "Weekend Kitchen", excerpt: "Flour, water and patience: a schedule that fits around a working week.",
			paragraphs: []string{"Day one is the easiest: equal parts flour and water, a loose lid and a warm spot in the kitchen."},
			words:      960, savedAgo: 26 * time.Hour},
		{url: "https://example.org/travel/north-sea-islands", title: "Five North Sea islands for a long weekend",
			site: "Coastal Travel", excerpt: "From Norderney to Spiekeroog: which island suits which kind of trip.",
			paragraphs: []string{"Each of the East Frisian islands has its own pace. Norderney feels like a small town, Spiekeroog has no cars at all."},
			words:      1320, savedAgo: 3 * 24 * time.Hour, read: true},
		{url: "https://example.org/tech/self-hosting-checklist", title: "A checklist before you expose a service to the internet",
			site: "Self-Hosted Weekly", excerpt: "Reverse proxy, updates, backups and two-factor authentication: the boring parts that matter.",
			paragraphs: []string{"Before a service leaves your home network, put it behind a reverse proxy with TLS and turn on automatic updates."},
			words:      2210, savedAgo: 5 * 24 * time.Hour},
	}
	for _, da := range articles {
		a, err := domain.NewArticle(ids.NewID(), da.url, now.Add(-da.savedAgo))
		if err != nil {
			return err
		}
		var b strings.Builder
		for _, p := range da.paragraphs {
			b.WriteString("<p>" + p + "</p>")
		}
		a.ApplyPage(da.title, da.site, da.excerpt, b.String(), da.words)
		a.Bytes = int64(b.Len())
		if da.read {
			a.Status, a.StatusAt = domain.StatusRead, now.Add(-da.savedAgo/2)
		}
		if err := repo.Save(ctx, a, nil); err != nil {
			return fmt.Errorf("save article: %w", err)
		}
	}
	return nil
}
