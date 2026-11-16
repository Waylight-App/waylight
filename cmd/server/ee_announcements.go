//go:build ee

package main

import (
	"context"
	"log/slog"

	"trisse-home/ee/announcements"
	"trisse-home/ee/announcements/usecase"
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/auth"
	"trisse-home/internal/modules/notifications"
	"trisse-home/internal/shared/kernel"
)

// accountNames shows authors by their account name; the owner has none in
// the member list and is shown as "owner".
type accountNames struct{ p auth.Provisioning }

func (n accountNames) Name(ctx context.Context, id string) string {
	if id == kernel.SystemAccountID {
		return id
	}
	m, err := n.p.Member(ctx, id)
	if err != nil {
		return "" // a deleted account
	}
	return m.Name
}

// chatPoster lets announcements post to the notification module's Slack and
// Teams channels (#332).
type chatPoster struct{ n *notifications.Module }

func (c chatPoster) ChatChannels(ctx context.Context) ([]usecase.ChatChannel, error) {
	list, err := c.n.ChatChannels(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]usecase.ChatChannel, 0, len(list))
	for _, ch := range list {
		out = append(out, usecase.ChatChannel{ID: ch.ID, Name: ch.Name, Kind: string(ch.Kind)})
	}
	return out, nil
}

func (c chatPoster) PostToChat(ctx context.Context, channelID, title, body, link string) error {
	return c.n.PostToChat(ctx, channelID, title, body, link)
}

// Announcements (#297), part of the ee/ features.
func init() {
	registerModule("announcements", func(d moduleDeps) modules.Module {
		var chat usecase.Chat
		if d.notifications != nil {
			chat = chatPoster{d.notifications}
		}
		m := announcements.New(announcements.Deps{
			Chat: chat, PublicURL: d.cfg.PublicURL,
			DB: d.db, Licensing: licenceModule(d).Licensing(), Names: accountNames{d.auth.Provisioning()},
			Alerts: d.alerts, IDs: d.ids, Clock: d.clock,
			Audit: func(ctx context.Context, action, id string) {
				if changeSink != nil {
					p := kernel.PrincipalFrom(ctx)
					changeSink.RecordChange(ctx, modules.Change{AccountID: p.AccountID, Role: p.Role, Module: "announcements",
						Route: action, ObjectID: id})
				}
			},
		})
		d.auth.OnAccountDeleted(func(ctx context.Context, accountID string) {
			if err := m.ForgetAccount(ctx, accountID); err != nil {
				slog.Default().Warn("announcements: forget read marks", "err", err)
			}
		})
		return m
	})
}
