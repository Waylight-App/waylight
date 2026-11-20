// Package store keeps the device token bindings and the mailbox of the push
// relay in SQLite.
package store

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"time"

	"waylight/internal/platform/database"
	"waylight/internal/pushrelay"
	"waylight/internal/pushrelay/stream"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Store implements pushrelay.Bindings and stream.Mailbox.
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

// Put implements stream.Mailbox: it stores the message and drops the
// oldest messages of the device beyond limit, in one transaction.
func (s *Store) Put(ctx context.Context, device string, payload []byte, now time.Time, limit int) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `INSERT INTO mailbox (device, payload, created) VALUES (?, ?, ?)`, device, payload, now.Unix())
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM mailbox WHERE device = ? AND id NOT IN (
			SELECT id FROM mailbox WHERE device = ? ORDER BY id DESC LIMIT ?)`, device, device, limit); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// After implements stream.Mailbox.
func (s *Store) After(ctx context.Context, device string, after int64) ([]stream.Event, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, payload FROM mailbox WHERE device = ? AND id > ? ORDER BY id`, device, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []stream.Event
	for rows.Next() {
		var e stream.Event
		if err := rows.Scan(&e.ID, &e.Payload); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Ack implements stream.Mailbox.
func (s *Store) Ack(ctx context.Context, device string, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM mailbox WHERE device = ? AND id <= ?`, device, id)
	return err
}

// PruneMailbox deletes messages stored before before and returns how many.
func (s *Store) PruneMailbox(ctx context.Context, before time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM mailbox WHERE created < ?`, before.Unix())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
