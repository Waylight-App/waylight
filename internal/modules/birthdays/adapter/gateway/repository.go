// Package gateway stores birthdays and reads CardDAV address books (#287).
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"waylight/internal/modules/birthdays/domain"
	"waylight/internal/modules/birthdays/usecase"
	"waylight/internal/shared/kernel"
	"waylight/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are the tables included in data exports, without the
// encrypted address book password.
var TransferTables = []sqlutil.Table{
	{Name: "birthdays"},
	{Name: "birthday_settings", Exclude: []string{"password_encrypted"}},
}

// Cipher encrypts secrets before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// Repository implements usecase.Repository.
type Repository struct {
	db     *sql.DB
	cipher Cipher
}

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB, cipher Cipher) *Repository { return &Repository{db: db, cipher: cipher} }

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Birthday, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, month, day, year, source FROM birthdays WHERE `+cond+` ORDER BY month, day, name`, args...)
	if err != nil {
		return nil, fmt.Errorf("list birthdays: %w", err)
	}
	defer rows.Close()
	var out []domain.Birthday
	for rows.Next() {
		var b domain.Birthday
		var m int
		if err := rows.Scan(&b.ID, &b.Name, &m, &b.Day, &b.Year, &b.Source); err != nil {
			return nil, fmt.Errorf("scan birthday: %w", err)
		}
		b.Month = time.Month(m)
		out = append(out, b)
	}
	return out, rows.Err()
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// save inserts a birthday owned by the principal or updates one it owns
// (ADR 0018); a foreign one returns kernel.ErrNotOwner.
func save(ctx context.Context, db execer, b domain.Birthday) error {
	p := kernel.PrincipalFrom(ctx)
	owned, args := sqlutil.OwnedBy(p, "birthdays")
	res, err := db.ExecContext(ctx, `INSERT INTO birthdays (id, name, month, day, year, source, owner_id, visibility) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, month = excluded.month, day = excluded.day, year = excluded.year, source = excluded.source
		WHERE `+owned,
		append([]any{b.ID, b.Name, int(b.Month), b.Day, b.Year, b.Source, p.AccountID, string(kernel.DefaultVisibility(p))}, args...)...)
	if err != nil {
		return fmt.Errorf("save birthday: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	return nil
}

// Save implements usecase.Repository.
func (r *Repository) Save(ctx context.Context, b domain.Birthday) error { return save(ctx, r.db, b) }

// Delete implements usecase.Repository for the owner. A visible birthday of
// someone else returns kernel.ErrNotOwner, an invisible one ErrNotFound.
func (r *Repository) Delete(ctx context.Context, id string) error {
	p := kernel.PrincipalFrom(ctx)
	owned, args := sqlutil.OwnedBy(p, "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM birthdays WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete birthday: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		cond, vargs := sqlutil.VisibleTo(p, "")
		var one int
		if r.db.QueryRowContext(ctx, `SELECT 1 FROM birthdays WHERE id = ? AND `+cond, append([]any{id}, vargs...)...).Scan(&one) == nil {
			return kernel.ErrNotOwner
		}
		return domain.ErrNotFound
	}
	return nil
}

// ReplaceSource implements usecase.Repository in one transaction.
func (r *Repository) ReplaceSource(ctx context.Context, source string, list []domain.Birthday) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	// The address book sync runs as the system; an account only replaces
	// its own synced birthdays.
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	if _, err := tx.ExecContext(ctx, `DELETE FROM birthdays WHERE source = ? AND `+owned, append([]any{source}, args...)...); err != nil {
		return fmt.Errorf("clear %s birthdays: %w", source, err)
	}
	for _, b := range list {
		if err := save(ctx, tx, b); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// SyncAccounts implements usecase.Repository: the accounts with an address
// book to synchronise.
func (r *Repository) SyncAccounts(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT account_id FROM birthday_settings WHERE carddav_url != '' ORDER BY account_id`)
	if err != nil {
		return nil, fmt.Errorf("select address books: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// Settings implements usecase.Repository; no row means defaults.
func (r *Repository) Settings(ctx context.Context) (usecase.Settings, error) {
	var s usecase.Settings
	var enabled bool
	var enc string
	// The own settings only; the address book credentials are personal.
	err := r.db.QueryRowContext(ctx, `SELECT remind_enabled, remind_at, last_reminded, carddav_url, username, password_encrypted FROM birthday_settings WHERE account_id = ?`,
		kernel.PrincipalFrom(ctx).SettingsAccount()).
		Scan(&enabled, &s.Reminder.At, &s.LastReminded, &s.CardDAVURL, &s.Username, &enc)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.Settings{Reminder: domain.Reminder{At: "08:00"}}, nil
	}
	if err != nil {
		return s, fmt.Errorf("read birthday settings: %w", err)
	}
	s.Reminder.Enabled = enabled
	if enc != "" {
		if s.Password, err = r.cipher.Decrypt(enc); err != nil {
			return s, fmt.Errorf("decrypt CardDAV password (was SECRET_KEY changed?): %w", err)
		}
	}
	return s, nil
}

// SaveSettings implements usecase.Repository.
func (r *Repository) SaveSettings(ctx context.Context, s usecase.Settings) error {
	enc := ""
	if s.Password != "" {
		var err error
		if enc, err = r.cipher.Encrypt(s.Password); err != nil {
			return fmt.Errorf("encrypt CardDAV password: %w", err)
		}
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO birthday_settings (account_id, remind_enabled, remind_at, last_reminded, carddav_url, username, password_encrypted)
		VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(account_id) DO UPDATE SET remind_enabled = excluded.remind_enabled, remind_at = excluded.remind_at,
		last_reminded = excluded.last_reminded, carddav_url = excluded.carddav_url, username = excluded.username, password_encrypted = excluded.password_encrypted`,
		kernel.PrincipalFrom(ctx).SettingsAccount(), s.Reminder.Enabled, s.Reminder.At, s.LastReminded, s.CardDAVURL, s.Username, enc)
	if err != nil {
		return fmt.Errorf("save birthday settings: %w", err)
	}
	return nil
}
