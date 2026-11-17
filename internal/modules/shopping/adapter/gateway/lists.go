package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"waylight/internal/modules/shopping/domain"
	"waylight/internal/shared/kernel"
	"waylight/internal/shared/sqlutil"
)

// ListRepository stores the shopping lists kept on waylight (#471). Every
// statement carries the visibility filter of ADR 0018 itself: items are only
// reached through a list the principal may see, and only a list's owner may
// rename, share or delete it.
type ListRepository struct{ db *sql.DB }

// NewListRepository returns the repository.
func NewListRepository(db *sql.DB) *ListRepository { return &ListRepository{db: db} }

// visibleLists is the condition "list_id belongs to a list p may see".
func visibleLists(p kernel.Principal) (string, []any) {
	cond, args := sqlutil.VisibleTo(p, "")
	return `list_id IN (SELECT id FROM shopping_lists WHERE ` + cond + `)`, args
}

const listColumns = `l.id, l.name, l.owner_id, l.visibility, l.created_at,
	(SELECT COUNT(*) FROM shopping_items i WHERE i.list_id = l.id AND i.done = 0),
	(SELECT COUNT(*) FROM shopping_items i WHERE i.list_id = l.id AND i.done = 1)`

func scanList(row interface{ Scan(...any) error }) (domain.List, error) {
	var l domain.List
	var owner, vis, created string
	if err := row.Scan(&l.ID, &l.Name, &owner, &vis, &created, &l.Open, &l.Done); err != nil {
		return domain.List{}, err
	}
	l.Access = kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}
	t, err := sqlutil.ParseTime(created)
	if err != nil {
		return domain.List{}, err
	}
	l.Created = t
	return l, nil
}

// Lists returns the lists the principal may see, own first, oldest first.
func (r *ListRepository) Lists(ctx context.Context) ([]domain.List, error) {
	p := kernel.PrincipalFrom(ctx)
	cond, args := sqlutil.VisibleTo(p, "l")
	//nolint:gosec // cond comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+listColumns+` FROM shopping_lists l WHERE `+cond+
		` ORDER BY l.owner_id = ? DESC, l.created_at`, append(args, p.AccountID)...)
	if err != nil {
		return nil, fmt.Errorf("query shopping lists: %w", err)
	}
	defer rows.Close()
	var out []domain.List
	for rows.Next() {
		l, err := scanList(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// Get returns a list the principal may see.
func (r *ListRepository) Get(ctx context.Context, id string) (domain.List, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "l")
	l, err := scanList(r.db.QueryRowContext(ctx, `SELECT `+listColumns+` FROM shopping_lists l WHERE l.id = ? AND `+cond,
		append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.List{}, domain.ErrListNotFound
	}
	return l, err
}

// CountOwn returns how many lists the principal owns.
func (r *ListRepository) CountOwn(ctx context.Context) (int, error) {
	cond, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM shopping_lists WHERE `+cond, args...).Scan(&n)
	return n, err
}

// Create stores a new list owned by the principal.
func (r *ListRepository) Create(ctx context.Context, l domain.List) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO shopping_lists (id, name, owner_id, visibility, created_at) VALUES (?, ?, ?, ?, ?)`,
		l.ID, l.Name, kernel.PrincipalFrom(ctx).AccountID, string(l.Access.Visibility), sqlutil.FormatTime(l.Created))
	return err
}

// Update renames and shares one of the principal's own lists.
func (r *ListRepository) Update(ctx context.Context, id, name string, v kernel.Visibility) error {
	cond, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // cond comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `UPDATE shopping_lists SET name = ?, visibility = ? WHERE id = ? AND `+cond,
		append([]any{name, string(v), id}, args...)...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotYourList
	}
	return nil
}

