package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"waylight/internal/modules/auth/domain"
	"waylight/internal/shared/sqlutil"
)

// PairingCodeRepository stores the pairing codes of app sessions in SQLite.
type PairingCodeRepository struct{ db *sql.DB }

// NewPairingCodeRepository returns a repository backed by db.
func NewPairingCodeRepository(db *sql.DB) *PairingCodeRepository {
	return &PairingCodeRepository{db: db}
}

// Create implements usecase.PairingCodeRepository. Expired codes are removed
// on the way, so the table never grows beyond the codes of a few minutes.
func (r *PairingCodeRepository) Create(ctx context.Context, c domain.PairingCode, now time.Time) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_pairing_codes WHERE expires_at <= ?`, sqlutil.FormatTime(now)); err != nil {
		return fmt.Errorf("delete expired pairing codes: %w", err)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO auth_pairing_codes (code_hash, account_id, expires_at, challenge) VALUES (?, ?, ?, ?)`,
		string(c.Hash), c.AccountID, sqlutil.FormatTime(c.ExpiresAt), c.Challenge)
	if err != nil {
		return fmt.Errorf("insert pairing code: %w", err)
	}
	return nil
}

// Take implements usecase.PairingCodeRepository. The code is deleted in the
// same step; only the request whose DELETE removed the row gets it, so two
// concurrent requests cannot both redeem it.
func (r *PairingCodeRepository) Take(ctx context.Context, hash domain.TokenHash, now time.Time) (domain.PairingCode, error) {
	var accountID, expires, challenge string
	err := r.db.QueryRowContext(ctx, `SELECT account_id, expires_at, challenge FROM auth_pairing_codes WHERE code_hash = ?`, string(hash)).
		Scan(&accountID, &expires, &challenge)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PairingCode{}, domain.ErrInvalidPairingCode
	}
	if err != nil {
		return domain.PairingCode{}, fmt.Errorf("select pairing code: %w", err)
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_pairing_codes WHERE code_hash = ?`, string(hash))
	if err != nil {
		return domain.PairingCode{}, fmt.Errorf("delete pairing code: %w", err)
	}
	if n, err := res.RowsAffected(); err != nil || n != 1 {
		return domain.PairingCode{}, domain.ErrInvalidPairingCode
	}
	expiresAt, err := sqlutil.ParseTime(expires)
	if err != nil {
		return domain.PairingCode{}, err
	}
	if !now.Before(expiresAt) {
		return domain.PairingCode{}, domain.ErrInvalidPairingCode
	}
	return domain.PairingCode{Hash: hash, AccountID: accountID, ExpiresAt: expiresAt, Challenge: challenge}, nil
}
