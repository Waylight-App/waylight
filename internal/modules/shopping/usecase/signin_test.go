package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/shopping/domain"
	"waylight/internal/shared/kernel"
)

type oneSignIn struct {
	id    string
	grant kernel.SignInGrant
}

func (o *oneSignIn) Redeem(_ context.Context, id, service string) (kernel.SignInGrant, error) {
	if o.id == "" || id != o.id || service != kernel.SignInNextcloud {
		return kernel.SignInGrant{}, kernel.ErrSignInUnknown
	}
	o.id = ""
	return o.grant, nil
}

// TestSaveWithSignIn: the app password of a sign-in is only used for a list
// on the Nextcloud that issued it.
func TestSaveWithSignIn(t *testing.T) {
	ctx := context.Background()
	grant := kernel.SignInGrant{URL: "https://cloud.example", Username: "anna", Secret: "app-pw"}
	repo := &memRepo{}
	signIns := &oneSignIn{id: "s1", grant: grant}
	uc := NewShopping(repo, &memStore{items: map[string]domain.Item{}}, ids{}, signIns)

	if _, err := uc.SaveSettings(ctx, SettingsInput{ListURL: "https://evil.example/remote.php/dav/calendars/anna/shopping/", SignIn: "s1"}); !errors.Is(err, errListElsewhere) || repo.s != nil {
		t.Fatalf("list on another server: %v", err)
	}
	signIns.id = "s2"
	if _, err := uc.SaveSettings(ctx, SettingsInput{ListURL: "https://cloud.example.evil/remote.php/x/", SignIn: "s2"}); !errors.Is(err, errListElsewhere) {
		t.Fatalf("a host that only starts like the signed-in one: %v", err)
	}
	signIns.id = "s3"
	if _, err := uc.SaveSettings(ctx, SettingsInput{ListURL: "https://cloud.example/remote.php/dav/calendars/anna/shopping/", SignIn: "s3"}); err != nil {
		t.Fatal(err)
	}
	if repo.s.Username != "anna" || repo.s.Password != "app-pw" {
		t.Fatalf("saved %+v", repo.s)
	}
}
