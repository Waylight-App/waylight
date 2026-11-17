package presenter

import (
	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/modules/appearance/usecase"
)

// AppearanceDTO is returned by GET/PUT /api/appearance and is also the body
// of PUT /api/appearance.
type AppearanceDTO struct {
	Theme           string `json:"theme" tstype:"'system' | 'light' | 'dark'"`
	AccentColor     string `json:"accentColor"`
	BackgroundImage string `json:"backgroundImage"`
	CustomCSS       string `json:"customCss"`
	Language        string `json:"language" tstype:"'' | 'en' | 'de'"`
	// GreetingName is the name in the greeting (#189); may be empty.
	GreetingName string `json:"greetingName"`
	// GreetingMoment adds a line by time of day to the greeting (#512).
	GreetingMoment bool `json:"greetingMoment"`
	// Launcher is what the launcher lists (#189).
	Launcher string `json:"launcher" tstype:"'both' | 'services' | 'bookmarks'"`
	// Profile is what the start page shows (#281); empty on PUT keeps it.
	Profile string `json:"profile" tstype:"'everyday' | 'homelab'"`
	// ThemeLocked: the organisation set theme and accent colour (#305); a
	// change to them is ignored. Never read from a PUT.
	ThemeLocked bool `json:"themeLocked,omitempty"`
}

// Appearance maps an AppearanceResponse.
func Appearance(r usecase.AppearanceResponse) AppearanceDTO {
	return AppearanceDTO{Theme: r.Theme, AccentColor: r.AccentColor, BackgroundImage: r.BackgroundImage, CustomCSS: r.CustomCSS, Language: r.Language,
		GreetingName: r.GreetingName, GreetingMoment: r.GreetingMoment, Launcher: r.Launcher, Profile: r.Profile, ThemeLocked: r.ThemeLocked}
}

// ThemeDTO is a gallery theme.
type ThemeDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author"`
	// Scheme is empty for a theme with variants.
	Scheme string            `json:"scheme" tstype:"'light' | 'dark' | ''"`
	Tokens map[string]string `json:"tokens"`
	// Variants are the light and dark tokens of a theme that has both (#505).
	Variants *ThemeVariantsDTO `json:"variants,omitempty"`
	BuiltIn  bool              `json:"builtIn"`
}

// ThemeVariantsDTO holds the tokens of the light and the dark variant.
type ThemeVariantsDTO struct {
	Light map[string]string `json:"light"`
	Dark  map[string]string `json:"dark"`
}

// GalleryDTO is the theme gallery (#111).
type GalleryDTO struct {
	Themes []ThemeDTO `json:"themes"`
	// Applied is the id of the theme in use ("" = own or no custom CSS).
	Applied    string `json:"applied"`
	CanRestore bool   `json:"canRestore"`
}

// Gallery maps the gallery.
func Gallery(g usecase.Gallery) GalleryDTO {
	out := GalleryDTO{Themes: make([]ThemeDTO, 0, len(g.Themes)), Applied: g.Applied, CanRestore: g.CanRestore}
	for _, t := range g.Themes {
		out.Themes = append(out.Themes, Theme(t))
	}
	return out
}

// Theme maps a theme.
func Theme(t domain.ThemeDef) ThemeDTO {
	out := ThemeDTO{ID: t.ID, Name: t.Name, Author: t.Author, Scheme: string(t.Scheme), Tokens: t.Tokens, BuiltIn: t.BuiltIn}
	if t.HasVariants() {
		out.Variants = &ThemeVariantsDTO{Light: t.Variants[domain.ThemeLight], Dark: t.Variants[domain.ThemeDark]}
	}
	return out
}

// BackgroundDTO is an uploaded background picture (#277); url is the value
// to store as background, thumb a small preview.
type BackgroundDTO struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Thumb  string `json:"thumb"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Size   int64  `json:"size"`
}

// BackgroundListDTO lists the uploaded backgrounds with the limits.
type BackgroundListDTO struct {
	Items     []BackgroundDTO `json:"items"`
	Max       int             `json:"max"`
	MaxUpload int64           `json:"maxUploadBytes"`
}

// Background maps one picture.
func Background(b domain.Background) BackgroundDTO {
	return BackgroundDTO{ID: b.ID, URL: b.URL(), Thumb: b.URL() + "?thumb=1", Width: b.Width, Height: b.Height, Size: b.Size}
}

// Backgrounds maps the list.
func Backgrounds(list []domain.Background) BackgroundListDTO {
	items := make([]BackgroundDTO, 0, len(list))
	for _, b := range list {
		items = append(items, Background(b))
	}
	return BackgroundListDTO{Items: items, Max: domain.MaxBackgrounds, MaxUpload: domain.MaxBackgroundUploadBytes}
}
