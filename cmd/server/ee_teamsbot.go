//go:build ee

package main

import (
	"context"
	"os"
	"strings"

	"trisse-home/ee/teamsbot"
	"trisse-home/ee/teamsbot/adapter/botauth"
	"trisse-home/ee/teamsbot/domain"
	"trisse-home/internal/modules"
)

// Endpoints of the Teams command box search; tests point them at fakes.
var (
	teamsBotMetadata = botauth.MetadataURL
	teamsGraphAPI    = "https://graph.microsoft.com/v1.0"
)

// teamsFinder reuses the start page search of the Slack command (#333).
type teamsFinder struct{ startPageFinder }

func (f teamsFinder) Find(ctx context.Context, accountID, query string) ([]domain.Hit, error) {
	hits, err := f.startPageFinder.Find(ctx, accountID, query)
	out := make([]domain.Hit, 0, len(hits))
	for _, h := range hits {
		out = append(out, domain.Hit{Source: h.Source, Title: h.Title, URL: h.URL})
	}
	return out, err
}

// The Teams command box search (#334), an opt-in of the chat apps in the
// ee/ features (ADR 0023): TEAMS_BOT_APP_ID is the organisation's own bot
// registration; without it the route does not exist.
func init() {
	registerModule("teamsbot", func(d moduleDeps) modules.Module {
		return teamsbot.New(teamsbot.Deps{
			AppID: strings.TrimSpace(os.Getenv("TEAMS_BOT_APP_ID")), Client: d.out.integration,
			MetadataURL: teamsBotMetadata, GraphAPI: teamsGraphAPI,
			Connections: connectionsModule(d), Finder: teamsFinder{startPageFinder{d.registry}},
			Licensing: licenceModule(d).Licensing(),
		})
	})
}
