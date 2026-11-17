// Package configfile maps the appearance to its section of waylight.yaml
// (#88) and back.
//
// Ring: Interface Adapters.
package configfile

import "waylight/internal/modules/appearance/usecase"

// Section is the appearance section. Empty fields fall back to the
// defaults, like in the settings dialog.
type Section struct {
	Theme           string `yaml:"theme,omitempty"`
	AccentColor     string `yaml:"accentColor,omitempty"`
	BackgroundImage string `yaml:"backgroundImage,omitempty"`
	Language        string `yaml:"language,omitempty"`
	GreetingName    string `yaml:"greetingName,omitempty"`
	GreetingMoment  bool   `yaml:"greetingMoment,omitempty"`
	Launcher        string `yaml:"launcher,omitempty"`
	Profile         string `yaml:"profile,omitempty"`
	CustomCSS       string `yaml:"customCss,omitempty"`
}

// defaultTheme follows the operating system, as on a fresh install.
const defaultTheme = "system"

// FromResponse maps the stored appearance.
func FromResponse(r usecase.AppearanceResponse) *Section {
	return &Section{Theme: r.Theme, AccentColor: r.AccentColor, BackgroundImage: r.BackgroundImage, Language: r.Language,
		GreetingName: r.GreetingName, GreetingMoment: r.GreetingMoment, Launcher: r.Launcher, Profile: r.Profile, CustomCSS: r.CustomCSS}
}

// Request maps the section to the update request.
func (s Section) Request() usecase.UpdateAppearanceRequest {
	theme := s.Theme
	if theme == "" {
		theme = defaultTheme
	}
	return usecase.UpdateAppearanceRequest{Theme: theme, AccentColor: s.AccentColor, BackgroundImage: s.BackgroundImage,
		Language: s.Language, GreetingName: s.GreetingName, GreetingMoment: s.GreetingMoment, Launcher: s.Launcher, Profile: s.Profile, CustomCSS: s.CustomCSS}
}
