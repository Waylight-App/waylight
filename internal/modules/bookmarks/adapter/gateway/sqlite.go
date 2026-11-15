package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"trisse-home/internal/modules/bookmarks/domain"
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

// bookmarkRow is the persistence model of a bookmark.
type bookmarkRow struct {
	id, title, url, description, createdAt string
	folderID                               sql.NullString
	tags                                   sql.NullString
	ownerID, visibility                    string
}

func (r bookmarkRow) toDomain() (domain.Bookmark, error) {
	created, err := sqlutil.ParseTime(r.createdAt)
	if err != nil {
		return domain.Bookmark{}, err
	}
	var tags []string
	if r.tags.Valid && r.tags.String != "" {
		tags = strings.Split(r.tags.String, ",")
	}
	return domain.RestoreBookmark(r.id, r.folderID.String, r.title, r.url, r.description, tags, created).
		WithAccess(kernel.Access{Owner: r.ownerID, Visibility: kernel.Visibility(r.visibility)}), nil
}

const bookmarkQuery = `SELECT b.id, b.folder_id, b.title, b.url, b.description, b.created_at,
	(SELECT group_concat(tag, ',') FROM (SELECT tag FROM bookmark_tags t WHERE t.bookmark_id = b.id ORDER BY tag)),
	b.owner_id, b.visibility
	FROM bookmarks b`

func scanBookmark(s interface{ Scan(...any) error }) (domain.Bookmark, error) {
	var r bookmarkRow
	if err := s.Scan(&r.id, &r.folderID, &r.title, &r.url, &r.description, &r.createdAt, &r.tags, &r.ownerID, &r.visibility); err != nil {
		return domain.Bookmark{}, err
	}
	return r.toDomain()
}

