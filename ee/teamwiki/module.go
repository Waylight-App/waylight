// Package teamwiki adds Confluence, SharePoint/OneDrive and Notion
// (#327–#329) and messages from Slack and Microsoft Teams (#330, #331) to
// the wiki search, each searched with the signed-in person's own account
// (part of workplace search in the ee/ features).
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module; it contains no business logic and stores nothing.
package teamwiki

import (
	"net/http"
	"time"

	"waylight/ee/licence"
	"waylight/ee/teamwiki/adapter/gateway"
	"waylight/ee/teamwiki/usecase"
	"waylight/internal/modules"
	"waylight/internal/modules/wiki"
)

// Module wires the sources into the wiki search.
type Module struct{ modules.Base }

// APIs are the service endpoints; tests replace them.
type APIs struct{ Atlassian, Graph, Notion, Slack string }

// slackPerMinute stays below Slack's limit of about ten searches per person
// and minute for the Real-time Search API.
const slackPerMinute = 8

// DefaultAPIs are the services' public endpoints.
var DefaultAPIs = APIs{Atlassian: "https://api.atlassian.com", Graph: "https://graph.microsoft.com/v1.0", Notion: "https://api.notion.com/v1",
	Slack: "https://slack.com/api"}

// New adds the sources to the wiki search's member sources.
func New(sources *wiki.MemberSources, tokens usecase.Tokens, licensing licence.Licensing, client *http.Client, apis APIs) *Module {
	sources.Add(usecase.NewSource("Confluence", "atlassian", tokens, licensing, gateway.Confluence{Client: client, API: apis.Atlassian}))
	sources.Add(usecase.NewSource("SharePoint and OneDrive", "microsoft", tokens, licensing, gateway.Graph{Client: client, API: apis.Graph}))
	sources.Add(usecase.NewSource("Notion", "notion", tokens, licensing, gateway.Notion{Client: client, API: apis.Notion}))
	sources.Add(usecase.NewSource("Slack", "slack", tokens, licensing, gateway.Slack{Client: client, API: apis.Slack}).
		WithLimit(slackPerMinute, time.Now))
	sources.Add(usecase.NewSource("Teams", "microsoft", tokens, licensing, gateway.Teams{Client: client, API: apis.Graph}))
	return &Module{}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "teamwiki" }

// DataKind implements modules.Module: nothing is stored.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }
