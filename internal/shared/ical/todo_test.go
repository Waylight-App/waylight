package ical

import (
	"strings"
	"testing"
	"time"
)

func TestTodos(t *testing.T) {
	now := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	raw := NewTodo("u1", "Milch, 2×", now)
	todos := ParseTodos(raw)
	if len(todos) != 1 || todos[0].Summary != "Milch, 2×" || todos[0].Done || !todos[0].Created.Equal(now) {
		t.Fatalf("%+v", todos)
	}
	// Other apps' properties survive and VALARM sub-components are untouched.
	foreign := strings.Replace(raw, "END:VTODO", "X-APPLE-SORT-ORDER:42\r\nPRIORITY:1\r\nBEGIN:VALARM\r\nACTION:DISPLAY\r\nEND:VALARM\r\nEND:VTODO", 1)
	done := SetTodoDone(foreign, true, now.Add(time.Hour))
	if got := ParseTodos(done); len(got) != 1 || !got[0].Done || !got[0].Completed.Equal(now.Add(time.Hour)) {
		t.Fatalf("done: %+v", got)
	}
	for _, keep := range []string{"X-APPLE-SORT-ORDER:42", "PRIORITY:1", "BEGIN:VALARM", "SUMMARY:Milch\\, 2×"} {
		if !strings.Contains(done, keep) {
			t.Fatalf("lost %q:\n%s", keep, done)
		}
	}
	if strings.Count(done, "STATUS:") != 1 {
		t.Fatalf("duplicate STATUS:\n%s", done)
	}
	undone := SetTodoDone(done, false, now)
	if got := ParseTodos(undone); got[0].Done || strings.Contains(undone, "COMPLETED:") {
		t.Fatalf("undone: %s", undone)
	}
	cancelled := strings.Replace(raw, "STATUS:NEEDS-ACTION", "STATUS:CANCELLED", 1)
	if len(ParseTodos(cancelled)) != 0 {
		t.Fatal("cancelled todo listed")
	}
}

func TestTodoDueAndEdit(t *testing.T) {
	now := time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC)
	due := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	raw := NewTodoDue("u2", "Steuer", due, now)
	if got := ParseTodos(raw); len(got) != 1 || !got[0].Due.Equal(due) || got[0].UID != "u2" {
		t.Fatalf("%+v\n%s", got, raw)
	}
	// A due time from another app is kept as long as the date stays the same.
	timed := strings.Replace(raw, "DUE;VALUE=DATE:20261201", "DUE:20261201T120000Z\r\nX-OTHER:1", 1)
	same := EditTodo(timed, TodoEdit{Summary: "Steuer erklären", Due: due}, now)
	if !strings.Contains(same, "DUE:20261201T120000Z") || !strings.Contains(same, "X-OTHER:1") {
		t.Fatalf("lost the due time or a foreign property:\n%s", same)
	}
	if got := ParseTodos(same); got[0].Summary != "Steuer erklären" || !got[0].Due.Equal(due) || got[0].Done {
		t.Fatalf("%+v", got)
	}
	moved := EditTodo(same, TodoEdit{Summary: "Steuer erklären", Due: due.AddDate(0, 0, 3), Done: true}, now)
	got := ParseTodos(moved)
	if !got[0].Due.Equal(due.AddDate(0, 0, 3)) || !got[0].Done || !got[0].Completed.Equal(now) || strings.Count(moved, "DUE") != 1 {
		t.Fatalf("%+v\n%s", got, moved)
	}
	// Editing a done to-do keeps its completion time; removing the date drops DUE.
	later := EditTodo(moved, TodoEdit{Summary: "Steuer", Done: true}, now.Add(time.Hour))
	if got := ParseTodos(later); !got[0].Completed.Equal(now) || !got[0].Due.IsZero() || strings.Contains(later, "DUE") {
		t.Fatalf("%+v\n%s", got, later)
	}
}
