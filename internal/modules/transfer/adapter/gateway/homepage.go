package gateway

import (
	"fmt"
	"net/url"
	"strings"

	"go.yaml.in/yaml/v3"

	"waylight/internal/modules/transfer/domain"
	"waylight/internal/shared/kernel"
)

// Homepage reads the configuration of gethomepage.dev (#77): services.yaml,
// bookmarks.yaml and widgets.yaml, each optional. Files are recognised by
// name, so a whole config folder can be uploaded.
type Homepage struct{}

// item is the shape Homepage uses everywhere: a list of one-key mappings,
// "- Name: value", which keeps the order of groups and entries.
type item struct {
	name  string
	value *yaml.Node
}

func items(n *yaml.Node) []item {
	if n == nil || n.Kind != yaml.SequenceNode {
		return nil
	}
	var out []item
	for _, e := range n.Content {
		if e.Kind == yaml.MappingNode && len(e.Content) == 2 {
			out = append(out, item{name: e.Content[0].Value, value: e.Content[1]})
		}
	}
	return out
}

// parseFile reads one YAML file; Homepage files are plain YAML without
// anchors in practice, so aliases are resolved by the library as usual.
func parseFile(name string, data []byte) (*yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, kernel.NewValidationErrorf("file", "{name} is not valid YAML: {reason}", "name", name, "reason", err.Error())
	}
	if len(doc.Content) == 0 {
		return nil, nil
	}
	return doc.Content[0], nil
}

// Parse implements usecase.ImportSource.
func (Homepage) Parse(files map[string][]byte) (domain.ImportPlan, error) {
	plan := domain.ImportPlan{Source: "Homepage"}
	known := false
	for name, data := range files {
		base := strings.ToLower(name[strings.LastIndexAny(name, `/\`)+1:])
		root, err := parseFile(name, data)
		if err != nil {
			return domain.ImportPlan{}, err
		}
		switch strings.TrimSuffix(strings.TrimSuffix(base, ".yaml"), ".yml") {
		case "services":
			known = true
			homepageServices(root, &plan)
		case "bookmarks":
			known = true
			homepageBookmarks(root, &plan)
		case "widgets":
			known = true
			homepageWidgets(root, &plan)
		default:
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("%s ignored: only services.yaml, bookmarks.yaml and widgets.yaml are read.", name))
		}
	}
	if !known {
		return domain.ImportPlan{}, kernel.NewValidationError("file", "Upload services.yaml, bookmarks.yaml or widgets.yaml from Homepage.")
	}
	return plan, nil
}

// homepageServices walks the groups; nested groups (Homepage layouts) are
// flattened, since Waylight tiles have no groups.
func homepageServices(root *yaml.Node, plan *domain.ImportPlan) {
	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		for _, it := range items(n) {
			if it.value.Kind == yaml.SequenceNode { // a (nested) group
				walk(it.value)
				continue
			}
			var s struct {
				Href        string         `yaml:"href"`
				Description string         `yaml:"description"`
				Icon        string         `yaml:"icon"`
				Widget      map[string]any `yaml:"widget"`
			}
			if err := it.value.Decode(&s); err != nil {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("Service %s skipped: %v", it.name, err))
				continue
			}
			if !isHTTP(s.Href) {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("Service %s skipped: it has no http(s) link.", it.name))
				continue
			}
			plan.Tiles = append(plan.Tiles, domain.ImportTile{Name: it.name, URL: s.Href, Icon: iconURL(s.Icon, it.name, plan), Description: s.Description})
			if t, ok := s.Widget["type"].(string); ok {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("The %s widget of %s is not imported; connect the service in the settings instead.", t, it.name))
			}
		}
	}
	walk(root)
}

func homepageBookmarks(root *yaml.Node, plan *domain.ImportPlan) {
	for _, group := range items(root) {
		f := domain.ImportFolder{Name: group.name}
		for _, b := range items(group.value) {
			// "- Name: [ {href: …} ]": the value is a list with one mapping.
			var entries []struct {
				Href string `yaml:"href"`
			}
			if err := b.value.Decode(&entries); err != nil || len(entries) == 0 || !isHTTP(entries[0].Href) {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("Bookmark %s skipped: it has no http(s) link.", b.name))
				continue
			}
			f.Links = append(f.Links, domain.ImportLink{Title: b.name, URL: entries[0].Href})
		}
		if len(f.Links) > 0 {
			plan.Folders = append(plan.Folders, f)
		}
	}
}

// homepageWidgets maps the information widgets that have a counterpart.
func homepageWidgets(root *yaml.Node, plan *domain.ImportPlan) {
	for _, w := range items(root) {
		switch w.name {
		case "datetime":
			plan.Widgets = append(plan.Widgets, domain.ImportWidget{Kind: "clock"})
		case "resources":
			plan.Widgets = append(plan.Widgets, domain.ImportWidget{Kind: "system"})
		case "openmeteo", "openweathermap", "weatherapi":
			var o struct {
				Label     string  `yaml:"label"`
				Latitude  float64 `yaml:"latitude"`
				Longitude float64 `yaml:"longitude"`
			}
			if err := w.value.Decode(&o); err != nil || (o.Latitude == 0 && o.Longitude == 0) {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("Widget %s skipped: it has no coordinates.", w.name))
				continue
			}
			plan.Widgets = append(plan.Widgets, weather(o.Label, o.Latitude, o.Longitude))
		case "search":
			plan.Warnings = append(plan.Warnings, "Widget search skipped: the search field is always at the top of Waylight.")
		default:
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("Widget %s skipped: Waylight has no counterpart.", w.name))
		}
	}
}

// weatherDays is the forecast length of an imported weather widget,
// the same as in the sample layout.
const weatherDays = 4

func weather(label string, lat, lon float64) domain.ImportWidget {
	cfg := map[string]any{"lat": lat, "lon": lon, "days": weatherDays}
	if label != "" {
		cfg["label"] = label
	}
	return domain.ImportWidget{Kind: "weather", Config: cfg}
}

func isHTTP(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// iconURL keeps icons given as http(s) URLs. Icon names ("sonarr.png",
// "mdi-…", "si-…") would make the browser load images from a public CDN,
// which Waylight avoids by default, so they are reported instead.
func iconURL(icon, service string, plan *domain.ImportPlan) string {
	if icon == "" {
		return ""
	}
	if isHTTP(icon) {
		return icon
	}
	plan.Warnings = append(plan.Warnings, fmt.Sprintf("Icon %q of %s not imported: only icons given as URL are taken.", icon, service))
	return ""
}
