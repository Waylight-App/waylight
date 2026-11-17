package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/wiki/domain"
	"waylight/internal/shared/kernel"
)

// personalSource answers only for the account it was connected by.
type personalSource struct {
	name    string
	account string
	fail    bool
}

func (p personalSource) Name() string { return p.name }
func (p personalSource) Search(ctx context.Context, q string, _ int) ([]domain.Page, error) {
	if p.fail {
		return nil, errors.New("service down")
	}
	if kernel.PrincipalFrom(ctx).AccountID != p.account {
		return nil, ErrNotConnected
	}
	return []domain.Page{{Title: "Private plan " + q, URL: "https://wiki.example/p"}}, nil
}

func TestMemberSourcesArePersonalAndNeverCached(t *testing.T) {
	members := &MemberSources{}
	members.Add(personalSource{name: "Confluence", account: "m1"})
	members.Add(personalSource{name: "Notion", fail: true})
	uc := New(Deps{Repo: &memRepo{}, Searchers: map[domain.Kind]PageSearcher{}, IDs: &ids{}, Clock: &clk{t: time.Now()}, Members: members})
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "m1", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "m2", Role: "member"})

	res, err := uc.Search(anna, "roadmap")
	if err != nil || len(res.Pages) != 1 || res.Pages[0].Wiki != "Confluence" {
		t.Fatalf("anna: %+v %v", res, err)
	}
	if len(res.Failed) != 1 || res.Failed[0] != "Notion" {
		t.Fatalf("failed sources: %v", res.Failed)
	}
	// Ben searches the same words right after: anna's page must not show up.
	res, _ = uc.Search(ben, "roadmap")
	if len(res.Pages) != 0 {
		t.Fatalf("ben sees anna's results: %+v", res.Pages)
	}
}
