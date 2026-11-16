//go:build ee

package main

import (
	"context"
	"os"
	"strings"

	"trisse-home/ee/slackcmd"
	"trisse-home/ee/slackcmd/domain"
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/bookmarks"
	"trisse-home/internal/modules/services"
	"trisse-home/internal/modules/wiki"
	"trisse-home/internal/shared/kernel"
)

// slackAPI points the command at a fake Slack in tests.
var slackAPI = "https://slack.com/api"

// startPageFinder searches services, bookmarks and wikis as one account for
// the Slack command (#333). Modules are looked up when a command arrives,
// so the build order does not matter; a switched-off module adds nothing.
type startPageFinder struct{ registry *modules.Registry }

func (f startPageFinder) Find(ctx context.Context, accountID, query string) ([]domain.Hit, error) {
	// The search runs with the member's rights: what they could find on the
	// start page, and nothing more.
	ctx = kernel.WithPrincipal(ctx, kernel.PrincipalFor(accountID))
	q := strings.ToLower(query)
	var hits []domain.Hit
	var firstErr error
	for _, m := range f.registry.All() {
		switch mod := m.(type) {
		case *services.Module:
			list, err := mod.Services(ctx)
			if err != nil {
				firstErr = err
				continue
			}
			for _, s := range list {
				if strings.Contains(strings.ToLower(s.Name+" "+s.Description), q) {
					hits = append(hits, domain.Hit{Source: "Service", Title: s.Name, URL: s.URL})
				}
			}
		case *bookmarks.Module:
			list, err := mod.Find(ctx, query, domain.MaxHits)
			if err != nil {
				firstErr = err
				continue
			}
			for _, b := range list {
				hits = append(hits, domain.Hit{Source: "Bookmark", Title: b.Title, URL: b.URL})
			}
		case *wiki.Module:
			pages, err := mod.Find(ctx, query)
			if err != nil {
				firstErr = err
				continue
			}
			for _, p := range pages {
				hits = append(hits, domain.Hit{Source: p.Wiki, Title: p.Title, URL: p.URL})
			}
		}
	}
	return hits, firstErr
}

// The Slack slash command (#333), part of the chat apps in the ee/
// features. SLACK_APP_TOKEN is the app-level token of the workspace's own
// app; without it nothing connects.
func init() {
	registerModule("slackcmd", func(d moduleDeps) modules.Module {
		settings := ""
		if d.cfg.PublicURL != nil {
			settings = strings.TrimSuffix(d.cfg.PublicURL.String(), "/") + "/?settings=plugin-connections"
		}
		return slackcmd.New(slackcmd.Deps{
			AppToken: strings.TrimSpace(os.Getenv("SLACK_APP_TOKEN")), Client: d.out.integration, API: slackAPI,
			Connections: connectionsModule(d), Finder: startPageFinder{d.registry}, Licensing: licenceModule(d).Licensing(),
			SettingsURL: settings,
		})
	})
}
