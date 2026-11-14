package gateway

import (
	"cmp"
	"context"
	"errors"
	"hash/fnv"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/modules/todos/usecase"
	"trisse-home/internal/shared/caldav"
	"trisse-home/internal/shared/ical"
)

// CalDAVProviders implements usecase.Providers for CalDAV task lists
// (Nextcloud Tasks, Radicale, Baïkal, Synology, Fastmail …). It keeps the
// parsed objects of each list in memory: a refresh asks for the list's
// CTag first and, when it changed, for the ETags, and then downloads only
// the objects whose ETag changed (calendar-multiget).
type CalDAVProviders struct {
	client caldav.Client
	now    func() time.Time

	mu    sync.Mutex
	cache map[string]*listCache
}

type listCache struct {
	ctag    string
	objects map[string]cachedObject
}

type cachedObject struct {
	etag string
	todo ical.Todo
	ok   bool // the object holds a usable VTODO
}

// NewCalDAVProviders returns the CalDAV provider.
func NewCalDAVProviders(client *http.Client) *CalDAVProviders {
	return &CalDAVProviders{client: caldav.Client{HTTP: client}, now: time.Now, cache: map[string]*listCache{}}
}

// Open implements usecase.Providers.
func (p *CalDAVProviders) Open(s domain.Source) usecase.Repository {
	return &calDAVList{p: p, s: s, acc: caldav.Account{URL: s.URL, Username: s.Username, Password: s.Password},
		key: s.ID + "\x00" + s.URL + "\x00" + s.Username, read: map[string]caldav.Object{}}
}

// Discover implements usecase.Providers.
func (p *CalDAVProviders) Discover(ctx context.Context, kind domain.SourceKind, url, username, password string) ([]usecase.DiscoveredList, error) {
	found, err := p.client.Discover(ctx, caldav.Account{URL: url, Username: username, Password: password})
	if err != nil {
		return nil, err
	}
	out := make([]usecase.DiscoveredList, 0, len(found))
	for _, c := range found {
		out = append(out, usecase.DiscoveredList{URL: c.URL, Name: c.Name})
	}
	return out, nil
}

// calDAVList is one list for the duration of a request.
type calDAVList struct {
	p    *CalDAVProviders
	s    domain.Source
	acc  caldav.Account
	key  string
	read map[string]caldav.Object // objects read by Get, for conditional writes
}

// version derives the to-do version from the ETag, so a client that edited
// an older state gets domain.ErrVersionConflict ("server wins").
func version(etag string) int {
	if etag == "" {
		return 1
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(etag))
	return int(h.Sum32()&0x7fffffff) | 1
}

func toTodo(name, etag string, t ical.Todo) domain.Todo {
	return domain.RestoreTodo(name, t.Summary, t.Due, 0, t.Created, t.Completed).WithVersion(version(etag))
}

func parse(o caldav.Object) cachedObject {
	todos := ical.ParseTodos(o.Data)
	if len(todos) == 0 {
		return cachedObject{etag: o.ETag}
	}
	t := todos[0]
	if t.Done && t.Completed.IsZero() {
		t.Completed = time.Unix(0, 0).UTC() // done without a date: keep it done
	}
	return cachedObject{etag: o.ETag, todo: t, ok: true}
}

// objectName maps the ID of a new to-do (a client-generated ID) to its
// object name.
func objectName(id string) string {
	if strings.HasSuffix(strings.ToLower(id), ".ics") {
		return id
	}
	return id + ".ics"
}

// List implements usecase.Repository.
func (l *calDAVList) List(ctx context.Context) ([]domain.Todo, error) {
	ctag, err := l.p.client.CTag(ctx, l.acc)
	if err != nil {
		return nil, err
	}
	l.p.mu.Lock()
	c := l.p.cache[l.key]
	fresh := c != nil && ctag != "" && c.ctag == ctag
	l.p.mu.Unlock()
	if !fresh {
		if c, err = l.refresh(ctx, c, ctag); err != nil {
			return nil, err
		}
	}
	return l.todos(c), nil
}

