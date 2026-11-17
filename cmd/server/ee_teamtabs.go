//go:build ee

package main

import (
	"context"

	"waylight/ee/teamtabs"
	"waylight/ee/teamtabs/usecase"
	"waylight/internal/modules"
	"waylight/internal/modules/dashboard"
	"waylight/internal/shared/kernel"
)

// tabSharing adapts the dashboard's Sharing to the team tabs port.
type tabSharing struct{ s *dashboard.Sharing }

func (t tabSharing) OwnTabs(ctx context.Context) ([]usecase.Tab, error) {
	tabs, err := t.s.OwnTabs(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]usecase.Tab, 0, len(tabs))
	for _, x := range tabs {
		out = append(out, usecase.Tab{ID: x.ID, Name: x.Name, Visibility: x.Visibility})
	}
	return out, nil
}

func (t tabSharing) Share(ctx context.Context, id string, v kernel.Visibility) error {
	return t.s.Share(ctx, id, v)
}

// Team tabs (#295), part of the ee/ features: the dashboard shows tabs that
// admins shared only while the licence covers them.
func init() {
	registerModule("teamtabs", func(d moduleDeps) modules.Module {
		m := teamtabs.New(d.db, tabSharing{d.tabSharing}, licenceModule(d).Licensing())
		d.tabSharing.Enable(m.SharedTabsOn)
		d.tabSharing.SetGroupDefault(m.DefaultTab)
		return m
	})
}
