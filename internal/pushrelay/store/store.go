// Package store keeps the device token bindings of the push relay in SQLite.
package store

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"time"

	"waylight/internal/platform/database"
	"waylight/internal/pushrelay"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Store implements pushrelay.Bindings.
type Store struct{ db *sql.DB }

// Open opens the database at path and applies the migrations.
func Open(ctx context.Context, path string) (*Store, error) {
	db, err := database.Open(ctx, path)
	if err != nil {
		return nil, err
	}
	sub, _ := fs.Sub(migrations, "migrations") // embedded, so Sub cannot fail
	if err := database.Migrate(ctx, db, "pushrelay", sub); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// Bind keeps the first instance of a token and records the use.
func (s *Store) Bind(ctx context.Context, h pushrelay.TokenHash, key string, now time.Time) (string, error) {
	var bound string
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO bindings (token_hash, instance, last_used) VALUES (?, ?, ?)
		ON CONFLICT (token_hash) DO UPDATE SET last_used = excluded.last_used
		RETURNING instance`, h[:], key, now.Unix()).Scan(&bound)
	return bound, err
}

// Delete removes a binding, e.g. after FCM or APNs dropped the token.
func (s *Store) Delete(ctx context.Context, h pushrelay.TokenHash) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM bindings WHERE token_hash = ?`, h[:])
	return err
}

// Prune deletes bindings unused since before and returns how many.
func (s *Store) Prune(ctx context.Context, before time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM bindings WHERE last_used < ?`, before.Unix())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
