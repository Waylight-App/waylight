package domain

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Health of a container as reported by its health check.
type Health string

// Health values.
const (
	HealthNone      Health = "none"
	HealthHealthy   Health = "healthy"
	HealthUnhealthy Health = "unhealthy"
	HealthStarting  Health = "starting"
)

// Container is a Docker container as shown on the dashboard.
type Container struct {
	ID      string
	Name    string
	Image   string
	State   string // running, exited, restarting, paused, created, dead
	Status  string // human readable, e.g. "Up 3 hours (healthy)"
	Project string // compose project, if any
	Created time.Time
	// Labels are the container labels (used by service discovery, #78).
	Labels map[string]string
	// Ports are the TCP ports the container exposes, and Addresses its IPs
	// in the Docker networks; the widget setup finds services by them (#590).
	Ports     []Port
	Addresses []string
}

// Port is an exposed TCP port; Public is 0 when it is not published on the
// host.
type Port struct {
	Private, Public int
}

// Health derives the health check state from the status text.
func (c Container) Health() Health {
	switch {
	case strings.Contains(c.Status, "(healthy)"):
		return HealthHealthy
	case strings.Contains(c.Status, "(unhealthy)"):
		return HealthUnhealthy
	case strings.Contains(c.Status, "(health: starting)"):
		return HealthStarting
	default:
		return HealthNone
	}
}

var exitCode = regexp.MustCompile(`^Exited \((-?\d+)\)`)

// ExitCode returns the exit code of an exited container from its status
// ("Exited (137) 2 hours ago"); ok is false when it has none.
func (c Container) ExitCode() (code int, ok bool) {
	m := exitCode.FindStringSubmatch(c.Status)
	if m == nil {
		return 0, false
	}
	code, err := strconv.Atoi(m[1])
	return code, err == nil
}

// NeedsAttention reports whether the container failed: unhealthy,
// restarting, dead, or exited with a non-zero code. A one-off container
// that finished cleanly (Exited (0)) or one that was only created is no
// incident (#313); the dashboard still lists it as stopped.
func (c Container) NeedsAttention() bool {
	switch c.State {
	case "running":
		return c.Health() == HealthUnhealthy
	case "restarting", "dead":
		return true
	case "exited":
		code, ok := c.ExitCode()
		return !ok || code != 0
	}
	return false // created, paused, removing
}

// Summary counts containers by condition.
type Summary struct {
	Total, Running, Stopped, Unhealthy int
}

// Summarise counts the containers.
func Summarise(cs []Container) Summary {
	s := Summary{Total: len(cs)}
	for _, c := range cs {
		switch {
		case c.State == "running" && c.Health() == HealthUnhealthy:
			s.Running++
			s.Unhealthy++
		case c.State == "running":
			s.Running++
		default:
			s.Stopped++
		}
	}
	return s
}

// SortForDisplay puts containers needing attention first, then sorts by name.
func SortForDisplay(cs []Container) {
	slices.SortStableFunc(cs, func(a, b Container) int {
		if a.NeedsAttention() != b.NeedsAttention() {
			if a.NeedsAttention() {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})
}
