//go:build ee

package main

import (
	"context"
	"log/slog"

	"waylight/ee/absences"
	"waylight/ee/presence"
	presencedomain "waylight/ee/presence/domain"
	presenceusecase "waylight/ee/presence/usecase"
	"waylight/internal/modules"
	"waylight/internal/modules/auth"
	"waylight/internal/shared/kernel"
)

// Live presence (#346), part of the ee/ features.
func init() {
	registerModule("presence", func(d moduleDeps) modules.Module {
		deps := presence.Deps{DB: d.db, Licensing: licenceModule(d).Licensing(), Directory: presenceDirectory{d.auth.Provisioning()},
			Clock: d.clock}
		if d.cfg.ModuleEnabled("absences") {
			deps.Absences = plannedAbsences{absencesModule(d)}
		}
		m := presence.New(deps)
		d.auth.OnAccountDeleted(func(ctx context.Context, accountID string) {
			if err := m.ForgetAccount(ctx, accountID); err != nil {
				slog.Default().Warn("presence: remove a deleted account's status", "err", err)
			}
		})
		return m
	})
}

// presenceDirectory lists the people and groups of the auth module: the
// owner, admins and members who are not disabled.
type presenceDirectory struct{ p auth.Provisioning }

func (d presenceDirectory) People(ctx context.Context) (map[string]string, error) {
	list, err := d.p.Members(ctx)
	if err != nil {
		return nil, err
	}
	out := map[string]string{kernel.SystemAccountID: kernel.SystemAccountID}
	for _, m := range list {
		if !m.Disabled {
			out[m.ID] = m.Name
		}
	}
	return out, nil
}

func (d presenceDirectory) GroupMembers(ctx context.Context, groupID string) ([]string, error) {
	groups, err := d.p.Groups(ctx)
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		if g.ID() == groupID {
			return g.Accounts(), nil
		}
	}
	return nil, presencedomain.ErrNoGroup
}

// plannedAbsences reads who is away today from the absences module (#299).
type plannedAbsences struct{ m *absences.Module }

func (a plannedAbsences) AwayOn(ctx context.Context, day string) (map[string]presencedomain.Absent, error) {
	list, err := a.m.AwayOn(ctx, day)
	if err != nil {
		return nil, err
	}
	out := make(map[string]presencedomain.Absent, len(list))
	for _, w := range list {
		out[w.AccountID] = presencedomain.Absent{Until: w.Until, Remote: w.Remote}
	}
	return out, nil
}

var _ presenceusecase.Absences = plannedAbsences{}
