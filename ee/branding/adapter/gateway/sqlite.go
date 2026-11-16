// Package gateway stores the branding in SQLite (#305); the images are in
// the database, so backups and restores include them.
//
// Ring: Interface adapters (gateway).
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"time"

	"trisse-home/ee/branding/domain"
	"trisse-home/ee/branding/usecase"
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

// Settings implements usecase.Store; no row means nothing set.
func (s *Store) Settings(ctx context.Context) (domain.Settings, error) {
	var set domain.Settings
	var hide int
	err := s.db.QueryRowContext(ctx, `SELECT name, accent, secondary, hide_powered_by FROM branding_settings WHERE id = 1`).
		Scan(&set.Name, &set.Accent, &set.Secondary, &hide)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, nil
	}
	set.HidePoweredBy = hide == 1
	return set, err
}

// SaveSettings implements usecase.Store.
func (s *Store) SaveSettings(ctx context.Context, set domain.Settings) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO branding_settings (id, name, accent, secondary, hide_powered_by) VALUES (1, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, accent = excluded.accent, secondary = excluded.secondary,
		hide_powered_by = excluded.hide_powered_by`, set.Name, set.Accent, set.Secondary, sqlutil.BoolToInt(set.HidePoweredBy))
	return err
}

// Asset implements usecase.Store.
func (s *Store) Asset(ctx context.Context, kind string) (usecase.Asset, bool, error) {
	var a usecase.Asset
	var at string
	err := s.db.QueryRowContext(ctx, `SELECT data, content_type, updated_at FROM branding_assets WHERE kind = ?`, kind).
		Scan(&a.Data, &a.ContentType, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.Asset{}, false, nil
	}
	if err != nil {
		return usecase.Asset{}, false, err
	}
	if a.UpdatedAt, err = sqlutil.ParseTime(at); err != nil {
		return usecase.Asset{}, false, err
	}
	return a, true, nil
}

// Assets implements usecase.Store.
func (s *Store) Assets(ctx context.Context) (map[string]time.Time, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT kind, updated_at FROM branding_assets`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]time.Time{}
	for rows.Next() {
		var kind, at string
		if err := rows.Scan(&kind, &at); err != nil {
			return nil, err
		}
		t, err := sqlutil.ParseTime(at)
		if err != nil {
			return nil, err
		}
		out[kind] = t
	}
	return out, rows.Err()
}

// SaveAsset implements usecase.Store.
func (s *Store) SaveAsset(ctx context.Context, kind string, a usecase.Asset) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO branding_assets (kind, data, content_type, updated_at) VALUES (?, ?, ?, ?)
		ON CONFLICT (kind) DO UPDATE SET data = excluded.data, content_type = excluded.content_type, updated_at = excluded.updated_at`,
		kind, a.Data, a.ContentType, sqlutil.FormatTime(a.UpdatedAt))
	return err
}

// DeleteAsset implements usecase.Store.
func (s *Store) DeleteAsset(ctx context.Context, kind string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM branding_assets WHERE kind = ?`, kind)
	return err
}
