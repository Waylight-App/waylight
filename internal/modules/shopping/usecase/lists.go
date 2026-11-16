package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/shopping/domain"
	"trisse-home/internal/shared/kernel"
)

// ListRepository stores the lists kept on trisse-home (output port, #471).
type ListRepository interface {
	Lists(ctx context.Context) ([]domain.List, error)
	Get(ctx context.Context, id string) (domain.List, error)
	CountOwn(ctx context.Context) (int, error)
	Create(ctx context.Context, l domain.List) error
	Update(ctx context.Context, id, name string, v kernel.Visibility) error
	Delete(ctx context.Context, id string) error
	Items(ctx context.Context, listID string) ([]domain.Item, error)
	AddItem(ctx context.Context, listID string, it domain.Item, max int) error
	SetDone(ctx context.Context, listID, itemID string, done bool, at time.Time) error
	DeleteItem(ctx context.Context, listID, itemID string) error
	ClearCompleted(ctx context.Context, listID string) (int, error)
	Current(ctx context.Context) (string, error)
	SetCurrent(ctx context.Context, v string) error
}

// ListsView is what the widget and the settings need.
type ListsView struct {
	Lists []ListView
	// CalDAV tells whether the account has a CalDAV list connected.
	CalDAV bool
	// Current is what the widget shows: a list ID, "caldav", or "".
	Current string
}

// ListView is a list as the principal sees it.
type ListView struct {
	domain.List
	Mine bool
}

// Export is a file for download.
type Export struct {
	Name, ContentType, Body string
}

// Lists implements the use cases of lists on trisse-home.
type Lists struct {
	lists    ListRepository
	settings SettingsRepository
	store    Store
	ids      kernel.IDGenerator
	clock    kernel.Clock
	host     string
}

// NewLists returns the use cases; host makes exported UIDs unique.
func NewLists(lists ListRepository, settings SettingsRepository, store Store, ids kernel.IDGenerator, clock kernel.Clock, host string) *Lists {
	return &Lists{lists: lists, settings: settings, store: store, ids: ids, clock: clock, host: host}
}

func (uc *Lists) view(ctx context.Context, l domain.List) ListView {
	return ListView{List: l, Mine: l.Access.Owner == kernel.PrincipalFrom(ctx).AccountID}
}

// All returns the visible lists, whether CalDAV is connected and the
// current choice.
func (uc *Lists) All(ctx context.Context) (ListsView, error) {
	list, err := uc.lists.Lists(ctx)
	if err != nil {
		return ListsView{}, err
	}
	out := ListsView{Lists: make([]ListView, 0, len(list))}
	for _, l := range list {
		out.Lists = append(out.Lists, uc.view(ctx, l))
	}
	_, err = uc.settings.Load(ctx)
	switch {
	case err == nil:
		out.CalDAV = true
	case !errors.Is(err, domain.ErrNotConfigured):
		return ListsView{}, err
	}
	if out.Current, err = uc.lists.Current(ctx); err != nil {
		return ListsView{}, err
	}
	return out, nil
}

// Create adds a list owned by the principal, private for members like a
// new note unless visibility says otherwise.
func (uc *Lists) Create(ctx context.Context, name, visibility string) (ListView, error) {
	p := kernel.PrincipalFrom(ctx)
	name, err := domain.NormaliseListName(name)
	if err != nil {
		return ListView{}, err
	}
	access := kernel.NewAccess(p)
	if visibility != "" {
		if access.Visibility, err = kernel.CheckShare(p, visibility); err != nil {
			return ListView{}, err
		}
	}
	n, err := uc.lists.CountOwn(ctx)
	if err != nil {
		return ListView{}, err
	}
	if n >= domain.MaxLists {
		return ListView{}, domain.ErrTooManyLists
	}
	l := domain.List{ID: uc.ids.NewID(), Name: name, Access: access, Created: uc.clock.Now()}
	if err := uc.lists.Create(ctx, l); err != nil {
		return ListView{}, fmt.Errorf("create list: %w", err)
	}
	return uc.view(ctx, l), nil
}

// Update renames and shares one of the principal's own lists.
func (uc *Lists) Update(ctx context.Context, id, name, visibility string) (ListView, error) {
	p := kernel.PrincipalFrom(ctx)
	l, err := uc.lists.Get(ctx, id)
	if err != nil {
		return ListView{}, err
	}
	if l.Access.Owner != p.AccountID {
		return ListView{}, domain.ErrNotYourList
	}
	if l.Name, err = domain.NormaliseListName(name); err != nil {
		return ListView{}, err
	}
	if l.Access.Visibility, err = kernel.CheckShare(p, visibility); err != nil {
		return ListView{}, err
	}
	if err := uc.lists.Update(ctx, id, l.Name, l.Access.Visibility); err != nil {
		return ListView{}, err
	}
	return uc.view(ctx, l), nil
}

