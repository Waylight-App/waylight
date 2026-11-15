// Package gateway stores countdowns in SQLite.
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

	"trisse-home/internal/modules/countdown/domain"
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

// TransferTables are the tables included in data exports (no secrets).
var TransferTables = []sqlutil.Table{{Name: "countdown_items"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, title, emoji, day, minute, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Countdown, error) {
	var c domain.Countdown
	var day, created string
	if err := sc.Scan(&c.ID, &c.Title, &c.Emoji, &day, &c.Minute, &created); err != nil {
		return c, err
	}
	var err error
	if c.Day, err = time.Parse(time.DateOnly, day); err != nil {
		return c, err
	}
	c.CreatedAt, err = sqlutil.ParseTime(created)
	return c, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Countdown, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM countdown_items WHERE `+cond+` ORDER BY day, minute`, args...)
	if err != nil {
		return nil, fmt.Errorf("query countdowns: %w", err)
	}
	defer rows.Close()
	var out []domain.Countdown
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan countdown: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Countdown, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	c, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM countdown_items WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Countdown{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Countdown{}, fmt.Errorf("select countdown: %w", err)
	}
	return c, nil
}

// Save implements usecase.Repository (upsert). New countdowns belong to the
// principal; an existing one is changed only by its owner (ADR 0018).
func (r *Repository) Save(ctx context.Context, c domain.Countdown) error {
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "countdown_items")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO countdown_items (`+columns+`, owner_id, visibility) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET title = excluded.title, emoji = excluded.emoji, day = excluded.day, minute = excluded.minute
		WHERE `+owned,
		append([]any{c.ID, c.Title, c.Emoji, c.Date(), c.Minute, sqlutil.FormatTime(c.CreatedAt), p.AccountID, string(kernel.DefaultVisibility(p))}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert countdown: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	return nil
}

// Delete implements usecase.Repository for the owner; a visible countdown
// of someone else returns kernel.ErrNotOwner, an unknown one is no error.
func (r *Repository) Delete(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM countdown_items WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete countdown: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := r.Get(ctx, id); err == nil {
			return kernel.ErrNotOwner
		}
	}
	return nil
}
