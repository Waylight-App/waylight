// Package gateway stores the smart home settings in SQLite and talks to
// Home Assistant.
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
	"time"

	"waylight/internal/modules/smarthome/domain"
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

// SettingsRepository stores the settings with an encrypted token.
type SettingsRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewSettingsRepository returns a repository backed by db.
func NewSettingsRepository(db *sql.DB, cipher Cipher) *SettingsRepository {
	return &SettingsRepository{db: db, cipher: cipher}
}

// Load implements usecase.SettingsRepository.
func (r *SettingsRepository) Load(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	var enc, refreshEnc string
	var expiry int64
	err := r.db.QueryRowContext(ctx, `SELECT url, token_encrypted, refresh_encrypted, client_id, token_expiry
		FROM smarthome_settings WHERE id = 1`).Scan(&s.URL, &enc, &refreshEnc, &s.ClientID, &expiry)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	if s.Token, err = r.cipher.Decrypt(enc); err != nil {
		return domain.Settings{}, fmt.Errorf("decrypt token (was SECRET_KEY changed?): %w", err)
	}
	if refreshEnc != "" {
		if s.Refresh, err = r.cipher.Decrypt(refreshEnc); err != nil {
			return domain.Settings{}, fmt.Errorf("decrypt refresh token (was SECRET_KEY changed?): %w", err)
		}
	}
	if expiry > 0 {
		s.Expiry = time.Unix(expiry, 0)
	}
	return s, nil
}

// Save implements usecase.SettingsRepository.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	enc, err := r.cipher.Encrypt(s.Token)
	if err != nil {
		return fmt.Errorf("encrypt token: %w", err)
	}
	refreshEnc := ""
	if s.Refresh != "" {
		if refreshEnc, err = r.cipher.Encrypt(s.Refresh); err != nil {
			return fmt.Errorf("encrypt refresh token: %w", err)
		}
	}
	var expiry int64
	if !s.Expiry.IsZero() {
		expiry = s.Expiry.Unix()
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO smarthome_settings (id, url, token_encrypted, refresh_encrypted, client_id, token_expiry)
		VALUES (1, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET url = excluded.url, token_encrypted = excluded.token_encrypted,
			refresh_encrypted = excluded.refresh_encrypted, client_id = excluded.client_id, token_expiry = excluded.token_expiry`,
		s.URL, enc, refreshEnc, s.ClientID, expiry)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}
