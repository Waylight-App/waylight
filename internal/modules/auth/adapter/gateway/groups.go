package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/sqlutil"
)

// GroupRepository implements usecase.GroupRepository (#294).
type GroupRepository struct{ db *sql.DB }

// NewGroupRepository returns a repository backed by db.
func NewGroupRepository(db *sql.DB) *GroupRepository { return &GroupRepository{db: db} }

// List implements usecase.GroupRepository.
func (r *GroupRepository) List(ctx context.Context) ([]domain.Group, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, created_at FROM auth_groups ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("select groups: %w", err)
	}
	defer rows.Close()
	type row struct{ id, name, created string }
	var found []row
	for rows.Next() {
		var g row
		if err := rows.Scan(&g.id, &g.name, &g.created); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		found = append(found, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select groups: %w", err)
	}
	members, err := r.allAccounts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Group, 0, len(found))
	for _, g := range found {
		created, err := sqlutil.ParseTime(g.created)
		if err != nil {
			return nil, err
		}
		accounts := members[g.id]
		if accounts == nil {
			accounts = []string{}
		}
		out = append(out, domain.RestoreGroup(g.id, g.name, accounts, created))
	}
	return out, nil
}

// allAccounts returns the accounts of every group in one query.
func (r *GroupRepository) allAccounts(ctx context.Context) (map[string][]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT group_id, account_id FROM auth_group_accounts ORDER BY account_id`)
	if err != nil {
		return nil, fmt.Errorf("select group accounts: %w", err)
	}
	defer rows.Close()
	out := map[string][]string{}
	for rows.Next() {
		var g, a string
		if err := rows.Scan(&g, &a); err != nil {
			return nil, fmt.Errorf("scan group account: %w", err)
		}
		out[g] = append(out[g], a)
	}
	return out, rows.Err()
}

// Get implements usecase.GroupRepository.
func (r *GroupRepository) Get(ctx context.Context, id string) (domain.Group, error) {
	var name, created string
	err := r.db.QueryRowContext(ctx, `SELECT name, created_at FROM auth_groups WHERE id = ?`, id).Scan(&name, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Group{}, domain.ErrGroupNotFound
	}
	if err != nil {
		return domain.Group{}, fmt.Errorf("select group: %w", err)
	}
	t, err := sqlutil.ParseTime(created)
	if err != nil {
		return domain.Group{}, err
	}
	members, err := r.allAccounts(ctx)
	if err != nil {
		return domain.Group{}, err
	}
	accounts := members[id]
	if accounts == nil {
		accounts = []string{}
	}
	return domain.RestoreGroup(id, name, accounts, t), nil
}

// Create implements usecase.GroupRepository.
func (r *GroupRepository) Create(ctx context.Context, g domain.Group) error {
	return r.write(ctx, g, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO auth_groups (id, name, created_at) VALUES (?, ?, ?)`,
			g.ID(), g.Name(), sqlutil.FormatTime(g.CreatedAt()))
		return err
	})
}

// Update implements usecase.GroupRepository.
func (r *GroupRepository) Update(ctx context.Context, g domain.Group) error {
	return r.write(ctx, g, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `UPDATE auth_groups SET name = ? WHERE id = ?`, g.Name(), g.ID())
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrGroupNotFound
		}
		return nil
	})
}

// write runs the change of the group row and replaces its accounts in one
// transaction.
func (r *GroupRepository) write(ctx context.Context, g domain.Group, change func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := change(tx); err != nil {
		if sqlutil.IsUniqueViolation(err) {
			return domain.ErrGroupNameTaken
		}
		if errors.Is(err, domain.ErrGroupNotFound) {
			return err
		}
		return fmt.Errorf("save group: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM auth_group_accounts WHERE group_id = ?`, g.ID()); err != nil {
		return fmt.Errorf("clear group accounts: %w", err)
	}
	for _, a := range g.Accounts() {
		if _, err := tx.ExecContext(ctx, `INSERT INTO auth_group_accounts (group_id, account_id) VALUES (?, ?)`, g.ID(), a); err != nil {
			return fmt.Errorf("add group account: %w", err)
		}
	}
	return tx.Commit()
}

// Delete implements usecase.GroupRepository.
func (r *GroupRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `DELETE FROM auth_groups WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrGroupNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM auth_group_accounts WHERE group_id = ?`, id); err != nil {
		return fmt.Errorf("delete group accounts: %w", err)
	}
	return tx.Commit()
}

// GroupsOf implements usecase.GroupRepository.
func (r *GroupRepository) GroupsOf(ctx context.Context, accountID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT group_id FROM auth_group_accounts WHERE account_id = ? ORDER BY group_id`, accountID)
	if err != nil {
		return nil, fmt.Errorf("select groups of account: %w", err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan group id: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// RemoveAccount implements usecase.GroupRepository.
func (r *GroupRepository) RemoveAccount(ctx context.Context, accountID string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_group_accounts WHERE account_id = ?`, accountID); err != nil {
		return fmt.Errorf("remove account from groups: %w", err)
	}
	return nil
}
