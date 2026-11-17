package gateway

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"waylight/internal/modules/todos/domain"
	"waylight/internal/modules/todos/usecase"
	"waylight/internal/shared/kernel"
)

// TodoistAPI is the base URL of the Todoist API v1.
const TodoistAPI = "https://api.todoist.com/api/v1"

const (
	// todoistMinSync is the shortest time between two syncs of one account,
	// so several widgets on a page share one request.
	todoistMinSync = 15 * time.Second
	// todoistBackoff is the pause after 429 when Todoist names no time.
	todoistBackoff = time.Minute
	maxTodoistBody = 4 << 20
)

// TodoistProviders implements usecase.Providers for Todoist (#170). Lists
// are read through the incremental sync endpoint: the first request loads
// all active tasks, later ones only what changed since the last sync token.
// Writes use the REST endpoints. A 429 pauses requests for the time
// Todoist asks for; meanwhile the last known list is shown.
type TodoistProviders struct {
	client *http.Client
	base   string
	now    func() time.Time

	mu       sync.Mutex
	accounts map[string]*todoistAccount // by token hash
}

type todoistAccount struct {
	mu      sync.Mutex
	token   string // sync token, "" before the first sync
	tasks   map[string]todoistTask
	synced  time.Time
	retryAt time.Time
}

type todoistTask struct {
	ID          string  `json:"id"`
	Content     string  `json:"content"`
	ProjectID   string  `json:"project_id"`
	Checked     bool    `json:"checked"`
	IsDeleted   bool    `json:"is_deleted"`
	AddedAt     string  `json:"added_at"`
	CompletedAt *string `json:"completed_at"`
	UpdatedAt   string  `json:"updated_at"`
	Due         *struct {
		Date string `json:"date"`
	} `json:"due"`
}

// NewTodoistProviders returns the Todoist provider.
func NewTodoistProviders(client *http.Client) *TodoistProviders {
	return &TodoistProviders{client: client, base: TodoistAPI, now: time.Now, accounts: map[string]*todoistAccount{}}
}

func (p *TodoistProviders) account(token string) *todoistAccount {
	h := fnv.New64a()
	_, _ = h.Write([]byte(token))
	key := strconv.FormatUint(h.Sum64(), 16)
	p.mu.Lock()
	defer p.mu.Unlock()
	a := p.accounts[key]
	if a == nil {
		a = &todoistAccount{tasks: map[string]todoistTask{}}
		p.accounts[key] = a
	}
	return a
}

// Open implements usecase.Providers.
func (p *TodoistProviders) Open(s domain.Source) usecase.Repository {
	return &todoistList{p: p, s: s, acc: p.account(s.Password), read: map[string]todoistTask{}, created: map[string]string{}}
}

// Discover implements usecase.Providers: the projects of the account.
func (p *TodoistProviders) Discover(ctx context.Context, _ domain.SourceKind, _, _, token string) ([]usecase.DiscoveredList, error) {
	var out []usecase.DiscoveredList
	cursor := ""
	for range 20 { // at most 20 pages of 200 projects
		q := url.Values{"limit": {"200"}}
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		var page struct {
			Results []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				IsArchived bool   `json:"is_archived"`
				IsDeleted  bool   `json:"is_deleted"`
			} `json:"results"`
			NextCursor *string `json:"next_cursor"`
		}
		if err := p.call(ctx, nil, token, http.MethodGet, "/projects?"+q.Encode(), nil, &page); err != nil {
			return nil, err
		}
		for _, pr := range page.Results {
			if !pr.IsArchived && !pr.IsDeleted {
				out = append(out, usecase.DiscoveredList{ListID: pr.ID, Name: pr.Name})
			}
		}
		if page.NextCursor == nil || *page.NextCursor == "" {
			break
		}
		cursor = *page.NextCursor
	}
	return out, nil
}

// call sends one request; body is JSON (or url.Values for the sync
// endpoint) and dst receives the JSON answer. a, if given, records a 429.
func (p *TodoistProviders) call(ctx context.Context, a *todoistAccount, token, method, path string, body any, dst any) error {
	var r io.Reader
	contentType := ""
	switch b := body.(type) {
	case nil:
	case url.Values:
		r, contentType = strings.NewReader(b.Encode()), "application/x-www-form-urlencoded"
	default:
		raw, err := json.Marshal(b)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		r, contentType = bytes.NewReader(raw), "application/json"
	}
	req, err := http.NewRequestWithContext(ctx, method, p.base+path, r)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	res, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, maxTodoistBody))
	if err != nil {
		return fmt.Errorf("read answer: %w", err)
	}
	switch {
	case res.StatusCode == http.StatusTooManyRequests:
		if a != nil {
			a.retryAt = p.now().Add(retryAfter(res, raw))
		}
		return domain.ErrRateLimited
	case res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden:
		return fmt.Errorf("todoist: %w", kernel.ErrTargetAuth)
	case res.StatusCode == http.StatusNotFound:
		return domain.ErrTodoNotFound
	case res.StatusCode >= 300:
		return fmt.Errorf("todoist answered %s", res.Status)
	}
	if dst == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("parse answer: %w", err)
	}
	return nil
}

