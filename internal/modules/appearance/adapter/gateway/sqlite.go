package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/shared/kernel"
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

// Load returns the principal's appearance (ADR 0018). An account without
// its own row starts from the admin's, so a new member sees the household's
// look and language; guests, kiosk displays and background work always use
// the admin's. Without any row it is the default.
func (r *Repository) Load(ctx context.Context) (domain.Appearance, error) {
	account := kernel.PrincipalFrom(ctx).SettingsAccount()
	var theme, accent, bg, css, lang, greetingName, launcher, profile string
	err := r.db.QueryRowContext(ctx,
		`SELECT theme, accent_color, background_image, custom_css, language, greeting_name, launcher, profile FROM appearance_settings
		WHERE account_id IN (?, ?) ORDER BY account_id = ? DESC LIMIT 1`, account, kernel.SystemAccountID, account).
		Scan(&theme, &accent, &bg, &css, &lang, &greetingName, &launcher, &profile)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Default(), nil
	}
	if err != nil {
		return domain.Appearance{}, fmt.Errorf("select appearance: %w", err)
	}
	a, err := domain.NewAppearance(domain.Theme(theme), accent, bg, css)
	if err == nil {
		err = a.SetLanguage(lang)
	}
	if err == nil {
		err = a.SetGreetingName(greetingName)
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
	_, err := r.db.ExecContext(ctx, `INSERT INTO appearance_settings (account_id, theme, accent_color, background_image, custom_css, language, greeting_name, launcher, profile)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (account_id) DO UPDATE SET theme = excluded.theme, accent_color = excluded.accent_color,
			background_image = excluded.background_image, custom_css = excluded.custom_css, language = excluded.language,
			greeting_name = excluded.greeting_name, launcher = excluded.launcher, profile = excluded.profile`,
		account, string(a.Theme()), a.AccentColor(), a.BackgroundImage(), a.CustomCSS(), a.Language(), a.GreetingName(), string(a.Launcher()), string(a.Profile()))
	if err != nil {
		return fmt.Errorf("upsert appearance: %w", err)
	}
	return nil
}
