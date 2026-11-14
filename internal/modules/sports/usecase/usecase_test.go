package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/sports/domain"
	"trisse-home/internal/shared/kernel"
)

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

type fake struct {
	calls  int
	fail   bool
	season int
}

var bayern, bvb = domain.Team{ID: 40, Name: "FC Bayern München"}, domain.Team{ID: 7, Name: "Borussia Dortmund"}

func (f *fake) Teams(_ context.Context, _ string, season int) ([]domain.Team, error) {
	f.calls++
	f.season = season
	if f.fail {
		return nil, errors.New("timeout")
	}
	return []domain.Team{bayern, bvb}, nil
}
func (f *fake) Matches(context.Context, string, int) ([]domain.Match, error) {
	return []domain.Match{{Kickoff: time.Date(2026, 11, 8, 15, 30, 0, 0, time.UTC), Home: bvb, Away: bayern, Finished: true, HomeGoals: 1, AwayGoals: 2, HasScore: true}}, nil
}
func (f *fake) Table(context.Context, string, int) ([]domain.Standing, error) {
	return []domain.Standing{{Position: 1, Team: bayern, Points: 30}, {Position: 2, Team: bvb, Points: 25}}, nil
}

func TestOverview(t *testing.T) {
	ctx := context.Background()
	c := &clk{t: time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)}
	p := &fake{}
	uc := New(p, c)
	o, err := uc.Overview(ctx, "bl1", 40)
	if err != nil || o.Team.ID != 40 || o.Last == nil || o.Next != nil || o.Position.Points != 30 || p.season != 2026 {
		t.Fatalf("%+v %v", o, err)
	}
	if o, _ := uc.Overview(ctx, "bl1", 0); len(o.Table) != 2 || o.Team != nil {
		t.Fatalf("league view: %+v", o)
	}
	if p.calls != 1 {
		t.Fatal("not cached")
	}
	if _, err := uc.Overview(ctx, "bl1", 999); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := uc.Overview(ctx, "nba", 0); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
	// After the cache time a failing provider keeps the last data.
	c.t = c.t.Add(time.Hour)
	p.fail = true
	if _, err := uc.Overview(ctx, "bl1", 40); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Teams(ctx, "bl2"); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatal(err)
	}
}
