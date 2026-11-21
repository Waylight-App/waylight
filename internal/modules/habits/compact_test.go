package habits

import (
	"testing"
	"time"

	"waylight/internal/modules/habits/domain"
	"waylight/internal/modules/habits/usecase"
)

func TestNextHabitAction(t *testing.T) {
	today := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)
	view := func(id string, due, done bool) usecase.HabitView {
		return usecase.HabitView{
			Habit: domain.Habit{ID: id, Name: "Habit " + id},
			Stats: domain.Stats{DoneToday: done, Heatmap: []domain.HeatCell{{Day: today, Due: due, Done: done}}},
		}
	}
	if a := nextHabitAction([]usecase.HabitView{view("a", true, true), view("b", false, false)}); a != nil {
		t.Fatalf("nothing open and due: %+v", a)
	}
	a := nextHabitAction([]usecase.HabitView{view("a", true, true), view("b", false, false), view("c", true, false)})
	if a == nil || a.Path != "/api/habits/c/toggle" || a.Method != "POST" || a.Body["date"] != "2026-11-21" || a.Label.String() != "Done: Habit c" {
		t.Fatalf("next habit: %+v", a)
	}
}
