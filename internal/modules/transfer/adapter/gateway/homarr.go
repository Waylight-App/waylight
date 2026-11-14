package gateway

import (
	"encoding/json"
	"fmt"

	"trisse-home/internal/modules/transfer/domain"
	"trisse-home/internal/shared/kernel"
)

// Homarr reads a board configuration of Homarr 0.x (#77), the JSON file in
// Homarr's data/configs folder. Homarr 1.x only exports an encrypted
// migration archive, which is not readable here.
type Homarr struct{}

type homarrConfig struct {
	SchemaVersion int `json:"schemaVersion"`
	Categories    []struct {
		Name string `json:"name"`
	} `json:"categories"`
	Apps []struct {
		Name      string `json:"name"`
		URL       string `json:"url"`
		Behaviour struct {
			ExternalURL string `json:"externalUrl"`
		} `json:"behaviour"`
		Appearance struct {
			IconURL string `json:"iconUrl"`
		} `json:"appearance"`
	} `json:"apps"`
	Widgets []struct {
		Type       string `json:"type"`
		Properties struct {
			Location *struct {
				Name      string  `json:"name"`
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
		} `json:"properties"`
	} `json:"widgets"`
}

// Parse implements usecase.ImportSource; it expects exactly one file.
func (Homarr) Parse(files map[string][]byte) (domain.ImportPlan, error) {
	if len(files) != 1 {
		return domain.ImportPlan{}, kernel.NewValidationError("file", "Upload one Homarr board file (JSON).")
	}
	var data []byte
	for _, d := range files {
		data = d
	}
	var c homarrConfig
	if err := json.Unmarshal(data, &c); err != nil || c.SchemaVersion == 0 {
		return domain.ImportPlan{}, kernel.NewValidationError("file", "This is not a Homarr board file. Homarr 1.x exports cannot be read; use the board JSON of Homarr 0.x.")
	}
	plan := domain.ImportPlan{Source: "Homarr"}
	for _, a := range c.Apps {
		// The external URL is what the browser opens; url is the internal
		// address Homarr pings.
		link := a.Behaviour.ExternalURL
		if !isHTTP(link) {
			link = a.URL
		}
		if !isHTTP(link) {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("App %s skipped: it has no http(s) link.", a.Name))
			continue
		}
		plan.Tiles = append(plan.Tiles, domain.ImportTile{Name: a.Name, URL: link, Icon: iconURL(a.Appearance.IconURL, a.Name, &plan)})
	}
	if len(c.Categories) > 0 {
		plan.Warnings = append(plan.Warnings, "Categories are not carried over: service tiles in trisse-home have no groups.")
	}
	for _, w := range c.Widgets {
		switch w.Type {
		case "date":
			plan.Widgets = append(plan.Widgets, domain.ImportWidget{Kind: "clock"})
		case "weather":
			if l := w.Properties.Location; l != nil && (l.Latitude != 0 || l.Longitude != 0) {
				plan.Widgets = append(plan.Widgets, weather(l.Name, l.Latitude, l.Longitude))
				continue
			}
			plan.Warnings = append(plan.Warnings, "Widget weather skipped: it has no location.")
		default:
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("Widget %s skipped: trisse-home has no counterpart.", w.Type))
		}
	}
	return plan, nil
}
