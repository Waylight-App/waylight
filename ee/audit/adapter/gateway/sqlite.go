// Package gateway stores the audit log (#303).
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
	"strings"
	"time"

	"trisse-home/ee/audit/domain"
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
type Store struct{ db *sql.DB }

// NewStore returns the store.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// Append implements usecase.Store.
func (s *Store) Append(ctx context.Context, e domain.Entry) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_entries (id, at, account_id, role, module, route, object_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.ID, sqlutil.FormatTime(e.At), e.AccountID, e.Role, e.Module, e.Route, e.ObjectID)
	if err != nil {
		return fmt.Errorf("insert audit entry: %w", err)
	}
	return nil
}

// List implements usecase.Store; Limit 0 returns everything.
func (s *Store) List(ctx context.Context, f domain.Filter) ([]domain.Entry, error) {
	var where []string
	var args []any
	if f.AccountID != "" {
		where, args = append(where, "account_id = ?"), append(args, f.AccountID)
	}
	if f.Module != "" {
		where, args = append(where, "module = ?"), append(args, f.Module)
	}
	if !f.From.IsZero() {
		where, args = append(where, "at >= ?"), append(args, sqlutil.FormatTime(f.From))
	}
	if !f.To.IsZero() {
		where, args = append(where, "at < ?"), append(args, sqlutil.FormatTime(f.To))
	}
	q := `SELECT id, at, account_id, role, module, route, object_id FROM audit_entries`
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ") //nolint:gosec // fixed clauses; values are bound parameters
	}
	q += " ORDER BY at DESC, id DESC"
	if f.Limit > 0 {
		q, args = q+" LIMIT ?", append(args, f.Limit)
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list audit entries: %w", err)
	}
	defer rows.Close()
	out := []domain.Entry{}
	for rows.Next() {
		var e domain.Entry
		var at string
		if err := rows.Scan(&e.ID, &at, &e.AccountID, &e.Role, &e.Module, &e.Route, &e.ObjectID); err != nil {
			return nil, fmt.Errorf("scan audit entry: %w", err)
		}
		if e.At, err = sqlutil.ParseTime(at); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// DeleteBefore implements usecase.Store.
func (s *Store) DeleteBefore(ctx context.Context, t time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM audit_entries WHERE at < ?`, sqlutil.FormatTime(t))
	if err != nil {
		return 0, fmt.Errorf("prune audit entries: %w", err)
	}
	return res.RowsAffected()
}

// Retention implements usecase.Store.
func (s *Store) Retention(ctx context.Context) (int, error) {
	var days int
	err := s.db.QueryRowContext(ctx, `SELECT retention_days FROM audit_settings WHERE id = 1`).Scan(&days)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DefaultRetentionDays, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read retention: %w", err)
	}
	return days, nil
}

// SetRetention implements usecase.Store.
func (s *Store) SetRetention(ctx context.Context, days int) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_settings (id, retention_days) VALUES (1, ?) ON CONFLICT (id) DO UPDATE SET retention_days = excluded.retention_days`, days)
	if err != nil {
		return fmt.Errorf("save retention: %w", err)
	}
	return nil
}