// Delete removes one of the principal's own lists.
func (uc *Lists) Delete(ctx context.Context, id string) error {
	l, err := uc.lists.Get(ctx, id)
	if err != nil {
		return err
	}
	if l.Access.Owner != kernel.PrincipalFrom(ctx).AccountID {
		return domain.ErrNotYourList
	}
	return uc.lists.Delete(ctx, id)
}

// SetCurrent chooses what the widget shows: a visible list or the CalDAV
// list.
func (uc *Lists) SetCurrent(ctx context.Context, v string) error {
	if v != domain.CurrentCalDAV {
		if _, err := uc.lists.Get(ctx, v); err != nil {
			return err
		}
	}
	return uc.lists.SetCurrent(ctx, v)
}

// Items returns a visible list's items as the widget shows them.
func (uc *Lists) Items(ctx context.Context, listID string) ([]domain.Item, error) {
	items, err := uc.lists.Items(ctx, listID)
	if err != nil {
		return nil, err
	}
	if _, err := uc.lists.Get(ctx, listID); err != nil {
		return nil, err
	}
	return domain.ForDisplay(items), nil
}

// Add adds an item to a visible list.
func (uc *Lists) Add(ctx context.Context, listID, summary string) (domain.Item, error) {
	summary, err := domain.NormaliseSummary(summary)
	if err != nil {
		return domain.Item{}, err
	}
	if _, err := uc.lists.Get(ctx, listID); err != nil {
		return domain.Item{}, err
	}
	it := domain.Item{ID: uc.ids.NewID(), Summary: summary, Created: uc.clock.Now()}
	if err := uc.lists.AddItem(ctx, listID, it, domain.MaxItemsPerList); err != nil {
		return domain.Item{}, err
	}
	return it, nil
}

// SetDone ticks an item off or back on.
func (uc *Lists) SetDone(ctx context.Context, listID, itemID string, done bool) error {
	return uc.lists.SetDone(ctx, listID, itemID, done, uc.clock.Now())
}

// DeleteItem removes an item.
func (uc *Lists) DeleteItem(ctx context.Context, listID, itemID string) error {
	return uc.lists.DeleteItem(ctx, listID, itemID)
}

// ClearCompleted removes the ticked items of a visible list.
func (uc *Lists) ClearCompleted(ctx context.Context, listID string) (int, error) {
	if _, err := uc.lists.Get(ctx, listID); err != nil {
		return 0, err
	}
	return uc.lists.ClearCompleted(ctx, listID)
}

// Import copies the open items of the principal's CalDAV list into a
// visible list, skipping texts the list already has open.
func (uc *Lists) Import(ctx context.Context, listID string) (int, error) {
	if _, err := uc.lists.Get(ctx, listID); err != nil {
		return 0, err
	}
	s, err := uc.settings.Load(ctx)
	if err != nil {
		return 0, err
	}
	remote, err := uc.store.List(ctx, s)
	if err != nil {
		return 0, unavailable(err)
	}
	have, err := uc.lists.Items(ctx, listID)
	if err != nil {
		return 0, err
	}
	open := map[string]bool{}
	for _, it := range have {
		if !it.Done {
			open[it.Summary] = true
		}
	}
	n := 0
	for _, it := range remote {
		if it.Done || open[it.Summary] {
			continue
		}
		summary, err := domain.NormaliseSummary(it.Summary)
		if err != nil {
			continue
		}
		item := domain.Item{ID: uc.ids.NewID(), Summary: summary, Created: uc.clock.Now()}
		if err := uc.lists.AddItem(ctx, listID, item, domain.MaxItemsPerList); err != nil {
			if errors.Is(err, domain.ErrListFull) && n > 0 {
				return n, nil
			}
			return n, err
		}
		open[summary] = true
		n++
	}
	return n, nil
}

// Export returns a visible list as text or as an iCalendar file of VTODOs.
func (uc *Lists) Export(ctx context.Context, listID, format string) (Export, error) {
	l, err := uc.lists.Get(ctx, listID)
	if err != nil {
		return Export{}, err
	}
	items, err := uc.lists.Items(ctx, listID)
	if err != nil {
		return Export{}, err
	}
	switch format {
	case "", "text":
		return Export{Name: l.Name + ".txt", ContentType: "text/plain; charset=utf-8", Body: domain.ExportText(l.Name, items)}, nil
	case "ics":
		return Export{Name: l.Name + ".ics", ContentType: "text/calendar; charset=utf-8",
			Body: domain.ExportICS(l, items, uc.host, uc.clock.Now())}, nil
	default:
		return Export{}, kernel.NewValidationError("format", "Choose text or ics.")
	}
}
