// Package gateway stores mail accounts and reads IMAP inboxes (#283).
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

	"trisse-home/internal/modules/mail/domain"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Cipher encrypts secrets before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// Repository implements usecase.AccountRepository.
type Repository struct {
	db     *sql.DB
	cipher Cipher
}

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB, cipher Cipher) *Repository { return &Repository{db: db, cipher: cipher} }

const columns = `id, name, host, port, username, password_encrypted, folder, webmail_url`

func (r *Repository) scan(row interface{ Scan(...any) error }) (domain.Account, error) {
	var a domain.Account
	var enc string
	if err := row.Scan(&a.ID, &a.Name, &a.Host, &a.Port, &a.Username, &enc, &a.Folder, &a.WebmailURL); err != nil {
		return domain.Account{}, err
	}
	pw, err := r.cipher.Decrypt(enc)
	if err != nil {
		return domain.Account{}, fmt.Errorf("decrypt password of %s (was SECRET_KEY changed?): %w", a.ID, err)
	}
	a.Password = pw
	return a, nil
}

// List implements usecase.AccountRepository (oldest first).
func (r *Repository) List(ctx context.Context) ([]domain.Account, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM mail_accounts ORDER BY created_at, name`)
	if err != nil {
		return nil, fmt.Errorf("select accounts: %w", err)
	}
	defer rows.Close()
	var out []domain.Account
	for rows.Next() {
		a, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Get implements usecase.AccountRepository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Account, error) {
	a, err := r.scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM mail_accounts WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	return a, err
}

// Save implements usecase.AccountRepository (insert or update).
func (r *Repository) Save(ctx context.Context, a domain.Account) error {
	enc, err := r.cipher.Encrypt(a.Password)
	if err != nil {
		return fmt.Errorf("encrypt password: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO mail_accounts (id, name, host, port, username, password_encrypted, folder, webmail_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT (id) DO UPDATE SET name = excluded.name, host = excluded.host, port = excluded.port,
		username = excluded.username, password_encrypted = excluded.password_encrypted, folder = excluded.folder, webmail_url = excluded.webmail_url`,
		a.ID, a.Name, a.Host, a.Port, a.Username, enc, a.Folder, a.WebmailURL)
	if err != nil {
		return fmt.Errorf("upsert account: %w", err)
	}
	return nil
}

// Delete implements usecase.AccountRepository.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM mail_accounts WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}
