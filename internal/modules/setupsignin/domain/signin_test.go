package domain

import (
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

func TestNormaliseURL(t *testing.T) {
	if got, err := NormaliseURL(" https://cloud.example/nc/ "); err != nil || got != "https://cloud.example/nc" {
		t.Fatalf("%q %v", got, err)
	}
	for _, bad := range []string{"", "cloud.example", "ftp://cloud.example", "https://user:pw@cloud.example", "https://cloud.example/?x=1", "javascript:alert(1)"} {
		if _, err := NormaliseURL(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	if ValidPage("javascript:alert(1)") || ValidPage("data:text/html,x") || !ValidPage("https://cloud.example/login/v2/flow/a?x=1") {
		t.Fatal("sign-in page check")
	}
}

// TestLifetime: a finished sign-in forgets its secrets and waits at least
// RedeemWindow for the module, even when it finished late.
func TestLifetime(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	s := New("id", "anna", Nextcloud, "https://cloud.example", now)
	s.PollToken = "poll"
	if !s.PollDue(now) || s.Expired(now.Add(TTL-time.Second)) || !s.Expired(now.Add(TTL)) {
		t.Fatal("pending lifetime")
	}
	late := now.Add(TTL - time.Second)
	s.Finish(kernel.SignInGrant{Secret: "pw"}, late)
	if s.PollToken != "" || s.Status != Ready || s.PollDue(late) || s.Expired(late.Add(RedeemWindow-time.Second)) {
		t.Fatalf("finished %+v", s)
	}
}
