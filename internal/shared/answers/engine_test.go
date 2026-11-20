package answers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

type fakeSource struct {
	err      error
	entities []Entity
}

func (*fakeSource) ID() string { return "fake" }
func (*fakeSource) Intents() []Intent {
	return []Intent{
		{ID: "fake.overview", Source: "fake", Answer: List, WidgetKind: "fake", Terms: Terms(3, "things")},
		{ID: "fake.status", Source: "fake", Answer: Status, Prefixes: []string{"thing"},
			Slots: []Slot{{Name: "thing", Type: SlotEntity, EntityKind: "thing", Required: true}}},
	}
}
func (f *fakeSource) Entities(context.Context) ([]Entity, error) { return f.entities, f.err }
func (f *fakeSource) Resolve(_ context.Context, r Request) (Answer, error) {
	if f.err != nil {
		return Answer{}, f.err
	}
	return Answer{Kind: List, Title: string(r.Lang)}, nil
}

var quiet = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestEngineAnswersInTheLanguage(t *testing.T) {
	e := NewEngine(quiet, &fakeSource{})
	got := e.Answer(context.Background(), "things", German, time.Now())
	if len(got) != 1 || got[0].Title != "de" || got[0].Intent != "fake.overview" || got[0].WidgetKind != "fake" {
		t.Fatalf("answers = %+v", got)
	}
}

func TestEngineAsksBack(t *testing.T) {
	e := NewEngine(quiet, &fakeSource{entities: []Entity{{Kind: "thing", ID: "1", Name: "Lamp"}}})
	got := e.Answer(context.Background(), "/thing", German, time.Now())
	if len(got) != 1 || got[0].Title != "Welcher?" || len(got[0].Suggestions) != 1 || got[0].Suggestions[0] != "Lamp" {
		t.Fatalf("answers = %+v", got)
	}
}

func TestEngineSaysWhenASourceFails(t *testing.T) {
	e := NewEngine(quiet, &fakeSource{err: errors.New("down")})
	got := e.Answer(context.Background(), "things", English, time.Now())
	if len(got) != 1 || got[0].Kind != Text || got[0].Title != "Not available right now" {
		t.Fatalf("answers = %+v", got)
	}
}
