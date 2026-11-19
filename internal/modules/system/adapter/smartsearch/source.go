// Package smartsearch answers questions about the host in the smart search
// (ADR 0026): CPU, memory, disks, load, uptime and temperature.
package smartsearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"waylight/internal/modules/system/usecase"
	"waylight/internal/shared/answers"
)

var t = answers.Terms

// intents of the system module; terms are folded (no umlauts, lower case).
var intents = []answers.Intent{
	{ID: "system.cpu", Source: "system", Prefixes: []string{"cpu"}, Answer: answers.Metric, WidgetKind: "system",
		Terms: concat(t(3, "cpu", "prozessor", "processor"), t(2, "auslastung", "ausgelastet", "usage", "utilization", "busy"))},
	{ID: "system.memory", Source: "system", Prefixes: []string{"ram"}, Answer: answers.Metric, WidgetKind: "system",
		Terms: concat(t(3, "ram", "arbeitsspeicher", "memory", "mem"), t(1, "speicher", "frei", "free"))},
	{ID: "system.disk", Source: "system", Prefixes: []string{"disk"}, Answer: answers.List, WidgetKind: "system",
		Terms: concat(
			t(3, "disk", "disks", "festplatte", "platte", "speicherplatz", "storage", "laufwerk"),
			t(1.5, "voll", "full", "space", "platz"),
			t(1, "frei", "free", "speicher"))},
	{ID: "system.load", Source: "system", Prefixes: []string{"load"}, Answer: answers.Metric, WidgetKind: "system",
		Terms: concat(t(3, "load", "loadavg", "systemlast"), t(1.5, "last"))},
	{ID: "system.uptime", Source: "system", Prefixes: []string{"uptime"}, Answer: answers.Metric, WidgetKind: "system",
		Terms: concat(
			t(3, "uptime", "laufzeit", "betriebszeit"),
			t(1.5, "neustart", "reboot", "seit", "since", "lange", "long"),
			t(1, "lauft", "laeuft", "running", "server"))},
	{ID: "system.temperature", Source: "system", Answer: answers.Metric, WidgetKind: "system",
		// "cpu temperatur" asks for the temperature, not the load.
		Terms: concat(t(1.5, "temperatur", "temperature", "temp"), t(1, "heiss", "hot", "server"), t(2, "cpu"))},
}

func concat(parts ...[]answers.Term) []answers.Term {
	var out []answers.Term
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

var messages = answers.Messages{
	answers.English: {
		"cpu": "CPU", "cpu_detail": "{n} cores · load {load}", "memory": "Memory", "of": "{used} of {total}",
		"disks": "Disks", "no_disks": "No disks found.", "free": "{free} free", "load": "Load (1, 5, 15 min)",
		"load_detail": "{n} cores", "uptime": "Uptime", "temperature": "Temperature",
		"no_temperature": "This host reports no temperature.",
	},
	answers.German: {
		"cpu": "CPU", "cpu_detail": "{n} Kerne · Last {load}", "memory": "Arbeitsspeicher", "of": "{used} von {total}",
		"disks": "Festplatten", "no_disks": "Keine Festplatten gefunden.", "free": "{free} frei",
		"load": "Last (1, 5, 15 Min.)", "load_detail": "{n} Kerne", "uptime": "Laufzeit", "temperature": "Temperatur",
		"no_temperature": "Dieser Rechner meldet keine Temperatur.",
	},
}

// Source answers from the latest host snapshot.
type Source struct{ monitor usecase.Monitor }

// New returns the source on monitor.
func New(monitor usecase.Monitor) *Source { return &Source{monitor: monitor} }

// ID implements answers.Source.
func (*Source) ID() string { return "system" }

// Intents implements answers.Source.
func (*Source) Intents() []answers.Intent { return intents }

// Entities implements answers.Source; the host has no named parts.
func (*Source) Entities(context.Context) ([]answers.Entity, error) { return nil, nil }

// Resolve implements answers.Source.
func (s *Source) Resolve(ctx context.Context, r answers.Request) (answers.Answer, error) {
	snap, err := s.monitor.Current(ctx)
	if err != nil {
		return answers.Answer{}, fmt.Errorf("system snapshot: %w", err)
	}
	l := r.Lang
	a := answers.Answer{Kind: answers.Metric, At: snap.TakenAt}
	switch r.Match.Intent.ID {
	case "system.cpu":
		a.Title, a.Value, a.Percent = messages.T(l, "cpu"), answers.FormatPercent(snap.CPUPercent), answers.Percent(snap.CPUPercent)
		a.Detail = messages.T(l, "cpu_detail", "n", snap.CPUCount, "load", fmt.Sprintf("%.2f", snap.Load1))
	case "system.memory":
		m := snap.Memory
		a.Title, a.Value, a.Percent = messages.T(l, "memory"), answers.FormatPercent(m.Percent), answers.Percent(m.Percent)
		a.Detail = messages.T(l, "of", "used", answers.FormatBytes(m.Used), "total", answers.FormatBytes(m.Total))
	case "system.disk":
		a.Kind, a.Title, a.Empty = answers.List, messages.T(l, "disks"), messages.T(l, "no_disks")
		for _, d := range snap.Disks {
			a.Items = append(a.Items, answers.Item{
				Label: d.Label,
				Value: answers.FormatPercent(d.Usage.Percent) + " · " +
					messages.T(l, "free", "free", answers.FormatBytes(d.Usage.Total-d.Usage.Used)),
				Percent: answers.Percent(d.Usage.Percent),
			})
		}
	case "system.load":
		a.Title = messages.T(l, "load")
		a.Value = strings.Join([]string{fmt.Sprintf("%.2f", snap.Load1), fmt.Sprintf("%.2f", snap.Load5), fmt.Sprintf("%.2f", snap.Load15)}, " · ")
		a.Detail = messages.T(l, "load_detail", "n", snap.CPUCount)
	case "system.uptime":
		a.Title, a.Value = messages.T(l, "uptime"), formatUptime(snap.Uptime)
	default:
		a.Title = messages.T(l, "temperature")
		if snap.Temperature > 0 {
			a.Value = fmt.Sprintf("%.0f °C", snap.Temperature)
		} else {
			a.Kind, a.Text = answers.Text, messages.T(l, "no_temperature")
		}
	}
	return a, nil
}

// formatUptime writes "3d 4h", "4h 12m" or "12m", as the web interface.
func formatUptime(d time.Duration) string {
	days, hours, mins := int(d.Hours())/24, int(d.Hours())%24, int(d.Minutes())%60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, mins)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}
