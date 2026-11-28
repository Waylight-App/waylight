package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"waylight/internal/modules/appearance/domain"
	"waylight/internal/shared/kernel"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// HasOwn reports whether the principal stored its own appearance, rather
// than starting from the admin's or the default (#305).
func (r *Repository) HasOwn(ctx context.Context) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM appearance_settings WHERE account_id = ?`,
		kernel.PrincipalFrom(ctx).SettingsAccount()).Scan(&n)
	return n > 0, err
}

// Load returns the principal's appearance (ADR 0018). An account without
// its own row starts from the admin's, so a new member sees the household's
// look and language; guests, kiosk displays and background work always use
// the admin's. Without any row it is the default.
func (r *Repository) Load(ctx context.Context) (domain.Appearance, error) {
	account := kernel.PrincipalFrom(ctx).SettingsAccount()
	var rowAccount, theme, accent, bg, css, lang, greetingName, launcher, profile, textSize, sunID string
	var greetingMoment, showAdvanced, highContrast, simpleView bool
	var sunLat, sunLon float64
	err := r.db.QueryRowContext(ctx,
		`SELECT account_id, theme, accent_color, background_image, custom_css, language, greeting_name, greeting_moment, launcher, profile, show_advanced,
			text_size, high_contrast, simple_view, sun_place_id, sun_lat, sun_lon
		FROM appearance_settings
		WHERE account_id IN (?, ?) ORDER BY account_id = ? DESC LIMIT 1`, account, kernel.SystemAccountID, account).
		Scan(&rowAccount, &theme, &accent, &bg, &css, &lang, &greetingName, &greetingMoment, &launcher, &profile, &showAdvanced,
			&textSize, &highContrast, &simpleView, &sunID, &sunLat, &sunLon)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Default(), nil
	}
	if err != nil {
		return domain.Appearance{}, fmt.Errorf("select appearance: %w", err)
	}
	// Light or dark is a choice of the account itself (#863): a member who
	// starts from the admin's look follows the system, unless that look is
	// a theme made for one scheme, which brings its scheme along.
	if rowAccount != account && !domain.IsSingleSchemeThemeCSS(css) {
		theme = string(domain.ThemeSystem)
	}
	a, err := domain.NewAppearance(domain.Theme(theme), accent, bg, css)
	if err == nil {
		err = a.SetLanguage(lang)
	}
	if err == nil {
		err = a.SetGreetingName(greetingName)
	}
	a.SetGreetingMoment(greetingMoment)
	// Advanced settings are a choice of the account itself; a member who
	// starts from the admin's look does not inherit it.
	a.SetShowAdvanced(showAdvanced && rowAccount == account)
	// Accessibility options are personal too (#681).
	if rowAccount == account {
		access, aerr := domain.NewAccessibility(textSize, highContrast, simpleView)
		if aerr == nil {
			a.SetAccessibility(access)
		}
		// The place is one of the account's own saved places.
		if sun, serr := domain.NewSunPlace(sunID, sunLat, sunLon); serr == nil {
			a.SetSunPlace(sun)
		}
	}
	if err == nil {
		err = a.SetLauncher(launcher)
	}
	if err == nil {
		err = a.SetProfile(profile)
	}
	if err != nil {
		// Stored data that no longer satisfies the rules falls back to the
		// default instead of locking the user out of the dashboard.
		return domain.Default(), nil //nolint:nilerr // intentional fallback
	}
	return a, nil
}

// Save stores the principal's appearance; it never changes another
// account's row.
func (r *Repository) Save(ctx context.Context, a domain.Appearance) error {
	account := kernel.PrincipalFrom(ctx).SettingsAccount()
	access, sun := a.Accessibility(), a.SunPlace()
	_, err := r.db.ExecContext(ctx, `INSERT INTO appearance_settings (account_id, theme, accent_color, background_image, custom_css, language, greeting_name, greeting_moment, launcher, profile, show_advanced,
			text_size, high_contrast, simple_view, sun_place_id, sun_lat, sun_lon)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (account_id) DO UPDATE SET theme = excluded.theme, accent_color = excluded.accent_color,
			background_image = excluded.background_image, custom_css = excluded.custom_css, language = excluded.language,
			greeting_name = excluded.greeting_name, greeting_moment = excluded.greeting_moment, launcher = excluded.launcher, profile = excluded.profile,
			show_advanced = excluded.show_advanced, text_size = excluded.text_size, high_contrast = excluded.high_contrast, simple_view = excluded.simple_view,
			sun_place_id = excluded.sun_place_id, sun_lat = excluded.sun_lat, sun_lon = excluded.sun_lon`,
		account, string(a.Theme()), a.AccentColor(), a.BackgroundImage(), a.CustomCSS(), a.Language(), a.GreetingName(), a.GreetingMoment(), string(a.Launcher()), string(a.Profile()), a.ShowAdvanced(),
		string(access.TextSize), access.HighContrast, access.SimpleView, sun.ID, sun.Lat, sun.Lon)
	if err != nil {
		return fmt.Errorf("upsert appearance: %w", err)
	}
	return nil
}
