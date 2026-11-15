package domain

import (
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestChatChannels(t *testing.T) {
	ok := []Channel{
		{Kind: ChannelSlack, Name: "Ops", Secret: " https://hooks.slack.com/services/T000/B000/XXXX "},
		{Kind: ChannelTeams, Name: "IT", Secret: "https://prod-12.westeurope.logic.azure.com:443/workflows/abc/triggers/manual/paths/invoke?api-version=2016-06-01&sig=xyz"},
	}
	for _, c := range ok {
		c.URL = "https://ignored.example"
		if err := c.Validate(); err != nil {
			t.Errorf("%s rejected: %v", c.Kind, err)
		}
		if c.URL != "" || c.Secret == "" || c.Secret[0] == ' ' {
			t.Errorf("%s: the webhook belongs in Secret only: %+v", c.Kind, c)
		}
	}
	for _, c := range []Channel{
		{Kind: ChannelSlack, Name: "x", Secret: ""},
		{Kind: ChannelSlack, Name: "x", Secret: "https://evil.example/services/T/B/X"},
		{Kind: ChannelTeams, Name: "x", Secret: "http://insecure.example/hook"},
		{Kind: ChannelTeams, Name: "x", Secret: "not a url"},
	} {
		if err := c.Validate(); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%+v accepted", c)
		}
	}
}
