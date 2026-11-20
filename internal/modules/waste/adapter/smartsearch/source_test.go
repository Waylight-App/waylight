package smartsearch

import (
	"context"
	"testing"

	"waylight/internal/modules/waste/domain"
	"waylight/internal/modules/waste/usecase"
	"waylight/internal/shared/answers"
)

type fakeWaste struct {
	usecase.Waste
	o usecase.Overview
}

func (f fakeWaste) Upcoming(context.Context) (usecase.Overview, error) { return f.o, nil }

func TestWasteAnswer(t *testing.T) {
	s := New(fakeWaste{o: usecase.Overview{Types: []domain.Upcoming{
		{Style: domain.Style{Label: "Gelber Sack", Icon: "♻️"}, DaysUntil: 1},
		{Style: domain.Style{Label: "Restmüll"}, DaysUntil: 5},
	}}})
	a, _ := s.Resolve(context.Background(), answers.Request{Match: answers.Match{Intent: intents[0]}, Lang: answers.German})
	if a.Title != "Müllabfuhr" || a.Items[0].Label != "♻️ Gelber Sack" || a.Items[0].Value != "Morgen" ||
		a.Items[0].State != answers.StateWarning || a.Items[1].Value != "in 5 Tagen" || a.Items[1].State != "" {
		t.Errorf("answer = %+v", a)
	}
	none, _ := New(fakeWaste{o: usecase.Overview{Error: "The calendar cannot be loaded."}}).
		Resolve(context.Background(), answers.Request{Match: answers.Match{Intent: intents[0]}, Lang: answers.English})
	if none.Empty != "The calendar cannot be loaded." {
		t.Errorf("error = %q", none.Empty)
	}
}
