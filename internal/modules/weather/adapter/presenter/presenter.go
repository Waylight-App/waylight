package presenter

import (
	"math"
	"time"

	"waylight/internal/modules/weather/usecase"
)

// DayDTO is one forecast day.
type DayDTO struct {
	Date                     string  `json:"date"`
	TemperatureMin           float64 `json:"temperatureMin"`
	TemperatureMax           float64 `json:"temperatureMax"`
	WeatherCode              int     `json:"weatherCode"`
	PrecipitationProbability float64 `json:"precipitationProbability"`
	// Sunrise and Sunset are null on polar days and nights.
	Sunrise *time.Time `json:"sunrise"`
	Sunset  *time.Time `json:"sunset"`
}

// ForecastDTO is returned by GET /api/weather.
type ForecastDTO struct {
	Lat                 float64 `json:"lat"`
	Lon                 float64 `json:"lon"`
	Timezone            string  `json:"timezone"`
	Temperature         float64 `json:"temperature"`
	ApparentTemperature float64 `json:"apparentTemperature"`
	Humidity            float64 `json:"humidity"`
	WindSpeed           float64 `json:"windSpeed"`
	// WindDirection is where the wind comes from in whole degrees (0 = north); null when unknown.
	WindDirection *int      `json:"windDirection"`
	WeatherCode   int       `json:"weatherCode"`
	IsDay         bool      `json:"isDay"`
	Days          []DayDTO  `json:"days"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// PlaceDTO is a geocoding result returned by GET /api/weather/places.
type PlaceDTO struct {
	Name    string  `json:"name"`
	Region  string  `json:"region"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// Forecast maps a ForecastResponse.
func Forecast(r usecase.ForecastResponse) ForecastDTO {
	dto := ForecastDTO{
		Lat: r.Lat, Lon: r.Lon, Timezone: r.Timezone, Temperature: r.Temperature,
		ApparentTemperature: r.ApparentTemperature, Humidity: r.Humidity, WindSpeed: r.WindSpeed,
		WeatherCode: r.WeatherCode, IsDay: r.IsDay, UpdatedAt: r.UpdatedAt, Days: make([]DayDTO, 0, len(r.Days)),
	}
	for _, d := range r.Days {
		dto.Days = append(dto.Days, DayDTO{
			Date: d.Date, TemperatureMin: d.TemperatureMin, TemperatureMax: d.TemperatureMax, WeatherCode: d.WeatherCode,
			PrecipitationProbability: d.PrecipitationProbability, Sunrise: instant(d.Sunrise), Sunset: instant(d.Sunset),
		})
	}
	if r.HasWindDirection {
		// 359.6° rounds to 360°, which is north again.
		deg := int(math.Round(r.WindDirection)) % 360
		dto.WindDirection = &deg
	}
	return dto
}

// instant is nil for an unknown time, so clients read null instead of year 1.
func instant(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	utc := t.UTC()
	return &utc
}

// Places maps geocoding results.
func Places(ps []usecase.PlaceView) []PlaceDTO {
	out := make([]PlaceDTO, 0, len(ps))
	for _, p := range ps {
		out = append(out, PlaceDTO(p))
	}
	return out
}
