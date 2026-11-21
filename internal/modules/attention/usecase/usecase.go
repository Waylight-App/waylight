package usecase

import (
	"context"
	"sync"
	"time"

	"waylight/internal/modules/attention/domain"
	"waylight/internal/shared/kernel"
)

// Widget is a widget on one of the account's own tabs.
type Widget struct {
	ID, TabID, Kind, Config string
}

// Source is a module that can say whether it needs attention: how its
// connection to an outside service is doing (Status) and what a widget of
// its kinds misses (Check). Either may be nil.
type Source struct {
	// Module is the module's name, e.g. "immich"; it is also the title the
	// clients show, translated through the widget names they know.
	Module string
	// Kinds are the widget kinds the module backs.
	Kinds []string
	// Instance reports whether the connection is the instance's (admins
	// see it even without a widget) rather than the account's.
	Instance bool
	Status   func(ctx context.Context) domain.Status
	Check    func(kind, config string) *domain.Need
}

// Deps bundles the ports of the attention list.
type Deps struct {
	// Widgets lists the widgets on the account's own tabs.
	Widgets func(ctx context.Context) ([]Widget, error)
	// MayRead reports whether the account may use a module's API; entries
	// of modules it may not use are left out.
	MayRead func(ctx context.Context, module string) bool
	Sources []Source
	Clock   kernel.Clock
}

// Attention is the input port of the module.
type Attention interface {
	// List returns what needs the account's attention, sorted (#599).
	// fresh asks the services again instead of answering from the kept
	// statuses, for a "Check again" a person just pressed.
	List(ctx context.Context, fresh bool) ([]domain.Item, error)
}

// statusTimeout bounds one status call; a source that takes longer counts
// as not answering this time.
const statusTimeout = 3 * time.Second

// statusTTL keeps a status for a while, so a badge that is polled does not
// ask every service on every request.
const statusTTL = 2 * time.Minute

// freshAfter is how old a kept status must be before a fresh list asks the
// service again, so repeated clicks on "Check again" do not hammer it.
const freshAfter = 10 * time.Second

type cached struct {
	status domain.Status
	at     time.Time
}

type attention struct {
	d      Deps
	byKind map[string]*Source

	mu        sync.Mutex
	statuses  map[string]cached
	firstSeen map[string]time.Time
}

// New returns the use case.
func New(d Deps) Attention {
	a := &attention{d: d, byKind: map[string]*Source{}, statuses: map[string]cached{}, firstSeen: map[string]time.Time{}}
	for i := range d.Sources {
		for _, k := range d.Sources[i].Kinds {
			a.byKind[k] = &d.Sources[i]
		}
	}
	return a
}

func (a *attention) List(ctx context.Context, fresh bool) ([]domain.Item, error) {
	p := kernel.PrincipalFrom(ctx)
	// Guests, kiosk displays and API tokens can change nothing, so nothing
	// needs their attention.
	if p.AccountID == "" || p.IsGuest() || p.IsKiosk() || p.IsToken() || p.IsSystem() {
		return []domain.Item{}, nil
	}
	widgets, err := a.d.Widgets(ctx)
	if err != nil {
		return nil, err
	}
	var items []domain.Item
	reported := map[string]bool{}
	for _, w := range widgets {
		src := a.byKind[w.Kind]
		if src == nil || !a.mayRead(ctx, src.Module) {
			continue
		}
		if src.Check != nil {
			if need := src.Check(w.Kind, w.Config); need != nil {
				items = append(items, domain.Item{
					ID: "widget:" + w.ID, Kind: domain.KindSetup, Integration: src.Module,
					WidgetID: w.ID, TabID: w.TabID, Title: src.Module, Reason: need.Reason,
					Action: domain.ActionChoose, Field: need.Field,
				})
				continue
			}
		}
		if src.Status == nil || reported[src.Module] {
			continue
		}
		reported[src.Module] = true
		if item, ok := a.fromStatus(ctx, p, src, fresh); ok {
			item.WidgetID, item.TabID = w.ID, w.TabID
			items = append(items, item)
		}
	}
	// Admins also learn about instance connections that broke although no
	// widget of theirs uses them; integrations never set up are not listed
	// there, that would only be noise.
	if p.IsAdmin() {
		for i := range a.d.Sources {
			src := &a.d.Sources[i]
			if !src.Instance || src.Status == nil || reported[src.Module] {
				continue
			}
			if item, ok := a.fromStatus(ctx, p, src, fresh); ok && item.Kind == domain.KindConnection {
				items = append(items, item)
			}
		}
	}
	a.stamp(p.AccountID, items)
	domain.Sort(items)
	if items == nil {
		items = []domain.Item{}
	}
	return items, nil
}

func (a *attention) mayRead(ctx context.Context, module string) bool {
	return a.d.MayRead == nil || a.d.MayRead(ctx, module)
}

func (a *attention) fromStatus(ctx context.Context, p kernel.Principal, src *Source, fresh bool) (domain.Item, bool) {
	st := a.status(ctx, p, src, fresh)
	if st.State == domain.StateOK || st.State == "" {
		return domain.Item{}, false
	}
	kind := domain.KindConnection
	if st.State == domain.StateSetup {
		kind = domain.KindSetup
	}
	return domain.Item{
		ID: "module:" + src.Module, Kind: kind, Integration: src.Module,
		Title: src.Module, Reason: st.Reason, Action: st.Action,
	}, true
}

// status asks the source with a timeout and keeps the answer for statusTTL
// (freshAfter when fresh); instance connections are shared, personal ones
// are kept per account.
func (a *attention) status(ctx context.Context, p kernel.Principal, src *Source, fresh bool) domain.Status {
	key := src.Module
	if !src.Instance {
		key += "|" + p.AccountID
	}
	now := a.d.Clock.Now()
	a.mu.Lock()
	c, ok := a.statuses[key]
	a.mu.Unlock()
	ttl := statusTTL
	if fresh {
		ttl = freshAfter
	}
	if ok && now.Sub(c.at) < ttl {
		return c.status
	}
	callCtx, cancel := context.WithTimeout(ctx, statusTimeout)
	defer cancel()
	done := make(chan domain.Status, 1)
	go func() { done <- src.Status(callCtx) }()
	var st domain.Status
	select {
	case st = <-done:
	case <-callCtx.Done():
		st = domain.Status{State: domain.StateFailing, Action: domain.ActionCheck,
			Reason: domain.Message{Text: "The service does not answer.", Template: "The service does not answer."}}
	}
	a.mu.Lock()
	a.statuses[key] = cached{status: st, at: now}
	a.mu.Unlock()
	return st
}

// stamp keeps when each entry was first seen, so "since" stays stable while
// the problem lasts and the list does not reshuffle; entries that are gone
// are forgotten.
func (a *attention) stamp(account string, items []domain.Item) {
	now := a.d.Clock.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	present := map[string]bool{}
	for i := range items {
		key := account + "|" + items[i].ID
		present[key] = true
		since, ok := a.firstSeen[key]
		if !ok {
			since = now
			a.firstSeen[key] = now
		}
		items[i].Since = since
	}
	prefix := account + "|"
	for key := range a.firstSeen {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix && !present[key] {
			delete(a.firstSeen, key)
		}
	}
}
