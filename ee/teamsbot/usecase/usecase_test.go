package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	licence "trisse-home/ee/licence/domain"
	"trisse-home/ee/teamsbot/domain"
)

type conns map[string]string

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

type dir struct{}

func (dir) ObjectID(_ context.Context, token string) (string, error) { return "OID-" + token, nil }

type finder struct{ asked string }

func (f *finder) Find(_ context.Context, account, _ string) ([]domain.Hit, error) {
	f.asked = account
	return []domain.Hit{{Source: "Service", Title: "VPN", URL: "https://vpn.example"}}, nil
}

type level licence.Level

func (l level) Level(context.Context, string) licence.Level { return licence.Level(l) }

func TestQueryForTheMemberBehindTheTeamsUser(t *testing.T) {
	f := &finder{}
	s := New(Deps{Connections: conns{"m1": "a", "m2": "b"}, Directory: dir{}, Finder: f, Licensing: level(licence.LevelFull), Now: time.Now})
	res := s.Query(context.Background(), "oid-B", "vpn")
	if res.ComposeExtension.Type != "result" || f.asked != "m2" || len(res.ComposeExtension.Attachments) != 1 ||
		res.ComposeExtension.Attachments[0].Content.Tap.Value != "https://vpn.example" {
		t.Fatalf("%+v asked %q", res, f.asked)
	}
	if res := s.Query(context.Background(), "OID-unknown", "vpn"); res.ComposeExtension.Text != domain.NotConnected {
		t.Fatalf("unknown: %+v", res)
	}
	s = New(Deps{Connections: conns{}, Directory: dir{}, Finder: f, Licensing: level(licence.LevelNone), Now: time.Now})
	if res := s.Query(context.Background(), "OID-a", "vpn"); res.ComposeExtension.Text != domain.Unlicensed {
		t.Fatalf("unlicensed: %+v", res)
	}
}
