package gateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/sqlutil"
)

// AccountLinkRepository implements usecase.AccountLinkRepository (#296).
type AccountLinkRepository struct{ db *sql.DB }

// NewAccountLinkRepository returns a repository backed by db.
func NewAccountLinkRepository(db *sql.DB) *AccountLinkRepository {
	return &AccountLinkRepository{db: db}
}

const linkColumns = `id, token_hash, purpose, account_id, name, role, tabs, created_at, expires_at`

func scanLink(row interface{ Scan(...any) error }) (domain.AccountLink, error) {
	var l domain.AccountLink
	var hash, purpose, role, tabs, created, expires string
	err := row.Scan(&l.ID, &hash, &purpose, &l.AccountID, &l.Name, &role, &tabs, &created, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.AccountLink{}, domain.ErrInvalidAccountLink
	}
	if err != nil {
		return domain.AccountLink{}, fmt.Errorf("select account link: %w", err)
	}
	l.TokenHash, l.Purpose, l.Role = domain.TokenHash(hash), domain.LinkPurpose(purpose), domain.Role(role)
	if err := json.Unmarshal([]byte(tabs), &l.Tabs); err != nil {
		return domain.AccountLink{}, fmt.Errorf("decode link tabs: %w", err)
	}
	if l.CreatedAt, err = sqlutil.ParseTime(created); err != nil {
		return domain.AccountLink{}, err
	}
	if l.ExpiresAt, err = sqlutil.ParseTime(expires); err != nil {
		return domain.AccountLink{}, err
	}
	return l, nil
}

// Save implements usecase.AccountLinkRepository.
func (r *AccountLinkRepository) Save(ctx context.Context, l domain.AccountLink) error {
	tabs, err := json.Marshal(append([]string{}, l.Tabs...))
	if err != nil {
		return fmt.Errorf("encode link tabs: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO auth_account_links (`+linkColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.ID, string(l.TokenHash), string(l.Purpose), l.AccountID, l.Name, string(l.Role), string(tabs),
		sqlutil.FormatTime(l.CreatedAt), sqlutil.FormatTime(l.ExpiresAt))
	if err != nil {
		return fmt.Errorf("insert account link: %w", err)
	}
	return nil
}

// FindByTokenHash implements usecase.AccountLinkRepository.
func (r *AccountLinkRepository) FindByTokenHash(ctx context.Context, hash domain.TokenHash) (domain.AccountLink, error) {
	return scanLink(r.db.QueryRowContext(ctx, `SELECT `+linkColumns+` FROM auth_account_links WHERE token_hash = ?`, string(hash)))
}

// List implements usecase.AccountLinkRepository.
func (r *AccountLinkRepository) List(ctx context.Context, purpose domain.LinkPurpose) ([]domain.AccountLink, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+linkColumns+` FROM auth_account_links WHERE purpose = ? ORDER BY created_at`, string(purpose))
	if err != nil {
		return nil, fmt.Errorf("list account links: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []domain.AccountLink{}
	for rows.Next() {
		l, err := scanLink(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// Delete implements usecase.AccountLinkRepository. It reports whether a
// link was removed, so a link is consumed exactly once even when two
// requests race.
func (r *AccountLinkRepository) Delete(ctx context.Context, id string) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_account_links WHERE id = ?`, id)
	if err != nil {
		return false, fmt.Errorf("delete account link: %w", err)
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

// DeleteForAccount implements usecase.AccountLinkRepository.
func (r *AccountLinkRepository) DeleteForAccount(ctx context.Context, accountID string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_account_links WHERE account_id = ? AND account_id != ''`, accountID); err != nil {
		return fmt.Errorf("delete account links: %w", err)
	}
	return nil
}

// DeleteExpired implements usecase.AccountLinkRepository.
func (r *AccountLinkRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_account_links WHERE expires_at <= ?`, sqlutil.FormatTime(now)); err != nil {
		return fmt.Errorf("delete expired account links: %w", err)
	}
	return nil
}