// Delete removes one of the principal's own lists and its items.
func (r *ListRepository) Delete(ctx context.Context, id string) error {
	cond, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	//nolint:gosec // cond comes from sqlutil with bound arguments
	res, err := tx.ExecContext(ctx, `DELETE FROM shopping_lists WHERE id = ? AND `+cond, append([]any{id}, args...)...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotYourList
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM shopping_items WHERE list_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM shopping_current WHERE list_id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteOwnedBy removes all lists of an account and their items, when the
// account is deleted.
//
//archtest:all-owners a deleted account's lists go with it, whoever asks
func (r *ListRepository) DeleteOwnedBy(ctx context.Context, accountID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, q := range []string{
		`DELETE FROM shopping_items WHERE list_id IN (SELECT id FROM shopping_lists WHERE owner_id = ?)`,
		`DELETE FROM shopping_current WHERE list_id IN (SELECT id FROM shopping_lists WHERE owner_id = ?)`,
		`DELETE FROM shopping_lists WHERE owner_id = ?`,
		`DELETE FROM shopping_current WHERE account_id = ?`,
	} {
		if _, err := tx.ExecContext(ctx, q, accountID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Items returns the items of a list the principal may see.
func (r *ListRepository) Items(ctx context.Context, listID string) ([]domain.Item, error) {
	cond, args := visibleLists(kernel.PrincipalFrom(ctx))
	//nolint:gosec // cond comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT id, summary, done, created_at, completed_at FROM shopping_items
		WHERE list_id = ? AND `+cond, append([]any{listID}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("query shopping items: %w", err)
	}
	defer rows.Close()
	out := []domain.Item{}
	for rows.Next() {
		var it domain.Item
		var done int
		var created, completed string
		if err := rows.Scan(&it.ID, &it.Summary, &done, &created, &completed); err != nil {
			return nil, err
		}
		it.Done = done == 1
		if it.Created, err = sqlutil.ParseTime(created); err != nil {
			return nil, err
		}
		if completed != "" {
			if it.Completed, err = sqlutil.ParseTime(completed); err != nil {
				return nil, err
			}
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// AddItem adds an item to a list the principal may see, as long as the
// list holds fewer than max items; domain.ErrListFull otherwise. The caller
// checks first that the list is visible, so a refusal here means full.
func (r *ListRepository) AddItem(ctx context.Context, listID string, it domain.Item, max int) error {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // cond comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO shopping_items (id, list_id, summary, done, created_at, completed_at)
		SELECT ?, ?, ?, 0, ?, ''
		WHERE (SELECT COUNT(*) FROM shopping_items WHERE list_id = ?) < ?
		AND EXISTS (SELECT 1 FROM shopping_lists WHERE id = ? AND `+cond+`)`,
		append([]any{it.ID, listID, it.Summary, sqlutil.FormatTime(it.Created), listID, max, listID}, args...)...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrListFull
	}
	return nil
}

// SetDone ticks an item off or back on.
func (r *ListRepository) SetDone(ctx context.Context, listID, itemID string, done bool, at time.Time) error {
	cond, args := visibleLists(kernel.PrincipalFrom(ctx))
	completed := ""
	if done {
		completed = sqlutil.FormatTime(at)
	}
	//nolint:gosec // cond comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `UPDATE shopping_items SET done = ?, completed_at = ? WHERE id = ? AND list_id = ? AND `+cond,
		append([]any{sqlutil.BoolToInt(done), completed, itemID, listID}, args...)...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

// DeleteItem removes an item.
func (r *ListRepository) DeleteItem(ctx context.Context, listID, itemID string) error {
	cond, args := visibleLists(kernel.PrincipalFrom(ctx))
	//nolint:gosec // cond comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM shopping_items WHERE id = ? AND list_id = ? AND `+cond,
		append([]any{itemID, listID}, args...)...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

// ClearCompleted removes the ticked items of a list and says how many.
func (r *ListRepository) ClearCompleted(ctx context.Context, listID string) (int, error) {
	cond, args := visibleLists(kernel.PrincipalFrom(ctx))
	//nolint:gosec // cond comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM shopping_items WHERE done = 1 AND list_id = ? AND `+cond,
		append([]any{listID}, args...)...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// Current returns what the widget shows for the principal: a list ID,
// domain.CurrentCalDAV, or "" when nothing was chosen.
func (r *ListRepository) Current(ctx context.Context) (string, error) {
	var v string
	err := r.db.QueryRowContext(ctx, `SELECT list_id FROM shopping_current WHERE account_id = ?`,
		kernel.PrincipalFrom(ctx).SettingsAccount()).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}

// SetCurrent stores what the widget shows for the principal.
func (r *ListRepository) SetCurrent(ctx context.Context, v string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO shopping_current (account_id, list_id) VALUES (?, ?)
		ON CONFLICT (account_id) DO UPDATE SET list_id = excluded.list_id`, kernel.PrincipalFrom(ctx).SettingsAccount(), v)
	return err
}
