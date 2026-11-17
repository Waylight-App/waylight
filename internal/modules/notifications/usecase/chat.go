package usecase

import (
	"context"
	"fmt"
	"strings"

	"waylight/internal/modules/notifications/domain"
	"waylight/internal/shared/kernel"
)

// ChatChannel is a Slack or Teams channel other modules can post to (#332).
type ChatChannel struct {
	ID   string
	Name string
	Kind domain.ChannelKind
}

// ErrNotAChatChannel refuses posting to a channel that is not an enabled
// Slack or Teams channel.
var ErrNotAChatChannel = kernel.NewPublicError(kernel.ErrNotFound, "The channel is not an enabled Slack or Teams channel.")

func isChat(c domain.Channel) bool {
	return c.Enabled && (c.Kind == domain.ChannelSlack || c.Kind == domain.ChannelTeams)
}

// ChatChannels lists the enabled Slack and Teams channels.
func (uc *notifier) ChatChannels(ctx context.Context) ([]ChatChannel, error) {
	list, err := uc.d.Channels.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	out := []ChatChannel{}
	for _, c := range list {
		if isChat(c) {
			out = append(out, ChatChannel{ID: c.ID, Name: c.Name, Kind: c.Kind})
		}
	}
	return out, nil
}

// PostToChat posts one message to a Slack or Teams channel, outside the
// routing rules: the author chose the channel, so severity, types and quiet
// hours do not apply. link, when set, is added below the text.
func (uc *notifier) PostToChat(ctx context.Context, channelID, title, body, link string) error {
	c, err := uc.channel(ctx, channelID)
	if err != nil || !isChat(c) {
		return ErrNotAChatChannel
	}
	p, ok := uc.d.Publishers[c.Kind]
	if !ok {
		return ErrNotAChatChannel
	}
	text := strings.TrimSpace(body)
	if link != "" {
		text = strings.TrimSpace(text + "\n\n" + link)
	}
	if err := p.Publish(ctx, c, domain.Message{Title: title, Body: text, Type: "announcement", Severity: domain.SeverityInfo,
		Priority: domain.PriorityDefault}); err != nil {
		return kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	return nil
}
