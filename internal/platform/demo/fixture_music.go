package demo

import (
	"net/http"
	"strings"
	"time"
)

// Artist is an example artist the new releases widget follows; the
// names are made up, the IDs have the MusicBrainz format.
type Artist struct{ ID, Name string }

// Artists are followed in the showcase.
var Artists = []Artist{
	{"3f1d6c2a-7b4e-4c9a-9e21-5a6b7c8d9e01", "Nordlicht Ensemble"},
	{"4a2e7d3b-8c5f-4dab-8f32-6b7c8d9e0f12", "The Harbour Lights"},
	{"5b3f8e4c-9d6a-4ebc-9a43-7c8d9e0f1a23", "Moorland Radio"},
}

func init() {
	handle(hostIs("musicbrainz.org"), musicBrainz)
	handle(hostIs("coverartarchive.org"), func(req *http.Request, _ time.Time) (int, string, string) {
		return http.StatusOK, "image/png", picture(req.URL.Path, 250, 250)
	})
}

func musicBrainz(req *http.Request, now time.Time) (int, string, string) {
	q := req.URL.Query().Get("query")
	if strings.HasSuffix(req.URL.Path, "/artist") {
		var out []map[string]string
		for _, a := range Artists {
			out = append(out, map[string]string{"id": a.ID, "name": a.Name, "disambiguation": "demo artist"})
		}
		return jsonOK(mustJSON(map[string]any{"artists": out}))
	}
	day := func(d int) string { return now.AddDate(0, 0, -d).Format("2006-01-02") }
	groups := map[string][]map[string]string{
		Artists[0].ID: {{"id": "a1b2c3d4-0000-4000-8000-000000000101", "title": "Polar Night", "primary-type": "Album", "first-release-date": day(2)},
			{"id": "a1b2c3d4-0000-4000-8000-000000000102", "title": "Low Sun", "primary-type": "Single", "first-release-date": day(19)}},
		Artists[1].ID: {{"id": "a1b2c3d4-0000-4000-8000-000000000201", "title": "Ferry Songs", "primary-type": "EP", "first-release-date": day(6)}},
		Artists[2].ID: {{"id": "a1b2c3d4-0000-4000-8000-000000000301", "title": "Heathland Sessions (Live)", "primary-type": "Album", "first-release-date": day(11)}},
	}
	for id, g := range groups {
		if strings.Contains(q, id) {
			return jsonOK(mustJSON(map[string]any{"release-groups": g}))
		}
	}
	return jsonOK(`{"release-groups":[]}`)
}
