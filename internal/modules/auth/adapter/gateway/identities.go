package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/sqlutil"
)

// IdentityRepository implements usecase.IdentityRepository in SQLite (#527).
type IdentityRepository struct{ db *sql.DB }

// NewIdentityRepository returns the store of linked sign-ins.
func NewIdentityRepository(db *sql.DB) *IdentityRepository { return &IdentityRepository{db: db} }

// Add implements usecase.IdentityRepository.
func (r *IdentityRepository) Add(ctx context.Context, l domain.IdentityLink) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO auth_identities (issuer, subject, account_id, email, created_at, only_sign_in) VALUES (?, ?, ?, ?, ?, ?)`,
		l.Issuer, l.Subject, l.AccountID, l.Email, sqlutil.FormatTime(l.CreatedAt), l.OnlySignIn)
	if sqlutil.IsUniqueViolation(err) {
		return domain.ErrIdentityTaken
	}
	if err != nil {
		return fmt.Errorf("insert identity: %w", err)
	}
	return nil
}

// Find implements usecase.IdentityRepository.
func (r *IdentityRepository) Find(ctx context.Context, issuer, subject string) (string, error) {
	var account string
	err := r.db.QueryRowContext(ctx, `SELECT account_id FROM auth_identities WHERE issuer = ? AND subject = ?`, issuer, subject).Scan(&account)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrIdentityNotFound
	}
	if err != nil {
		return "", fmt.Errorf("select identity: %w", err)
	}
	return account, nil
}

// ListByAccount implements usecase.IdentityRepository.
func (r *IdentityRepository) ListByAccount(ctx context.Context, accountID string) ([]domain.IdentityLink, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT issuer, subject, account_id, email, created_at, only_sign_in FROM auth_identities
		WHERE account_id = ? ORDER BY created_at`, accountID)
	if err != nil {
		return nil, fmt.Errorf("query identities: %w", err)
	}
	defer rows.Close()
	var out []domain.IdentityLink
	for rows.Next() {
		var l domain.IdentityLink
		var created string
		if err := rows.Scan(&l.Issuer, &l.Subject, &l.AccountID, &l.Email, &created, &l.OnlySignIn); err != nil {
			return nil, fmt.Errorf("scan identity: %w", err)
		}
		if l.CreatedAt, err = sqlutil.ParseTime(created); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// Remove implements usecase.IdentityRepository.
func (r *IdentityRepository) Remove(ctx context.Context, accountID, issuer, subject string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_identities WHERE account_id = ? AND issuer = ? AND subject = ?`, accountID, issuer, subject)
	if err != nil {
		return fmt.Errorf("delete identity: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrIdentityNotFound
	}
	return nil
}

// RemoveAccount implements usecase.IdentityRepository.
func (r *IdentityRepository) RemoveAccount(ctx context.Context, accountID string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_identities WHERE account_id = ?`, accountID); err != nil {
		return fmt.Errorf("delete identities: %w", err)
	}
	return nil
}

// PasswordSet implements usecase.IdentityRepository: the account has a known
// password now, so its sign-ins may be disconnected.
func (r *IdentityRepository) PasswordSet(ctx context.Context, accountID string) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE auth_identities SET only_sign_in = 0 WHERE account_id = ?`, accountID); err != nil {
		return fmt.Errorf("update identities: %w", err)
	}
	return nil
}
