package domain

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits of user-provided values.
const (
	MaxCustomCSSBytes = 64 * 1024
)

// Theme is the color scheme preference.
type Theme string

// Supported themes.
const (
	ThemeSystem Theme = "system"
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Appearance holds the visual preferences. The zero value is not valid; use
// Default or NewAppearance.
type Appearance struct {
	theme           Theme
	accentColor     string
	backgroundImage string
	customCSS       string
	language        string
	greetingName    string
	launcher        Launcher
	profile         Profile
}

// Profile says what the start page is about (#281): everyday topics only,
// or also the home network and server. It only shows or hides module
// groups; it never deletes anything.
type Profile string

// Profiles; homelab is the default of existing installations.
const (
	ProfileEveryday Profile = "everyday"
	ProfileHomelab  Profile = "homelab"
)

// SetProfile chooses the profile; empty means homelab.
func (a *Appearance) SetProfile(p string) error {
	switch Profile(p) {
	case "", ProfileHomelab:
		a.profile = ProfileHomelab
	case ProfileEveryday:
		a.profile = ProfileEveryday
	default:
		return kernel.NewValidationError("profile", "The start page shows everyday topics or also the home network.")
	}
	return nil
}

// Profile returns what the start page is about.
func (a Appearance) Profile() Profile {
	if a.profile == "" {
		return ProfileHomelab
	}
	return a.profile
}

// Launcher says what the launcher of the top bar lists (#189).
type Launcher string

// Launcher contents; both is the default.
const (
	LauncherBoth      Launcher = "both"
	LauncherServices  Launcher = "services"
	LauncherBookmarks Launcher = "bookmarks"
)

// SetLauncher chooses the launcher content; empty means both.
func (a *Appearance) SetLauncher(l string) error {
	switch Launcher(l) {
	case "", LauncherBoth:
		a.launcher = LauncherBoth
	case LauncherServices, LauncherBookmarks:
		a.launcher = Launcher(l)
	default:
		return kernel.NewValidationError("launcher", "The launcher shows both, services or bookmarks.")
	}
	return nil
}

// Launcher returns what the launcher lists.
func (a Appearance) Launcher() Launcher {
	if a.launcher == "" {
		return LauncherBoth
	}
	return a.launcher
}

// MaxGreetingNameLength limits the name in the greeting (#167, #189).
const MaxGreetingNameLength = 40

// SetGreetingName sets the name the greeting addresses (may be empty).
func (a *Appearance) SetGreetingName(name string) error {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) > MaxGreetingNameLength || strings.ContainsFunc(name, unicode.IsControl) {
		return kernel.NewValidationError("greetingName", "The name must be at most 40 characters long.")
	}
	a.greetingName = name
	return nil
}

// GreetingName is the name used in the greeting.
func (a Appearance) GreetingName() string { return a.greetingName }

// Languages lists the supported UI languages; "" follows the browser.
var Languages = []string{"", "en", "de"}

// SetLanguage sets the UI language ("" = browser language).
func (a *Appearance) SetLanguage(lang string) error {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if !slices.Contains(Languages, lang) {
		return kernel.NewValidationError("language", "The language must be en, de or empty (browser).")
	}
	a.language = lang
	return nil
}

// Language returns the UI language or "" for the browser language.
func (a Appearance) Language() string { return a.language }

// Default returns the appearance of a fresh installation.
func Default() Appearance { return Appearance{theme: ThemeSystem} }

// WithScheme returns a copy with another theme and accent colour, for an
// organisation's preset (#305); the rest stays.
func (a Appearance) WithScheme(theme Theme, accentColor string) (Appearance, error) {
	b, err := NewAppearance(theme, accentColor, a.backgroundImage, a.customCSS)
	if err != nil {
		return Appearance{}, err
	}
	b.language, b.greetingName, b.launcher, b.profile = a.language, a.greetingName, a.launcher, a.profile
	return b, nil
}

// NewAppearance validates and creates an Appearance. Empty accentColor and
// backgroundImage mean "use the built-in default".
func NewAppearance(theme Theme, accentColor, backgroundImage, customCSS string) (Appearance, error) {
	switch theme {
	case ThemeSystem, ThemeLight, ThemeDark:
	default:
		return Appearance{}, kernel.NewValidationError("theme", "The theme must be system, light or dark.")
	}
	accentColor = strings.TrimSpace(accentColor)
	if accentColor != "" && !hexColor.MatchString(accentColor) {
		return Appearance{}, kernel.NewValidationError("accentColor", "The accent color must look like #3b6ef5.")
	}
	backgroundImage, err := kernel.ParseBackground("backgroundImage", backgroundImage)
	if err != nil {
		return Appearance{}, err
	}
	if len(customCSS) > MaxCustomCSSBytes {
		return Appearance{}, kernel.NewValidationError("customCss", "The custom CSS must be at most 64 KiB.")
	}
	return Appearance{theme: theme, accentColor: strings.ToLower(accentColor), backgroundImage: backgroundImage, customCSS: customCSS}, nil
}

// Theme returns the color scheme.
func (a Appearance) Theme() Theme { return a.theme }

// AccentColor returns the accent color (#rrggbb) or "".
func (a Appearance) AccentColor() string { return a.accentColor }

// BackgroundImage returns the background image URL or "".
func (a Appearance) BackgroundImage() string { return a.backgroundImage }

// SetBackgroundImage changes only the background (see kernel.ParseBackground).
func (a *Appearance) SetBackgroundImage(raw string) error {
	bg, err := kernel.ParseBackground("backgroundImage", raw)
	if err != nil {
		return err
	}
	a.backgroundImage = bg
	return nil
}

// CustomCSS returns the user's CSS.
func (a Appearance) CustomCSS() string { return a.customCSS }
