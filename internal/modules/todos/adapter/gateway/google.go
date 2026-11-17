package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"waylight/internal/modules/todos/domain"
	"waylight/internal/modules/todos/usecase"
)

// GoogleTasksAPI is the base URL of the Google Tasks API.
const GoogleTasksAPI = "https://tasks.googleapis.com/tasks/v1"

// GoogleProvider implements usecase.Providers for Google Tasks (#171).
type GoogleProvider struct {
	api  *oauthAPI
	base string
}

// NewGoogleProvider returns the Google Tasks provider.
func NewGoogleProvider(client *http.Client, tokens usecase.TokenSource) *GoogleProvider {
	return &GoogleProvider{api: newOAuthAPI(client, tokens), base: GoogleTasksAPI}
}

type googleTask struct {
	ID        string `json:"id"`
	ETag      string `json:"etag"`
	Title     string `json:"title"`
	Status    string `json:"status"` // needsAction or completed
	Due       string `json:"due"`    // RFC 3339; only the date is used by Google
	Completed string `json:"completed"`
	Updated   string `json:"updated"`
	Deleted   bool   `json:"deleted"`
}

func (t googleTask) dueDate() time.Time {
	d, err := time.Parse(time.RFC3339Nano, t.Due)
	if err != nil {
		return time.Time{}
	}
	y, m, dd := d.UTC().Date()
	return time.Date(y, m, dd, 0, 0, 0, 0, time.UTC)
}

func (t googleTask) todo() domain.Todo {
	var done time.Time
	if t.Status == "completed" {
		c, err := time.Parse(time.RFC3339Nano, t.Completed)
		if done = c.UTC(); err != nil || done.IsZero() {
			done = time.Unix(0, 0).UTC()
		}
	}
	// Google has no creation time; the last update orders undated tasks.
	updated, _ := time.Parse(time.RFC3339Nano, t.Updated)
	return domain.RestoreTodo(t.ID, t.Title, t.dueDate(), 0, updated.UTC(), done).WithVersion(hashVersion(t.ETag, t.Updated, t.Title, t.Status))
}

// Open implements usecase.Providers.
func (p *GoogleProvider) Open(s domain.Source) usecase.Repository {
	return &googleList{p: p, s: s, read: map[string]googleTask{}, created: map[string]string{}}
}

// Discover implements usecase.Providers: the task lists of the account.
func (p *GoogleProvider) Discover(ctx context.Context, _ domain.SourceKind, _, _, token string) ([]usecase.DiscoveredList, error) {
	var out []usecase.DiscoveredList
	pageToken := ""
	for range 20 {
		q := url.Values{"maxResults": {"100"}}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		var page struct {
			Items []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"items"`
			Next string `json:"nextPageToken"`
		}
		if _, err := p.api.do(ctx, domain.Source{}, token, http.MethodGet, p.base+"/users/@me/lists?"+q.Encode(), nil, &page, nil); err != nil {
			return nil, err
		}
		for _, l := range page.Items {
			out = append(out, usecase.DiscoveredList{ListID: l.ID, Name: l.Title})
		}
		if pageToken = page.Next; pageToken == "" {
			break
		}
	}
	return out, nil
}

type googleList struct {
	p       *GoogleProvider
	s       domain.Source
	read    map[string]googleTask
	created map[string]string
}

func (l *googleList) tasksURL() string {
	return l.p.base + "/lists/" + url.PathEscape(l.s.ListID) + "/tasks"
}

func (l *googleList) List(ctx context.Context) ([]domain.Todo, error) {
	if todos, ok := l.p.api.cached(l.s.ID); ok {
		return todos, nil
	}
	var todos []domain.Todo
	pageToken := ""
	for range 50 {
		q := url.Values{"maxResults": {"100"}, "showCompleted": {"true"}, "showHidden": {"true"}}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		var page struct {
			Items []googleTask `json:"items"`
			Next  string       `json:"nextPageToken"`
		}
		if _, err := l.p.api.do(ctx, l.s, "", http.MethodGet, l.tasksURL()+"?"+q.Encode(), nil, &page, nil); err != nil {
			return nil, err
		}
		for _, t := range page.Items {
			if !t.Deleted {
				todos = append(todos, t.todo())
			}
		}
		if pageToken = page.Next; pageToken == "" {
			break
		}
	}
	todos = sortTodos(todos)
	l.p.api.store(l.s.ID, todos)
	return todos, nil
}

func (l *googleList) Get(ctx context.Context, id string) (domain.Todo, error) {
	if real, ok := l.created[id]; ok {
		id = real
	}
	var t googleTask
	if _, err := l.p.api.do(ctx, l.s, "", http.MethodGet, l.tasksURL()+"/"+url.PathEscape(id), nil, &t, nil); err != nil {
		return domain.Todo{}, err
	}
	if t.Deleted {
		return domain.Todo{}, domain.ErrTodoNotFound
	}
	l.read[t.ID] = t
	return t.todo(), nil
}

func googleFields(t domain.Todo) map[string]any {
	body := map[string]any{"title": t.Title(), "status": "needsAction", "completed": nil, "due": nil}
	if t.Done() {
		body["status"] = "completed"
		delete(body, "completed") // Google sets the completion time
	}
	if !t.Due().IsZero() {
		body["due"] = t.Due().Format("2006-01-02") + "T00:00:00.000Z"
	}
	return body
}

// Save implements usecase.Repository (PATCH sends only these fields).
func (l *googleList) Save(ctx context.Context, t domain.Todo) error {
	defer l.p.api.forget(l.s.ID)
	if old, ok := l.read[t.ID()]; ok {
		body := googleFields(t)
		if old.Status == "completed" && t.Done() {
			delete(body, "status") // keep Google's completion time
		}
		_, err := l.p.api.do(ctx, l.s, "", http.MethodPatch, l.tasksURL()+"/"+url.PathEscape(old.ID), body, nil, nil)
		return err
	}
	var created googleTask
	if _, err := l.p.api.do(ctx, l.s, "", http.MethodPost, l.tasksURL(), googleFields(t), &created, nil); err != nil {
		return err
	}
	l.created[t.ID()] = created.ID
	return nil
}

func (l *googleList) Delete(ctx context.Context, ids ...string) error {
	defer l.p.api.forget(l.s.ID)
	for _, id := range ids {
		if _, err := l.p.api.do(ctx, l.s, "", http.MethodDelete, l.tasksURL()+"/"+url.PathEscape(id), nil, nil, nil); err != nil && !errors.Is(err, domain.ErrTodoNotFound) {
			return err
		}
	}
	return nil
}
