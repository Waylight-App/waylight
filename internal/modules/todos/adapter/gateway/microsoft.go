package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"waylight/internal/modules/todos/domain"
	"waylight/internal/modules/todos/usecase"
)

// GraphAPI is the base URL of Microsoft Graph.
const GraphAPI = "https://graph.microsoft.com/v1.0"

// MicrosoftProvider implements usecase.Providers for Microsoft To Do via
// Microsoft Graph (#171).
type MicrosoftProvider struct {
	api  *oauthAPI
	base string
}

// NewMicrosoftProvider returns the Microsoft To Do provider.
func NewMicrosoftProvider(client *http.Client, tokens usecase.TokenSource) *MicrosoftProvider {
	return &MicrosoftProvider{api: newOAuthAPI(client, tokens), base: GraphAPI}
}

type graphDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type graphTask struct {
	ID          string         `json:"id"`
	ETag        string         `json:"@odata.etag"`
	Title       string         `json:"title"`
	Status      string         `json:"status"`
	Created     string         `json:"createdDateTime"`
	Modified    string         `json:"lastModifiedDateTime"`
	Due         *graphDateTime `json:"dueDateTime"`
	CompletedAt *graphDateTime `json:"completedDateTime"`
}

func graphTime(d *graphDateTime) time.Time {
	if d == nil || d.DateTime == "" {
		return time.Time{}
	}
	s := strings.TrimSuffix(d.DateTime, "Z")
	if i := strings.IndexByte(s, '.'); i >= 0 {
		s = s[:i]
	}
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func (t graphTask) todo() domain.Todo {
	var due time.Time
	if d := graphTime(t.Due); !d.IsZero() {
		y, m, dd := d.Date()
		due = time.Date(y, m, dd, 0, 0, 0, 0, time.UTC)
	}
	var done time.Time
	if t.Status == "completed" {
		if done = graphTime(t.CompletedAt); done.IsZero() {
			done = time.Unix(0, 0).UTC()
		}
	}
	created, _ := time.Parse(time.RFC3339Nano, t.Created)
	v := hashVersion(t.ETag, t.Modified, t.Title, t.Status)
	return domain.RestoreTodo(t.ID, t.Title, due, 0, created.UTC(), done).WithVersion(v)
}

// Open implements usecase.Providers.
func (p *MicrosoftProvider) Open(s domain.Source) usecase.Repository {
	return &graphList{p: p, s: s, read: map[string]graphTask{}, created: map[string]string{}}
}

// Discover implements usecase.Providers: the To Do lists of the account.
func (p *MicrosoftProvider) Discover(ctx context.Context, _ domain.SourceKind, _, _, token string) ([]usecase.DiscoveredList, error) {
	var out []usecase.DiscoveredList
	next := p.base + "/me/todo/lists"
	for range 20 {
		var page struct {
			Value []struct {
				ID          string `json:"id"`
				DisplayName string `json:"displayName"`
			} `json:"value"`
			Next string `json:"@odata.nextLink"`
		}
		if _, err := p.api.do(ctx, domain.Source{}, token, http.MethodGet, next, nil, &page, nil); err != nil {
			return nil, err
		}
		for _, l := range page.Value {
			out = append(out, usecase.DiscoveredList{ListID: l.ID, Name: l.DisplayName})
		}
		if page.Next == "" || !strings.HasPrefix(page.Next, p.base+"/") {
			break
		}
		next = page.Next
	}
	return out, nil
}

type graphList struct {
	p       *MicrosoftProvider
	s       domain.Source
	read    map[string]graphTask
	created map[string]string // client ID → Graph ID
}

func (l *graphList) tasksURL() string {
	return l.p.base + "/me/todo/lists/" + url.PathEscape(l.s.ListID) + "/tasks"
}

func (l *graphList) List(ctx context.Context) ([]domain.Todo, error) {
	if todos, ok := l.p.api.cached(l.s.ID); ok {
		return todos, nil
	}
	var todos []domain.Todo
	next := l.tasksURL() + "?$top=100"
	for range 50 {
		var page struct {
			Value []graphTask `json:"value"`
			Next  string      `json:"@odata.nextLink"`
		}
		if _, err := l.p.api.do(ctx, l.s, "", http.MethodGet, next, nil, &page, nil); err != nil {
			return nil, err
		}
		for _, t := range page.Value {
			todos = append(todos, t.todo())
		}
		if page.Next == "" || !strings.HasPrefix(page.Next, l.p.base+"/") {
			break
		}
		next = page.Next
	}
	todos = sortTodos(todos)
	l.p.api.store(l.s.ID, todos)
	return todos, nil
}

func (l *graphList) Get(ctx context.Context, id string) (domain.Todo, error) {
	if real, ok := l.created[id]; ok {
		id = real
	}
	var t graphTask
	if _, err := l.p.api.do(ctx, l.s, "", http.MethodGet, l.tasksURL()+"/"+url.PathEscape(id), nil, &t, nil); err != nil {
		return domain.Todo{}, err
	}
	l.read[t.ID] = t
	return t.todo(), nil
}

func graphFields(t domain.Todo) map[string]any {
	body := map[string]any{"title": t.Title(), "status": "notStarted", "dueDateTime": nil}
	if t.Done() {
		body["status"] = "completed"
	}
	if !t.Due().IsZero() {
		body["dueDateTime"] = graphDateTime{DateTime: t.Due().Format("2006-01-02") + "T00:00:00.0000000", TimeZone: "UTC"}
	}
	return body
}

// Save implements usecase.Repository; edits use If-Match with the ETag
// read before, so a change made in To Do meanwhile wins.
func (l *graphList) Save(ctx context.Context, t domain.Todo) error {
	defer l.p.api.forget(l.s.ID)
	if old, ok := l.read[t.ID()]; ok {
		body := graphFields(t)
		if old.Due != nil && graphTime(old.Due).Format("2006-01-02") == t.Due().Format("2006-01-02") {
			delete(body, "dueDateTime") // unchanged: keep a due time or reminder set in To Do
		}
		var hdr map[string]string
		if old.ETag != "" {
			hdr = map[string]string{"If-Match": old.ETag}
		}
		_, err := l.p.api.do(ctx, l.s, "", http.MethodPatch, l.tasksURL()+"/"+url.PathEscape(old.ID), body, nil, hdr)
		return err
	}
	var created graphTask
	if _, err := l.p.api.do(ctx, l.s, "", http.MethodPost, l.tasksURL(), graphFields(t), &created, nil); err != nil {
		return err
	}
	l.created[t.ID()] = created.ID // the read-back finds it under the client ID
	return nil
}

func (l *graphList) Delete(ctx context.Context, ids ...string) error {
	defer l.p.api.forget(l.s.ID)
	for _, id := range ids {
		if _, err := l.p.api.do(ctx, l.s, "", http.MethodDelete, l.tasksURL()+"/"+url.PathEscape(id), nil, nil, nil); err != nil && !errors.Is(err, domain.ErrTodoNotFound) {
			return err
		}
	}
	return nil
}
