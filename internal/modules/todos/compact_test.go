package todos

import (
	"net/http"
	"testing"
	"time"

	"waylight/internal/modules/todos/usecase"
)

func TestUrgentTodoAction(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 11, d, 0, 0, 0, 0, time.UTC) }
	if a := urgentTodoAction([]usecase.TodoView{{ID: "x", Done: true}}); a != nil {
		t.Fatalf("all done: %+v", a)
	}
	list := []usecase.TodoView{
		{ID: "plain", Title: "Plain", Position: 0},
		{ID: "later", Title: "Later", Due: day(25), Position: 1},
		{ID: "soon", Title: "Soon", Due: day(22), Position: 2, Version: 3},
	}
	a := urgentTodoAction(list)
	if a == nil || a.Path != "/api/todos/soon" || a.Method != http.MethodPut || a.Body["due"] != "2026-11-22" || a.Body["version"] != 3 || a.Body["done"] != true {
		t.Fatalf("earliest due: %+v", a)
	}
	list = append(list, usecase.TodoView{ID: "late", Title: "Late", Due: day(30), Overdue: true, Position: 3})
	if a = urgentTodoAction(list); a.Path != "/api/todos/late" || a.Label.String() != "Done: Late" {
		t.Fatalf("overdue first: %+v", a)
	}
}
