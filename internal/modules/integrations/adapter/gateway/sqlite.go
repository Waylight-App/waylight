// Package gateway stores connections in SQLite with their credential
// encrypted.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	"waylight/internal/modules/integrations/domain"
	"waylight/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Cipher encrypts credentials with SECRET_KEY.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

const columns = `id, service, name, url, credential_encrypted, created_at`

// Repository implements usecase.ConfigRepository.
type Repository struct {
	db     *sql.DB
	cipher Cipher
}

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB, cipher Cipher) *Repository { return &Repository{db: db, cipher: cipher} }

// scan reads one row. A credential that cannot be decrypted (SECRET_KEY was
// changed) reads as empty, so the list still opens and the admin can enter
// it again instead of every connection failing.
func (r *Repository) scan(sc interface{ Scan(...any) error }) (domain.Connection, error) {
	var c domain.Connection
	var enc, created string
	if err := sc.Scan(&c.ID, &c.Service, &c.Name, &c.URL, &enc, &created); err != nil {
		return c, err
	}
	c.Credential = domain.Credential{}
	if plain, err := r.cipher.Decrypt(enc); err == nil {
		_ = json.Unmarshal([]byte(plain), &c.Credential)
	}
	var err error
	c.CreatedAt, err = sqlutil.ParseTime(created)
	return c, err
}

func (r *Repository) encrypt(c domain.Connection) (string, error) {
	plain, err := json.Marshal(c.Credential)
	if err != nil {
		return "", fmt.Errorf("encode credential: %w", err)
	}
	enc, err := r.cipher.Encrypt(string(plain))
	if err != nil {
		return "", fmt.Errorf("encrypt credential: %w", err)
	}
	return enc, nil
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Connection, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM integrations_connections ORDER BY service, name COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("query connections: %w", err)
	}
	defer rows.Close()
	var out []domain.Connection
	for rows.Next() {
		c, err := r.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan connection: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Connection, error) {
	c, err := r.scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM integrations_connections WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Connection{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Connection{}, fmt.Errorf("select connection: %w", err)
	}
	return c, nil
}

// upsert keeps created_at of an existing row.
const upsert = `INSERT INTO integrations_connections (` + columns + `) VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT (id) DO UPDATE SET service = excluded.service, name = excluded.name, url = excluded.url,
		credential_encrypted = excluded.credential_encrypted`

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (r *Repository) save(ctx context.Context, db execer, c domain.Connection) error {
	enc, err := r.encrypt(c)
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, upsert, c.ID, c.Service, c.Name, c.URL, enc, sqlutil.FormatTime(c.CreatedAt)); err != nil {
		return fmt.Errorf("upsert connection: %w", err)
	}
	return nil
}

// Save implements usecase.Repository (upsert).
func (r *Repository) Save(ctx context.Context, c domain.Connection) error {
	return r.save(ctx, r.db, c)
}

// Delete implements usecase.Repository; an unknown ID is no error.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM integrations_connections WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}
	return nil
}

// ReplaceAll implements usecase.ConfigRepository in one transaction. All
// rows are written again, so names may move between connections (a swap
// would trip the unique index in an update in place).
func (r *Repository) ReplaceAll(ctx context.Context, list []domain.Connection) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM integrations_connections`); err != nil {
		return fmt.Errorf("delete connections: %w", err)
	}
	for _, c := range list {
		if err := r.save(ctx, tx, c); err != nil {
			return err
		}
	}
	return tx.Commit()
}
