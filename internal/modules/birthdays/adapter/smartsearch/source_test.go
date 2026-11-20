package smartsearch

import (
	"context"
	"testing"

	"waylight/internal/modules/birthdays/domain"
	"waylight/internal/modules/birthdays/usecase"
	"waylight/internal/shared/answers"
)

type fakeBirthdays struct {
	usecase.Birthdays
	list []domain.Upcoming
}

func (f fakeBirthdays) Upcoming(context.Context, int) ([]domain.Upcoming, error) { return f.list, nil }

func TestBirthdayAnswer(t *testing.T) {
	s := New(fakeBirthdays{list: []domain.Upcoming{
		{Birthday: domain.Birthday{Name: "Anna"}, InDays: 0, Age: 30},
		{Birthday: domain.Birthday{Name: "Ben"}, InDays: 3},
	}})
	a, err := s.Resolve(context.Background(), answers.Request{Match: answers.Match{Intent: intents[0]}, Lang: answers.German})
	if err != nil || a.Title != "Geburtstage" || len(a.Items) != 2 {
		t.Fatalf("answer = %+v, %v", a, err)
	}
	if a.Items[0].Value != "Heute · wird 30" || a.Items[0].State != answers.StateOK || a.Items[1].Value != "in 3 Tagen" {
		t.Errorf("items = %+v", a.Items)
	}
}
