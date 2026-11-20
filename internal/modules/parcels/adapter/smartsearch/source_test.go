package smartsearch

import (
	"context"
	"testing"
	"time"

	"waylight/internal/modules/parcels/domain"
	"waylight/internal/modules/parcels/usecase"
	"waylight/internal/shared/answers"
)

type fakeParcels struct {
	usecase.Parcels
	list []domain.Parcel
}

func (f fakeParcels) List(context.Context) ([]domain.Parcel, error) { return f.list, nil }

func TestParcelAnswer(t *testing.T) {
	at := time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC)
	s := New(fakeParcels{list: []domain.Parcel{
		{Label: "Shoes", Status: domain.Status{Code: domain.StatusTransit, At: at}, EstimatedDelivery: time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)},
		{TrackingNumber: "00340434", Status: domain.Status{Code: domain.StatusDelivered, Location: "Packstation 101"}},
		{Label: "Broken", Error: "Unknown tracking number"},
	}})
	a, err := s.Resolve(context.Background(), answers.Request{Match: answers.Match{Intent: intents[0]}, Lang: answers.German})
	if err != nil || len(a.Items) != 3 || !a.At.Equal(at) {
		t.Fatalf("answer = %+v, %v", a, err)
	}
	want := []answers.Item{
		{Label: "Shoes", Value: "Unterwegs · erwartet am 21.11."},
		{Label: "00340434", Value: "Zugestellt · Packstation 101", State: answers.StateOK},
		{Label: "Broken", Value: "Unknown tracking number", State: answers.StateError},
	}
	for i, w := range want {
		if a.Items[i].Label != w.Label || a.Items[i].Value != w.Value || a.Items[i].State != w.State {
			t.Errorf("item %d = %+v, want %+v", i, a.Items[i], w)
		}
	}
	en, _ := s.Resolve(context.Background(), answers.Request{Match: answers.Match{Intent: intents[0]}, Lang: answers.English})
	if en.Items[0].Value != "On the way · expected Nov 21" {
		t.Errorf("english = %q", en.Items[0].Value)
	}
}
