package usecase

import (
	"context"
	"testing"

	"waylight/internal/modules/services/domain"
	"waylight/internal/shared/kernel"
)

func TestSeedSampleData(t *testing.T) {
	repo := &memRepo{}
	uc := NewSeedSampleData(Deps{Repo: repo, Status: &memStatus{m: map[string]domain.Status{}}, Observations: kernel.DiscardObservations{}, Clock: clk{}, IDs: &ids{}})
	if created, err := uc.Execute(context.Background()); err != nil || !created || len(repo.items) != 3 {
		t.Fatalf("first seed: %v %v %d", created, err, len(repo.items))
	}
	if created, _ := uc.Execute(context.Background()); created || len(repo.items) != 3 {
		t.Fatal("seeding must not run twice")
	}
	for _, s := range repo.items {
		if !s.Example {
			t.Fatalf("sample tile not marked as example: %+v", s)
		}
	}
}

func TestDiscardExamples(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{}
	deps := Deps{Repo: repo, Status: &memStatus{m: map[string]domain.Status{}}, Checker: fakeChecker{}, History: &memHistory{},
		Observations: kernel.DiscardObservations{}, Clock: clk{}, IDs: &ids{}}
	if _, err := NewSeedSampleData(deps).Execute(ctx); err != nil {
		t.Fatal(err)
	}
	uc := NewServices(deps)
	own, _ := uc.Create(ctx, ServiceInput{Name: "Wiki", URL: "https://wiki.example"})
	// An edited example tile is an own service and stays.
	edited, err := uc.Update(ctx, repo.items[0].ID, ServiceInput{Name: "Search", URL: "https://duckduckgo.com/"})
	if err != nil || edited.Example {
		t.Fatalf("edit: %+v %v", edited, err)
	}
	if err := uc.DiscardExamples(ctx); err != nil {
		t.Fatal(err)
	}
	list, _ := uc.List(ctx)
	if len(list) != 2 || list[0].ID != edited.ID && list[1].ID != edited.ID || list[0].ID != own.ID && list[1].ID != own.ID {
		t.Fatalf("after discard: %+v", list)
	}
}
