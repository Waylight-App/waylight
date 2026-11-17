package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"waylight/internal/modules/todos/domain"
	"waylight/internal/shared/kernel"
)

type fakeOAuth struct {
	redirect  string
	n         int
	refreshed int
	revoked   []string
	rejected  bool
}

func (f *fakeOAuth) RedirectURL() string { return f.redirect }
func (f *fakeOAuth) NewState() string    { f.n++; return fmt.Sprintf("state%d", f.n) }
func (f *fakeOAuth) AuthURL(kind domain.SourceKind, clientID, state, verifier string) string {
	return "https://login.example/" + string(kind) + "?client_id=" + clientID + "&state=" + state
}
func (f *fakeOAuth) Exchange(_ context.Context, _ domain.SourceKind, _, secret, code, _ string) (domain.OAuthGrant, error) {
	if code != "good" {
		return domain.OAuthGrant{}, errors.New("invalid_grant")
	}
	return domain.OAuthGrant{ClientSecret: secret, Access: "a1", Refresh: "r1", Expiry: time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC)}, nil
}
func (f *fakeOAuth) Refresh(_ context.Context, _ domain.SourceKind, _ string, g domain.OAuthGrant) (domain.OAuthGrant, error) {
	if f.rejected {
		return domain.OAuthGrant{}, fmt.Errorf("%w: expired", ErrGrantRejected)
	}
	f.refreshed++
	g.Access, g.Expiry = fmt.Sprintf("a%d", f.refreshed+1), g.Expiry.Add(time.Hour)
	return g, nil
}
func (f *fakeOAuth) Revoke(_ context.Context, _ domain.SourceKind, _ string, g domain.OAuthGrant) error {
	f.revoked = append(f.revoked, g.Refresh)
	return nil
}

func TestOAuthSignIn(t *testing.T) {
	ctx := context.Background()
	d, p, src := externalDeps()
	o := &fakeOAuth{}
	d.OAuth = o
	uc := NewSources(d)
	in := SourceInput{Kind: "microsoft", Username: "client-1", Password: "secret"}

	if _, err := uc.StartOAuth(ctx, in); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("signing in without PUBLIC_URL", err)
	}
	o.redirect = "https://home.example/api/todos/settings/oauth/callback"
	if info := uc.OAuthInfo(); info.RedirectURL != o.redirect {
		t.Fatal(info)
	}
	if _, err := uc.StartOAuth(ctx, SourceInput{Kind: "caldav", Username: "x"}); err == nil {
		t.Fatal("caldav is not signed in")
	}
	st, err := uc.StartOAuth(ctx, in)
	if err != nil || st.State == "" || st.URL == "" {
		t.Fatal(st, err)
	}
	if s, _ := uc.OAuthStatus(st.State); s != "pending" {
		t.Fatal(s)
	}
	if _, err := uc.OAuthStatus("unknown"); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal(err)
	}
	// A refused sign-in is reported, and its lists cannot be read.
	bad, _ := uc.StartOAuth(ctx, in)
	if err := uc.FinishOAuth(ctx, bad.State, "", "access_denied"); err == nil {
		t.Fatal("provider error accepted")
	}
	if s, _ := uc.OAuthStatus(bad.State); s != "failed" {
		t.Fatal(s)
	}
	if _, err := uc.Discover(ctx, SourceInput{Kind: "microsoft", OAuthState: bad.State}); err == nil {
		t.Fatal("discovered with a failed sign-in")
	}
	if err := uc.FinishOAuth(ctx, "forged", "good", ""); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal("unknown state accepted", err)
	}

	if err := uc.FinishOAuth(ctx, st.State, "good", ""); err != nil {
		t.Fatal(err)
	}
	if s, _ := uc.OAuthStatus(st.State); s != "ready" {
		t.Fatal(s)
	}
	if _, err := uc.Discover(ctx, SourceInput{Kind: "microsoft", OAuthState: st.State}); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Save(ctx, SourceInput{Kind: "google", Name: "x", ListID: "l1", OAuthState: st.State}); err == nil {
		t.Fatal("a Microsoft sign-in used for Google")
	}
	v, err := uc.Save(ctx, SourceInput{Kind: "microsoft", Name: "To Do", ListID: "AAMkAD=", OAuthState: st.State})
	if err != nil || v.Username != "client-1" {
		t.Fatalf("%+v %v", v, err)
	}
	s := src.m[v.ID]
	if s.Grant.Access != "a1" || s.Grant.Refresh != "r1" || s.Grant.ClientSecret != "secret" || s.ListID != "AAMkAD=" || len(p.opened) == 0 {
		t.Fatalf("stored: %+v", s)
	}
	// Renaming keeps the sign-in.
	if _, err := uc.Save(ctx, SourceInput{ID: v.ID, Kind: "microsoft", Name: "Aufgaben", ListID: "AAMkAD="}); err != nil || src.m[v.ID].Grant.Access != "a1" {
		t.Fatal(err)
	}

	// Tokens: valid ones are used, expiring ones refreshed and stored.
	ts := NewTokenSource(d)
	clock := d.Clock.(*clk)
	if tok, _ := ts.AccessToken(ctx, src.m[v.ID]); tok != "a1" || o.refreshed != 0 {
		t.Fatal(tok)
	}
	clock.t = time.Date(2026, 11, 14, 9, 59, 0, 0, time.UTC)
	if tok, err := ts.AccessToken(ctx, src.m[v.ID]); err != nil || tok != "a2" || src.m[v.ID].Grant.Access != "a2" || src.m[v.ID].Grant.Refresh != "r1" {
		t.Fatalf("refresh: %q %v %+v", tok, err, src.m[v.ID].Grant)
	}
	clock.t = clock.t.Add(2 * time.Hour)
	o.rejected = true
	if _, err := ts.AccessToken(ctx, src.m[v.ID]); !errors.Is(err, domain.ErrSignInAgain) {
		t.Fatal(err)
	}
	// Disconnecting revokes the sign-in at the provider.
	if err := uc.Delete(ctx, v.ID); err != nil || len(o.revoked) != 1 || o.revoked[0] != "r1" {
		t.Fatal(err, o.revoked)
	}
	// Sign-ins expire.
	old, _ := uc.StartOAuth(ctx, in)
	clock.t = clock.t.Add(signInTTL + time.Minute)
	_, _ = uc.StartOAuth(ctx, in) // prunes
	if _, err := uc.OAuthStatus(old.State); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal("expired sign-in still known", err)
	}
}
