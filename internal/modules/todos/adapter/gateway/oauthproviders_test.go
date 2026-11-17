package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"waylight/internal/modules/todos/domain"
	"waylight/internal/platform/database/dbtest"
)

type staticTokens struct{ token string }

func (s staticTokens) AccessToken(context.Context, domain.Source) (string, error) {
	return s.token, nil
}

// fakeAPI is a tiny JSON REST server: handlers by "METHOD /path" prefix,
// with the bearer token checked and every call recorded.
type fakeAPI struct {
	mu     sync.Mutex
	calls  []string
	bodies []map[string]any
	routes func(w http.ResponseWriter, r *http.Request, body map[string]any)
}

func (f *fakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, r.Method+" "+r.URL.Path)
	if r.Header.Get("Authorization") != "Bearer tok" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.bodies = append(f.bodies, body)
	w.Header().Set("Content-Type", "application/json")
	f.routes(w, r, body)
}

func (f *fakeAPI) last() map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.bodies[len(f.bodies)-1]
}

func (f *fakeAPI) count(call string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if c == call {
			n++
		}
	}
	return n
}

func serve(t *testing.T, f *fakeAPI) string {
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	return srv.URL
}

func write(w http.ResponseWriter, v any) { _ = json.NewEncoder(w).Encode(v) }

