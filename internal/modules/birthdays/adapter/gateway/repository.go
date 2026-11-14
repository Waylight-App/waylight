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

	"trisse-home/internal/modules/birthdays/domain"
	"trisse-home/internal/modules/birthdays/usecase"
	"trisse-home/internal/shared/sqlutil"
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
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, month, day, year, source FROM birthdays ORDER BY month, day, name`)
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

func save(ctx context.Context, db execer, b domain.Birthday) error {
	_, err := db.ExecContext(ctx, `INSERT INTO birthdays (id, name, month, day, year, source) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, month = excluded.month, day = excluded.day, year = excluded.year, source = excluded.source`,
		b.ID, b.Name, int(b.Month), b.Day, b.Year, b.Source)
	if err != nil {
		return fmt.Errorf("save birthday: %w", err)
	}
	return nil
}

// Save implements usecase.Repository.
func (r *Repository) Save(ctx context.Context, b domain.Birthday) error { return save(ctx, r.db, b) }

// Delete implements usecase.Repository.
func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM birthdays WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete birthday: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
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
	if _, err := tx.ExecContext(ctx, `DELETE FROM birthdays WHERE source = ?`, source); err != nil {
		return fmt.Errorf("clear %s birthdays: %w", source, err)
	}
	for _, b := range list {
		if err := save(ctx, tx, b); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Settings implements usecase.Repository; no row means defaults.
func (r *Repository) Settings(ctx context.Context) (usecase.Settings, error) {
	var s usecase.Settings
	var enabled bool
	var enc string
	err := r.db.QueryRowContext(ctx, `SELECT remind_enabled, remind_at, last_reminded, carddav_url, username, password_encrypted FROM birthday_settings WHERE id = 1`).
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
	_, err := r.db.ExecContext(ctx, `INSERT INTO birthday_settings (id, remind_enabled, remind_at, last_reminded, carddav_url, username, password_encrypted)
		VALUES (1, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET remind_enabled = excluded.remind_enabled, remind_at = excluded.remind_at,
		last_reminded = excluded.last_reminded, carddav_url = excluded.carddav_url, username = excluded.username, password_encrypted = excluded.password_encrypted`,
		s.Reminder.Enabled, s.Reminder.At, s.LastReminded, s.CardDAVURL, s.Username, enc)
	if err != nil {
		return fmt.Errorf("save birthday settings: %w", err)
	}
	return nil
}
