// Package gateway stores meters and their readings in SQLite.
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

	"waylight/internal/modules/meters/domain"
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

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, name, kind, unit, owner_id, visibility, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Meter, error) {
	var m domain.Meter
	var kind, owner, vis, created string
	if err := sc.Scan(&m.ID, &m.Name, &kind, &m.Unit, &owner, &vis, &created); err != nil {
		return m, err
	}
	m.Kind = domain.Kind(kind)
	m.Access = kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}
	var err error
	m.CreatedAt, err = sqlutil.ParseTime(created)
	return m, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Meter, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM meters_items WHERE `+cond+` ORDER BY name, id`, args...)
	if err != nil {
		return nil, fmt.Errorf("query meters: %w", err)
	}
	defer rows.Close()
	var out []domain.Meter
	for rows.Next() {
		m, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan meter: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Meter, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	m, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM meters_items WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Meter{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Meter{}, fmt.Errorf("select meter: %w", err)
	}
	return m, nil
}

// Save implements usecase.Repository (upsert). A new meter belongs to the
// principal; an existing one is changed only by its owner (ADR 0018).
func (r *Repository) Save(ctx context.Context, m domain.Meter) error {
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "meters_items")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO meters_items (id, name, kind, unit, owner_id, visibility, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, kind = excluded.kind, unit = excluded.unit,
		visibility = excluded.visibility
		WHERE `+owned,
		append([]any{m.ID, m.Name, string(m.Kind), m.Unit, p.AccountID, string(m.Access.Visibility), sqlutil.FormatTime(m.CreatedAt)}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert meter: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	return nil
}

// Delete implements usecase.Repository for the owner; a visible meter of
// someone else returns kernel.ErrNotOwner, an unknown one is no error.
func (r *Repository) Delete(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	if _, err := tx.ExecContext(ctx, `DELETE FROM meters_readings WHERE meter_id IN (SELECT id FROM meters_items WHERE id = ? AND `+owned+`)`,
		append([]any{id}, args...)...); err != nil {
		return fmt.Errorf("delete readings: %w", err)
	}
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := tx.ExecContext(ctx, `DELETE FROM meters_items WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete meter: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := r.Get(ctx, id); err == nil {
			return kernel.ErrNotOwner
		}
	}
	return nil
}

// LastReadings implements usecase.Repository.
func (r *Repository) LastReadings(ctx context.Context) (map[string]domain.Reading, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "i")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT r.meter_id, r.day, r.value FROM meters_readings r
		JOIN meters_items i ON i.id = r.meter_id
		WHERE `+cond+` AND r.day = (SELECT MAX(day) FROM meters_readings WHERE meter_id = r.meter_id)`, args...)
	if err != nil {
		return nil, fmt.Errorf("query last readings: %w", err)
	}
	defer rows.Close()
	out := map[string]domain.Reading{}
	for rows.Next() {
		var id, day string
		var rd domain.Reading
		if err := rows.Scan(&id, &day, &rd.Value); err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		if rd.Day, err = domain.ParseDay("day", day); err != nil {
			return nil, err
		}
		out[id] = rd
	}
	return out, rows.Err()
}

// Readings implements usecase.Repository.
func (r *Repository) Readings(ctx context.Context, id string) ([]domain.Reading, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "i")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT r.day, r.value FROM meters_readings r
		JOIN meters_items i ON i.id = r.meter_id
		WHERE r.meter_id = ? AND `+cond+` ORDER BY r.day`, append([]any{id}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("query readings: %w", err)
	}
	defer rows.Close()
	var out []domain.Reading
	for rows.Next() {
		var day string
		var rd domain.Reading
		if err := rows.Scan(&day, &rd.Value); err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		if rd.Day, err = domain.ParseDay("day", day); err != nil {
			return nil, err
		}
		out = append(out, rd)
	}
	return out, rows.Err()
}

// SaveReading implements usecase.Repository for everyone who sees the meter.
func (r *Repository) SaveReading(ctx context.Context, id string, rd domain.Reading) error {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "meters_items")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO meters_readings (meter_id, day, value)
		SELECT ?, ?, ? WHERE EXISTS (SELECT 1 FROM meters_items WHERE id = ? AND `+cond+`)
		ON CONFLICT (meter_id, day) DO UPDATE SET value = excluded.value`,
		append([]any{id, domain.Day(rd.Day), rd.Value, id}, args...)...)
	if err != nil {
		return fmt.Errorf("upsert reading: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// DeleteReading implements usecase.Repository for the owner of the meter.
func (r *Repository) DeleteReading(ctx context.Context, id string, day time.Time) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM meters_readings WHERE meter_id = ? AND day = ?
		AND meter_id IN (SELECT id FROM meters_items WHERE `+owned+`)`,
		append([]any{id, domain.Day(day)}, args...)...)
	if err != nil {
		return fmt.Errorf("delete reading: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNoReading
	}
	return nil
}
