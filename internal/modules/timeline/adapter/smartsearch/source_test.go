package smartsearch

import (
	"context"
	"testing"
	"time"

	"waylight/internal/modules/timeline/domain"
	"waylight/internal/modules/timeline/usecase"
	"waylight/internal/shared/answers"
)

type fakeTimeline struct {
	usecase.Timeline
	events []domain.Event
	filter usecase.Filter
}

func (f *fakeTimeline) List(_ context.Context, filter usecase.Filter) ([]domain.Event, error) {
	f.filter = filter
	return f.events, nil
}

func TestTimelineAnswer(t *testing.T) {
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	tl := &fakeTimeline{events: []domain.Event{
		{At: now.Add(-20 * time.Minute), Severity: domain.SeverityCritical, Subject: "NAS", Message: "offline"},
		{At: now.Add(-72 * time.Hour), Severity: domain.SeverityInfo, Subject: "Backup"},
	}}
	a, err := New(tl).Resolve(context.Background(), answers.Request{Match: answers.Match{Intent: intents[0]}, Lang: answers.German, Now: now})
	if err != nil || tl.filter.Limit != maxItems || a.Title != "Letzte Ereignisse" {
		t.Fatalf("answer = %+v, %v", a, err)
	}
	if a.Items[0].Value != "vor 20 Min. · offline" || a.Items[0].State != answers.StateError || a.Items[1].Value != "vor 3 T." {
		t.Errorf("items = %+v", a.Items)
	}
}
