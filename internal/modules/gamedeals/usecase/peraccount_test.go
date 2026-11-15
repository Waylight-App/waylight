package usecase

import (
	"context"
	"testing"

	"trisse-home/internal/modules/gamedeals/domain"
	"trisse-home/internal/shared/kernel"
)

// accountRepo keeps settings per account like the gateway, with the
// admin's as fallback.
type accountRepo map[string]domain.Settings

func (r accountRepo) Load(ctx context.Context) (domain.Settings, error) {
	if s, ok := r[kernel.PrincipalFrom(ctx).SettingsAccount()]; ok {
		return s, nil
	}
	return r[kernel.SystemAccountID], nil
}

func (r accountRepo) Save(ctx context.Context, s domain.Settings) error {
	r[kernel.PrincipalFrom(ctx).SettingsAccount()] = s
	return nil
}

// wishlistByID returns one deal named after the Steam ID.
type wishlistByID struct{ calls map[string]int }

func (w *wishlistByID) Wishlist(_ context.Context, steamID, _ string) ([]domain.Deal, error) {
	w.calls[steamID]++
	return []domain.Deal{{AppID: 1, Name: steamID, DiscountPct: 90}}, nil
}

// TestDealsPerAccount: each account sees the offers of its own wishlist;
// accounts without settings share the admin's (ADR 0018).
func TestDealsPerAccount(t *testing.T) {
	const adminID, annaID = "76561198000000001", "76561198000000002"
	repo := accountRepo{kernel.SystemAccountID: {SteamID: adminID, Country: "DE"}}
	steam := &wishlistByID{calls: map[string]int{}}
	uc := NewGameDeals(Deps{Repo: repo, Wishlist: steam, Free: &fakeEpic{}, Clock: clk{}})
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})

	if _, err := uc.SaveSettings(anna, domain.Settings{SteamID: annaID, Country: "DE"}); err != nil {
		t.Fatal(err)
	}
	name := func(ctx context.Context) string {
		o, err := uc.Overview(ctx)
		if err != nil || len(o.Deals) != 1 {
			t.Fatalf("overview %+v %v", o, err)
		}
		return o.Deals[0].Name
	}
	if name(anna) != annaID || name(ben) != adminID || name(context.Background()) != adminID {
		t.Fatal("offers are mixed between accounts")
	}
	before := steam.calls[annaID]
	if err := uc.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if steam.calls[annaID] != before+1 || steam.calls[adminID] < 2 {
		t.Fatalf("the job must refresh every known wishlist: %v", steam.calls)
	}
}