// refresh loads the changed objects of the list into a new cache entry.
func (l *calDAVList) refresh(ctx context.Context, old *listCache, ctag string) (*listCache, error) {
	next := &listCache{ctag: ctag, objects: map[string]cachedObject{}}
	etags, err := l.p.client.ETags(ctx, l.acc)
	if err != nil {
		return nil, err
	}
	var changed []string
	for _, o := range etags {
		if old != nil {
			if c, ok := old.objects[o.Name]; ok && c.etag == o.ETag && o.ETag != "" {
				next.objects[o.Name] = c
				continue
			}
		}
		changed = append(changed, o.Name)
	}
	if len(changed) > 0 {
		objs, err := l.p.client.Multiget(ctx, l.acc, changed)
		if err != nil || len(objs) == 0 {
			// Some servers lack calendar-multiget: load everything once.
			if objs, err = l.p.client.Todos(ctx, l.acc); err != nil {
				return nil, err
			}
		}
		for _, o := range objs {
			next.objects[o.Name] = parse(o)
		}
	}
	l.p.mu.Lock()
	l.p.cache[l.key] = next
	l.p.mu.Unlock()
	return next, nil
}

// todos returns the cached to-dos, open ones by due date (none last) and
// creation; positions follow that order.
func (l *calDAVList) todos(c *listCache) []domain.Todo {
	type entry struct {
		name string
		o    cachedObject
	}
	var entries []entry
	for name, o := range c.objects {
		if o.ok {
			entries = append(entries, entry{name, o})
		}
	}
	slices.SortFunc(entries, func(a, b entry) int {
		ad, bd := a.o.todo.Due, b.o.todo.Due
		switch {
		case ad.IsZero() != bd.IsZero():
			if ad.IsZero() {
				return 1
			}
			return -1
		case !ad.Equal(bd):
			return ad.Compare(bd)
		}
		return cmp.Or(a.o.todo.Created.Compare(b.o.todo.Created), strings.Compare(a.name, b.name))
	})
	out := make([]domain.Todo, 0, len(entries))
	for i, e := range entries {
		t := toTodo(e.name, e.o.etag, e.o.todo)
		t.MoveTo(i)
		out = append(out, t)
	}
	return out
}

// invalidate makes the next List compare ETags again.
func (l *calDAVList) invalidate() {
	l.p.mu.Lock()
	if c := l.p.cache[l.key]; c != nil {
		c2 := *c
		c2.ctag = ""
		l.p.cache[l.key] = &c2
	}
	l.p.mu.Unlock()
}

func domainError(err error) error {
	switch {
	case errors.Is(err, caldav.ErrNotFound):
		return domain.ErrTodoNotFound
	case errors.Is(err, caldav.ErrPreconditionFailed):
		return domain.ErrVersionConflict
	}
	return err
}

// Get implements usecase.Repository.
// The ID is the object name; a client-generated ID of a to-do created here
// lacks ".ics", so that name is tried as well.
func (l *calDAVList) Get(ctx context.Context, id string) (domain.Todo, error) {
	name := id
	o, err := l.p.client.Get(ctx, l.acc, name)
	if errors.Is(err, caldav.ErrNotFound) && name != objectName(id) {
		name = objectName(id)
		o, err = l.p.client.Get(ctx, l.acc, name)
	}
	if err != nil {
		return domain.Todo{}, domainError(err)
	}
	c := parse(o)
	if !c.ok {
		return domain.Todo{}, domain.ErrTodoNotFound
	}
	l.read[name] = o
	return toTodo(name, o.ETag, c.todo), nil
}

// Save implements usecase.Repository: a to-do read by Get is edited in
// place (other apps' properties stay) and written with If-Match; a new one
// is created with If-None-Match.
func (l *calDAVList) Save(ctx context.Context, t domain.Todo) error {
	defer l.invalidate()
	now := l.p.now()
	if o, ok := l.read[t.ID()]; ok {
		name := t.ID()
		data := ical.EditTodo(o.Data, ical.TodoEdit{Summary: t.Title(), Due: t.Due(), Done: t.Done()}, now)
		_, err := l.p.client.Put(ctx, l.acc, name, data, o.ETag, false)
		return domainError(err)
	}
	name := objectName(t.ID())
	uid := strings.TrimSuffix(name, ".ics") + "@trisse-home"
	data := ical.NewTodoDue(uid, t.Title(), t.Due(), now)
	if t.Done() {
		data = ical.SetTodoDone(data, true, now)
	}
	_, err := l.p.client.Put(ctx, l.acc, name, data, "", true)
	return domainError(err)
}

// Delete implements usecase.Repository.
func (l *calDAVList) Delete(ctx context.Context, ids ...string) error {
	defer l.invalidate()
	for _, id := range ids {
		if err := l.p.client.Delete(ctx, l.acc, id); err != nil {
			return domainError(err)
		}
	}
	return nil
}
