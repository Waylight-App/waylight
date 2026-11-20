package answers_test

import (
	"slices"
	"testing"
	"time"

	birthdaysearch "waylight/internal/modules/birthdays/adapter/smartsearch"
	dockersearch "waylight/internal/modules/docker/adapter/smartsearch"
	parcelsearch "waylight/internal/modules/parcels/adapter/smartsearch"
	servicessearch "waylight/internal/modules/services/adapter/smartsearch"
	systemsearch "waylight/internal/modules/system/adapter/smartsearch"
	timelinesearch "waylight/internal/modules/timeline/adapter/smartsearch"
	transitsearch "waylight/internal/modules/transit/adapter/smartsearch"
	wastesearch "waylight/internal/modules/waste/adapter/smartsearch"
	"waylight/internal/shared/answers"
)

// The intents of the modules on the server, with the regression table of
// the web interface's smart search for them.
var (
	intents = slices.Concat(
		systemsearch.New(nil).Intents(), dockersearch.New(nil).Intents(), servicessearch.New(nil).Intents(),
		parcelsearch.New(nil).Intents(), birthdaysearch.New(nil).Intents(),
		timelinesearch.New(nil).Intents(), wastesearch.New(nil).Intents(), transitsearch.New(nil).Intents(),
	)
	entities = []answers.Entity{
		{Kind: "container", ID: "c1", Name: "jellyfin"},
		{Kind: "container", ID: "c2", Name: "nextcloud-app"},
		{Kind: "container", ID: "c3", Name: "home-assistant"},
		{Kind: "container", ID: "c4", Name: "pihole"},
		{Kind: "container", ID: "c5", Name: "postgres"},
		{Kind: "service", ID: "s1", Name: "Router"},
		{Kind: "service", ID: "s2", Name: "NAS"},
		{Kind: "service", ID: "s3", Name: "Nextcloud"},
		{Kind: "service", ID: "s4", Name: "Grafana"},
	}
	// Thursday, 12 November 2026.
	today = time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
)

func top(query string) (answers.Match, bool) {
	ranked := answers.Rank(answers.Interpret(query, intents, entities, today))
	if len(ranked) == 0 {
		return answers.Match{}, false
	}
	return ranked[0], true
}

func TestRegressionTable(t *testing.T) {
	table := []struct{ query, want string }{
		{"cpu", "system.cpu"}, {"CPU Auslastung", "system.cpu"}, {"wie hoch ist die cpu", "system.cpu"},
		{"prozessor", "system.cpu"}, {"processor usage", "system.cpu"}, {"/cpu", "system.cpu"},
		{"ram", "system.memory"}, {"arbeitsspeicher", "system.memory"}, {"wie viel ram ist frei", "system.memory"},
		{"memory usage", "system.memory"}, {"/ram", "system.memory"},
		{"wie voll ist die platte", "system.disk"}, {"festplatte", "system.disk"}, {"speicherplatz", "system.disk"},
		{"disk", "system.disk"}, {"is the disk full", "system.disk"}, {"how much storage is free", "system.disk"},
		{"plate voll", "system.disk"}, {"platte voll", "system.disk"}, {"/disk", "system.disk"},
		{"load", "system.load"}, {"systemlast", "system.load"}, {"load average", "system.load"},
		{"uptime", "system.uptime"}, {"laufzeit", "system.uptime"}, {"seit wann läuft der server", "system.uptime"},
		{"wann war der letzte neustart", "system.uptime"}, {"temperatur server", "system.temperature"},
		{"wie heiß ist die cpu", "system.cpu"},
		{"docker", "docker.overview"}, {"welche container laufen", "docker.overview"}, {"containers", "docker.overview"},
		{"jellyfin", "docker.status"}, {"jelyfin", "docker.status"}, {"jel", "docker.status"},
		{"läuft jellyfin", "docker.status"}, {"pihole status", "docker.status"}, {"home assistant", "docker.status"},
		{"/docker postgres", "docker.status"},
		{"dienste", "services.overview"}, {"welche dienste sind offline", "services.overview"},
		{"what is down", "services.overview"},
		{"router", "services.status"}, {"ist das nas erreichbar", "services.status"}, {"grafana ping", "services.status"},
		{"/service router", "services.status"},
		{"paket", "parcels.status"}, {"wo ist mein paket", "parcels.status"}, {"dhl", "parcels.status"},
		{"dhl paket", "parcels.status"},
		{"wer hat geburtstag", "birthdays.upcoming"}, {"wer hat heute geburtstag", "birthdays.upcoming"},
		{"birthdays", "birthdays.upcoming"},
		{"wann kommt die gelbe tonne", "waste.next"}, {"müll", "waste.next"}, {"abfall", "waste.next"},
		{"when is the trash collected", "waste.next"}, {"/müll", "waste.next"},
		{"was ist passiert", "timeline.recent"}, {"vorfälle", "timeline.recent"}, {"what happened", "timeline.recent"},
		{"timeline", "timeline.recent"},
		{"abfahrt", "transit.departures"}, {"nächster bus", "transit.departures"},
		{"abfahrten hauptbahnhof", "transit.departures"}, {"wann fährt die nächste bahn", "transit.departures"},
		{"/bus oldenburg hbf", "transit.departures"}, {"Abfahrt Oldenburg Hbf", "transit.departures"},
		{"", ""}, {"hallo", ""}, {"xyz", ""}, {"der die das", ""},
	}
	for _, c := range table {
		m, ok := top(c.query)
		got := ""
		if ok {
			got = m.Intent.ID
		}
		if got != c.want {
			t.Errorf("%q: got %q, want %q", c.query, got, c.want)
		}
	}
}

func TestSlots(t *testing.T) {
	m, _ := top("nas")
	if e, ok := m.EntitySlot("service"); !ok || e.Name != "NAS" {
		t.Errorf("service slot = %+v", m.Slots)
	}
	m, _ = top("/docker")
	if !slices.Equal(m.Missing, []string{"container"}) || len(m.Suggestions) == 0 {
		t.Errorf("asks back: missing %v, suggestions %v", m.Missing, m.Suggestions)
	}
}

func TestDayOffset(t *testing.T) {
	for word, want := range map[string]int{"morgen": 1, "heute": 0, "samstag": 2, "thursday": 0, "ubermorgen": 2} {
		if got, ok := answers.DayOffset(word, today); !ok || got != want {
			t.Errorf("DayOffset(%q) = %d, %v; want %d", word, got, ok, want)
		}
	}
	if _, ok := answers.DayOffset("so", today); ok {
		t.Error("abbreviations are no days")
	}
}
