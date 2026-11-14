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
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, album, position, width, height, size, created_at FROM photos WHERE album = ? ORDER BY position, created_at`, album)
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
	rows, err := r.db.QueryContext(ctx, `SELECT album, COUNT(*), SUM(size) FROM photos GROUP BY album ORDER BY album COLLATE NOCASE`)
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
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), -1) + 1 FROM photos WHERE album = ?`, p.Album).Scan(&p.Position); err != nil {
		return p, fmt.Errorf("next position: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO photos (id, album, position, width, height, size, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Album, p.Position, p.Width, p.Height, p.Size, p.CreatedAt.UTC().Format(time.RFC3339Nano)); err != nil {
		return p, fmt.Errorf("insert photo: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO photo_files (photo_id, image, thumb) VALUES (?, ?, ?)`, p.ID, image, thumb); err != nil {
		return p, fmt.Errorf("insert photo file: %w", err)
	}
	return p, tx.Commit()
}

// Delete implements usecase.Repository; the file goes with the row (cascade).
func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM photos WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete photo: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
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
	for i, id := range ids {
		if _, err := tx.ExecContext(ctx, `UPDATE photos SET position = ? WHERE id = ? AND album = ?`, i, id, album); err != nil {
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
	err := r.db.QueryRowContext(ctx, `SELECT `+col+` FROM photo_files WHERE photo_id = ?`, id).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read photo: %w", err)
	}
	return data, nil
}
