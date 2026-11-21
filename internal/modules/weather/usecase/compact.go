package usecase

import (
	"math"
	"strconv"

	"waylight/internal/shared/kernel"
)

// conditions names the WMO weather codes Open-Meteo reports, with the same
// wording as the widget.
var conditions = map[int]kernel.Text{
	0: kernel.NewText("Clear sky"), 1: kernel.NewText("Mainly clear"),
	2: kernel.NewText("Partly cloudy"), 3: kernel.NewText("Overcast"),
	45: kernel.NewText("Fog"), 48: kernel.NewText("Rime fog"),
	51: kernel.NewText("Light drizzle"), 53: kernel.NewText("Drizzle"), 55: kernel.NewText("Dense drizzle"),
	56: kernel.NewText("Freezing drizzle"), 57: kernel.NewText("Freezing drizzle"),
	61: kernel.NewText("Light rain"), 63: kernel.NewText("Rain"), 65: kernel.NewText("Heavy rain"),
	66: kernel.NewText("Freezing rain"), 67: kernel.NewText("Freezing rain"),
	71: kernel.NewText("Light snow"), 73: kernel.NewText("Snow"), 75: kernel.NewText("Heavy snow"),
	77: kernel.NewText("Snow grains"),
	80: kernel.NewText("Rain showers"), 81: kernel.NewText("Rain showers"), 82: kernel.NewText("Violent showers"),
	85: kernel.NewText("Snow showers"), 86: kernel.NewText("Snow showers"),
	95: kernel.NewText("Thunderstorm"), 96: kernel.NewText("Thunderstorm with hail"), 99: kernel.NewText("Thunderstorm with hail"),
}

// CompactState shows the current temperature with the sky for the compact
// view (#641), e.g. "7°" and "Overcast". The weather never needs action, so
// the level stays neutral.
func CompactState(f ForecastResponse) kernel.CompactState {
	condition, ok := conditions[f.WeatherCode]
	if !ok {
		condition = kernel.NewText("Unknown")
	}
	return kernel.CompactState{
		Title:     kernel.NewTextf("{temp}°", "temp", degrees(f.Temperature)),
		Detail:    condition,
		Level:     kernel.CompactNone,
		UpdatedAt: f.UpdatedAt,
	}
}

// degrees rounds to whole degrees like the widget.
func degrees(v float64) string { return strconv.Itoa(int(math.Round(v))) }
