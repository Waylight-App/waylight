// Package gateway stores pictures in SQLite and scales uploads (#286).
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

	"trisse-home/internal/modules/photos/domain"
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

// Repository implements usecase.Repository. Pictures live in the database,
// so they are part of every backup without extra steps.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context, album string) ([]domain.Photo, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	rows, err := r.db.QueryContext(ctx,
		//nolint:gosec // the condition comes from sqlutil with bound arguments
		`SELECT id, album, position, width, height, size, created_at FROM photos WHERE album = ? AND `+cond+` ORDER BY position, created_at`,
		append([]any{album}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("list photos: %w", err)
	}
	defer rows.Close()
	var out []domain.Photo
	for rows.Next() {
		var p domain.Photo
		var created string
		if err := rows.Scan(&p.ID, &p.Album, &p.Position, &p.Width, &p.Height, &p.Size, &created); err != nil {
			return nil, fmt.Errorf("scan photo: %w", err)
		}
		p.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		out = append(out, p)
	}
	return out, rows.Err()
}

// Albums implements usecase.Repository.
func (r *Repository) Albums(ctx context.Context) ([]domain.Album, error) {
	// Albums as the principal sees them: only visible photos count.
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT album, COUNT(*), SUM(size) FROM photos WHERE `+cond+` GROUP BY album ORDER BY album COLLATE NOCASE`, args...)
	if err != nil {
		return nil, fmt.Errorf("list albums: %w", err)
	}
	defer rows.Close()
	var out []domain.Album
	for rows.Next() {
		var a domain.Album
		if err := rows.Scan(&a.Name, &a.Count, &a.Bytes); err != nil {
			return nil, fmt.Errorf("scan album: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Usage implements usecase.Repository.
//
//archtest:all-owners the storage limit applies to the whole instance
func (r *Repository) Usage(ctx context.Context) (domain.Usage, error) {
	var u domain.Usage
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(size), 0) FROM photos`).Scan(&u.Count, &u.Bytes)
	if err != nil {
		return u, fmt.Errorf("photo usage: %w", err)
	}
	return u, nil
}

// Add implements usecase.Repository.
func (r *Repository) Add(ctx context.Context, p domain.Photo, image, thumb []byte) (domain.Photo, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return p, err
	}
	defer func() { _ = tx.Rollback() }()
	// The new photo goes after the principal's own photos in the album.
	owner := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(owner, "")
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), -1) + 1 FROM photos WHERE album = ? AND `+owned,
		append([]any{p.Album}, ownedArgs...)...).Scan(&p.Position); err != nil {
		return p, fmt.Errorf("next position: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO photos (id, album, position, width, height, size, created_at, owner_id, visibility) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Album, p.Position, p.Width, p.Height, p.Size, p.CreatedAt.UTC().Format(time.RFC3339Nano),
		owner.AccountID, string(kernel.DefaultVisibility(owner))); err != nil {
		return p, fmt.Errorf("insert photo: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO photo_files (photo_id, image, thumb) VALUES (?, ?, ?)`, p.ID, image, thumb); err != nil {
		return p, fmt.Errorf("insert photo file: %w", err)
	}
	return p, tx.Commit()
}

// Delete implements usecase.Repository for the photo's owner; the file goes
// with the row (cascade). A visible photo of someone else returns
// kernel.ErrNotOwner.
func (r *Repository) Delete(ctx context.Context, id string) error {
	p := kernel.PrincipalFrom(ctx)
	owned, args := sqlutil.OwnedBy(p, "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM photos WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete photo: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		cond, vargs := sqlutil.VisibleTo(p, "")
		var one int
		if r.db.QueryRowContext(ctx, `SELECT 1 FROM photos WHERE id = ? AND `+cond, append([]any{id}, vargs...)...).Scan(&one) == nil {
			return kernel.ErrNotOwner
		}
		return domain.ErrNotFound
	}
	return nil
}

// Reorder implements usecase.Repository.
func (r *Repository) Reorder(ctx context.Context, album string, ids []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	// Only the principal's own photos move; others keep their place.
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	for i, id := range ids {
		//nolint:gosec // the condition comes from sqlutil with bound arguments
		if _, err := tx.ExecContext(ctx, `UPDATE photos SET position = ? WHERE id = ? AND album = ? AND `+owned, append([]any{i, id, album}, args...)...); err != nil {
			return fmt.Errorf("reorder photos: %w", err)
		}
	}
	return tx.Commit()
}

// Image implements usecase.Repository.
func (r *Repository) Image(ctx context.Context, id string, thumb bool) ([]byte, error) {
	col := "image"
	if thumb {
		col = "thumb"
	}
	var data []byte
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	err := r.db.QueryRowContext(ctx, `SELECT `+col+` FROM photo_files WHERE photo_id = ? AND photo_id IN (SELECT id FROM photos WHERE `+cond+`)`,
		append([]any{id}, args...)...).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read photo: %w", err)
	}
	return data, nil
}
