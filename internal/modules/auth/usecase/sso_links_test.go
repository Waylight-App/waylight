package usecase

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/auth/domain"
)

type fakeIdentities struct{ links []domain.IdentityLink }

func (f *fakeIdentities) Add(_ context.Context, l domain.IdentityLink) error {
	for _, x := range f.links {
		if x.Issuer == l.Issuer && x.Subject == l.Subject {
			return domain.ErrIdentityTaken
		}
	}
	f.links = append(f.links, l)
	return nil
}

func (f *fakeIdentities) Find(_ context.Context, issuer, subject string) (string, error) {
	for _, x := range f.links {
		if x.Issuer == issuer && x.Subject == subject {
			return x.AccountID, nil
		}
	}
	return "", domain.ErrIdentityNotFound
}

func (f *fakeIdentities) ListByAccount(_ context.Context, accountID string) ([]domain.IdentityLink, error) {
	var out []domain.IdentityLink
	for _, x := range f.links {
		if x.AccountID == accountID {
			out = append(out, x)
		}
	}
	return out, nil
}

func (f *fakeIdentities) Remove(_ context.Context, accountID, issuer, subject string) error {
	for i, x := range f.links {
		if x.AccountID == accountID && x.Issuer == issuer && x.Subject == subject {
			f.links = append(f.links[:i], f.links[i+1:]...)
			return nil
		}
	}
	return domain.ErrIdentityNotFound
}

func (f *fakeIdentities) RemoveAccount(_ context.Context, accountID string) error {
	kept := f.links[:0]
	for _, x := range f.links {
		if x.AccountID != accountID {
			kept = append(kept, x)
		}
	}
	f.links = kept
	return nil
}

// licenceStub answers like the ee/ module: sign in with any licence, link
// with full use only.
type licenceStub struct{ signIn, link bool }

func (licenceStub) Resolve(context.Context, domain.ExternalIdentity) (string, error) {
	return "", domain.ErrSSONotAllowed
}

func (l licenceStub) SSOLicensed(context.Context) (bool, bool) { return l.signIn, l.link }

// Linking a sign-in (#527): the owner always, others with the licence; an
// expired licence keeps existing links working but allows no new ones.
func TestLinkedSignIns(t *testing.T) {
	e := newEnv().withAccount()
	member, _ := domain.NewMember("m1", "anna", "hash:x", e.clock.now)
	e.accounts.guests = map[string]domain.Account{"m1": member}
	ids := &fakeIdentities{}
	provider := &fakeOIDC{}
	uc := NewSSO(e.deps, SSOConfig{OIDC: provider, Issuer: "https://login.example", Identities: ids, LocalLogin: true})
	client := domain.Client{IP: "10.0.0.7"}

	link := func(account, subject string) error {
		_, state, err := uc.BeginLink(ctx, account)
		if err != nil {
			return err
		}
		provider.id = domain.ExternalIdentity{Subject: subject, Email: subject + "@outlook.com"}
		res, _, err := uc.CompleteOIDC(ctx, state, "good", client)
		if err == nil && !res.Linked {
			t.Fatalf("%s: no link reported", account)
		}
		return err
	}
	signIn := func(subject string) (SessionResponse, error) {
		_, state, err := uc.BeginOIDC(ctx)
		if err != nil {
			t.Fatal(err)
		}
		provider.id = domain.ExternalIdentity{Subject: subject}
		res, _, err := uc.CompleteOIDC(ctx, state, "good", client)
		return res, err
	}

	if err := link(domain.OwnerID, "owner-sub"); err != nil {
		t.Fatal(err)
	}
	if res, err := signIn("owner-sub"); err != nil || res.SessionToken == "" {
		t.Fatalf("owner: %+v %v", res, err)
	}
	if err := link("m1", "anna-sub"); !errors.Is(err, domain.ErrLinkNeedsLicence) {
		t.Fatalf("member without a licence: %v", err)
	}

	uc.SetExternalAccounts(licenceStub{signIn: true, link: true})
	if err := link("m1", "anna-sub"); err != nil {
		t.Fatal(err)
	}
	if err := link("m1", "owner-sub"); !errors.Is(err, domain.ErrIdentityTaken) {
		t.Fatalf("taken: %v", err)
	}
	if _, err := signIn("anna-sub"); err != nil {
		t.Fatalf("member: %v", err)
	}

	// Expired: the existing link still counts, a new one does not.
	uc.SetExternalAccounts(licenceStub{signIn: true, link: false})
	if _, err := signIn("anna-sub"); err != nil {
		t.Fatalf("member, expired licence: %v", err)
	}
	if _, _, err := uc.BeginLink(ctx, "m1"); !errors.Is(err, domain.ErrLinkNeedsLicence) {
		t.Fatalf("new link, expired licence: %v", err)
	}
	// No licence at all: the member's link no longer signs in.
	uc.SetExternalAccounts(nil)
	if _, err := signIn("anna-sub"); !errors.Is(err, domain.ErrSSONotAllowed) {
		t.Fatalf("member, no licence: %v", err)
	}

	if err := uc.Unlink(ctx, domain.OwnerID, "owner-sub", client); err != nil {
		t.Fatal(err)
	}
	if links, _ := uc.Links(ctx, domain.OwnerID); len(links) != 0 {
		t.Fatalf("after unlink: %v", links)
	}
}
