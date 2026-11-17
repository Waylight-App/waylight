package gateway

import (
	"context"
	"math"
	"time"

	"waylight/internal/modules/system/domain"
)

const gib = 1 << 30

// DemoHost is the machine of the public demo: a plausible home server whose
// values drift slowly over time. The demo must not show the name or load of
// the server it actually runs on.
type DemoHost struct {
	// Now is the clock; the values follow it so every sample differs a bit.
	Now func() time.Time
}

// wave returns a value between 0 and 1 that changes over a period.
func (h DemoHost) wave(period time.Duration, phase float64) float64 {
	t := float64(h.Now().UnixNano()) / float64(period)
	return (math.Sin(2*math.Pi*t+phase) + 1) / 2
}

// CPUTimes implements usecase.HostReader. The counters grow so that the
// busy share between two samples lies around 15–35 %.
func (h DemoHost) CPUTimes(context.Context) (domain.CPUTimes, error) {
	ticks := uint64(h.Now().UnixMilli() / 10)
	busy := 0.15 + 0.2*h.wave(17*time.Minute, 0)
	return domain.CPUTimes{Total: ticks, Idle: uint64(float64(ticks) * (1 - busy))}, nil
}

// CPUCount implements usecase.HostReader.
func (DemoHost) CPUCount(context.Context) (int, error) { return 8, nil }

// Load implements usecase.HostReader.
func (h DemoHost) Load(context.Context) (domain.Load, error) {
	one := 0.8 + 1.6*h.wave(13*time.Minute, 1)
	return domain.Load{One: one, Five: one*0.8 + 0.3, Fifteen: one*0.6 + 0.4}, nil
}

// Memory implements usecase.HostReader.
func (h DemoHost) Memory(context.Context) (domain.Usage, domain.Usage, error) {
	used := uint64((11 + 3*h.wave(41*time.Minute, 2)) * gib)
	return domain.Usage{Total: 32 * gib, Used: used}, domain.Usage{Total: 8 * gib, Used: gib / 2}, nil
}

// Uptime implements usecase.HostReader.
func (DemoHost) Uptime(context.Context) (time.Duration, error) {
	return 23*24*time.Hour + 5*time.Hour, nil
}

// Disks implements usecase.HostReader.
func (DemoHost) Disks(context.Context) ([]domain.Disk, error) {
	return []domain.Disk{
		{Label: "System", Usage: domain.Usage{Total: 476 * gib, Used: 138 * gib}},
		{Label: "Data", Usage: domain.Usage{Total: 7450 * gib, Used: 5180 * gib}},
	}, nil
}

// Hostname implements usecase.HostReader.
func (DemoHost) Hostname(context.Context) string { return "homeserver" }

// Temperature implements usecase.HostReader.
func (h DemoHost) Temperature(context.Context) (float64, error) {
	return 44 + 6*h.wave(29*time.Minute, 3), nil
}
