package gateway

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/search/domain"
	"trisse-home/internal/modules/search/usecase"
	"trisse-home/internal/shared/kernel"
	"trisse-home/internal/shared/sqlutil"
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

// Search settings are personal (ADR 0018). An account that has not changed
// anything yet uses the admin's settings and shortcuts, so a new member
// starts with the household's; the first change copies them into the
// account. Guests, kiosk displays and background work use the admin's.

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// readAccount returns the account whose settings apply: its own once it
// has a settings row, otherwise the admin's.
func readAccount(ctx context.Context, q queryer) (string, error) {
	account := kernel.PrincipalFrom(ctx).SettingsAccount()
	if account == kernel.SystemAccountID {
		return account, nil
	}
	var one int
	err := q.QueryRowContext(ctx, `SELECT 1 FROM search_settings WHERE account_id = ?`, account).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return kernel.SystemAccountID, nil
	}
	if err != nil {
		return "", fmt.Errorf("check search settings: %w", err)
	}
	return account, nil
}

// writeAccount returns the principal's account and, the first time it
// changes anything, copies the admin's settings and shortcuts into it.
func writeAccount(ctx context.Context, q queryer) (string, error) {
	account := kernel.PrincipalFrom(ctx).SettingsAccount()
	if account == kernel.SystemAccountID {
		return account, nil
	}
	var one int
	err := q.QueryRowContext(ctx, `SELECT 1 FROM search_settings WHERE account_id = ?`, account).Scan(&one)
	if err == nil {
		return account, nil // has its own settings already
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("check search settings: %w", err)
	}
	res, err := q.ExecContext(ctx, `INSERT INTO search_settings (account_id, default_provider, open_in_new_tab, capture_prefixes)
		SELECT ?, default_provider, open_in_new_tab, capture_prefixes FROM search_settings WHERE account_id = ?`, account, kernel.SystemAccountID)
	if err != nil {
		return "", fmt.Errorf("copy search settings: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := q.ExecContext(ctx, `INSERT INTO search_settings (account_id, default_provider) VALUES (?, '')`, account); err != nil {
			return "", fmt.Errorf("create search settings: %w", err)
		}
	}
	if _, err := q.ExecContext(ctx, `INSERT INTO search_shortcuts (account_id, keyword, name, url_template)
		SELECT ?, keyword, name, url_template FROM search_shortcuts WHERE account_id = ?`, account, kernel.SystemAccountID); err != nil {
		return "", fmt.Errorf("copy shortcuts: %w", err)
	}
	if _, err := q.ExecContext(ctx, `INSERT INTO search_keywords (account_id, keyword, kind, target, label)
		SELECT ?, keyword, kind, target, label FROM search_keywords WHERE account_id = ?`, account, kernel.SystemAccountID); err != nil {
		return "", fmt.Errorf("copy keywords: %w", err)
	}
	return account, nil
}

// DefaultProvider implements usecase.Repository.
func (r *Repository) DefaultProvider(ctx context.Context) (string, error) {
	account, err := readAccount(ctx, r.db)
	if err != nil {
		return "", err
	}
	var id string
	err = r.db.QueryRowContext(ctx, `SELECT default_provider FROM search_settings WHERE account_id = ?`, account).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("select default provider: %w", err)
	}
	return id, nil
}

