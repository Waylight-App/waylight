// Package gateway stores absences in SQLite (#299).
//
// Ring: Interface adapters (gateway).
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"

	"trisse-home/ee/absences/domain"
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

const columns = `id, account_id, from_day, to_day, kind, note, private, created_by, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Absence, error) {
	var a domain.Absence
	var kind, created string
	var private int
	if err := sc.Scan(&a.ID, &a.AccountID, &a.From, &a.To, &kind, &a.Note, &private, &a.CreatedBy, &created); err != nil {
		return domain.Absence{}, err
	}
	a.Kind, a.Private = domain.Kind(kind), private == 1
	t, err := sqlutil.ParseTime(created)
	if err != nil {
		return domain.Absence{}, err
	}
	a.CreatedAt = t
	return a, nil
}

// Between implements usecase.Store.
func (s *Store) Between(ctx context.Context, from, to string) ([]domain.Absence, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+columns+` FROM absences WHERE from_day <= ? AND to_day >= ?`, to, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Absence
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
func (s *Store) Get(ctx context.Context, id string) (domain.Absence, error) {
	a, err := scan(s.db.QueryRowContext(ctx, `SELECT `+columns+` FROM absences WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Absence{}, domain.ErrNotFound
	}
	return a, err
}

// Save implements usecase.Store.
func (s *Store) Save(ctx context.Context, a domain.Absence) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO absences (`+columns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET from_day = excluded.from_day, to_day = excluded.to_day, kind = excluded.kind,
		note = excluded.note, private = excluded.private`,
		a.ID, a.AccountID, a.From, a.To, string(a.Kind), a.Note, sqlutil.BoolToInt(a.Private), a.CreatedBy, sqlutil.FormatTime(a.CreatedAt))
	return err
}

// Delete implements usecase.Store.
func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM absences WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// DeleteAccount implements usecase.Store.
func (s *Store) DeleteAccount(ctx context.Context, accountID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM absences WHERE account_id = ?`, accountID)
	return err
}
