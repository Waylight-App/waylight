package domain

import "testing"

func TestContainerHealthAndSummary(t *testing.T) {
	cs := []Container{
		{Name: "web", State: "running", Status: "Up 2 hours (healthy)"},
		{Name: "db", State: "running", Status: "Up 2 hours (unhealthy)"},
		{Name: "job", State: "exited", Status: "Exited (0) 5 minutes ago"},
		{Name: "app", State: "running", Status: "Up 1 second (health: starting)"},
	}
	if cs[0].Health() != HealthHealthy || cs[1].Health() != HealthUnhealthy || cs[2].Health() != HealthNone || cs[3].Health() != HealthStarting {
		t.Fatal("health parsing wrong")
	}
	if s := Summarise(cs); s != (Summary{Total: 4, Running: 3, Stopped: 1, Unhealthy: 1}) {
		t.Fatalf("summary %+v", s)
	}
	SortForDisplay(cs)
	// Only the unhealthy one needs attention; the cleanly exited job sorts by name (#313).
	if cs[0].Name != "db" || cs[1].Name != "app" || cs[2].Name != "job" {
		t.Fatalf("order %v %v %v", cs[0].Name, cs[1].Name, cs[2].Name)
	}
}

func TestNeedsAttention(t *testing.T) {
	for _, tc := range []struct {
		state, status string
		want          bool
	}{
		{"running", "Up 3 hours (healthy)", false},
		{"running", "Up 3 hours (unhealthy)", true},
		{"exited", "Exited (0) 35 seconds ago", false}, // one-off job finished cleanly (#313)
		{"created", "Created", false},
		{"exited", "Exited (137) 2 hours ago", true},
		{"exited", "Exited (1) 5 minutes ago", true},
		{"exited", "Exited", true}, // no code: better report it
		{"restarting", "Restarting (1) 3 seconds ago", true},
		{"dead", "Dead", true},
		{"paused", "Up 2 hours (Paused)", false},
	} {
		if got := (Container{State: tc.state, Status: tc.status}).NeedsAttention(); got != tc.want {
			t.Errorf("%s %q: got %v", tc.state, tc.status, got)
		}
	}
	if code, ok := (Container{Status: "Exited (137) 2 hours ago"}).ExitCode(); !ok || code != 137 {
		t.Fatal(code, ok)
	}
}
