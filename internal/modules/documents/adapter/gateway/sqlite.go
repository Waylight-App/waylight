// Package gateway stores documents and their files in SQLite.
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

	"waylight/internal/modules/documents/domain"
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

// Repository implements usecase.Repository. The files live in the database
// next to everything else, so every backup and restore carries them.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, title, purchased_on, warranty_end, note, file_name, content_type, size, owner_id, visibility, created_at`

func parseDay(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.DateOnly, s)
}

func scan(sc interface{ Scan(...any) error }) (domain.Document, error) {
	var d domain.Document
	var bought, end, owner, vis, created string
	if err := sc.Scan(&d.ID, &d.Title, &bought, &end, &d.Note, &d.File.Name, &d.File.ContentType, &d.File.Size, &owner, &vis, &created); err != nil {
		return d, err
	}
	var err error
	if d.PurchasedOn, err = parseDay(bought); err != nil {
		return d, err
	}
	if d.WarrantyEnd, err = parseDay(end); err != nil {
		return d, err
	}
	d.Access = kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}
	d.CreatedAt, err = sqlutil.ParseTime(created)
	return d, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Document, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM documents_items WHERE `+cond+` ORDER BY title`, args...)
	if err != nil {
		return nil, fmt.Errorf("query documents: %w", err)
	}
	defer rows.Close()
	var out []domain.Document
	for rows.Next() {
		d, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Document, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	d, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM documents_items WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Document{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Document{}, fmt.Errorf("select document: %w", err)
	}
	return d, nil
}

// Save implements usecase.Repository (upsert). A new document belongs to the
// principal; an existing one is changed only by its owner (ADR 0018). The
// file columns are left to SaveFile.
func (r *Repository) Save(ctx context.Context, d domain.Document) error {
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "documents_items")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO documents_items (id, title, purchased_on, warranty_end, note, owner_id, visibility, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET title = excluded.title, purchased_on = excluded.purchased_on,
		warranty_end = excluded.warranty_end, note = excluded.note, visibility = excluded.visibility
		WHERE `+owned,
		append([]any{d.ID, d.Title, domain.Day(d.PurchasedOn), domain.Day(d.WarrantyEnd), d.Note, p.AccountID,
			string(d.Access.Visibility), sqlutil.FormatTime(d.CreatedAt)}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	return nil
}

// SaveFile implements usecase.Repository for the owner.
func (r *Repository) SaveFile(ctx context.Context, id string, f domain.File, data []byte) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := tx.ExecContext(ctx, `UPDATE documents_items SET file_name = ?, content_type = ?, size = ? WHERE id = ? AND `+owned,
		append([]any{f.Name, f.ContentType, f.Size, id}, args...)...)
	if err != nil {
		return fmt.Errorf("update file info: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO documents_files (item_id, data) VALUES (?, ?)
		ON CONFLICT (item_id) DO UPDATE SET data = excluded.data`, id, data); err != nil {
		return fmt.Errorf("store file: %w", err)
	}
	return tx.Commit()
}

// File implements usecase.Repository.
func (r *Repository) File(ctx context.Context, id string) ([]byte, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	var data []byte
	err := r.db.QueryRowContext(ctx, `SELECT data FROM documents_files WHERE item_id IN (SELECT id FROM documents_items WHERE id = ? AND `+cond+`)`,
		append([]any{id}, args...)...).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNoFile
	}
	if err != nil {
		return nil, fmt.Errorf("select file: %w", err)
	}
	return data, nil
}

// Delete implements usecase.Repository for the owner; a visible document of
// someone else returns kernel.ErrNotOwner, an unknown one is no error. The
// file goes with the document.
func (r *Repository) Delete(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	if _, err := tx.ExecContext(ctx, `DELETE FROM documents_files WHERE item_id IN (SELECT id FROM documents_items WHERE id = ? AND `+owned+`)`,
		append([]any{id}, args...)...); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := tx.ExecContext(ctx, `DELETE FROM documents_items WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
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
