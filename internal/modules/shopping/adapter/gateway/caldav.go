package gateway

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"trisse-home/internal/modules/shopping/domain"
	"trisse-home/internal/shared/caldav"
	"trisse-home/internal/shared/ical"
)

// CalDAVStore implements usecase.Store on the shared CalDAV client.
type CalDAVStore struct {
	client caldav.Client
	now    func() time.Time
}

// NewCalDAVStore returns a store.
func NewCalDAVStore(client *http.Client) *CalDAVStore {
	return &CalDAVStore{client: caldav.Client{HTTP: client}, now: time.Now}
}

func account(s domain.Settings) caldav.Account {
	return caldav.Account{URL: s.ListURL, Username: s.Username, Password: s.Password}
}

// domainError maps CalDAV errors to the module's errors.
func domainError(err error) error {
	switch {
	case errors.Is(err, caldav.ErrPreconditionFailed):
		return domain.ErrConflict
	case errors.Is(err, caldav.ErrNotFound):
		return domain.ErrItemNotFound
	}
	return err
}

// List implements usecase.Store.
func (c *CalDAVStore) List(ctx context.Context, s domain.Settings) ([]domain.Item, error) {
	objs, err := c.client.Todos(ctx, account(s))
	if err != nil {
		return nil, domainError(err)
	}
	var items []domain.Item
	for _, o := range objs {
		if !domain.ValidItemID(o.Name) {
			continue
		}
		for _, t := range ical.ParseTodos(o.Data) {
			items = append(items, domain.Item{ID: o.Name, Summary: t.Summary, Done: t.Done, Created: t.Created, Completed: t.Completed})
		}
	}
	return items, nil
}

// Add implements usecase.Store; If-None-Match never overwrites an object.
func (c *CalDAVStore) Add(ctx context.Context, s domain.Settings, uid, summary string) (domain.Item, error) {
	id := strings.ReplaceAll(uid, "@", "-") + ".ics" // the client escapes it in the URL
	now := c.now()
	if _, err := c.client.Put(ctx, account(s), id, ical.NewTodo(uid, summary, now), "", true); err != nil {
		return domain.Item{}, domainError(err)
	}
	return domain.Item{ID: id, Summary: summary, Created: now}, nil
}

// SetDone implements usecase.Store: the object is read, only its completion
// properties are changed and it is written back with If-Match, so changes
// made meanwhile by other apps are never overwritten.
func (c *CalDAVStore) SetDone(ctx context.Context, s domain.Settings, id string, done bool) error {
	o, err := c.client.Get(ctx, account(s), id)
	if err != nil {
		return domainError(err)
	}
	_, err = c.client.Put(ctx, account(s), id, ical.SetTodoDone(o.Data, done, c.now()), o.ETag, false)
	return domainError(err)
}

// Delete implements usecase.Store.
func (c *CalDAVStore) Delete(ctx context.Context, s domain.Settings, id string) error {
	return domainError(c.client.Delete(ctx, account(s), id))
}
