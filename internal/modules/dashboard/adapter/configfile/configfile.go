// Package configfile maps the dashboard to its section of waylight.yaml
// (#88) and back.
//
// Ring: Interface Adapters.
package configfile

import (
	"encoding/json"

	"waylight/internal/modules/dashboard/domain"
	"waylight/internal/modules/dashboard/usecase"
	"waylight/internal/shared/kernel"
)

// Section is the dashboard section.
type Section struct {
	Tabs []Tab `yaml:"tabs"`
}

// Tab is one tab with its widgets.
type Tab struct {
	Name       string   `yaml:"name"`
	Background string   `yaml:"background,omitempty"`
	Rules      []Rule   `yaml:"rules,omitempty"`
	Widgets    []Widget `yaml:"widgets"`
}

// Rule opens the tab by default on these weekdays (0 = Sunday), in this
// time range and/or from these networks; empty parts always match.
type Rule struct {
	Days     []int    `yaml:"days,omitempty,flow"`
	From     string   `yaml:"from,omitempty"`
	To       string   `yaml:"to,omitempty"`
	Networks []string `yaml:"networks,omitempty,flow"`
}

// Widget is one tile; config is whatever the widget's settings store.
type Widget struct {
	Kind   string         `yaml:"kind"`
	At     Rect           `yaml:"at,flow"`
	Phone  *Phone         `yaml:"phone,omitempty,flow"`
	Config map[string]any `yaml:"config,omitempty"`
}

// Rect is the position in the 12-column grid.
type Rect struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
	W int `yaml:"w"`
	H int `yaml:"h"`
}

// Phone is the place in the phone layout: order (from 1), height in rows
// and whether the widget is hidden on phones.
type Phone struct {
	Order  int  `yaml:"order"`
	Rows   int  `yaml:"rows"`
	Hidden bool `yaml:"hidden,omitempty"`
}

var errConfig = kernel.NewValidationError("config", "A widget config must be a mapping of settings.")

// FromTabs maps the use case tabs.
func FromTabs(tabs []usecase.ConfigTab) (*Section, error) {
	out := &Section{Tabs: []Tab{}}
	for _, t := range tabs {
		ct := Tab{Name: t.Name, Background: t.Background, Widgets: []Widget{}}
		for _, r := range t.Rules {
			ct.Rules = append(ct.Rules, Rule{Days: r.Days, From: r.From, To: r.To, Networks: r.Networks})
		}
		for _, w := range t.Widgets {
			var cfg map[string]any
			if w.Config != "" {
				if err := json.Unmarshal([]byte(w.Config), &cfg); err != nil {
					return nil, err
				}
			}
			if len(cfg) == 0 {
				cfg = nil
			}
			cw := Widget{Kind: w.Kind, At: Rect(w.Rect), Config: cfg}
			if w.Phone != nil {
				cw.Phone = &Phone{Order: w.Phone.Order, Rows: w.Phone.Rows, Hidden: w.Phone.Hidden}
			}
			ct.Widgets = append(ct.Widgets, cw)
		}
		out.Tabs = append(out.Tabs, ct)
	}
	return out, nil
}

// ToTabs maps back to the use case tabs.
func (s Section) ToTabs() ([]usecase.ConfigTab, error) {
	out := make([]usecase.ConfigTab, 0, len(s.Tabs))
	for _, t := range s.Tabs {
		ct := usecase.ConfigTab{Name: t.Name, Background: t.Background}
		for _, r := range t.Rules {
			ct.Rules = append(ct.Rules, domain.RuleInput{Days: r.Days, From: r.From, To: r.To, Networks: r.Networks})
		}
		for _, w := range t.Widgets {
			cfg := "{}"
			if len(w.Config) > 0 {
				b, err := json.Marshal(w.Config)
				if err != nil {
					return nil, errConfig
				}
				cfg = string(b)
			}
			cw := usecase.ConfigWidget{Kind: w.Kind, Rect: domain.Rect(w.At), Config: cfg}
			if w.Phone != nil {
				cw.Phone = &usecase.ConfigPhone{Order: w.Phone.Order, Rows: w.Phone.Rows, Hidden: w.Phone.Hidden}
			}
			ct.Widgets = append(ct.Widgets, cw)
		}
		out = append(out, ct)
	}
	return out, nil
}
