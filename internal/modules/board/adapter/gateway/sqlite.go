// Package gateway stores the family board in SQLite.
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

	"waylight/internal/modules/board/domain"
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
var TransferTables = []sqlutil.Table{{Name: "board_notes"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, text, pinned, expires_on, owner_id, visibility, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Note, error) {
	var n domain.Note
	var pinned int
	var expires, owner, vis, created string
	if err := sc.Scan(&n.ID, &n.Text, &pinned, &expires, &owner, &vis, &created); err != nil {
		return n, err
	}
	n.Pinned = pinned == 1
	var err error
	if expires != "" {
		if n.ExpiresOn, err = time.Parse(time.DateOnly, expires); err != nil {
			return n, err
		}
	}
	n.Access = kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}
	n.CreatedAt, err = sqlutil.ParseTime(created)
	return n, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Note, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM board_notes WHERE `+cond+` ORDER BY pinned DESC, created_at DESC`, args...)
	if err != nil {
		return nil, fmt.Errorf("query notes: %w", err)
	}
	defer rows.Close()
	var out []domain.Note
	for rows.Next() {
		n, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Note, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	n, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM board_notes WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Note{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("select note: %w", err)
	}
	return n, nil
}

// Save implements usecase.Repository (upsert). A new note belongs to the
// principal; an existing one is changed only by its owner (ADR 0018).
func (r *Repository) Save(ctx context.Context, n domain.Note) error {
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "board_notes")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO board_notes (id, text, pinned, expires_on, owner_id, visibility, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET text = excluded.text, pinned = excluded.pinned, expires_on = excluded.expires_on,
		visibility = excluded.visibility
		WHERE `+owned,
		append([]any{n.ID, n.Text, sqlutil.BoolToInt(n.Pinned), n.ExpiryDate(), p.AccountID, string(n.Access.Visibility), sqlutil.FormatTime(n.CreatedAt)}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert note: %w", err)
	}
	if k, _ := res.RowsAffected(); k == 0 {
		return kernel.ErrNotOwner
	}
	return nil
}

// Delete implements usecase.Repository for the owner; a visible note of
// someone else returns kernel.ErrNotOwner, an unknown one is no error.
func (r *Repository) Delete(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM board_notes WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	if k, _ := res.RowsAffected(); k == 0 {
		if _, err := r.Get(ctx, id); err == nil {
			return kernel.ErrNotOwner
		}
	}
	return nil
}

// DeleteExpired implements usecase.Repository.
//
//archtest:all-owners the daily cleanup removes expired notes of every account
func (r *Repository) DeleteExpired(ctx context.Context, day time.Time) (int, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM board_notes WHERE expires_on != '' AND expires_on < ?`, day.Format(time.DateOnly))
	if err != nil {
		return 0, fmt.Errorf("delete expired notes: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
