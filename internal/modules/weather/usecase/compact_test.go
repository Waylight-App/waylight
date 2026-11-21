package usecase

import (
	"testing"

	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	s := CompactState(ForecastResponse{Temperature: 6.6, WeatherCode: 3})
	if s.Title.String() != "7°" || s.Detail.String() != "Overcast" || s.Level != kernel.CompactNone {
		t.Fatalf("%+v", s)
	}
	if s := CompactState(ForecastResponse{Temperature: -0.3, WeatherCode: 42}); s.Title.String() != "0°" || s.Detail.String() != "Unknown" {
		t.Fatalf("unknown code: %+v", s)
	}
}
