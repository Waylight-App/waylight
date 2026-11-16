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
var TransferTables = []sqlutil.Table{{Name: "phonelist_entries"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, name, role, phone, email, team, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Entry, error) {
	var e domain.Entry
	var created string
	if err := sc.Scan(&e.ID, &e.Name, &e.Role, &e.Phone, &e.Email, &e.Team, &created); err != nil {
		return e, err
	}
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
	_, err := r.db.ExecContext(ctx, `INSERT INTO phonelist_entries (`+columns+`) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, role = excluded.role, phone = excluded.phone,
			email = excluded.email, team = excluded.team`,
		e.ID, e.Name, e.Role, e.Phone, e.Email, e.Team, sqlutil.FormatTime(e.CreatedAt))
	if err != nil {
		return fmt.Errorf("upsert entry: %w", err)
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
