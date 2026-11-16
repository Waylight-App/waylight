// Package gateway stores the licence key (#307).
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

// Store implements usecase.Store with SQLite.
type Store struct {
	db    *sql.DB
	clock kernel.Clock
}

// NewStore returns the store.
func NewStore(db *sql.DB, clock kernel.Clock) *Store { return &Store{db: db, clock: clock} }

// Load implements usecase.Store.
func (s *Store) Load(ctx context.Context) (string, error) {
	var key string
	err := s.db.QueryRowContext(ctx, `SELECT key FROM licence_key WHERE id = 1`).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read licence key: %w", err)
	}
	return key, nil
}

// Save implements usecase.Store.
func (s *Store) Save(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO licence_key (id, key, updated_at) VALUES (1, ?, ?)
		 ON CONFLICT (id) DO UPDATE SET key = excluded.key, updated_at = excluded.updated_at`,
		key, sqlutil.FormatTime(s.clock.Now()))
	if err != nil {
		return fmt.Errorf("store licence key: %w", err)
	}
	return nil
}

// Delete implements usecase.Store.
func (s *Store) Delete(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM licence_key`); err != nil {
		return fmt.Errorf("delete licence key: %w", err)
	}
	return nil
}
