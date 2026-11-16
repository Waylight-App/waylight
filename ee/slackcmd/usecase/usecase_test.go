package usecase

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	licence "trisse-home/ee/licence/domain"
	"trisse-home/ee/slackcmd/domain"
)

type conns map[string]string // account → token

func (c conns) AccountsWith(context.Context, string) ([]string, error) {
	var out []string
	for a := range c {
		out = append(out, a)
	}
	return out, nil
}
func (c conns) AccessToken(_ context.Context, a, _ string) (string, error) {
	if t, ok := c[a]; ok {
		return t, nil
	}
	return "", errors.New("not connected")
}

type slack struct{ calls atomic.Int32 }

func (s *slack) Identity(_ context.Context, token string) (string, string, error) {
	s.calls.Add(1)
	return "T1", strings.TrimPrefix(token, "xoxp-"), nil
}

type finder struct{ asked string }

func (f *finder) Find(_ context.Context, account, q string) ([]domain.Hit, error) {
	f.asked = account
	return []domain.Hit{{Source: "Service", Title: "VPN", URL: "https://vpn.example"}}, nil
}

type level licence.Level

func (l level) Level(context.Context, string) licence.Level { return licence.Level(l) }

func TestCommandAnswersForTheMemberBehindTheSlackUser(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	sl, f := &slack{}, &finder{}
	c := conns{"m1": "xoxp-U1", "m2": "xoxp-U2"}
	s := New(Deps{Connections: c, Slack: sl, Finder: f, Licensing: level(licence.LevelFull), Now: func() time.Time { return now },
		SettingsURL: "https://home.example/?settings=plugin-connections"})

	if got := s.Handle(context.Background(), Command{TeamID: "T1", UserID: "U2", Command: "/home", Text: " vpn "}); !strings.Contains(got, "VPN") || f.asked != "m2" {
		t.Fatalf("reply %q for %q", got, f.asked)
	}
	// Someone unknown gets the link to connect and no results.
	if got := s.Handle(context.Background(), Command{TeamID: "T1", UserID: "U9", Text: "vpn"}); !strings.Contains(got, "Connect your Slack account") ||
		!strings.Contains(got, "settings=plugin-connections") {
		t.Fatalf("unknown: %q", got)
	}
	// Unknown users do not make the service ask Slack again right away.
	before := sl.calls.Load()
	s.Handle(context.Background(), Command{TeamID: "T1", UserID: "U9", Text: "vpn"})
	if sl.calls.Load() != before {
		t.Fatal("rescanned within the pause")
	}
	// A new connection is found after the pause.
	c["m3"] = "xoxp-U9"
	now = now.Add(time.Minute)
	if s.Handle(context.Background(), Command{TeamID: "T1", UserID: "U9", Text: "vpn"}); f.asked != "m3" {
		t.Fatalf("after rescan: %q", f.asked)
	}
	// Same user ID in another workspace is someone else.
	if got := s.Handle(context.Background(), Command{TeamID: "T2", UserID: "U2", Text: "vpn"}); !strings.Contains(got, "Connect") {
		t.Fatalf("other team: %q", got)
	}
}

func TestCommandNeedsTheLicenceAndWords(t *testing.T) {
	s := New(Deps{Connections: conns{}, Slack: &slack{}, Finder: &finder{}, Licensing: level(licence.LevelRead), Now: time.Now})
	if got := s.Handle(context.Background(), Command{Text: "vpn"}); got != domain.Unlicensed {
		t.Fatalf("expired: %q", got)
	}
	s = New(Deps{Connections: conns{}, Slack: &slack{}, Finder: &finder{}, Licensing: level(licence.LevelFull), Now: time.Now})
	if got := s.Handle(context.Background(), Command{Command: "/home"}); !strings.HasPrefix(got, "Search the start page") {
		t.Fatalf("help: %q", got)
	}
}
