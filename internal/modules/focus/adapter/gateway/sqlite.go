// Package gateway stores focus sessions in SQLite.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"time"

	"trisse-home/internal/modules/focus/domain"
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
var TransferTables = []sqlutil.Table{{Name: "focus_sessions"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// mine is the condition for the sessions that make up the principal's
// statistics. Focus time is personal: an account counts only its own
// sessions, also when the admin's are visible to everyone. Guests and
// kiosk displays have none and see what is shared with them (ADR 0018).
func mine(ctx context.Context) (string, []any) {
	p := kernel.PrincipalFrom(ctx)
	if p.TabScoped() {
		return sqlutil.VisibleTo(p, "")
	}
	return sqlutil.OwnedBy(p, "")
}

// Save implements usecase.Repository; the session belongs to the principal.
func (r *Repository) Save(ctx context.Context, s domain.Session) error {
	p := kernel.PrincipalFrom(ctx)
	_, err := r.db.ExecContext(ctx, `INSERT INTO focus_sessions (id, minutes, ended_at, owner_id, visibility) VALUES (?, ?, ?, ?, ?)`,
		s.ID, s.Minutes, sqlutil.FormatTime(s.EndedAt), p.AccountID, string(kernel.DefaultVisibility(p)))
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// Since implements usecase.Repository. Stored times are UTC with a fixed
// width, so they compare correctly as text.
func (r *Repository) Since(ctx context.Context, t time.Time) ([]domain.Session, error) {
	cond, args := mine(ctx)
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT id, minutes, ended_at FROM focus_sessions WHERE ended_at >= ? AND `+cond+` ORDER BY ended_at`,
		append([]any{sqlutil.FormatTime(t)}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()
	var out []domain.Session
	for rows.Next() {
		var s domain.Session
		var ended string
		if err := rows.Scan(&s.ID, &s.Minutes, &ended); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		if s.EndedAt, err = sqlutil.ParseTime(ended); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// DeleteBefore implements usecase.Repository: the pruning job (the system)
// removes old sessions of everyone, an account only its own.
func (r *Repository) DeleteBefore(ctx context.Context, t time.Time) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	if _, err := r.db.ExecContext(ctx, `DELETE FROM focus_sessions WHERE ended_at < ? AND `+owned, append([]any{sqlutil.FormatTime(t)}, args...)...); err != nil {
		return fmt.Errorf("prune sessions: %w", err)
	}
	return nil
}
