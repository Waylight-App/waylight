// Package gateway stores live presence in SQLite (#346).
//
// Ring: Interface adapters (gateway).
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"

	"trisse-home/ee/presence/domain"
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

// Entries implements usecase.Store.
func (s *Store) Entries(ctx context.Context) ([]domain.Entry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT account_id, status, day, return_on FROM presence_status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Entry
	for rows.Next() {
		var e domain.Entry
		var status string
		if err := rows.Scan(&e.AccountID, &status, &e.Day, &e.ReturnOn); err != nil {
			return nil, err
		}
		e.Status = domain.Status(status)
		out = append(out, e)
	}
	return out, rows.Err()
}

// Save implements usecase.Store.
func (s *Store) Save(ctx context.Context, e domain.Entry) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO presence_status (account_id, status, day, return_on) VALUES (?, ?, ?, ?)
		ON CONFLICT (account_id) DO UPDATE SET status = excluded.status, day = excluded.day, return_on = excluded.return_on`,
		e.AccountID, string(e.Status), e.Day, e.ReturnOn)
	return err
}

// Delete implements usecase.Store; deleting a missing status is no error.
func (s *Store) Delete(ctx context.Context, accountID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM presence_status WHERE account_id = ?`, accountID)
	return err
}

// DeleteBefore implements usecase.Store.
func (s *Store) DeleteBefore(ctx context.Context, day string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM presence_status WHERE day < ?`, day)
	return err
}

// Enabled implements usecase.Store.
func (s *Store) Enabled(ctx context.Context) (bool, error) {
	var on int
	err := s.db.QueryRowContext(ctx, `SELECT enabled FROM presence_settings WHERE id = 1`).Scan(&on)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	return on == 1, err
}

// SetEnabled implements usecase.Store.
func (s *Store) SetEnabled(ctx context.Context, on bool) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO presence_settings (id, enabled) VALUES (1, ?)
		ON CONFLICT (id) DO UPDATE SET enabled = excluded.enabled`, sqlutil.BoolToInt(on))
	return err
}
