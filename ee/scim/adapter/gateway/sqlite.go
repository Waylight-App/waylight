// Package gateway stores the directory's view of members and groups (#396).
//
// Ring: Interface adapters (gateway).
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"waylight/ee/scim/usecase"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Store implements usecase.Store with SQLite.
type Store struct{ db *sql.DB }

// NewStore returns the store.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// User implements usecase.Store.
func (s *Store) User(ctx context.Context, accountID string) (usecase.UserRecord, bool, error) {
	r := usecase.UserRecord{AccountID: accountID}
	err := s.db.QueryRowContext(ctx, `SELECT user_name, external_id, display_name FROM scim_users WHERE account_id = ?`, accountID).
		Scan(&r.UserName, &r.ExternalID, &r.DisplayName)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.UserRecord{}, false, nil
	}
	if err != nil {
		return usecase.UserRecord{}, false, fmt.Errorf("read scim user: %w", err)
	}
	return r, true, nil
}

// SaveUser implements usecase.Store.
func (s *Store) SaveUser(ctx context.Context, r usecase.UserRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO scim_users (account_id, user_name, external_id, display_name) VALUES (?, ?, ?, ?)
		 ON CONFLICT (account_id) DO UPDATE SET user_name = excluded.user_name, external_id = excluded.external_id,
		   display_name = excluded.display_name`,
		r.AccountID, r.UserName, r.ExternalID, r.DisplayName)
	if err != nil {
		return fmt.Errorf("save scim user: %w", err)
	}
	return nil
}

// DeleteUser implements usecase.Store.
func (s *Store) DeleteUser(ctx context.Context, accountID string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM scim_users WHERE account_id = ?`, accountID); err != nil {
		return fmt.Errorf("delete scim user: %w", err)
	}
	return nil
}

// GroupExternalID implements usecase.Store.
func (s *Store) GroupExternalID(ctx context.Context, groupID string) (string, error) {
	var ext string
	err := s.db.QueryRowContext(ctx, `SELECT external_id FROM scim_groups WHERE group_id = ?`, groupID).Scan(&ext)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read scim group: %w", err)
	}
	return ext, nil
}

// SaveGroupExternalID implements usecase.Store.
func (s *Store) SaveGroupExternalID(ctx context.Context, groupID, externalID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO scim_groups (group_id, external_id) VALUES (?, ?)
		 ON CONFLICT (group_id) DO UPDATE SET external_id = excluded.external_id`, groupID, externalID)
	if err != nil {
		return fmt.Errorf("save scim group: %w", err)
	}
	return nil
}

// DeleteGroup implements usecase.Store.
func (s *Store) DeleteGroup(ctx context.Context, groupID string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM scim_groups WHERE group_id = ?`, groupID); err != nil {
		return fmt.Errorf("delete scim group: %w", err)
	}
	return nil
}
