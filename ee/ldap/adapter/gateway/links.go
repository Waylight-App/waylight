// Package gateway stores links and talks to the directory (#396).
//
// Ring: Interface adapters (gateway).
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	"waylight/ee/ldap/usecase"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Links implements usecase.Links with SQLite.
type Links struct{ db *sql.DB }

// NewLinks returns the store.
func NewLinks(db *sql.DB) *Links { return &Links{db: db} }

// Get implements usecase.Links.
func (s *Links) Get(ctx context.Context, name string) (usecase.Link, bool, error) {
	l := usecase.Link{Name: name}
	var groups string
	var disabled int
	err := s.db.QueryRowContext(ctx, `SELECT account_id, groups, disabled_by_dir FROM ldap_links WHERE name = ?`, name).
		Scan(&l.AccountID, &groups, &disabled)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.Link{}, false, nil
	}
	if err != nil {
		return usecase.Link{}, false, fmt.Errorf("read link: %w", err)
	}
	if err := json.Unmarshal([]byte(groups), &l.Groups); err != nil {
		return usecase.Link{}, false, fmt.Errorf("decode groups: %w", err)
	}
	l.DisabledByDir = disabled == 1
	return l, true, nil
}

// Save implements usecase.Links.
func (s *Links) Save(ctx context.Context, l usecase.Link) error {
	groups := []byte("[]")
	if l.Groups != nil {
		var err error
		if groups, err = json.Marshal(l.Groups); err != nil {
			return fmt.Errorf("encode groups: %w", err)
		}
	}
	disabled := 0
	if l.DisabledByDir {
		disabled = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ldap_links (name, account_id, groups, disabled_by_dir) VALUES (?, ?, ?, ?)
		 ON CONFLICT (name) DO UPDATE SET account_id = excluded.account_id, groups = excluded.groups,
		   disabled_by_dir = excluded.disabled_by_dir`,
		l.Name, l.AccountID, string(groups), disabled)
	if err != nil {
		return fmt.Errorf("save link: %w", err)
	}
	return nil
}

// Delete implements usecase.Links.
func (s *Links) Delete(ctx context.Context, name string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM ldap_links WHERE name = ?`, name); err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	return nil
}
