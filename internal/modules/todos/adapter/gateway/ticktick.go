package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"waylight/internal/modules/todos/domain"
	"waylight/internal/modules/todos/usecase"
)

// TickTickAPI is the base URL of the TickTick Open API.
const TickTickAPI = "https://api.ticktick.com/open/v1"

// TickTickProvider implements usecase.Providers for TickTick (#171). The
// Open API lists only open tasks of a project, so completed tasks leave the
// list.
type TickTickProvider struct {
	api  *oauthAPI
	base string
}

// NewTickTickProvider returns the TickTick provider.
func NewTickTickProvider(client *http.Client, tokens usecase.TokenSource) *TickTickProvider {
	return &TickTickProvider{api: newOAuthAPI(client, tokens), base: TickTickAPI}
}

type tickTask struct {
	ID            string `json:"id"`
	ProjectID     string `json:"projectId"`
	Title         string `json:"title"`
	Status        int    `json:"status"` // 0 open, 2 completed
	DueDate       string `json:"dueDate"`
	IsAllDay      bool   `json:"isAllDay"`
	TimeZone      string `json:"timeZone"`
	CompletedTime string `json:"completedTime"`
	SortOrder     int64  `json:"sortOrder"`
	ETag          string `json:"etag"`
}

// tickLayout is TickTick's date format, e.g. 2026-11-30T00:00:00.000+0000.
const tickLayout = "2006-01-02T15:04:05.000-0700"

func (t tickTask) dueDate() time.Time {
	d, err := time.Parse(tickLayout, t.DueDate)
	if err != nil {
		return time.Time{}
	}
	if t.TimeZone != "" {
		if loc, err := time.LoadLocation(t.TimeZone); err == nil {
			d = d.In(loc)
		}
	}
	y, m, dd := d.Date()
	return time.Date(y, m, dd, 0, 0, 0, 0, time.UTC)
}

func (t tickTask) todo() domain.Todo {
	var done time.Time
	if t.Status == 2 {
		c, err := time.Parse(tickLayout, t.CompletedTime)
		if done = c.UTC(); err != nil || done.IsZero() {
			done = time.Unix(0, 0).UTC()
		}
	}
	return domain.RestoreTodo(t.ID, t.Title, t.dueDate(), 0, time.Time{}, done).
		WithVersion(hashVersion(t.ETag, t.Title, t.DueDate, strconv.Itoa(t.Status)))
}

// Open implements usecase.Providers.
func (p *TickTickProvider) Open(s domain.Source) usecase.Repository {
	return &tickList{p: p, s: s, read: map[string]tickTask{}, created: map[string]string{}}
}

// Discover implements usecase.Providers: the open projects of the account.
func (p *TickTickProvider) Discover(ctx context.Context, _ domain.SourceKind, _, _, token string) ([]usecase.DiscoveredList, error) {
	var projects []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Closed bool   `json:"closed"`
	}
	if _, err := p.api.do(ctx, domain.Source{}, token, http.MethodGet, p.base+"/project", nil, &projects, nil); err != nil {
		return nil, err
	}
	var out []usecase.DiscoveredList
	for _, pr := range projects {
		if !pr.Closed {
			out = append(out, usecase.DiscoveredList{ListID: pr.ID, Name: pr.Name})
		}
	}
	return out, nil
}

type tickList struct {
	p       *TickTickProvider
	s       domain.Source
	read    map[string]tickTask
	created map[string]string
}

func (l *tickList) projectURL() string { return l.p.base + "/project/" + url.PathEscape(l.s.ListID) }

func (l *tickList) List(ctx context.Context) ([]domain.Todo, error) {
	if todos, ok := l.p.api.cached(l.s.ID); ok {
		return todos, nil
	}
	var data struct {
		Tasks []tickTask `json:"tasks"`
	}
	if _, err := l.p.api.do(ctx, l.s, "", http.MethodGet, l.projectURL()+"/data", nil, &data, nil); err != nil {
		return nil, err
	}
	todos := make([]domain.Todo, 0, len(data.Tasks))
	for _, t := range data.Tasks {
		todos = append(todos, t.todo())
	}
	todos = sortTodos(todos)
	l.p.api.store(l.s.ID, todos)
	return todos, nil
}

func (l *tickList) Get(ctx context.Context, id string) (domain.Todo, error) {
	if real, ok := l.created[id]; ok {
		id = real
	}
	var t tickTask
	if _, err := l.p.api.do(ctx, l.s, "", http.MethodGet, l.projectURL()+"/task/"+url.PathEscape(id), nil, &t, nil); err != nil {
		return domain.Todo{}, err
	}
	if t.ID == "" || t.ProjectID != l.s.ListID {
		return domain.Todo{}, domain.ErrTodoNotFound
	}
	l.read[t.ID] = t
	return t.todo(), nil
}

func tickFields(t domain.Todo, projectID string) map[string]any {
	body := map[string]any{"title": t.Title(), "projectId": projectID}
	if !t.Due().IsZero() {
		body["dueDate"] = t.Due().Format("2006-01-02") + "T00:00:00.000+0000"
		body["isAllDay"] = true
		body["timeZone"] = "UTC"
	}
	return body
}

// Save implements usecase.Repository. TickTick completes through its own
// endpoint and reopens through an update with status 0.
func (l *tickList) Save(ctx context.Context, t domain.Todo) error {
	defer l.p.api.forget(l.s.ID)
	if old, ok := l.read[t.ID()]; ok {
		body := tickFields(t, l.s.ListID)
		body["id"] = old.ID
		if old.dueDate().Equal(t.Due()) {
			delete(body, "dueDate") // unchanged: keep a time or rule set in TickTick
			delete(body, "isAllDay")
			delete(body, "timeZone")
		} else if t.Due().IsZero() {
			body["dueDate"] = nil
		}
		if old.Status == 2 && !t.Done() {
			body["status"] = 0
		}
		if _, err := l.p.api.do(ctx, l.s, "", http.MethodPost, l.p.base+"/task/"+url.PathEscape(old.ID), body, nil, nil); err != nil {
			return err
		}
		if old.Status != 2 && t.Done() {
			return l.complete(ctx, old.ID)
		}
		return nil
	}
	var created tickTask
	if _, err := l.p.api.do(ctx, l.s, "", http.MethodPost, l.p.base+"/task", tickFields(t, l.s.ListID), &created, nil); err != nil {
		return err
	}
	if created.ID == "" {
		return errors.New("ticktick: the created task has no ID")
	}
	l.created[t.ID()] = created.ID
	if t.Done() {
		return l.complete(ctx, created.ID)
	}
	return nil
}

func (l *tickList) complete(ctx context.Context, id string) error {
	_, err := l.p.api.do(ctx, l.s, "", http.MethodPost, l.projectURL()+"/task/"+url.PathEscape(id)+"/complete", nil, nil, nil)
	return err
}

func (l *tickList) Delete(ctx context.Context, ids ...string) error {
	defer l.p.api.forget(l.s.ID)
	for _, id := range ids {
		if _, err := l.p.api.do(ctx, l.s, "", http.MethodDelete, l.projectURL()+"/task/"+url.PathEscape(id), nil, nil, nil); err != nil && !errors.Is(err, domain.ErrTodoNotFound) {
			return err
		}
	}
	return nil
}
