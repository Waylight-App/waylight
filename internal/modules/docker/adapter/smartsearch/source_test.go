package smartsearch

import (
	"context"
	"testing"

	"waylight/internal/modules/docker/usecase"
	"waylight/internal/shared/answers"
)

type fakeContainers struct {
	usecase.Containers
	list usecase.ListResponse
}

func (f fakeContainers) List(context.Context) (usecase.ListResponse, error) { return f.list, nil }

func TestDockerAnswers(t *testing.T) {
	s := New(fakeContainers{list: usecase.ListResponse{Total: 2, Running: 1, Containers: []usecase.ContainerView{
		{ID: "1", Name: "jellyfin", State: "running", Status: "Up 3 hours", Image: "jellyfin/jellyfin"},
		{ID: "2", Name: "pihole", State: "exited", Status: "Exited (0)"},
	}}})
	entities, _ := s.Entities(context.Background())
	if len(entities) != 2 {
		t.Fatalf("entities = %v", entities)
	}
	in := s.Intents()
	overview, err := s.Resolve(context.Background(), answers.Request{Match: answers.Match{Intent: in[0]}, Lang: answers.German})
	if err != nil || overview.Title != "Container: 1 von 2 laufen" || len(overview.Items) != 1 || overview.Items[0].State != answers.StateOff {
		t.Errorf("overview = %+v, %v", overview, err)
	}
	m := answers.Match{Intent: in[1], Slots: map[string]answers.SlotValue{"container": {Type: answers.SlotEntity, Entity: entities[0]}}}
	status, _ := s.Resolve(context.Background(), answers.Request{Match: m, Lang: answers.English})
	if status.State != answers.StateOK || status.Detail != "jellyfin/jellyfin" {
		t.Errorf("status = %+v", status)
	}
	m.Slots["container"] = answers.SlotValue{Type: answers.SlotEntity, Entity: answers.Entity{ID: "x", Name: "gone"}}
	if gone, _ := s.Resolve(context.Background(), answers.Request{Match: m, Lang: answers.English}); gone.Kind != answers.Text {
		t.Errorf("gone = %+v", gone)
	}
}
