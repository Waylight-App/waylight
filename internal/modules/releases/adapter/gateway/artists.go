// Package gateway stores followed artists in SQLite and reads MusicBrainz
// and the Cover Art Archive.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/releases/domain"
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

// TransferTables are part of data exports; the follow list has no secrets.
var TransferTables = []sqlutil.Table{{Name: "releases_artists"}}

// ArtistRepository implements usecase.ArtistRepository.
type ArtistRepository struct{ db *sql.DB }

// NewArtistRepository returns a repository backed by db.
func NewArtistRepository(db *sql.DB) *ArtistRepository { return &ArtistRepository{db: db} }

// mine is the condition for the principal's followed artists. The list is
// personal like a watch list: an account sees only the artists it follows;
// guests and kiosk displays see what is shared with them (ADR 0018).
func mine(ctx context.Context) (string, []any) {
	p := kernel.PrincipalFrom(ctx)
	if p.TabScoped() {
		return sqlutil.VisibleTo(p, "")
	}
	return sqlutil.OwnedBy(p, "")
}

// List returns the followed artists in the order they were added. The
// background refresh (the system) gets every account's artists, possibly
// the same artist twice; callers deduplicate by ID.
func (r *ArtistRepository) List(ctx context.Context) ([]domain.Artist, error) {
	cond, args := mine(ctx)
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM releases_artists WHERE `+cond+` ORDER BY position`, args...)
	if err != nil {
		return nil, fmt.Errorf("query artists: %w", err)
	}
	defer rows.Close()
	var out []domain.Artist
	for rows.Next() {
		var a domain.Artist
		if err := rows.Scan(&a.ID, &a.Name); err != nil {
			return nil, fmt.Errorf("scan artist: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Add follows an artist for the principal (idempotent).
func (r *ArtistRepository) Add(ctx context.Context, a domain.Artist) error {
	p := kernel.PrincipalFrom(ctx)
	owned, args := sqlutil.OwnedBy(p, "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	_, err := r.db.ExecContext(ctx, `INSERT INTO releases_artists (owner_id, id, name, position, visibility)
		VALUES (?, ?, ?, (SELECT COALESCE(MAX(position), -1) + 1 FROM releases_artists WHERE `+owned+`), ?) ON CONFLICT (owner_id, id) DO NOTHING`,
		append(append([]any{p.AccountID, a.ID, a.Name}, args...), string(kernel.DefaultVisibility(p)))...)
	if err != nil {
		return fmt.Errorf("insert artist: %w", err)
	}
	return nil
}

// Remove deletes an artist.
func (r *ArtistRepository) Remove(ctx context.Context, id string) error {
	// Unfollowing only affects the principal's own list.
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	if _, err := r.db.ExecContext(ctx, `DELETE FROM releases_artists WHERE id = ? AND `+owned, append([]any{id}, args...)...); err != nil {
		return fmt.Errorf("delete artist: %w", err)
	}
	return nil
}
