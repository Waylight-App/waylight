package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/attention/domain"
	"waylight/internal/shared/kernel"
)

type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

func ctxAs(role string) context.Context {
	return kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "acc-" + role, Role: role})
}

func failing(calls *int) func(context.Context) domain.Status {
	return func(context.Context) domain.Status {
		*calls++
		return domain.Status{State: domain.StateFailing, Action: domain.ActionCheck,
			Reason: domain.Message{Text: "Immich does not answer.", Template: "Immich does not answer."}}
	}
}

func TestListCollectsWidgetsAndConnections(t *testing.T) {
	c := &clock{now: time.Unix(1_800_000_000, 0)}
	var immichCalls, haCalls int
	uc := New(Deps{
		Clock: c,
		Widgets: func(context.Context) ([]Widget, error) {
			return []Widget{
				{ID: "w1", TabID: "t1", Kind: "transit", Config: `{}`},
				{ID: "w2", TabID: "t1", Kind: "immich"},
				{ID: "w3", TabID: "t2", Kind: "immich"},
				{ID: "w4", TabID: "t2", Kind: "clock"},
			}, nil
		},
		Sources: []Source{
			{Module: "transit", Kinds: []string{"transit"}, Check: func(kind, config string) *domain.Need {
				return &domain.Need{Reason: domain.Message{Template: "Choose a stop for the departures."}, Field: "stopId"}
			}},
			{Module: "immich", Kinds: []string{"immich"}, Instance: true, Status: failing(&immichCalls)},
			{Module: "smarthome", Kinds: []string{"smarthome"}, Instance: true, Status: failing(&haCalls)},
			{Module: "fritzbox", Kinds: []string{"fritzbox"}, Instance: true, Status: func(context.Context) domain.Status {
				return domain.Status{State: domain.StateSetup, Action: domain.ActionConnect}
			}},
		},
	})

	items, err := uc.List(ctxAs("member"), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("member: %+v", items)
	}
	// What stopped working comes first, then what is not set up.
	if items[0].ID != "module:immich" || items[0].Kind != domain.KindConnection || items[0].WidgetID != "w2" {
		t.Fatalf("first %+v", items[0])
	}
	if items[1].ID != "widget:w1" || items[1].Action != domain.ActionChoose || items[1].Field != "stopId" {
		t.Fatalf("second %+v", items[1])
	}
	if haCalls != 0 {
		t.Fatal("members are not told about instance connections without a widget")
	}

	// Admins also see broken instance connections without a widget, but not
	// integrations that were never set up.
	items, _ = uc.List(ctxAs("admin"), false)
	ids := map[string]bool{}
	for _, it := range items {
		ids[it.ID] = true
	}
	if !ids["module:smarthome"] || ids["module:fritzbox"] {
		t.Fatalf("admin: %+v", items)
	}
	if immichCalls != 1 {
		t.Fatalf("the status is cached between requests: %d calls", immichCalls)
	}

	// Since stays while the problem lasts.
	first := items[0].Since
	c.now = c.now.Add(time.Minute)
	items, _ = uc.List(ctxAs("admin"), false)
	if !items[0].Since.Equal(first) {
		t.Fatal("since must not move")
	}
	c.now = c.now.Add(statusTTL)
	_, _ = uc.List(ctxAs("admin"), false)
	if immichCalls != 2 {
		t.Fatalf("asked again after the TTL: %d", immichCalls)
	}

	// A fresh list asks again, but not right after the last answer.
	_, _ = uc.List(ctxAs("admin"), true)
	if immichCalls != 2 {
		t.Fatalf("fresh right after an answer: %d", immichCalls)
	}
	c.now = c.now.Add(freshAfter)
	_, _ = uc.List(ctxAs("admin"), true)
	if immichCalls != 3 {
		t.Fatalf("fresh after freshAfter: %d", immichCalls)
	}
}

func TestListIsEmptyForReadOnlyAccounts(t *testing.T) {
	uc := New(Deps{Clock: &clock{}, Widgets: func(context.Context) ([]Widget, error) {
		t.Fatal("widgets must not be read")
		return nil, nil
	}})
	for _, role := range []string{"guest", "kiosk", "token"} {
		items, err := uc.List(ctxAs(role), false)
		if err != nil || items == nil || len(items) != 0 {
			t.Fatalf("%s: %v %v", role, items, err)
		}
	}
}

func TestListRespectsModuleAccessAndErrors(t *testing.T) {
	calls := 0
	uc := New(Deps{
		Clock:   &clock{},
		MayRead: func(_ context.Context, module string) bool { return module != "immich" },
		Widgets: func(context.Context) ([]Widget, error) { return []Widget{{ID: "w", Kind: "immich"}}, nil },
		Sources: []Source{{Module: "immich", Kinds: []string{"immich"}, Status: failing(&calls)}},
	})
	if items, _ := uc.List(ctxAs("member"), false); len(items) != 0 || calls != 0 {
		t.Fatalf("a module the account may not use is left out: %+v", items)
	}
	broken := New(Deps{Clock: &clock{}, Widgets: func(context.Context) ([]Widget, error) {
		return nil, errors.New("db")
	}})
	if _, err := broken.List(ctxAs("member"), false); err == nil {
		t.Fatal("widget errors are returned")
	}
}

func TestSlowSourceCountsAsNotAnswering(t *testing.T) {
	uc := New(Deps{
		Clock:   &clock{},
		Widgets: func(context.Context) ([]Widget, error) { return []Widget{{ID: "w", Kind: "slow"}}, nil },
		Sources: []Source{{Module: "slow", Kinds: []string{"slow"}, Status: func(ctx context.Context) domain.Status {
			<-ctx.Done()
			return domain.Status{State: domain.StateOK}
		}}},
	}).(*attention)
	start := time.Now()
	items, _ := uc.List(ctxAs("member"), false)
	if len(items) != 1 || items[0].Action != domain.ActionCheck {
		t.Fatalf("%+v", items)
	}
	if time.Since(start) > statusTimeout+time.Second {
		t.Fatal("the timeout bounds the call")
	}
}
