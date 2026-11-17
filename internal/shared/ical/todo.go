package ical

import (
	"strings"
	"time"
)

// Todo is one VTODO.
type Todo struct {
	UID       string
	Summary   string
	Done      bool
	Completed time.Time
	Created   time.Time
	// Due is the due calendar date at midnight UTC (zero = none); a DUE with
	// a time counts on its day in the server's time zone.
	Due time.Time
}

// ParseTodos extracts the VTODOs of an iCalendar object.
func ParseTodos(data string) []Todo {
	var out []Todo
	var props map[string]icalProperty
	depth := 0
	for _, line := range unfold(data) {
		switch {
		case line == "BEGIN:VTODO":
			props, depth = map[string]icalProperty{}, 0
		case strings.HasPrefix(line, "BEGIN:") && props != nil:
			depth++
		case strings.HasPrefix(line, "END:") && props != nil && depth > 0:
			depth--
		case line == "END:VTODO" && props != nil:
			t := Todo{UID: props["UID"].value, Summary: strings.TrimSpace(textUnescaper.Replace(props["SUMMARY"].value))}
			status := strings.ToUpper(props["STATUS"].value)
			t.Done = status == "COMPLETED" || props["COMPLETED"].value != ""
			if status == "CANCELLED" {
				props = nil
				continue
			}
			if c, _, err := parseDateTime(props["COMPLETED"], time.UTC); err == nil {
				t.Completed = c
			}
			if c, _, err := parseDateTime(props["CREATED"], time.UTC); err == nil {
				t.Created = c
			}
			t.Due = dueDate(props["DUE"])
			out = append(out, t)
			props = nil
		case props != nil && depth == 0:
			if p, ok := parseLine(line); ok {
				if _, seen := props[p.name]; !seen {
					props[p.name] = p
				}
			}
		}
	}
	return out
}

// dueDate reads a DUE property as a calendar date (zero if absent or invalid).
func dueDate(p icalProperty) time.Time {
	if p.value == "" {
		return time.Time{}
	}
	t, allDay, err := parseDateTime(p, time.Local)
	if err != nil {
		return time.Time{}
	}
	if !allDay {
		t = t.In(time.Local)
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// NewTodo serialises a new VTODO.
func NewTodo(uid, summary string, now time.Time) string {
	return NewTodoDue(uid, summary, time.Time{}, now)
}

// NewTodoDue serialises a new VTODO with a due date (zero = none).
func NewTodoDue(uid, summary string, due, now time.Time) string {
	stamp := now.UTC().Format("20060102T150405Z")
	lines := []string{
		"BEGIN:VCALENDAR", "VERSION:2.0", "PRODID:-//waylight//EN",
		"BEGIN:VTODO", "UID:" + uid, "DTSTAMP:" + stamp, "CREATED:" + stamp, "LAST-MODIFIED:" + stamp,
		"SUMMARY:" + EscapeText(summary), "STATUS:NEEDS-ACTION",
	}
	if !due.IsZero() {
		lines = append(lines, "DUE;VALUE=DATE:"+due.Format("20060102"))
	}
	lines = append(lines, "END:VTODO", "END:VCALENDAR")
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(FoldLine(l))
	}
	return b.String()
}

// SetTodoDone rewrites the completion properties of the first VTODO in raw
// and keeps every other property (also those of other apps) unchanged.
func SetTodoDone(raw string, done bool, now time.Time) string {
	stamp := now.UTC().Format("20060102T150405Z")
	replace := map[string]bool{"STATUS": true, "COMPLETED": true, "PERCENT-COMPLETE": true, "LAST-MODIFIED": true, "DTSTAMP": true}
	add := []string{"DTSTAMP:" + stamp, "LAST-MODIFIED:" + stamp}
	if done {
		add = append(add, "STATUS:COMPLETED", "COMPLETED:"+stamp, "PERCENT-COMPLETE:100")
	} else {
		add = append(add, "STATUS:NEEDS-ACTION")
	}
	return rewriteTodo(raw, replace, add)
}

// TodoEdit is the new state of a VTODO's text, due date and completion.
type TodoEdit struct {
	Summary string
	Due     time.Time // calendar date; zero removes DUE
	Done    bool
}

// EditTodo rewrites summary, due date and completion of the first VTODO in
// raw. Every other property (also those of other apps, alarms included)
// stays as it is; so do SUMMARY, DUE and COMPLETED when their value does
// not change, which keeps a due time or the original completion time.
func EditTodo(raw string, e TodoEdit, now time.Time) string {
	stamp := now.UTC().Format("20060102T150405Z")
	var old Todo
	if todos := ParseTodos(raw); len(todos) > 0 {
		old = todos[0]
	}
	replace := map[string]bool{"LAST-MODIFIED": true, "DTSTAMP": true}
	var add []string
	if old.Summary != strings.TrimSpace(e.Summary) {
		replace["SUMMARY"] = true
		add = append(add, "SUMMARY:"+EscapeText(e.Summary))
	}
	if !old.Due.Equal(e.Due) {
		replace["DUE"] = true
		if !e.Due.IsZero() {
			add = append(add, "DUE;VALUE=DATE:"+e.Due.Format("20060102"))
		}
	}
	if old.Done != e.Done {
		for _, n := range []string{"STATUS", "COMPLETED", "PERCENT-COMPLETE"} {
			replace[n] = true
		}
		if e.Done {
			add = append(add, "STATUS:COMPLETED", "COMPLETED:"+stamp, "PERCENT-COMPLETE:100")
		} else {
			add = append(add, "STATUS:NEEDS-ACTION")
		}
	}
	add = append([]string{"DTSTAMP:" + stamp, "LAST-MODIFIED:" + stamp}, add...)
	return rewriteTodo(raw, replace, add)
}

// rewriteTodo drops the properties named in replace from the first VTODO
// of raw and appends add before its END line.
func rewriteTodo(raw string, replace map[string]bool, add []string) string {
	var b strings.Builder
	inTodo, depth, handled := false, 0, false
	for _, line := range unfold(raw) {
		if line == "" {
			continue
		}
		switch {
		case line == "BEGIN:VTODO" && !handled:
			inTodo, depth = true, 0
		case inTodo && strings.HasPrefix(line, "BEGIN:"):
			depth++
		case inTodo && depth > 0 && strings.HasPrefix(line, "END:"):
			depth--
		case inTodo && depth == 0 && line == "END:VTODO":
			for _, l := range add {
				b.WriteString(FoldLine(l))
			}
			inTodo, handled = false, true
		case inTodo && depth == 0:
			if p, ok := parseLine(line); ok && replace[p.name] {
				continue
			}
		}
		b.WriteString(FoldLine(line))
	}
	return b.String()
}
