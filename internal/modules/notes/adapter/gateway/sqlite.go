package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/notes/domain"
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

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, title, body, pinned, created_at, updated_at, version`

func scan(s interface{ Scan(...any) error }) (domain.Note, error) {
	var id, title, body, created, updated string
	var pinned, version int
	if err := s.Scan(&id, &title, &body, &pinned, &created, &updated, &version); err != nil {
		return domain.Note{}, err
	}
	c, err := sqlutil.ParseTime(created)
	if err != nil {
		return domain.Note{}, err
	}
	u, err := sqlutil.ParseTime(updated)
	if err != nil {
		return domain.Note{}, err
	}
	return domain.RestoreNote(id, title, body, pinned == 1, c, u).WithVersion(version), nil
}

// List returns the notes the principal may see, pinned first, then by last
// edit (ADR 0018).
func (r *Repository) List(ctx context.Context) ([]domain.Note, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // cond comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM notes WHERE `+cond+` ORDER BY pinned DESC, updated_at DESC`, args...)
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

// Get returns one note the principal may see; others are not found.
func (r *Repository) Get(ctx context.Context, id string) (domain.Note, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	n, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM notes WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Note{}, domain.ErrNoteNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("select note: %w", err)
	}
	return n, nil
}

// Save inserts a note owned by the principal or updates one it owns. Owner
// and visibility of an existing note never change here.
func (r *Repository) Save(ctx context.Context, n domain.Note) error {
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "notes")
	// A change is only stored on top of the version it was made from (#80),
	// and only by the note's owner (ADR 0018).
	//nolint:gosec // owned comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO notes (`+columns+`, owner_id, visibility) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET title = excluded.title, body = excluded.body,
			pinned = excluded.pinned, updated_at = excluded.updated_at, version = excluded.version
		WHERE notes.version = excluded.version - 1 AND `+owned,
		append([]any{n.ID(), n.Title(), n.Body(), sqlutil.BoolToInt(n.Pinned()), sqlutil.FormatTime(n.CreatedAt()), sqlutil.FormatTime(n.UpdatedAt()), n.Version(),
			p.AccountID, string(kernel.DefaultVisibility(p))}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert note: %w", err)
	}
	if count, err := res.RowsAffected(); err == nil && count == 0 {
		return r.whyUnchanged(ctx, n.ID())
	}
	return nil
}

// whyUnchanged tells a version conflict from a note the principal does not
// own, after an update changed no row.
func (r *Repository) whyUnchanged(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	var one int
	err := r.db.QueryRowContext(ctx, `SELECT 1 FROM notes WHERE id = ? AND `+owned, append([]any{id}, args...)...).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return kernel.ErrNotOwner
	}
	if err != nil {
		return fmt.Errorf("check note owner: %w", err)
	}
	return domain.ErrVersionConflict
}

// Delete removes a note the principal owns. Deleting a note it can only
// see returns kernel.ErrNotOwner; an unknown note is not an error.
func (r *Repository) Delete(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // owned comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := r.Get(ctx, id); err == nil {
			return kernel.ErrNotOwner
		}
	}
	return nil
}