// retryAfter reads the pause Todoist asks for (Retry-After header or
// error_extra.retry_after), at least a few seconds and at most an hour.
func retryAfter(res *http.Response, body []byte) time.Duration {
	secs, err := strconv.Atoi(strings.TrimSpace(res.Header.Get("Retry-After")))
	if err != nil {
		var e struct {
			Extra struct {
				RetryAfter int `json:"retry_after"`
			} `json:"error_extra"`
		}
		if json.Unmarshal(body, &e) == nil {
			secs = e.Extra.RetryAfter
		}
	}
	if secs <= 0 {
		return todoistBackoff
	}
	return min(max(time.Duration(secs)*time.Second, 5*time.Second), time.Hour)
}

// todoistList is one project for the duration of a request.
type todoistList struct {
	p       *TodoistProviders
	s       domain.Source
	acc     *todoistAccount
	read    map[string]todoistTask
	created map[string]string // client ID → Todoist ID
}

func (t todoistTask) version() int {
	h := fnv.New32a()
	due := ""
	if t.Due != nil {
		due = t.Due.Date
	}
	_, _ = fmt.Fprintf(h, "%s|%s|%s|%t", t.UpdatedAt, t.Content, due, t.Checked)
	return int(h.Sum32()&0x7fffffff) | 1
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t.UTC()
}

func (t todoistTask) dueDate() time.Time {
	if t.Due == nil || len(t.Due.Date) < 10 {
		return time.Time{}
	}
	d, err := time.Parse(time.DateOnly, t.Due.Date[:10]) // a datetime keeps its day
	if err != nil {
		return time.Time{}
	}
	return d
}

func (t todoistTask) todo() domain.Todo {
	var completed time.Time
	if t.Checked {
		completed = time.Unix(0, 0).UTC()
		if t.CompletedAt != nil {
			if c := parseTime(*t.CompletedAt); !c.IsZero() {
				completed = c
			}
		}
	}
	return domain.RestoreTodo(t.ID, t.Content, t.dueDate(), 0, parseTime(t.AddedAt), completed).WithVersion(t.version())
}

// List implements usecase.Repository.
func (l *todoistList) List(ctx context.Context) ([]domain.Todo, error) {
	a := l.acc
	a.mu.Lock()
	defer a.mu.Unlock()
	now := l.p.now()
	switch {
	case now.Before(a.retryAt):
		if a.token == "" {
			return nil, domain.ErrRateLimited
		}
	case a.token != "" && now.Sub(a.synced) < todoistMinSync:
	default:
		if err := l.sync(ctx); err != nil {
			if !errors.Is(err, domain.ErrRateLimited) || a.token == "" {
				return nil, err
			}
		}
	}
	var tasks []todoistTask
	for _, t := range a.tasks {
		if t.ProjectID == l.s.ListID {
			tasks = append(tasks, t)
		}
	}
	slices.SortFunc(tasks, func(x, y todoistTask) int {
		xd, yd := x.dueDate(), y.dueDate()
		switch {
		case xd.IsZero() != yd.IsZero():
			if xd.IsZero() {
				return 1
			}
			return -1
		case !xd.Equal(yd):
			return xd.Compare(yd)
		}
		return cmp.Or(strings.Compare(x.AddedAt, y.AddedAt), strings.Compare(x.ID, y.ID))
	})
	out := make([]domain.Todo, 0, len(tasks))
	for i, t := range tasks {
		td := t.todo()
		td.MoveTo(i)
		out = append(out, td)
	}
	return out, nil
}

// sync merges the changes since the last sync token (a.mu is held).
func (l *todoistList) sync(ctx context.Context) error {
	a := l.acc
	token := a.token
	if token == "" {
		token = "*"
	}
	var res struct {
		SyncToken string        `json:"sync_token"`
		FullSync  bool          `json:"full_sync"`
		Items     []todoistTask `json:"items"`
	}
	form := url.Values{"sync_token": {token}, "resource_types": {`["items"]`}}
	if err := l.p.call(ctx, a, l.s.Password, http.MethodPost, "/sync", form, &res); err != nil {
		return err
	}
	if res.FullSync {
		a.tasks = map[string]todoistTask{}
	}
	for _, t := range res.Items {
		if t.IsDeleted {
			delete(a.tasks, t.ID)
			continue
		}
		a.tasks[t.ID] = t
	}
	a.token, a.synced = res.SyncToken, l.p.now()
	return nil
}

