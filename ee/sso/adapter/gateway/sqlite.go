// Package gateway stores the links between provider subjects and members (#302).
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

	"trisse-home/ee/sso/usecase"
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

// Links implements usecase.Links with SQLite.
type Links struct {
	db    *sql.DB
	clock kernel.Clock
}

// NewLinks returns the store.
func NewLinks(db *sql.DB, clock kernel.Clock) *Links { return &Links{db: db, clock: clock} }

// Get implements usecase.Links.
func (s *Links) Get(ctx context.Context, subject string) (usecase.Link, bool, error) {
	var l usecase.Link
	var groups string
	var disabled int
	err := s.db.QueryRowContext(ctx, `SELECT subject, account_id, groups, disabled_by_sso FROM sso_links WHERE subject = ?`, subject).
		Scan(&l.Subject, &l.AccountID, &groups, &disabled)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.Link{}, false, nil
	}
	if err != nil {
		return usecase.Link{}, false, fmt.Errorf("read link: %w", err)
	}
	if err := json.Unmarshal([]byte(groups), &l.Groups); err != nil {
		return usecase.Link{}, false, fmt.Errorf("decode groups: %w", err)
	}
	l.DisabledBySSO = disabled == 1
	return l, true, nil
}

// Save implements usecase.Links.
func (s *Links) Save(ctx context.Context, l usecase.Link) error {
	groups, err := json.Marshal(l.Groups)
	if err != nil {
		return fmt.Errorf("encode groups: %w", err)
	}
	if l.Groups == nil {
		groups = []byte("[]")
	}
	disabled := 0
	if l.DisabledBySSO {
		disabled = 1
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO sso_links (subject, account_id, groups, disabled_by_sso, updated_at) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (subject) DO UPDATE SET account_id = excluded.account_id, groups = excluded.groups,
		   disabled_by_sso = excluded.disabled_by_sso, updated_at = excluded.updated_at`,
		l.Subject, l.AccountID, string(groups), disabled, sqlutil.FormatTime(s.clock.Now()))
	if err != nil {
		return fmt.Errorf("save link: %w", err)
	}
	return nil
}

// Delete implements usecase.Links.
func (s *Links) Delete(ctx context.Context, subject string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM sso_links WHERE subject = ?`, subject); err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	return nil
}
