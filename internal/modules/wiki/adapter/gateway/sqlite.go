// Package gateway stores wiki connections and asks the wikis (#321).
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/wiki/domain"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Cipher encrypts tokens before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// Repository implements usecase.Repository on SQLite.
type Repository struct {
	db     *sql.DB
	cipher Cipher
}

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB, cipher Cipher) *Repository { return &Repository{db: db, cipher: cipher} }

func (r *Repository) scan(row interface{ Scan(...any) error }) (domain.Wiki, error) {
	var w domain.Wiki
	var kind, enc string
	if err := row.Scan(&w.ID, &w.Name, &kind, &w.URL, &enc); err != nil {
		return w, err
	}
	w.Kind = domain.Kind(kind)
	token, err := r.cipher.Decrypt(enc)
	if err != nil {
		return w, fmt.Errorf("decrypt wiki token: %w", err)
	}
	w.Token = token
	return w, nil
}

const columns = `id, name, kind, url, token_encrypted`

// List implements usecase.Repository, oldest first.
func (r *Repository) List(ctx context.Context) ([]domain.Wiki, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM wiki_connections ORDER BY created_at, name`)
	if err != nil {
		return nil, fmt.Errorf("list wikis: %w", err)
	}
	defer rows.Close()
	var out []domain.Wiki
	for rows.Next() {
		w, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Wiki, error) {
	w, err := r.scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM wiki_connections WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Wiki{}, domain.ErrNotFound
	}
	return w, err
}

// Save implements usecase.Repository (insert or update).
func (r *Repository) Save(ctx context.Context, w domain.Wiki) error {
	enc, err := r.cipher.Encrypt(w.Token)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO wiki_connections (`+columns+`) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, kind = excluded.kind, url = excluded.url,
		token_encrypted = excluded.token_encrypted`, w.ID, w.Name, string(w.Kind), w.URL, enc)
	if err != nil {
		return fmt.Errorf("save wiki: %w", err)
	}
	return nil
}

// Delete implements usecase.Repository.
func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM wiki_connections WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete wiki: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