// SetDefaultProvider implements usecase.Repository.
func (r *Repository) SetDefaultProvider(ctx context.Context, id string) error {
	account, err := writeAccount(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO search_settings (account_id, default_provider) VALUES (?, ?)
		ON CONFLICT (account_id) DO UPDATE SET default_provider = excluded.default_provider`, account, id)
	if err != nil {
		return fmt.Errorf("upsert default provider: %w", err)
	}
	return nil
}

// BarOptions implements usecase.Repository.
func (r *Repository) BarOptions(ctx context.Context) (domain.BarOptions, error) {
	account, err := readAccount(ctx, r.db)
	if err != nil {
		return domain.BarOptions{}, err
	}
	var newTab bool
	var raw string
	err = r.db.QueryRowContext(ctx, `SELECT open_in_new_tab, capture_prefixes FROM search_settings WHERE account_id = ?`, account).Scan(&newTab, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DefaultBarOptions(), nil
	}
	if err != nil {
		return domain.BarOptions{}, fmt.Errorf("select search field options: %w", err)
	}
	p := domain.DefaultCapturePrefixes
	if raw != "" {
		var stored struct{ Todo, Note, Bookmark string }
		if json.Unmarshal([]byte(raw), &stored) == nil {
			p = domain.CapturePrefixes{Todo: stored.Todo, Note: stored.Note, Bookmark: stored.Bookmark}
		}
	}
	o, err := domain.NewBarOptions(newTab, p)
	if err != nil {
		// A row that violates current rules falls back to the default prefixes
		// instead of breaking the search field.
		o, _ = domain.NewBarOptions(newTab, domain.DefaultCapturePrefixes)
	}
	return o, nil
}

// SetBarOptions implements usecase.Repository.
func (r *Repository) SetBarOptions(ctx context.Context, o domain.BarOptions) error {
	p := o.Prefixes()
	raw, err := json.Marshal(struct{ Todo, Note, Bookmark string }{p.Todo, p.Note, p.Bookmark})
	if err != nil {
		return fmt.Errorf("encode capture prefixes: %w", err)
	}
	account, err := writeAccount(ctx, r.db)
	if err != nil {
		return err
	}
	// No row yet: default_provider '' means "the default provider".
	_, err = r.db.ExecContext(ctx, `INSERT INTO search_settings (account_id, default_provider, open_in_new_tab, capture_prefixes)
		VALUES (?, '', ?, ?)
		ON CONFLICT (account_id) DO UPDATE SET open_in_new_tab = excluded.open_in_new_tab, capture_prefixes = excluded.capture_prefixes`,
		account, o.OpenInNewTab(), string(raw))
	if err != nil {
		return fmt.Errorf("upsert search field options: %w", err)
	}
	return nil
}

// Shortcuts implements usecase.Repository.
func (r *Repository) Shortcuts(ctx context.Context) ([]domain.Shortcut, error) {
	account, err := readAccount(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT keyword, name, url_template FROM search_shortcuts WHERE account_id = ? ORDER BY keyword`, account)
	if err != nil {
		return nil, fmt.Errorf("query shortcuts: %w", err)
	}
	defer rows.Close()
	var out []domain.Shortcut
	for rows.Next() {
		var k, n, t string
		if err := rows.Scan(&k, &n, &t); err != nil {
			return nil, fmt.Errorf("scan shortcut: %w", err)
		}
		s, err := domain.NewShortcut(k, n, t)
		if err != nil {
			continue // skip rows that violate current rules instead of failing every search
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// AddShortcut implements usecase.Repository.
func (r *Repository) AddShortcut(ctx context.Context, s domain.Shortcut) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	account, err := writeAccount(ctx, tx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO search_shortcuts (account_id, keyword, name, url_template) VALUES (?, ?, ?, ?)`,
		account, s.Keyword(), s.Name(), s.Template())
	if sqlutil.IsUniqueViolation(err) {
		return usecase.ErrShortcutExists
	}
	if err != nil {
		return fmt.Errorf("insert shortcut: %w", err)
	}
	return tx.Commit()
}

// ReplaceShortcuts implements usecase.ConfigRepository in one transaction;
// config as code runs as the system and replaces the admin's shortcuts.
func (r *Repository) ReplaceShortcuts(ctx context.Context, list []domain.Shortcut) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	account, err := writeAccount(ctx, tx)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO search_settings (account_id, default_provider) VALUES (?, '') ON CONFLICT (account_id) DO NOTHING`, account); err != nil {
		return fmt.Errorf("create search settings: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM search_shortcuts WHERE account_id = ?`, account); err != nil {
		return fmt.Errorf("clear shortcuts: %w", err)
	}
	for _, sc := range list {
		if _, err := tx.ExecContext(ctx, `INSERT INTO search_shortcuts (account_id, keyword, name, url_template) VALUES (?, ?, ?, ?)`,
			account, sc.Keyword(), sc.Name(), sc.Template()); err != nil {
			return fmt.Errorf("insert shortcut: %w", err)
		}
	}
	return tx.Commit()
}

// DeleteShortcut implements usecase.Repository.
func (r *Repository) DeleteShortcut(ctx context.Context, keyword string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	account, err := writeAccount(ctx, tx)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM search_shortcuts WHERE account_id = ? AND keyword = ?`, account, keyword)
	if err != nil {
		return fmt.Errorf("delete shortcut: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return usecase.ErrShortcutNotFound
	}
	return tx.Commit()
}