// ListBookmarks returns the bookmarks the principal may see, by title.
func (r *Repository) ListBookmarks(ctx context.Context) ([]domain.Bookmark, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "b")
	rows, err := r.db.QueryContext(ctx, bookmarkQuery+` WHERE `+cond+` ORDER BY b.title COLLATE NOCASE`, args...)
	if err != nil {
		return nil, fmt.Errorf("query bookmarks: %w", err)
	}
	defer rows.Close()
	var out []domain.Bookmark
	for rows.Next() {
		b, err := scanBookmark(rows)
		if err != nil {
			return nil, fmt.Errorf("scan bookmark: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// GetBookmark returns one bookmark.
func (r *Repository) GetBookmark(ctx context.Context, id string) (domain.Bookmark, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "b")
	b, err := scanBookmark(r.db.QueryRowContext(ctx, bookmarkQuery+` WHERE b.id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Bookmark{}, domain.ErrBookmarkNotFound
	}
	if err != nil {
		return domain.Bookmark{}, fmt.Errorf("select bookmark: %w", err)
	}
	return b, nil
}

// SaveBookmark inserts a bookmark owned by the principal or updates one it
// owns, and replaces its tags. A bookmark can only be filed into one of
// the principal's own folders (ADR 0018).
func (r *Repository) SaveBookmark(ctx context.Context, b domain.Bookmark) error {
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "")
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var folder any
	if b.FolderID() != "" {
		folder = b.FolderID()
	}
	ownedRow, rowArgs := sqlutil.OwnedBy(p, "bookmarks")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := tx.ExecContext(ctx, `INSERT INTO bookmarks (id, folder_id, title, url, description, created_at, owner_id, visibility)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET folder_id = excluded.folder_id, title = excluded.title,
			url = excluded.url, description = excluded.description
		WHERE `+ownedRow,
		append([]any{b.ID(), folder, b.Title(), b.URL(), b.Description(), sqlutil.FormatTime(b.CreatedAt()), p.AccountID, string(kernel.DefaultVisibility(p))}, rowArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert bookmark: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	// Checked after the upsert, in the same transaction, so a foreign
	// bookmark is refused as such first.
	if b.FolderID() != "" {
		var one int
		err := tx.QueryRowContext(ctx, `SELECT 1 FROM bookmark_folders WHERE id = ? AND `+owned, append([]any{b.FolderID()}, ownedArgs...)...).Scan(&one)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrFolderNotFound
		}
		if err != nil {
			return fmt.Errorf("check folder: %w", err)
		}
		// A filed bookmark is seen by whoever may see its folder (#298).
		if _, err := tx.ExecContext(ctx, `UPDATE bookmarks SET visibility = (SELECT visibility FROM bookmark_folders WHERE id = ?) WHERE id = ?`,
			b.FolderID(), b.ID()); err != nil {
			return fmt.Errorf("follow folder visibility: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM bookmark_tags WHERE bookmark_id = ?`, b.ID()); err != nil {
		return fmt.Errorf("clear tags: %w", err)
	}
	for _, t := range b.Tags() {
		if _, err := tx.ExecContext(ctx, `INSERT INTO bookmark_tags (bookmark_id, tag) VALUES (?, ?)`, b.ID(), t); err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}
	return tx.Commit()
}

// ReplaceAll implements usecase.ConfigRepository in one transaction. It
// works on the principal's own folders and bookmarks (config as code passes
// the admin, ADR 0018) and never takes over another account's IDs.
func (r *Repository) ReplaceAll(ctx context.Context, folders []domain.Folder, bookmarks []domain.Bookmark) error {
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "")
	vis := string(kernel.DefaultVisibility(p))
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	keepB, keepF := map[string]bool{}, map[string]bool{}
	for _, b := range bookmarks {
		keepB[b.ID()] = true
	}
	for _, f := range folders {
		keepF[f.ID()] = true
	}
	if err := deleteMissing(ctx, tx, `SELECT id FROM bookmarks WHERE `+owned, ownedArgs, `DELETE FROM bookmarks WHERE id = ?`, keepB); err != nil {
		return err
	}
	if err := deleteMissing(ctx, tx, `SELECT id FROM bookmark_folders WHERE `+owned, ownedArgs, `DELETE FROM bookmark_folders WHERE id = ?`, keepF); err != nil {
		return err
	}
	ownedFolder, folderArgs := sqlutil.OwnedBy(p, "bookmark_folders")
	for _, f := range folders {
		//nolint:gosec // the condition comes from sqlutil with bound arguments
		res, err := tx.ExecContext(ctx, `INSERT INTO bookmark_folders (id, name, position, owner_id, visibility) VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (id) DO UPDATE SET name = excluded.name, position = excluded.position WHERE `+ownedFolder,
			append([]any{f.ID(), f.Name(), f.Position(), p.AccountID, vis}, folderArgs...)...)
		if err != nil {
			return fmt.Errorf("upsert folder: %w", err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return kernel.ErrNotOwner
		}
	}
	ownedBookmark, bookmarkArgs := sqlutil.OwnedBy(p, "bookmarks")
	for _, b := range bookmarks {
		var folder any
		if b.FolderID() != "" {
			folder = b.FolderID()
		}
		//nolint:gosec // the condition comes from sqlutil with bound arguments
		res, err := tx.ExecContext(ctx, `INSERT INTO bookmarks (id, folder_id, title, url, description, created_at, owner_id, visibility)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (id) DO UPDATE SET folder_id = excluded.folder_id, title = excluded.title,
				url = excluded.url, description = excluded.description WHERE `+ownedBookmark,
			append([]any{b.ID(), folder, b.Title(), b.URL(), b.Description(), sqlutil.FormatTime(b.CreatedAt()), p.AccountID, vis}, bookmarkArgs...)...)
		if err != nil {
			return fmt.Errorf("upsert bookmark: %w", err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return kernel.ErrNotOwner
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM bookmark_tags WHERE bookmark_id = ?`, b.ID()); err != nil {
			return fmt.Errorf("clear tags: %w", err)
		}
		for _, t := range b.Tags() {
			if _, err := tx.ExecContext(ctx, `INSERT INTO bookmark_tags (bookmark_id, tag) VALUES (?, ?)`, b.ID(), t); err != nil {
				return fmt.Errorf("insert tag: %w", err)
			}
		}
	}
	return tx.Commit()
}

// deleteMissing deletes the ids returned by selectQuery (with args) that
// are not in keep, one by one with deleteQuery.
func deleteMissing(ctx context.Context, tx *sql.Tx, selectQuery string, args []any, deleteQuery string, keep map[string]bool) error {
	ids, err := queryIDs(ctx, tx, selectQuery, args...)
	if err != nil {
		return fmt.Errorf("list ids: %w", err)
	}
	for _, id := range ids {
		if keep[id] {
			continue
		}
		if _, err := tx.ExecContext(ctx, deleteQuery, id); err != nil {
			return fmt.Errorf("delete: %w", err)
		}
	}
	return nil
}

// queryIDs runs a query that selects one id column.
func queryIDs(ctx context.Context, tx *sql.Tx, query string, args ...any) ([]string, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// DeleteBookmark removes an own bookmark (tags cascade); a visible bookmark
// of someone else returns kernel.ErrNotOwner.
func (r *Repository) DeleteBookmark(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM bookmarks WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete bookmark: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := r.GetBookmark(ctx, id); err == nil {
			return kernel.ErrNotOwner
		}
	}
	return nil
}

// ListFolders returns the folders the principal may see by position.
func (r *Repository) ListFolders(ctx context.Context) ([]domain.Folder, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, position, owner_id, visibility FROM bookmark_folders WHERE `+cond+` ORDER BY position, name`, args...)
	if err != nil {
		return nil, fmt.Errorf("query folders: %w", err)
	}
	defer rows.Close()
	var out []domain.Folder
	for rows.Next() {
		var id, name, owner, vis string
		var pos int
		if err := rows.Scan(&id, &name, &pos, &owner, &vis); err != nil {
			return nil, fmt.Errorf("scan folder: %w", err)
		}
		out = append(out, domain.RestoreFolder(id, name, pos).WithAccess(kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}))
	}
	return out, rows.Err()
}

// GetFolder returns one folder.
func (r *Repository) GetFolder(ctx context.Context, id string) (domain.Folder, error) {
	var name, owner, vis string
	var pos int
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	err := r.db.QueryRowContext(ctx, `SELECT name, position, owner_id, visibility FROM bookmark_folders WHERE id = ? AND `+cond, append([]any{id}, args...)...).Scan(&name, &pos, &owner, &vis)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Folder{}, domain.ErrFolderNotFound
	}
	if err != nil {
		return domain.Folder{}, fmt.Errorf("select folder: %w", err)
	}
	return domain.RestoreFolder(id, name, pos).WithAccess(kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}), nil
}

// SetFolderVisibility implements usecase.Repository: the owner shares a
// folder, and the bookmarks in it follow in the same transaction (#298).
func (r *Repository) SetFolderVisibility(ctx context.Context, id string, v kernel.Visibility) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := tx.ExecContext(ctx, `UPDATE bookmark_folders SET visibility = ? WHERE id = ? AND `+owned, append([]any{string(v), id}, args...)...)
	if err != nil {
		return fmt.Errorf("update folder visibility: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		// End the transaction first: SQLite has one writer, and GetFolder
		// reads on its own connection.
		_ = tx.Rollback()
		if _, err := r.GetFolder(ctx, id); err != nil {
			return err
		}
		return kernel.ErrNotOwner
	}
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	if _, err := tx.ExecContext(ctx, `UPDATE bookmarks SET visibility = ? WHERE folder_id = ? AND `+owned, append([]any{string(v), id}, args...)...); err != nil {
		return fmt.Errorf("update bookmark visibility: %w", err)
	}
	return tx.Commit()
}

// SaveFolder inserts a folder owned by the principal or updates one it owns.
func (r *Repository) SaveFolder(ctx context.Context, f domain.Folder) error {
	p := kernel.PrincipalFrom(ctx)
	owned, args := sqlutil.OwnedBy(p, "bookmark_folders")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO bookmark_folders (id, name, position, owner_id, visibility) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, position = excluded.position WHERE `+owned,
		append([]any{f.ID(), f.Name(), f.Position(), p.AccountID, string(kernel.DefaultVisibility(p))}, args...)...)
	if err != nil {
		return fmt.Errorf("upsert folder: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	return nil
}

// DeleteFolder removes an own folder; ON DELETE SET NULL unfiles its
// bookmarks, which can only be the owner's own (see SaveBookmark).
func (r *Repository) DeleteFolder(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM bookmark_folders WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := r.GetFolder(ctx, id); err == nil {
			return kernel.ErrNotOwner
		}
	}
	return nil
}
