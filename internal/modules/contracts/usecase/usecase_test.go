package usecase

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"waylight/internal/modules/contracts/domain"
	"waylight/internal/shared/kernel"
)

// memRepo keeps contracts in memory and applies the visibility rule like
// the SQLite gateway.
type memRepo struct{ items map[string]domain.Contract }

func (m *memRepo) List(ctx context.Context) ([]domain.Contract, error) {
	var out []domain.Contract
	for _, c := range m.items {
		if c.Access.Visibility.VisibleTo(kernel.PrincipalFrom(ctx), c.Access.Owner) {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *memRepo) Get(ctx context.Context, id string) (domain.Contract, error) {
	c, ok := m.items[id]
	if !ok || !c.Access.Visibility.VisibleTo(kernel.PrincipalFrom(ctx), c.Access.Owner) {
		return domain.Contract{}, domain.ErrNotFound
	}
	return c, nil
}

func (m *memRepo) Save(_ context.Context, c domain.Contract) error { m.items[c.ID] = c; return nil }
func (m *memRepo) Delete(_ context.Context, id string) error       { delete(m.items, id); return nil }

func (m *memRepo) MarkReminded(_ context.Context, id, deadline string) error {
	c := m.items[id]
	c.RemindedFor = deadline
	m.items[id] = c
	return nil
}

type clk struct{ now *time.Time }

func (c clk) Now() time.Time { return *c.now }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "c" + strconv.Itoa(i.n) }

type alerts struct{ got []kernel.Alert }

func (a *alerts) Alert(_ context.Context, al kernel.Alert) { a.got = append(a.got, al) }

var (
	admin = kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna  = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben   = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
)

func input(name, end string, notice int) domain.Input {
	return domain.Input{Name: name, CostCents: 1000, Period: "monthly", TermEnd: end, NoticeDays: notice, RenewMonths: 12, RemindDays: 14}
}

func TestPrivateByDefaultForMembers(t *testing.T) {
	now := time.Date(2026, 11, 23, 9, 0, 0, 0, time.UTC)
	c := NewContracts(Deps{Repo: &memRepo{items: map[string]domain.Contract{}}, Clock: clk{&now}, IDs: &ids{}, Location: time.UTC})
	v, err := c.Create(anna, input("Gym", "2027-03-31", 90), "")
	if err != nil || v.Access.Visibility != kernel.Private {
		t.Fatalf("%+v %v", v, err)
	}
	if list, _ := c.List(ben); len(list) != 0 {
		t.Fatalf("ben sees %v", list)
	}
	shared, _ := c.Create(anna, input("Internet", "2027-06-30", 30), "everyone")
	if _, err := c.Update(ben, shared.ID, input("Mine", "2027-06-30", 30), ""); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed anna's contract: %v", err)
	}
}

func TestListByDeadline(t *testing.T) {
	now := time.Date(2026, 11, 23, 9, 0, 0, 0, time.UTC)
	c := NewContracts(Deps{Repo: &memRepo{items: map[string]domain.Contract{}}, Clock: clk{&now}, IDs: &ids{}, Location: time.UTC})
	ended := input("Ended", "2026-01-31", 30)
	ended.RenewMonths = 0
	for _, in := range []domain.Input{ended, input("Late", "2027-06-30", 30), input("Soon", "2026-12-10", 7)} {
		if _, err := c.Create(admin, in, ""); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := c.List(admin)
	if list[0].Name != "Soon" || !list[0].Soon || list[1].Name != "Late" || list[2].Name != "Ended" || list[2].Term.Open {
		t.Fatalf("%+v", list)
	}
}

func TestRemindOncePerDeadlineAndOnlyWhatTheAdminSees(t *testing.T) {
	now := time.Date(2026, 12, 20, 9, 0, 0, 0, time.UTC)
	repo := &memRepo{items: map[string]domain.Contract{}}
	sink := &alerts{}
	c := NewContracts(Deps{Repo: repo, Clock: clk{&now}, IDs: &ids{}, Location: time.UTC, Alerts: sink})
	if _, err := c.Create(admin, input("Mobile phone", "2027-03-31", 90), ""); err != nil { // deadline 2026-12-31
		t.Fatal(err)
	}
	if _, err := c.Create(anna, input("Anna's gym", "2027-03-31", 90), ""); err != nil { // private
		t.Fatal(err)
	}
	n, err := c.Remind(context.Background())
	if err != nil || n != 1 || len(sink.got) != 1 || sink.got[0].Kind != "reminder" || sink.got[0].Title != "📄 Mobile phone" {
		t.Fatalf("%d %v %+v", n, err, sink.got)
	}
	if want := "Cancel by 2026-12-31 to end the contract on 2027-03-31. Otherwise it renews for 12 months."; sink.got[0].Body != want {
		t.Fatalf("body %q", sink.got[0].Body)
	}
	if n, _ := c.Remind(context.Background()); n != 0 {
		t.Fatal("reminded twice")
	}
	// A year later the next deadline is reminded again.
	now = time.Date(2027, 12, 25, 9, 0, 0, 0, time.UTC)
	if n, _ := c.Remind(context.Background()); n != 1 {
		t.Fatal("next deadline not reminded")
	}
}

func TestSeedDemoOnce(t *testing.T) {
	now := time.Date(2026, 11, 23, 9, 0, 0, 0, time.UTC)
	repo := &memRepo{items: map[string]domain.Contract{}}
	d := Deps{Repo: repo, Clock: clk{&now}, IDs: &ids{}, Location: time.UTC}
	if err := SeedDemo(admin, d); err != nil || len(repo.items) == 0 {
		t.Fatalf("%v %d", err, len(repo.items))
	}
	n := len(repo.items)
	if err := SeedDemo(admin, d); err != nil || len(repo.items) != n {
		t.Fatalf("second seed: %d → %d %v", n, len(repo.items), err)
	}
}