func TestMicrosoftToDo(t *testing.T) {
	ctx := context.Background()
	task := map[string]any{"@odata.etag": `W/"e1"`, "id": "AAMkT1=", "title": "Steuer", "status": "notStarted",
		"createdDateTime": "2026-11-01T08:00:00.0000000Z", "lastModifiedDateTime": "2026-11-01T08:00:00.0000000Z",
		"dueDateTime": map[string]any{"dateTime": "2026-11-30T00:00:00.0000000", "timeZone": "UTC"}}
	f := &fakeAPI{}
	f.routes = func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		switch {
		case r.URL.Path == "/me/todo/lists":
			write(w, map[string]any{"value": []any{map[string]any{"id": "L1", "displayName": "Aufgaben", "wellknownListName": "defaultList"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/me/todo/lists/L1/tasks":
			write(w, map[string]any{"value": []any{task}})
		case r.Method == http.MethodGet && r.URL.Path == "/me/todo/lists/L1/tasks/AAMkT1=":
			write(w, task)
		case r.Method == http.MethodGet && r.URL.Path == "/me/todo/lists/L1/tasks/NEW":
			write(w, map[string]any{"id": "NEW", "title": "Brief", "status": "notStarted"})
		case r.Method == http.MethodPatch:
			if r.Header.Get("If-Match") != task["@odata.etag"] {
				w.WriteHeader(http.StatusPreconditionFailed)
				return
			}
			write(w, task)
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			write(w, map[string]any{"id": "NEW", "title": body["title"], "status": "notStarted"})
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
	p := NewMicrosoftProvider(http.DefaultClient, staticTokens{"tok"})
	p.base = serve(t, f)
	lists, err := p.Discover(ctx, domain.KindMicrosoft, "", "", "tok")
	if err != nil || len(lists) != 1 || lists[0].ListID != "L1" || lists[0].Name != "Aufgaben" {
		t.Fatalf("%+v %v", lists, err)
	}
	s := domain.Source{ID: "ms", Kind: domain.KindMicrosoft, ListID: "L1"}
	todos, err := p.Open(s).List(ctx)
	if err != nil || len(todos) != 1 || todos[0].Title() != "Steuer" || todos[0].Due().Format(time.DateOnly) != "2026-11-30" {
		t.Fatalf("%+v %v", todos, err)
	}
	if _, _ = p.Open(s).List(ctx); f.count("GET /me/todo/lists/L1/tasks") != 1 {
		t.Fatal("a second widget within 30 s must use the cache")
	}
	l := p.Open(s)
	got, _ := l.Get(ctx, "AAMkT1=")
	_ = got.Edit("Steuer erklären", got.Due())
	got.SetDone(true, time.Now())
	if err := l.Save(ctx, got); err != nil {
		t.Fatal(err)
	}
	if b := f.last(); b["title"] != "Steuer erklären" || b["status"] != "completed" {
		t.Fatalf("patch: %v", b)
	} else if _, sent := b["dueDateTime"]; sent {
		t.Fatal("an unchanged due date must not be sent (keeps a time set in To Do)")
	}
	// Changed in To Do meanwhile: If-Match fails, the change is refused.
	task["@odata.etag"] = `W/"e2"`
	if err := l.Save(ctx, got); !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatal(err)
	}
	n, _ := domain.NewTodo("0b7e9d2c-4f1a-4c8e-9a51-2d3c4b5a6f70", "Brief", time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), 0, time.Now())
	if err := l.Save(ctx, n); err != nil {
		t.Fatal(err)
	}
	if due := f.last()["dueDateTime"].(map[string]any); due["dateTime"] != "2026-12-01T00:00:00.0000000" || due["timeZone"] != "UTC" {
		t.Fatal(due)
	}
	if c, err := l.Get(ctx, n.ID()); err != nil || c.ID() != "NEW" {
		t.Fatalf("read-back by client ID: %+v %v", c, err)
	}
	if err := l.Delete(ctx, "NEW"); err != nil {
		t.Fatal(err)
	}
	bad := NewMicrosoftProvider(http.DefaultClient, staticTokens{"expired"})
	bad.base = p.base
	if _, err := bad.Open(domain.Source{ID: "x", ListID: "L1"}).List(ctx); !errors.Is(err, domain.ErrSignInAgain) {
		t.Fatal(err)
	}
}

func TestGoogleTasks(t *testing.T) {
	ctx := context.Background()
	f := &fakeAPI{}
	f.routes = func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		switch {
		case r.URL.Path == "/users/@me/lists":
			if r.URL.Query().Get("pageToken") == "" {
				write(w, map[string]any{"kind": "tasks#taskLists", "items": []any{map[string]any{"id": "MDE", "title": "Meine Aufgaben"}}, "nextPageToken": "p2"})
				return
			}
			write(w, map[string]any{"items": []any{map[string]any{"id": "MDI", "title": "Einkauf"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/lists/MDE/tasks":
			if r.URL.Query().Get("showCompleted") != "true" {
				t.Error("completed tasks must be asked for")
			}
			write(w, map[string]any{"items": []any{
				map[string]any{"id": "t1", "etag": `"x1"`, "title": "Steuer", "status": "needsAction", "due": "2026-11-30T00:00:00.000Z", "updated": "2026-11-01T08:00:00.000Z"},
				map[string]any{"id": "t2", "etag": `"x2"`, "title": "Erledigt", "status": "completed", "completed": "2026-11-10T08:00:00.000Z", "updated": "2026-11-10T08:00:00.000Z"},
				map[string]any{"id": "t3", "title": "Weg", "deleted": true},
			}})
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/lists/MDE/tasks/"):
			write(w, map[string]any{"id": strings.TrimPrefix(r.URL.Path, "/lists/MDE/tasks/"), "etag": `"x1"`, "title": "Steuer", "status": "needsAction", "due": "2026-11-30T00:00:00.000Z"})
		case r.Method == http.MethodPatch, r.Method == http.MethodPost:
			write(w, map[string]any{"id": "t9", "title": body["title"]})
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
	p := NewGoogleProvider(http.DefaultClient, staticTokens{"tok"})
	p.base = serve(t, f)
	lists, err := p.Discover(ctx, domain.KindGoogle, "", "", "tok")
	if err != nil || len(lists) != 2 || lists[1].Name != "Einkauf" {
		t.Fatalf("pages followed: %+v %v", lists, err)
	}
	s := domain.Source{ID: "g", Kind: domain.KindGoogle, ListID: "MDE"}
	todos, err := p.Open(s).List(ctx)
	if err != nil || len(todos) != 2 || todos[0].Title() != "Steuer" || !todos[1].Done() || todos[1].CompletedAt().Day() != 10 {
		t.Fatalf("%+v %v", todos, err)
	}
	l := p.Open(s)
	got, _ := l.Get(ctx, "t1")
	_ = got.Edit("Steuer", time.Time{})
	got.SetDone(true, time.Now())
	if err := l.Save(ctx, got); err != nil {
		t.Fatal(err)
	}
	if b := f.last(); b["status"] != "completed" || b["due"] != nil {
		t.Fatalf("patch: %v", b)
	}
	n, _ := domain.NewTodo("0b7e9d2c-4f1a-4c8e-9a51-2d3c4b5a6f70", "Brief", time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), 0, time.Now())
	if err := l.Save(ctx, n); err != nil || f.last()["due"] != "2026-12-01T00:00:00.000Z" {
		t.Fatal(err, f.last())
	}
	if c, err := l.Get(ctx, n.ID()); err != nil || c.ID() != "t9" {
		t.Fatalf("%+v %v", c, err)
	}
	if err := l.Delete(ctx, "t9"); err != nil {
		t.Fatal(err)
	}
}

func TestTickTick(t *testing.T) {
	ctx := context.Background()
	f := &fakeAPI{}
	f.routes = func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		switch {
		case r.URL.Path == "/project":
			write(w, []any{
				map[string]any{"id": "p1", "name": "Privat", "closed": false},
				map[string]any{"id": "p0", "name": "Archiv", "closed": true},
			})
		case r.URL.Path == "/project/p1/data":
			write(w, map[string]any{"project": map[string]any{"id": "p1"}, "tasks": []any{
				map[string]any{"id": "k1", "projectId": "p1", "title": "Steuer", "status": 0, "dueDate": "2026-11-29T23:00:00.000+0000", "isAllDay": true, "timeZone": "Europe/Berlin", "etag": "abc"},
			}})
		case r.Method == http.MethodGet && r.URL.Path == "/project/p1/task/k1":
			write(w, map[string]any{"id": "k1", "projectId": "p1", "title": "Steuer", "status": 0, "dueDate": "2026-11-29T23:00:00.000+0000", "timeZone": "Europe/Berlin"})
		case r.Method == http.MethodGet && r.URL.Path == "/project/p1/task/other":
			write(w, map[string]any{"id": "other", "projectId": "p2", "title": "Fremd"})
		case r.Method == http.MethodGet && r.URL.Path == "/project/p1/task/k9":
			write(w, map[string]any{"id": "k9", "projectId": "p1", "title": "Brief"})
		case r.Method == http.MethodPost && r.URL.Path == "/task":
			write(w, map[string]any{"id": "k9", "projectId": "p1", "title": body["title"]})
		case r.Method == http.MethodPost:
			write(w, map[string]any{})
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		}
	}
	p := NewTickTickProvider(http.DefaultClient, staticTokens{"tok"})
	p.base = serve(t, f)
	lists, err := p.Discover(ctx, domain.KindTickTick, "", "", "tok")
	if err != nil || len(lists) != 1 || lists[0].ListID != "p1" {
		t.Fatalf("closed projects skipped: %+v %v", lists, err)
	}
	s := domain.Source{ID: "tt", Kind: domain.KindTickTick, ListID: "p1"}
	todos, err := p.Open(s).List(ctx)
	if err != nil || len(todos) != 1 || todos[0].Due().Format(time.DateOnly) != "2026-11-30" {
		t.Fatalf("all-day date in its time zone: %+v %v", todos, err)
	}
	l := p.Open(s)
	if _, err := l.Get(ctx, "other"); !errors.Is(err, domain.ErrTodoNotFound) {
		t.Fatal("a task of another project", err)
	}
	got, _ := l.Get(ctx, "k1")
	_ = got.Edit("Steuer erklären", got.Due())
	got.SetDone(true, time.Now())
	if err := l.Save(ctx, got); err != nil {
		t.Fatal(err)
	}
	if f.count("POST /task/k1") != 1 || f.count("POST /project/p1/task/k1/complete") != 1 {
		t.Fatal(f.calls)
	}
	n, _ := domain.NewTodo("0b7e9d2c-4f1a-4c8e-9a51-2d3c4b5a6f70", "Brief", time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), 0, time.Now())
	if err := l.Save(ctx, n); err != nil {
		t.Fatal(err)
	}
	if b := f.last(); b["dueDate"] != "2026-12-01T00:00:00.000+0000" || b["isAllDay"] != true || b["projectId"] != "p1" {
		t.Fatal(b)
	}
	if c, err := l.Get(ctx, n.ID()); err != nil || c.ID() != "k9" {
		t.Fatalf("%+v %v", c, err)
	}
	if err := l.Delete(ctx, "k9"); err != nil || f.count("DELETE /project/p1/task/k9") != 1 {
		t.Fatal(err)
	}
}

func TestOAuthGrantStoredEncrypted(t *testing.T) {
	ctx := context.Background()
	repo := NewSourceRepository(dbtest.Open(t, "todos", Migrations()), prefixCipher{})
	exp := time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC)
	s := domain.Source{ID: "ms", Name: "To Do", Kind: domain.KindMicrosoft, Username: "client", ListID: "AAMk=",
		Grant: domain.OAuthGrant{ClientSecret: "cs", Access: "a1", Refresh: "r1", Expiry: exp}}
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	var enc string
	_ = repo.db.QueryRowContext(ctx, `SELECT password_encrypted FROM todos_sources`).Scan(&enc)
	if !strings.HasPrefix(enc, "enc:") || !strings.Contains(enc, `"refresh":"r1"`) {
		t.Fatal(enc)
	}
	got, err := repo.Get(ctx, "ms")
	if err != nil || got.Grant != s.Grant || got.Password != "" {
		t.Fatalf("%+v %v", got, err)
	}
}
