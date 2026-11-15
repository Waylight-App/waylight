package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// PreferenceRepository implements usecase.PreferenceRepository.
type PreferenceRepository struct{ db *sql.DB }

// NewPreferenceRepository returns a repository backed by db.
func NewPreferenceRepository(db *sql.DB) *PreferenceRepository { return &PreferenceRepository{db: db} }

// PrivateByDefault implements usecase.PreferenceRepository; false without a row.
func (r *PreferenceRepository) PrivateByDefault(ctx context.Context, accountID string) (bool, error) {
	var on bool
	err := r.db.QueryRowContext(ctx, `SELECT private_by_default FROM auth_preferences WHERE account_id = ?`, accountID).Scan(&on)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("select preferences: %w", err)
	}
	return on, nil
}

// SetPrivateByDefault implements usecase.PreferenceRepository.
func (r *PreferenceRepository) SetPrivateByDefault(ctx context.Context, accountID string, on bool) error {
	if _, err := r.db.ExecContext(ctx, `INSERT INTO auth_preferences (account_id, private_by_default) VALUES (?, ?)
		ON CONFLICT (account_id) DO UPDATE SET private_by_default = excluded.private_by_default`, accountID, on); err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}
	return nil
}
