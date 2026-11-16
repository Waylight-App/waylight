// Package gateway stores announcements in SQLite (#297).
//
// Ring: Interface adapters (gateway).
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"trisse-home/ee/announcements/domain"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Store implements usecase.Store.
type Store struct{ db *sql.DB }

// NewStore returns the store.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

const columns = `id, title, body, pinned, author_id, expires_at, created_at, updated_at`

func scan(sc interface{ Scan(...any) error }) (domain.Announcement, error) {
	var a domain.Announcement
	var pinned int
	var expires, created, updated string
	if err := sc.Scan(&a.ID, &a.Title, &a.Body, &pinned, &a.AuthorID, &expires, &created, &updated); err != nil {
		return domain.Announcement{}, err
	}
	a.Pinned = pinned == 1
	var err error
	if expires != "" {
		if a.ExpiresAt, err = sqlutil.ParseTime(expires); err != nil {
			return domain.Announcement{}, err
		}
	}
	if a.CreatedAt, err = sqlutil.ParseTime(created); err != nil {
		return domain.Announcement{}, err
	}
	if a.UpdatedAt, err = sqlutil.ParseTime(updated); err != nil {
		return domain.Announcement{}, err
	}
	return a, nil
}

func formatOptional(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return sqlutil.FormatTime(t)
}

// List implements usecase.Store.
func (s *Store) List(ctx context.Context) ([]domain.Announcement, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+columns+` FROM announcements`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Announcement
	for rows.Next() {
		a, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Get implements usecase.Store.
func (s *Store) Get(ctx context.Context, id string) (domain.Announcement, error) {
	a, err := scan(s.db.QueryRowContext(ctx, `SELECT `+columns+` FROM announcements WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Announcement{}, domain.ErrNotFound
	}
	return a, err
}

// Save implements usecase.Store.
func (s *Store) Save(ctx context.Context, a domain.Announcement) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO announcements (`+columns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET title = excluded.title, body = excluded.body, pinned = excluded.pinned,
		expires_at = excluded.expires_at, updated_at = excluded.updated_at`,
		a.ID, a.Title, a.Body, sqlutil.BoolToInt(a.Pinned), a.AuthorID, formatOptional(a.ExpiresAt),
		sqlutil.FormatTime(a.CreatedAt), sqlutil.FormatTime(a.UpdatedAt))
	return err
}

// Delete implements usecase.Store; read marks go with it.
func (s *Store) Delete(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM announcement_reads WHERE announcement_id = ?`, id); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM announcements WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return tx.Commit()
}

// ReadBy implements usecase.Store.
func (s *Store) ReadBy(ctx context.Context, accountID string) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT announcement_id FROM announcement_reads WHERE account_id = ?`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// MarkRead implements usecase.Store; marking twice keeps the first time.
func (s *Store) MarkRead(ctx context.Context, accountID, id string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO announcement_reads (announcement_id, account_id, read_at) VALUES (?, ?, ?)
		ON CONFLICT DO NOTHING`, id, accountID, sqlutil.FormatTime(at))
	return err
}

// ForgetAccount implements usecase.Store.
func (s *Store) ForgetAccount(ctx context.Context, accountID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM announcement_reads WHERE account_id = ?`, accountID)
	return err
}

// Settings implements usecase.Store; no row means admins only.
func (s *Store) Settings(ctx context.Context) (domain.Settings, error) {
	var set domain.Settings
	err := s.db.QueryRowContext(ctx, `SELECT writer_group FROM announcement_settings WHERE id = 1`).Scan(&set.WriterGroup)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, nil
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("load announcement settings: %w", err)
	}
	return set, nil
}

// SaveSettings implements usecase.Store.
func (s *Store) SaveSettings(ctx context.Context, set domain.Settings) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO announcement_settings (id, writer_group) VALUES (1, ?)
		ON CONFLICT (id) DO UPDATE SET writer_group = excluded.writer_group`, set.WriterGroup)
	return err
}
