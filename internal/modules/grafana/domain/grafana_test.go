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
