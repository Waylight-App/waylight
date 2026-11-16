// Package gateway stores phone list entries in SQLite.
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

	"trisse-home/internal/modules/phonelist/domain"
	"trisse-home/internal/modules/phonelist/usecase"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are the tables included in data exports (no secrets).
var TransferTables = []sqlutil.Table{
	{Name: "phonelist_entries"},
	{Name: "phonelist_settings", Exclude: []string{"password_encrypted"}},
}

// Cipher encrypts the address book password with the instance key.
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

const columns = `id, name, role, phone, email, team, created_at, source`

func scan(sc interface{ Scan(...any) error }) (domain.Entry, error) {
	var e domain.Entry
	var created, source string
	if err := sc.Scan(&e.ID, &e.Name, &e.Role, &e.Phone, &e.Email, &e.Team, &created, &source); err != nil {
		return e, err
	}
	e.Source = domain.Source(source)
	var err error
	e.CreatedAt, err = sqlutil.ParseTime(created)
	return e, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM phonelist_entries ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("query entries: %w", err)
	}
	defer rows.Close()
	out := []domain.Entry{}
	for rows.Next() {
		e, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan entry: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Entry, error) {
	e, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM phonelist_entries WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Entry{}, domain.ErrEntryNotFound
	}
	if err != nil {
		return domain.Entry{}, fmt.Errorf("select entry: %w", err)
	}
	return e, nil
}

// Save implements usecase.Repository (upsert).
func (r *Repository) Save(ctx context.Context, e domain.Entry) error {
	return save(ctx, r.db, e)
}

// save upserts an entry with db or a transaction.
func save(ctx context.Context, db interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, e domain.Entry) error {
	source := e.Source
	if source == "" {
		source = domain.SourceManual
	}
	_, err := db.ExecContext(ctx, `INSERT INTO phonelist_entries (`+columns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, role = excluded.role, phone = excluded.phone,
			email = excluded.email, team = excluded.team, source = excluded.source`,
		e.ID, e.Name, e.Role, e.Phone, e.Email, e.Team, sqlutil.FormatTime(e.CreatedAt), string(source))
	if err != nil {
		return fmt.Errorf("upsert entry: %w", err)
	}
	return nil
}

// ReplaceSource implements usecase.Repository.
func (r *Repository) ReplaceSource(ctx context.Context, source domain.Source, entries []domain.Entry) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM phonelist_entries WHERE source = ?`, string(source)); err != nil {
		return fmt.Errorf("clear %s entries: %w", source, err)
	}
	for _, e := range entries {
		e.Source = source
		if err := save(ctx, tx, e); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Settings implements usecase.Repository.
func (r *Repository) Settings(ctx context.Context) (usecase.Settings, error) {
	var s usecase.Settings
	var enc, last string
	err := r.db.QueryRowContext(ctx, `SELECT carddav_url, username, password_encrypted, last_sync, last_error FROM phonelist_settings WHERE id = 1`).
		Scan(&s.CardDAVURL, &s.Username, &enc, &last, &s.LastError)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.Settings{}, nil
	}
	if err != nil {
		return s, fmt.Errorf("read phone list settings: %w", err)
	}
	if enc != "" {
		if s.Password, err = r.cipher.Decrypt(enc); err != nil {
			return s, fmt.Errorf("decrypt CardDAV password: %w", err)
		}
	}
	if last != "" {
		if s.LastSync, err = sqlutil.ParseTime(last); err != nil {
			return s, err
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
	last := ""
	if !s.LastSync.IsZero() {
		last = sqlutil.FormatTime(s.LastSync)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO phonelist_settings (id, carddav_url, username, password_encrypted, last_sync, last_error)
		VALUES (1, ?, ?, ?, ?, ?) ON CONFLICT (id) DO UPDATE SET carddav_url = excluded.carddav_url, username = excluded.username,
		password_encrypted = excluded.password_encrypted, last_sync = excluded.last_sync, last_error = excluded.last_error`,
		s.CardDAVURL, s.Username, enc, last, s.LastError)
	if err != nil {
		return fmt.Errorf("save phone list settings: %w", err)
	}
	return nil
}

// Delete implements usecase.Repository.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM phonelist_entries WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete entry: %w", err)
	}
	return nil
}
