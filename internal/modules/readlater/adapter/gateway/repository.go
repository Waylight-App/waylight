// Package gateway stores articles in SQLite, fetches pages and images and
// extracts a sanitised reader copy.
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

	"waylight/internal/modules/readlater/domain"
	"waylight/internal/modules/readlater/usecase"
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

// TransferTables are exported with the copies but without the images
// (binary data does not fit the JSON export; backups include them).
var TransferTables = []sqlutil.Table{{Name: "readlater_articles"}, {Name: "readlater_settings"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, url, title, site, excerpt, content, words, status, saved_at, bytes, error, pending, status_at`

func scan(row interface{ Scan(...any) error }) (domain.Article, error) {
	var a domain.Article
	var status, saved string
	var statusAt *string
	if err := row.Scan(&a.ID, &a.URL, &a.Title, &a.Site, &a.Excerpt, &a.Content, &a.Words, &status, &saved, &a.Bytes, &a.Error, &a.Pending, &statusAt); err != nil {
		return domain.Article{}, err
	}
	a.Status = domain.Status(status)
	var err error
	if a.SavedAt, err = sqlutil.ParseTime(saved); err != nil {
		return domain.Article{}, err
	}
	a.StatusAt, err = sqlutil.ParseNullableTime(statusAt)
	return a, err
}

// mine is the condition for the principal's reading list. The list is personal:
// an account sees only its own, also when the admin's are visible to
// everyone; guests and kiosk displays see what is shared (ADR 0018).
func mine(ctx context.Context) (string, []any) {
	p := kernel.PrincipalFrom(ctx)
	if p.TabScoped() {
		return sqlutil.VisibleTo(p, "")
	}
	return sqlutil.OwnedBy(p, "")
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Article, error) {
	cond, args := mine(ctx)
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM readlater_articles WHERE `+cond, args...)
	if err != nil {
		return nil, fmt.Errorf("select articles: %w", err)
	}
	defer rows.Close()
	var out []domain.Article
	for rows.Next() {
		a, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Article, error) {
	cond, args := mine(ctx)
	a, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM readlater_articles WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Article{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Article{}, fmt.Errorf("select article: %w", err)
	}
	return a, nil
}

// Save implements usecase.Repository: the article and its images in one transaction.
func (r *Repository) Save(ctx context.Context, a domain.Article, images map[int]usecase.Image) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	// Upsert: an imported article gets its reader copy later (#99). New
	// articles belong to the principal; the fetch job (the system) and the
	// owner update one (ADR 0018).
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "readlater_articles")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := tx.ExecContext(ctx, `INSERT INTO readlater_articles (`+columns+`, owner_id, visibility) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET title = excluded.title, site = excluded.site, excerpt = excluded.excerpt,
			content = excluded.content, words = excluded.words, bytes = excluded.bytes, error = excluded.error, pending = excluded.pending
		WHERE `+owned,
		append([]any{a.ID, a.URL, a.Title, a.Site, a.Excerpt, a.Content, a.Words, string(a.Status), sqlutil.FormatTime(a.SavedAt), a.Bytes, a.Error,
			a.Pending, sqlutil.FormatNullableTime(a.StatusAt), p.AccountID, string(kernel.DefaultVisibility(p))}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("save article: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM readlater_images WHERE article_id = ?`, a.ID); err != nil {
		return fmt.Errorf("replace images: %w", err)
	}
	for n, img := range images {
		if _, err = tx.ExecContext(ctx, `INSERT INTO readlater_images (article_id, n, content_type, data) VALUES (?, ?, ?, ?)`,
			a.ID, n, img.ContentType, img.Bytes); err != nil {
			return fmt.Errorf("insert image: %w", err)
		}
	}
	return tx.Commit()
}

// SetStatus implements usecase.Repository for the article's owner.
func (r *Repository) SetStatus(ctx context.Context, id string, s domain.Status, at time.Time) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `UPDATE readlater_articles SET status = ?, status_at = ? WHERE id = ? AND `+owned,
		append([]any{string(s), sqlutil.FormatTime(at), id}, args...)...)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete implements usecase.Repository for the owner (images go with the
// article).
func (r *Repository) Delete(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	if _, err := r.db.ExecContext(ctx, `DELETE FROM readlater_images WHERE article_id = ? AND article_id IN (SELECT id FROM readlater_articles WHERE `+owned+`)`,
		append([]any{id}, args...)...); err != nil {
		return fmt.Errorf("delete images: %w", err)
	}
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	if _, err := r.db.ExecContext(ctx, `DELETE FROM readlater_articles WHERE id = ? AND `+owned, append([]any{id}, args...)...); err != nil {
		return fmt.Errorf("delete article: %w", err)
	}
	return nil
}

// Image implements usecase.Repository.
func (r *Repository) Image(ctx context.Context, id string, n int) (usecase.Image, error) {
	var img usecase.Image
	cond, args := mine(ctx)
	err := r.db.QueryRowContext(ctx, `SELECT content_type, data FROM readlater_images WHERE article_id = ? AND n = ?
		AND article_id IN (SELECT id FROM readlater_articles WHERE `+cond+`)`, append([]any{id, n}, args...)...).Scan(&img.ContentType, &img.Bytes)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.Image{}, domain.ErrNotFound
	}
	if err != nil {
		return usecase.Image{}, fmt.Errorf("select image: %w", err)
	}
	return img, nil
}

// TotalBytes implements usecase.Repository.
//
//archtest:all-owners the storage limit applies to the whole instance
func (r *Repository) TotalBytes(ctx context.Context) (int64, error) {
	var n sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `SELECT SUM(bytes) FROM readlater_articles`).Scan(&n); err != nil {
		return 0, fmt.Errorf("sum bytes: %w", err)
	}
	return n.Int64, nil
}

// Pending implements usecase.Repository.
func (r *Repository) Pending(ctx context.Context, limit int) ([]domain.Article, error) {
	cond, args := mine(ctx)
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM readlater_articles WHERE pending = 1 AND `+cond+` ORDER BY saved_at LIMIT ?`, append(args, limit)...)
	if err != nil {
		return nil, fmt.Errorf("select pending: %w", err)
	}
	defer rows.Close()
	var out []domain.Article
	for rows.Next() {
		a, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Settings implements usecase.Repository; defaults without a row.
func (r *Repository) Settings(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	// The own setting, or the admin's (ADR 0018).
	account := kernel.PrincipalFrom(ctx).SettingsAccount()
	err := r.db.QueryRowContext(ctx, `SELECT archive_days FROM readlater_settings WHERE account_id IN (?, ?) ORDER BY account_id = ? DESC LIMIT 1`,
		account, kernel.SystemAccountID, account).Scan(&s.ArchiveDays)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, nil
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	return s, nil
}

// SaveSettings implements usecase.Repository.
func (r *Repository) SaveSettings(ctx context.Context, s domain.Settings) error {
	if _, err := r.db.ExecContext(ctx, `INSERT INTO readlater_settings (account_id, archive_days) VALUES (?, ?)
		ON CONFLICT (account_id) DO UPDATE SET archive_days = excluded.archive_days`, kernel.PrincipalFrom(ctx).SettingsAccount(), s.ArchiveDays); err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}

// Owners implements usecase.Repository: the accounts that have articles, so
// the cleanup can apply each account's own setting.
func (r *Repository) Owners(ctx context.Context) ([]string, error) {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT owner_id FROM readlater_articles WHERE `+owned+` ORDER BY owner_id`, args...)
	if err != nil {
		return nil, fmt.Errorf("select owners: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan owner: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
