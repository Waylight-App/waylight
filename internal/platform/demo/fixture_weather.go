package demo

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"strings"
	"time"
)

// SolarURL is the JSON endpoint of the API widget's example: a balcony solar
// inverter.
const SolarURL = "http://solar.demo.invalid/api/live"

func init() {
	handle(hostIs("maps.dwd.de"), func(req *http.Request, now time.Time) (int, string, string) {
		if strings.HasSuffix(req.URL.Path, "/wms") {
			return http.StatusOK, "image/png", radarImage(now)
		}
		return jsonOK(dwdWarnings(now))
	})
	handle(hostIs("solar.demo.invalid"), func(_ *http.Request, now time.Time) (int, string, string) {
		// Output follows the sun: zero at night, about 600 W at noon.
		h := float64(now.Hour()) + float64(now.Minute())/60
		watts := math.Max(0, 640*math.Sin((h-6)/14*math.Pi))
		return jsonOK(fmt.Sprintf(`{"inverter":{"power_w":%.0f,"today_kwh":%.2f,"total_kwh":1834.6,"temperature_c":%.1f}}`,
			watts, 2.4+watts/400, 21+watts/60))
	})
}

func dwdWarnings(now time.Time) string {
	warning := map[string]any{"properties": map[string]string{
		"IDENTIFIER": "demo.warning.1", "EVENT": "STARKWIND", "HEADLINE": "Amtliche WARNUNG vor STARKEN WINDBÖEN",
		"DESCRIPTION": "Es treten Windböen mit Geschwindigkeiten um 55 km/h aus westlicher Richtung auf.",
		"SEVERITY":    "Minor", "ONSET": now.Add(-time.Hour).UTC().Format(time.RFC3339),
		"EXPIRES": now.Add(5 * time.Hour).UTC().Format(time.RFC3339),
	}}
	return mustJSON(map[string]any{"type": "FeatureCollection", "features": []any{warning}})
}

// radarImage draws a few soft rain cells on a transparent background, like
// the precipitation layer of a radar map; the cells drift with the hour.
func radarImage(now time.Time) string {
	const w, h = 480, 480
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	shift := float64(now.Minute()) * 2
	cells := []struct{ x, y, r float64 }{{140, 180, 70}, {230, 240, 95}, {330, 150, 50}, {380, 330, 60}}
	for y := range h {
		for x := range w {
			v := 0.0
			for _, c := range cells {
				d := math.Hypot(float64(x)-(c.x+shift), float64(y)-c.y) / c.r
				v = math.Max(v, 1-d)
			}
			if v <= 0 {
				continue
			}
			col := color.NRGBA{0x4a, 0xa8, 0xe8, uint8(120 + 100*v)}
			if v > 0.55 {
				col = color.NRGBA{0x2f, 0x7f, 0xd0, 220}
			}
			if v > 0.8 {
				col = color.NRGBA{0xf2, 0xb4, 0x41, 230}
			}
			img.Set(x, y, col)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img) // in-memory buffer, cannot fail
	return buf.String()
}
