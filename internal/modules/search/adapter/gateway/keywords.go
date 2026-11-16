package gateway

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/search/domain"
	"trisse-home/internal/modules/search/usecase"
	"trisse-home/internal/shared/sqlutil"
)

// Own keywords (#478) follow the shortcuts: read from the account's own
// settings once it has them, otherwise from the admin's, and copied into
// the account on its first change (see writeAccount).

// Keywords implements usecase.Repository.
func (r *Repository) Keywords(ctx context.Context) ([]domain.Keyword, error) {
	account, err := readAccount(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT keyword, kind, target, label FROM search_keywords WHERE account_id = ? ORDER BY keyword`, account)
	if err != nil {
		return nil, fmt.Errorf("query keywords: %w", err)
	}
	defer rows.Close()
	var out []domain.Keyword
	for rows.Next() {
		var word, kind, target, label string
		if err := rows.Scan(&word, &kind, &target, &label); err != nil {
			return nil, fmt.Errorf("scan keyword: %w", err)
		}
		k, err := domain.NewKeyword(word, domain.TargetKind(kind), target, label)
		if err != nil {
			continue // a row that violates current rules is skipped, not fatal
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// AddKeyword implements usecase.Repository.
func (r *Repository) AddKeyword(ctx context.Context, k domain.Keyword) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	account, err := writeAccount(ctx, tx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO search_keywords (account_id, keyword, kind, target, label) VALUES (?, ?, ?, ?, ?)`,
		account, k.Word(), string(k.Kind()), k.Target(), k.Label())
	if sqlutil.IsUniqueViolation(err) {
		return usecase.ErrKeywordExists
	}
	if err != nil {
		return fmt.Errorf("insert keyword: %w", err)
	}
	return tx.Commit()
}

// DeleteKeyword implements usecase.Repository.
func (r *Repository) DeleteKeyword(ctx context.Context, word string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	account, err := writeAccount(ctx, tx)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM search_keywords WHERE account_id = ? AND keyword = ?`, account, word)
	if err != nil {
		return fmt.Errorf("delete keyword: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return usecase.ErrKeywordNotFound
	}
	return tx.Commit()
}

// ReplaceKeywords implements usecase.ConfigRepository in one transaction,
// like ReplaceShortcuts.
func (r *Repository) ReplaceKeywords(ctx context.Context, list []domain.Keyword) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	account, err := writeAccount(ctx, tx)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO search_settings (account_id, default_provider) VALUES (?, '') ON CONFLICT (account_id) DO NOTHING`, account); err != nil {
		return fmt.Errorf("create search settings: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM search_keywords WHERE account_id = ?`, account); err != nil {
		return fmt.Errorf("clear keywords: %w", err)
	}
	for _, k := range list {
		if _, err := tx.ExecContext(ctx, `INSERT INTO search_keywords (account_id, keyword, kind, target, label) VALUES (?, ?, ?, ?, ?)`,
			account, k.Word(), string(k.Kind()), k.Target(), k.Label()); err != nil {
			return fmt.Errorf("insert keyword: %w", err)
		}
	}
	return tx.Commit()
}
