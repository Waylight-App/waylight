package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/shared/kernel"
)

// fakeTodoist answers like the Todoist API v1 (field names and shapes from
// recorded responses, anonymised): /sync with sync tokens, /projects with
// cursor pages, /tasks REST endpoints.
type fakeTodoist struct {
	mu       sync.Mutex
	tasks    map[string]map[string]any
	changed  map[string]bool // task IDs changed since the last sync token
	seq      int
	calls    []string
	tooMany  int // answer this many requests with 429
	lastSync string
}

func newFakeTodoist(t *testing.T) (*fakeTodoist, *TodoistProviders) {
	f := &fakeTodoist{tasks: map[string]map[string]any{}, changed: map[string]bool{}}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	p := NewTodoistProviders(srv.Client())
	p.base = srv.URL + "/api/v1"
	return f, p
}

func (f *fakeTodoist) add(id, content, project, due string, checked bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.put(map[string]any{"id": id, "content": content, "project_id": project, "checked": checked, "is_deleted": false,
		"added_at": "2026-11-01T08:00:00.000000Z", "updated_at": "2026-11-01T08:00:00.000000Z", "completed_at": nil,
		"due": dueObj(due), "priority": 1, "child_order": 1, "labels": []string{}})
}

func dueObj(date string) any {
	if date == "" {
		return nil
	}
	return map[string]any{"date": date, "string": date, "lang": "en", "is_recurring": false, "timezone": nil}
}

func (f *fakeTodoist) put(t map[string]any) {
	f.seq++
	t["updated_at"] = fmt.Sprintf("2026-11-14T09:00:%02d.000000Z", f.seq%60)
	f.tasks[t["id"].(string)] = t
	f.changed[t["id"].(string)] = true
}

