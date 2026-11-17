package gateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"waylight/internal/modules/todos/domain"
	"waylight/internal/shared/kernel"
	"waylight/internal/shared/sqlutil"
)

// Cipher encrypts secrets before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// SourceRepository implements usecase.SourceRepository.
type SourceRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewSourceRepository returns a repository backed by db.
func NewSourceRepository(db *sql.DB, cipher Cipher) *SourceRepository {
	return &SourceRepository{db: db, cipher: cipher}
}

const sourceColumns = `id, name, kind, url, username, password_encrypted, list_id`

func (r *SourceRepository) scan(row interface{ Scan(...any) error }) (domain.Source, error) {
	var s domain.Source
	var kind, enc string
	if err := row.Scan(&s.ID, &s.Name, &kind, &s.URL, &s.Username, &enc, &s.ListID); err != nil {
		return domain.Source{}, err
	}
	s.Kind = domain.SourceKind(kind)
	pw, err := r.cipher.Decrypt(enc)
	if err != nil {
		return domain.Source{}, fmt.Errorf("decrypt password of list %s (was SECRET_KEY changed?): %w", s.ID, err)
	}
	if !s.Kind.OAuth() {
		s.Password = pw
		return s, nil
	}
	var g grantJSON
	if err := json.Unmarshal([]byte(pw), &g); err != nil {
		return domain.Source{}, fmt.Errorf("read sign-in of list %s: %w", s.ID, err)
	}
	s.Grant = domain.OAuthGrant{ClientSecret: g.ClientSecret, Access: g.Access, Refresh: g.Refresh, Expiry: g.Expiry}
	return s, nil
}

// grantJSON is how the secret part of a signed-in list is stored
// (encrypted as a whole, #171).
type grantJSON struct {
	ClientSecret string    `json:"clientSecret,omitempty"`
	Access       string    `json:"access"`
	Refresh      string    `json:"refresh,omitempty"`
	Expiry       time.Time `json:"expiry,omitzero"`
}

// secret is what is encrypted into password_encrypted.
func secret(s domain.Source) (string, error) {
	if !s.Kind.OAuth() {
		return s.Password, nil
	}
	g := s.Grant
	//nolint:gosec // the JSON is encrypted with SECRET_KEY right after (Save) and never leaves the server
	raw, err := json.Marshal(grantJSON{ClientSecret: g.ClientSecret, Access: g.Access, Refresh: g.Refresh, Expiry: g.Expiry})
	return string(raw), err
}

// mine is the condition for the connected to-do lists of the account whose settings
// apply (ADR 0018): an account's own, and the admin's for guests, kiosk
// displays and background work. Members get no fallback to the admin's.
func mine(ctx context.Context) (string, []any) {
	return sqlutil.OwnedBy(kernel.PrincipalFor(kernel.PrincipalFrom(ctx).SettingsAccount()), "")
}

// List implements usecase.SourceRepository (oldest first).
func (r *SourceRepository) List(ctx context.Context) ([]domain.Source, error) {
	cond, args := mine(ctx)
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+sourceColumns+` FROM todos_sources WHERE `+cond+` ORDER BY created_at, name`, args...)
	if err != nil {
		return nil, fmt.Errorf("select sources: %w", err)
	}
	defer rows.Close()
	var out []domain.Source
	for rows.Next() {
		s, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// Get implements usecase.SourceRepository.
func (r *SourceRepository) Get(ctx context.Context, id string) (domain.Source, error) {
	cond, args := mine(ctx)
	s, err := r.scan(r.db.QueryRowContext(ctx, `SELECT `+sourceColumns+` FROM todos_sources WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Source{}, domain.ErrSourceNotFound
	}
	return s, err
}

// Save implements usecase.SourceRepository (insert or update).
func (r *SourceRepository) Save(ctx context.Context, s domain.Source) error {
	plain, err := secret(s)
	if err != nil {
		return fmt.Errorf("encode secret: %w", err)
	}
	enc, err := r.cipher.Encrypt(plain)
	if err != nil {
		return fmt.Errorf("encrypt password: %w", err)
	}
	account := kernel.PrincipalFrom(ctx).SettingsAccount()
	owned, ownedArgs := sqlutil.OwnedBy(kernel.PrincipalFor(account), "todos_sources")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	_, err = r.db.ExecContext(ctx, `INSERT INTO todos_sources (id, name, kind, url, username, password_encrypted, list_id, owner_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, kind = excluded.kind, url = excluded.url,
		username = excluded.username, password_encrypted = excluded.password_encrypted, list_id = excluded.list_id
		WHERE `+owned,
		append([]any{s.ID, s.Name, string(s.Kind), s.URL, s.Username, enc, s.ListID, account}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert source: %w", err)
	}
	return nil
}

// Delete implements usecase.SourceRepository.
func (r *SourceRepository) Delete(ctx context.Context, id string) error {
	cond, args := mine(ctx)
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	if _, err := r.db.ExecContext(ctx, `DELETE FROM todos_sources WHERE id = ? AND `+cond, append([]any{id}, args...)...); err != nil {
		return fmt.Errorf("delete source: %w", err)
	}
	return nil
}
