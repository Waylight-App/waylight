package smartsearch

import (
	"context"
	"testing"

	"waylight/internal/modules/services/usecase"
	"waylight/internal/shared/answers"
)

type fakeServices struct {
	usecase.Services
	list []usecase.ServiceView
}

func (f fakeServices) List(context.Context) ([]usecase.ServiceView, error) { return f.list, nil }

func TestServicesAnswers(t *testing.T) {
	s := New(fakeServices{list: []usecase.ServiceView{
		{ID: "1", Name: "Router", State: "online", LatencyMs: 3, URL: "http://router"},
		{ID: "2", Name: "NAS", State: "offline", Error: "timeout"},
		{ID: "3", Name: "Old", State: "offline", Hidden: true},
	}})
	entities, _ := s.Entities(context.Background())
	if len(entities) != 2 {
		t.Fatalf("hidden services count: %v", entities)
	}
	in := s.Intents()
	overview, _ := s.Resolve(context.Background(), answers.Request{Match: answers.Match{Intent: in[0]}, Lang: answers.German})
	if overview.Title != "Dienste: 1 von 2 online" || len(overview.Items) != 1 || overview.Items[0].Value != "timeout" {
		t.Errorf("overview = %+v", overview)
	}
	m := answers.Match{Intent: in[1], Slots: map[string]answers.SlotValue{"service": {Type: answers.SlotEntity, Entity: entities[0]}}}
	status, _ := s.Resolve(context.Background(), answers.Request{Match: m, Lang: answers.English})
	if status.State != answers.StateOK || status.Detail != "3 ms · http://router" || status.Text != "online" {
		t.Errorf("status = %+v", status)
	}
}
