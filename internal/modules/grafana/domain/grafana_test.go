package domain

import "testing"

func TestNewSettings(t *testing.T) {
	s, err := NewSettings(" http://grafana.local:3000/ ", " glsa_x ")
	if err != nil || s.URL != "http://grafana.local:3000" || s.Token != "glsa_x" {
		t.Fatalf("settings = %+v, %v", s, err)
	}
	for _, bad := range [][2]string{{"ftp://x", "t"}, {"http://", "t"}, {"http://g", " "}} {
		if _, err := NewSettings(bad[0], bad[1]); err == nil {
			t.Errorf("%v accepted", bad)
		}
	}
}

func TestRenderRequest(t *testing.T) {
	r, err := NewRenderRequest("abc-123", 4, 5000, 0, "", true)
	if err != nil || r.Width != 1600 || r.Height != 300 || r.Range != "6h" || !r.Dark {
		t.Fatalf("request = %+v, %v", r, err)
	}
	for _, bad := range []struct {
		uid   string
		panel int
		rng   string
	}{{"../x", 1, "6h"}, {"abc", 0, "6h"}, {"abc", 1, "now-6h"}, {"abc", 1, "0h"}} {
		if _, err := NewRenderRequest(bad.uid, bad.panel, 0, 0, bad.rng, false); err == nil {
			t.Errorf("%+v accepted", bad)
		}
	}
}

func TestParseAlertState(t *testing.T) {
	for in, want := range map[string]AlertState{"Alerting": Firing, "firing": Firing, "Pending": Pending} {
		if got, ok := ParseAlertState(in); !ok || got != want {
			t.Errorf("%q = %q, %v", in, got, ok)
		}
	}
	if _, ok := ParseAlertState("Normal"); ok {
		t.Error("normal alerts are not shown")
	}
}

func TestDownsampleKeepsTheNewestPoint(t *testing.T) {
	points := make([]Point, 500)
	for i := range points {
		points[i].Value = float64(i)
	}
	got := Downsample(points, MaxPoints)
	if len(got) != MaxPoints || got[0].Value != 0 || got[len(got)-1].Value != 499 {
		t.Fatalf("got %d points, first %v, last %v", len(got), got[0].Value, got[len(got)-1].Value)
	}
	for i := 1; i < len(got); i++ {
		if got[i].Value <= got[i-1].Value {
			t.Fatalf("not in order at %d", i)
		}
	}
	if short := Downsample(points[:10], MaxPoints); len(short) != 10 {
		t.Fatalf("short series changed: %d", len(short))
	}
}

func TestNewDataRequest(t *testing.T) {
	r, err := NewDataRequest("node", 2, "")
	if err != nil || r.Range != "6h" || r.PanelID != 2 {
		t.Fatalf("got %+v %v", r, err)
	}
	if _, err := NewDataRequest("../x", 2, "6h"); err == nil {
		t.Fatal("bad uid accepted")
	}
	if _, err := NewDataRequest("node", 2, "6x"); err == nil {
		t.Fatal("bad range accepted")
	}
}
