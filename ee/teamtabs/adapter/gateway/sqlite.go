// Package gateway stores the default team tab per group (#295).
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Repository implements usecase.Defaults.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// All implements usecase.Defaults: group ID → tab ID.
func (r *Repository) All(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT group_id, tab_id FROM teamtabs_defaults`)
	if err != nil {
		return nil, fmt.Errorf("query defaults: %w", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var g, t string
		if err := rows.Scan(&g, &t); err != nil {
			return nil, fmt.Errorf("scan default: %w", err)
		}
		out[g] = t
	}
	return out, rows.Err()
}

// Set implements usecase.Defaults; an empty tab ID removes the default.
func (r *Repository) Set(ctx context.Context, groupID, tabID string) error {
	var err error
	if tabID == "" {
		_, err = r.db.ExecContext(ctx, `DELETE FROM teamtabs_defaults WHERE group_id = ?`, groupID)
	} else {
		_, err = r.db.ExecContext(ctx, `INSERT INTO teamtabs_defaults (group_id, tab_id) VALUES (?, ?)
			ON CONFLICT (group_id) DO UPDATE SET tab_id = excluded.tab_id`, groupID, tabID)
	}
	if err != nil {
		return fmt.Errorf("save default: %w", err)
	}
	return nil
}
