// Package gateway stores the admin policies (#304).
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

	"trisse-home/ee/policies/domain"
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

// Load implements usecase.Store.
func (s *Store) Load(ctx context.Context) (domain.Settings, error) {
	var set domain.Settings
	var second int
	var modules sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT require_second_factor, min_password_length, max_session_hours, member_modules FROM policies WHERE id = 1`).
		Scan(&second, &set.MinPasswordLength, &set.MaxSessionHours, &modules)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, nil
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("read policies: %w", err)
	}
	set.RequireSecondFactor = second == 1
	if modules.Valid {
		set.MemberModules = []string{}
		if err := json.Unmarshal([]byte(modules.String), &set.MemberModules); err != nil {
			return domain.Settings{}, fmt.Errorf("decode member modules: %w", err)
		}
	}
	return set, nil
}

// Save implements usecase.Store.
func (s *Store) Save(ctx context.Context, set domain.Settings) error {
	var modules sql.NullString
	if set.MemberModules != nil {
		b, err := json.Marshal(set.MemberModules)
		if err != nil {
			return fmt.Errorf("encode member modules: %w", err)
		}
		modules = sql.NullString{String: string(b), Valid: true}
	}
	second := 0
	if set.RequireSecondFactor {
		second = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO policies (id, require_second_factor, min_password_length, max_session_hours, member_modules) VALUES (1, ?, ?, ?, ?)
		 ON CONFLICT (id) DO UPDATE SET require_second_factor = excluded.require_second_factor,
		   min_password_length = excluded.min_password_length, max_session_hours = excluded.max_session_hours,
		   member_modules = excluded.member_modules`,
		second, set.MinPasswordLength, set.MaxSessionHours, modules)
	if err != nil {
		return fmt.Errorf("save policies: %w", err)
	}
	return nil
}
