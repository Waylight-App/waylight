package gateway

import (
	"context"
	"fmt"

	"waylight/internal/modules/search/domain"
	"waylight/internal/modules/search/usecase"
	"waylight/internal/shared/kernel"
	"waylight/internal/shared/sqlutil"
)

// Own keywords (#478) are strictly personal, unlike the rest of the search
// settings: their targets and labels can name private bookmarks, so a
// member never sees the admin's keywords and nothing is copied on the
// first change. Guests and kiosk displays have none and cannot add any.

// errKeywordsNotForTabs refuses changes from guest links and kiosk displays,
// which act on someone else's account.
var errKeywordsNotForTabs = kernel.NewPublicError(kernel.ErrForbidden, "Own keywords need a signed-in account.")

// keywordAccount is the account whose keywords apply, and false for
// tab-scoped principals.
func keywordAccount(ctx context.Context) (string, bool) {
	p := kernel.PrincipalFrom(ctx)
	if p.TabScoped() {
		return "", false
	}
	return p.SettingsAccount(), true
}

// Keywords implements usecase.Repository.
func (r *Repository) Keywords(ctx context.Context) ([]domain.Keyword, error) {
	account, ok := keywordAccount(ctx)
	if !ok {
		return nil, nil
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
	account, ok := keywordAccount(ctx)
	if !ok {
		return errKeywordsNotForTabs
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO search_keywords (account_id, keyword, kind, target, label) VALUES (?, ?, ?, ?, ?)`,
		account, k.Word(), string(k.Kind()), k.Target(), k.Label())
	if sqlutil.IsUniqueViolation(err) {
		return usecase.ErrKeywordExists
	}
	if err != nil {
		return fmt.Errorf("insert keyword: %w", err)
	}
	return nil
}

// DeleteKeyword implements usecase.Repository.
func (r *Repository) DeleteKeyword(ctx context.Context, word string) error {
	account, ok := keywordAccount(ctx)
	if !ok {
		return errKeywordsNotForTabs
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM search_keywords WHERE account_id = ? AND keyword = ?`, account, word)
	if err != nil {
		return fmt.Errorf("delete keyword: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return usecase.ErrKeywordNotFound
	}
	return nil
}

// ReplaceKeywords implements usecase.ConfigRepository in one transaction.
// Config as code runs as the system, so it replaces the admin's keywords.
func (r *Repository) ReplaceKeywords(ctx context.Context, list []domain.Keyword) error {
	account, ok := keywordAccount(ctx)
	if !ok {
		return errKeywordsNotForTabs
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
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

// ForgetAccount removes the search settings, shortcuts and keywords of a
// deleted account. The admin's rows (the household defaults) are kept.
func (r *Repository) ForgetAccount(ctx context.Context, account string) error {
	if account == "" || account == kernel.SystemAccountID {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, table := range []string{"search_keywords", "search_shortcuts", "search_settings"} {
		// Table names are constants from the list above, not input.
		if _, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE account_id = ?`, account); err != nil { //nolint:gosec // constant table names
			return fmt.Errorf("forget account in %s: %w", table, err)
		}
	}
	return tx.Commit()
}
