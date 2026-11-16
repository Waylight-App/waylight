//go:build ee

package main

import (
	"trisse-home/ee/teamwiki"
	"trisse-home/internal/modules"
)

// teamwikiAPIs points the services at fakes in tests.
var teamwikiAPIs = teamwiki.DefaultAPIs

// Wiki search for teams (#327–#329), part of the ee/ features: Confluence,
// SharePoint/OneDrive and Notion, searched with each person's connected
// account.
func init() {
	registerModule("teamwiki", func(d moduleDeps) modules.Module {
		return teamwiki.New(d.wikiSources, connectionsModule(d), licenceModule(d).Licensing(), d.out.integration, teamwikiAPIs)
	})
}
