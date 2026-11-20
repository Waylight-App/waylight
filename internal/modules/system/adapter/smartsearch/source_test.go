package smartsearch

import (
	"context"
	"testing"
	"time"

	"waylight/internal/modules/system/usecase"
	"waylight/internal/shared/answers"
)

type fakeMonitor struct {
	usecase.Monitor
	snap usecase.SnapshotView
}

func (f fakeMonitor) Current(context.Context) (usecase.SnapshotView, error) { return f.snap, nil }

func resolve(t *testing.T, s *Source, id string, lang answers.Lang) answers.Answer {
	t.Helper()
	for _, in := range s.Intents() {
		if in.ID == id {
			a, err := s.Resolve(context.Background(), answers.Request{Match: answers.Match{Intent: in}, Lang: lang})
			if err != nil {
				t.Fatal(err)
			}
			return a
		}
	}
	t.Fatalf("no intent %s", id)
	return answers.Answer{}
}

func TestSystemAnswers(t *testing.T) {
	s := New(fakeMonitor{snap: usecase.SnapshotView{
		CPUPercent: 42.4, CPUCount: 4, Load1: 0.5, Load5: 0.25, Load15: 0.1,
		Memory: usecase.UsageView{Total: 8 << 30, Used: 2 << 30, Percent: 25},
		Disks:  []usecase.DiskView{{Label: "/", Usage: usecase.UsageView{Total: 100 << 30, Used: 60 << 30, Percent: 60}}},
		Uptime: 50 * time.Hour,
	}})
	if a := resolve(t, s, "system.cpu", answers.German); a.Value != "42 %" || a.Detail != "4 Kerne · Last 0.50" {
		t.Errorf("cpu = %+v", a)
	}
	if a := resolve(t, s, "system.memory", answers.English); a.Detail != "2.0 GiB of 8.0 GiB" {
		t.Errorf("memory = %+v", a)
	}
	if a := resolve(t, s, "system.disk", answers.German); len(a.Items) != 1 || a.Items[0].Value != "60 % · 40 GiB frei" {
		t.Errorf("disk = %+v", a)
	}
	if a := resolve(t, s, "system.load", answers.English); a.Value != "0.50 · 0.25 · 0.10" {
		t.Errorf("load = %+v", a)
	}
	if a := resolve(t, s, "system.uptime", answers.English); a.Value != "2d 2h" {
		t.Errorf("uptime = %+v", a)
	}
	if a := resolve(t, s, "system.temperature", answers.German); a.Kind != answers.Text {
		t.Errorf("no temperature = %+v", a)
	}
}