func (f *fakeTodoist) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, r.Method+" "+r.URL.Path)
	if r.Header.Get("Authorization") != "Bearer tok" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if f.tooMany > 0 {
		f.tooMany--
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"error":"Too many requests","error_code":35,"error_extra":{"retry_after":30},"http_code":429}`)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1")
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodGet && path == "/projects":
		if r.URL.Query().Get("cursor") == "" {
			_ = json.NewEncoder(w).Encode(map[string]any{"results": []any{
				map[string]any{"id": "6Jf8VQXxpwv56VQ7", "name": "Inbox", "is_archived": false, "is_deleted": false},
				map[string]any{"id": "old", "name": "Old", "is_archived": true, "is_deleted": false},
			}, "next_cursor": "c2"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"results": []any{map[string]any{"id": "6X7rM8997g3RQmvh", "name": "Haushalt"}}, "next_cursor": nil})
	case r.Method == http.MethodPost && path == "/sync":
		_ = r.ParseForm()
		full := r.Form.Get("sync_token") == "*"
		var items []any
		for id, t := range f.tasks {
			if full && (t["checked"] == true || t["is_deleted"] == true) {
				continue // a full sync has only active tasks
			}
			if full || f.changed[id] {
				items = append(items, t)
			}
		}
		f.changed = map[string]bool{}
		f.lastSync = r.Form.Get("sync_token")
		_ = json.NewEncoder(w).Encode(map[string]any{"sync_token": fmt.Sprintf("tok%d", f.seq), "full_sync": full, "items": items})
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/tasks/"):
		t, ok := f.tasks[strings.TrimPrefix(path, "/tasks/")]
		if !ok || t["is_deleted"] == true {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(t)
	case r.Method == http.MethodPost && path == "/tasks":
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		id := fmt.Sprintf("n%d", f.seq+1)
		due, _ := in["due_date"].(string)
		t := map[string]any{"id": id, "content": in["content"], "project_id": in["project_id"], "checked": false, "is_deleted": false,
			"added_at": "2026-11-14T09:00:00.000000Z", "completed_at": nil, "due": dueObj(due)}
		f.put(t)
		_ = json.NewEncoder(w).Encode(t)
	case r.Method == http.MethodPost && strings.HasSuffix(path, "/close"), r.Method == http.MethodPost && strings.HasSuffix(path, "/reopen"):
		id := strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(path, "/tasks/"), "/close"), "/reopen")
		t := f.tasks[id]
		t["checked"] = strings.HasSuffix(path, "/close")
		t["completed_at"] = nil
		if t["checked"] == true {
			t["completed_at"] = "2026-11-14T10:00:00.000000Z"
		}
		f.put(t)
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodPost && strings.HasPrefix(path, "/tasks/"):
		t, ok := f.tasks[strings.TrimPrefix(path, "/tasks/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		if c, ok := in["content"]; ok {
			t["content"] = c
		}
		if d, ok := in["due_date"].(string); ok {
			t["due"] = dueObj(d)
		}
		if in["due_string"] == "no date" {
			t["due"] = nil
		}
		f.put(t)
		_ = json.NewEncoder(w).Encode(t)
	case r.Method == http.MethodDelete && strings.HasPrefix(path, "/tasks/"):
		if t, ok := f.tasks[strings.TrimPrefix(path, "/tasks/")]; ok {
			t["is_deleted"] = true
			f.put(t)
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *fakeTodoist) count(prefix string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if strings.HasPrefix(c, prefix) {
			n++
		}
	}
	return n
}

func todoistSource() domain.Source {
	return domain.Source{ID: "td", Kind: domain.KindTodoist, Password: "tok", ListID: "inbox"}
}

func TestTodoistDiscover(t *testing.T) {
	_, p := newFakeTodoist(t)
	got, err := p.Discover(context.Background(), domain.KindTodoist, "", "", "tok")
	if err != nil || len(got) != 2 || got[0].ListID != "6Jf8VQXxpwv56VQ7" || got[1].Name != "Haushalt" {
		t.Fatalf("archived projects skipped, pages followed: %+v %v", got, err)
	}
	if _, err := p.Discover(context.Background(), domain.KindTodoist, "", "", "wrong"); !errors.Is(err, kernel.ErrTargetAuth) {
		t.Fatal(err)
	}
}

func TestTodoistIncrementalSync(t *testing.T) {
	ctx := context.Background()
	f, p := newFakeTodoist(t)
	now := time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC)
	p.now = func() time.Time { return now }
	f.add("a", "Steuer", "inbox", "2026-11-30", false)
	f.add("b", "Später", "inbox", "", false)
	f.add("x", "Other project", "work", "", false)

	todos, err := p.Open(todoistSource()).List(ctx)
	if err != nil || len(todos) != 2 || todos[0].Title() != "Steuer" || !todos[0].Due().Equal(time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)) || todos[1].Title() != "Später" {
		t.Fatalf("%+v %v", todos, err)
	}
	// Within 15 s a second widget gets the same data without a request.
	if _, _ = p.Open(todoistSource()).List(ctx); f.count("POST /api/v1/sync") != 1 {
		t.Fatal("synced twice within the minimum interval")
	}
	// Later: only the changes come (incremental sync token).
	now = now.Add(time.Minute)
	f.mu.Lock()
	f.tasks["b"]["content"] = "Gleich"
	f.put(f.tasks["b"])
	f.tasks["a"]["is_deleted"] = true
	f.put(f.tasks["a"])
	f.mu.Unlock()
	todos, _ = p.Open(todoistSource()).List(ctx)
	if f.lastSync == "*" || len(todos) != 1 || todos[0].Title() != "Gleich" {
		t.Fatalf("incremental sync: token %q, %+v", f.lastSync, todos)
	}
	// 429: the last list is shown and Todoist is left alone until retry_after has passed.
	now = now.Add(time.Minute)
	f.mu.Lock()
	f.tooMany = 1
	f.mu.Unlock()
	if todos, err = p.Open(todoistSource()).List(ctx); err != nil || len(todos) != 1 {
		t.Fatalf("rate limited: %+v %v", todos, err)
	}
	syncs := f.count("POST /api/v1/sync")
	now = now.Add(20 * time.Second)
	_, _ = p.Open(todoistSource()).List(ctx)
	if f.count("POST /api/v1/sync") != syncs {
		t.Fatal("asked again before retry_after")
	}
	now = now.Add(15 * time.Second)
	if _, _ = p.Open(todoistSource()).List(ctx); f.count("POST /api/v1/sync") != syncs+1 {
		t.Fatal("did not resume after retry_after")
	}
	// Without any cached list a 429 is an error the user can read.
	_, p2 := newFakeTodoist(t)
	p2.base = p.base
	f.mu.Lock()
	f.tooMany = 1
	f.mu.Unlock()
	src := todoistSource()
	src.Password = "tok" // same token, but a fresh provider has no cache
	if _, err := p2.Open(src).List(ctx); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatal(err)
	}
}

func TestTodoistWrites(t *testing.T) {
	ctx := context.Background()
	f, p := newFakeTodoist(t)
	f.add("a", "Steuer", "inbox", "2026-11-30", false)
	f.add("x", "Other project", "work", "", false)
	l := p.Open(todoistSource())

	if _, err := l.Get(ctx, "x"); !errors.Is(err, domain.ErrTodoNotFound) {
		t.Fatal("a task of another project must not be reachable", err)
	}
	got, err := l.Get(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	// Only the title changes: the due date (maybe recurring in Todoist) is not sent.
	_ = got.Edit("Steuer erklären", got.Due())
	got.SetDone(true, time.Now())
	if err := l.Save(ctx, got); err != nil {
		t.Fatal(err)
	}
	if f.tasks["a"]["content"] != "Steuer erklären" || f.tasks["a"]["checked"] != true || f.count("POST /api/v1/tasks/a/close") != 1 {
		t.Fatalf("%v %v", f.tasks["a"], f.calls)
	}
	if fresh, _ := p.Open(todoistSource()).Get(ctx, "a"); fresh.Version() == got.Version() || !fresh.Done() {
		t.Fatal("the version must change with the task")
	}
	// Reopen and remove the date.
	l2 := p.Open(todoistSource())
	again, _ := l2.Get(ctx, "a")
	_ = again.Edit(again.Title(), time.Time{})
	again.SetDone(false, time.Now())
	if err := l2.Save(ctx, again); err != nil || f.tasks["a"]["due"] != nil || f.tasks["a"]["checked"] != false {
		t.Fatalf("%v %v", f.tasks["a"], err)
	}
	// Create with a client ID; the read-back finds the Todoist ID.
	n, _ := domain.NewTodo("0b7e9d2c-4f1a-4c8e-9a51-2d3c4b5a6f70", "Brief", time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), 0, time.Now())
	l3 := p.Open(todoistSource())
	if err := l3.Save(ctx, n); err != nil {
		t.Fatal(err)
	}
	created, err := l3.Get(ctx, n.ID())
	if err != nil || created.Title() != "Brief" || created.ID() == n.ID() || created.Due().Format(time.DateOnly) != "2026-12-01" {
		t.Fatalf("%+v %v", created, err)
	}
	if err := l3.Delete(ctx, created.ID()); err != nil || f.tasks[created.ID()]["is_deleted"] != true {
		t.Fatal(err)
	}
	if err := l3.Delete(ctx, "gone"); err != nil {
		t.Fatal("deleting a missing task is fine", err)
	}
	bad := todoistSource()
	bad.Password = "wrong"
	if _, err := p.Open(bad).List(ctx); !errors.Is(err, kernel.ErrTargetAuth) {
		t.Fatal(err)
	}
}

func TestProvidersRouteByKind(t *testing.T) {
	p := NewProviders(http.DefaultClient, nil)
	if _, ok := p.Open(todoistSource()).(*todoistList); !ok {
		t.Fatal("todoist")
	}
	if _, ok := p.Open(domain.Source{Kind: domain.KindCalDAV}).(*calDAVList); !ok {
		t.Fatal("caldav")
	}
}