// changed makes the next List sync again (cheap: only the changes come).
func (l *todoistList) changed() {
	l.acc.mu.Lock()
	l.acc.synced = time.Time{}
	l.acc.mu.Unlock()
}

// Get implements usecase.Repository.
func (l *todoistList) Get(ctx context.Context, id string) (domain.Todo, error) {
	if real, ok := l.created[id]; ok {
		id = real
	}
	if !todoistID.MatchString(id) {
		return domain.Todo{}, domain.ErrTodoNotFound
	}
	var t todoistTask
	if err := l.p.call(ctx, l.acc, l.s.Password, http.MethodGet, "/tasks/"+url.PathEscape(id), nil, &t); err != nil {
		return domain.Todo{}, err
	}
	if t.ProjectID != l.s.ListID || t.IsDeleted {
		return domain.Todo{}, domain.ErrTodoNotFound
	}
	l.read[t.ID] = t
	return t.todo(), nil
}

// todoistID matches Todoist task IDs (letters, digits, "_" and "-").
var todoistID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// Save implements usecase.Repository. Only what changed is sent: a due date
// is left alone when it did not change, so a recurring Todoist due date
// keeps its rule. Completing a recurring task moves it to its next date in
// Todoist, as in the Todoist apps.
func (l *todoistList) Save(ctx context.Context, t domain.Todo) error {
	defer l.changed()
	due := func() map[string]any {
		if t.Due().IsZero() {
			return map[string]any{"due_string": "no date"}
		}
		return map[string]any{"due_date": t.Due().Format(time.DateOnly)}
	}
	if old, ok := l.read[t.ID()]; ok {
		body := map[string]any{}
		if old.Content != t.Title() {
			body["content"] = t.Title()
		}
		if !old.dueDate().Equal(t.Due()) {
			for k, v := range due() {
				body[k] = v
			}
		}
		if len(body) > 0 {
			if err := l.p.call(ctx, l.acc, l.s.Password, http.MethodPost, "/tasks/"+url.PathEscape(old.ID), body, nil); err != nil {
				return err
			}
		}
		if old.Checked != t.Done() {
			action := "/reopen"
			if t.Done() {
				action = "/close"
			}
			return l.p.call(ctx, l.acc, l.s.Password, http.MethodPost, "/tasks/"+url.PathEscape(old.ID)+action, nil, nil)
		}
		return nil
	}
	body := map[string]any{"content": t.Title(), "project_id": l.s.ListID}
	if !t.Due().IsZero() {
		body["due_date"] = t.Due().Format(time.DateOnly)
	}
	var created todoistTask
	if err := l.p.call(ctx, l.acc, l.s.Password, http.MethodPost, "/tasks", body, &created); err != nil {
		return err
	}
	if created.ID == "" {
		return errors.New("todoist: the created task has no ID")
	}
	l.created[t.ID()] = created.ID
	if t.Done() {
		return l.p.call(ctx, l.acc, l.s.Password, http.MethodPost, "/tasks/"+url.PathEscape(created.ID)+"/close", nil, nil)
	}
	return nil
}

// Delete implements usecase.Repository.
func (l *todoistList) Delete(ctx context.Context, ids ...string) error {
	defer l.changed()
	for _, id := range ids {
		err := l.p.call(ctx, l.acc, l.s.Password, http.MethodDelete, "/tasks/"+url.PathEscape(id), nil, nil)
		if err != nil && !errors.Is(err, domain.ErrTodoNotFound) {
			return err
		}
	}
	return nil
}

// Providers routes each list to the provider of its kind.
type Providers struct {
	byKind map[domain.SourceKind]usecase.Providers
}

// NewProviders returns the providers of all kinds on one HTTP client;
// tokens serves the lists that are connected by signing in (#171).
func NewProviders(client *http.Client, tokens usecase.TokenSource) Providers {
	return Providers{byKind: map[domain.SourceKind]usecase.Providers{
		domain.KindCalDAV:    NewCalDAVProviders(client),
		domain.KindTodoist:   NewTodoistProviders(client),
		domain.KindMicrosoft: NewMicrosoftProvider(client, tokens),
		domain.KindGoogle:    NewGoogleProvider(client, tokens),
		domain.KindTickTick:  NewTickTickProvider(client, tokens),
	}}
}

func (p Providers) of(kind domain.SourceKind) usecase.Providers {
	if pr, ok := p.byKind[kind]; ok {
		return pr
	}
	return p.byKind[domain.KindCalDAV]
}

// Open implements usecase.Providers.
func (p Providers) Open(s domain.Source) usecase.Repository { return p.of(s.Kind).Open(s) }

// Discover implements usecase.Providers.
func (p Providers) Discover(ctx context.Context, kind domain.SourceKind, u, user, pw string) ([]usecase.DiscoveredList, error) {
	return p.of(kind).Discover(ctx, kind, u, user, pw)
}
