package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"waylight/internal/modules/notifications/domain"
)

func TestPostToChatOnlyToEnabledSlackOrTeams(t *testing.T) {
	ctx := context.Background()
	slack, teams := &recordingPublisher{}, &recordingPublisher{}
	channels := &memChannels{list: []domain.Channel{
		{ID: "s1", Kind: domain.ChannelSlack, Name: "#general", Enabled: true, MinSeverity: domain.SeverityCritical, Types: []string{"service"}},
		{ID: "t1", Kind: domain.ChannelTeams, Name: "Team", Enabled: false},
		{ID: "w1", Kind: domain.ChannelWebhook, Name: "Hook", Enabled: true},
	}}
	n := NewNotifier(Deps{Repo: &memRepo{}, Sender: &fakeSender{}, Clock: &clk{}, IDs: &seq{}, Channels: channels,
		Maintenance: &memMaintenance{}, Pending: &memPending{},
		Publishers: map[domain.ChannelKind]Publisher{domain.ChannelSlack: slack, domain.ChannelTeams: teams, domain.ChannelWebhook: &recordingPublisher{}}})

	list, err := n.ChatChannels(ctx)
	if err != nil || len(list) != 1 || list[0].ID != "s1" {
		t.Fatalf("chat channels %v %v", list, err)
	}
	// The rules of the channel (critical only, type service) do not apply.
	if err := n.PostToChat(ctx, "s1", "Office closed", "On Friday.", "https://home.example/?announcement=a1"); err != nil {
		t.Fatal(err)
	}
	if len(slack.sent) != 1 || slack.sent[0].Title != "Office closed" ||
		!strings.HasSuffix(slack.sent[0].Body, "On Friday.\n\nhttps://home.example/?announcement=a1") {
		t.Fatalf("sent %+v", slack.sent)
	}
	for _, id := range []string{"t1", "w1", "nope"} {
		if err := n.PostToChat(ctx, id, "x", "y", ""); !errors.Is(err, ErrNotAChatChannel) {
			t.Fatalf("%s: %v", id, err)
		}
	}
}
