// Package gateway stores places in SQLite.
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

	"waylight/internal/modules/places/domain"
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

// TransferTables are the tables included in data exports (no secrets).
var TransferTables = []sqlutil.Table{{Name: "places_places"}}

const columns = `id, name, label, lat, lon, created_at`

// Repository implements usecase.Repository with ownership per record.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func scan(sc interface{ Scan(...any) error }) (domain.Place, error) {
	var p domain.Place
	var created string
	if err := sc.Scan(&p.ID, &p.Name, &p.Label, &p.Lat, &p.Lon, &created); err != nil {
		return p, err
	}
	var err error
	p.CreatedAt, err = sqlutil.ParseTime(created)
	return p, err
}

// List implements usecase.Repository: the places visible to the principal.
func (r *Repository) List(ctx context.Context) ([]domain.Place, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM places_places WHERE `+cond+` ORDER BY name COLLATE NOCASE, created_at`, args...)
	if err != nil {
		return nil, fmt.Errorf("query places: %w", err)
	}
	defer rows.Close()
	var out []domain.Place
	for rows.Next() {
		p, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan place: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Place, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	p, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM places_places WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Place{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Place{}, fmt.Errorf("select place: %w", err)
	}
	return p, nil
}

// Save implements usecase.Repository (upsert). A new place belongs to the
// principal; an existing one is changed only by its owner (ADR 0018).
func (r *Repository) Save(ctx context.Context, p domain.Place) error {
	pr := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(pr, "places_places")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO places_places (`+columns+`, owner_id, visibility) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, label = excluded.label, lat = excluded.lat, lon = excluded.lon
		WHERE `+owned,
		append([]any{p.ID, p.Name, p.Label, p.Lat, p.Lon, sqlutil.FormatTime(p.CreatedAt), pr.AccountID, string(kernel.DefaultVisibility(pr))}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert place: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	return nil
}

// Delete implements usecase.Repository for the owner; a visible place of
// someone else returns kernel.ErrNotOwner, an unknown one is no error.
func (r *Repository) Delete(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM places_places WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete place: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := r.Get(ctx, id); err == nil {
			return kernel.ErrNotOwner
		}
	}
	return nil
}
