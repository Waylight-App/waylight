// Package gateway stores contracts in SQLite.
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

	"waylight/internal/modules/contracts/domain"
	"waylight/internal/shared/kernel"
	"waylight/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are the tables included in data exports (no secrets).
var TransferTables = []sqlutil.Table{{Name: "contracts_items"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, name, provider, cost_cents, period, term_end, notice_days, renew_months, remind_days, note, reminded_for, owner_id, visibility, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Contract, error) {
	var c domain.Contract
	var period, end, owner, vis, created string
	if err := sc.Scan(&c.ID, &c.Name, &c.Provider, &c.CostCents, &period, &end, &c.NoticeDays, &c.RenewMonths, &c.RemindDays,
		&c.Note, &c.RemindedFor, &owner, &vis, &created); err != nil {
		return c, err
	}
	var err error
	if c.TermEnd, err = time.Parse(time.DateOnly, end); err != nil {
		return c, err
	}
	c.Period = domain.Period(period)
	c.Access = kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}
	c.CreatedAt, err = sqlutil.ParseTime(created)
	return c, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Contract, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM contracts_items WHERE `+cond+` ORDER BY name`, args...)
	if err != nil {
		return nil, fmt.Errorf("query contracts: %w", err)
	}
	defer rows.Close()
	var out []domain.Contract
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan contract: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Contract, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	c, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM contracts_items WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Contract{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Contract{}, fmt.Errorf("select contract: %w", err)
	}
	return c, nil
}

// Save implements usecase.Repository (upsert). A new contract belongs to
// the principal; an existing one is changed only by its owner (ADR 0018).
// The reminder bookkeeping is left to MarkReminded.
func (r *Repository) Save(ctx context.Context, c domain.Contract) error {
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "contracts_items")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO contracts_items (id, name, provider, cost_cents, period, term_end, notice_days,
		renew_months, remind_days, note, reminded_for, owner_id, visibility, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, provider = excluded.provider, cost_cents = excluded.cost_cents,
		period = excluded.period, term_end = excluded.term_end, notice_days = excluded.notice_days,
		renew_months = excluded.renew_months, remind_days = excluded.remind_days, note = excluded.note,
		visibility = excluded.visibility
		WHERE `+owned,
		append([]any{c.ID, c.Name, c.Provider, c.CostCents, string(c.Period), c.TermEndDate(), c.NoticeDays, c.RenewMonths, c.RemindDays,
			c.Note, p.AccountID, string(c.Access.Visibility), sqlutil.FormatTime(c.CreatedAt)}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert contract: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	return nil
}

// Delete implements usecase.Repository for the owner; a visible contract
// of someone else returns kernel.ErrNotOwner, an unknown one is no error.
func (r *Repository) Delete(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM contracts_items WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete contract: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := r.Get(ctx, id); err == nil {
			return kernel.ErrNotOwner
		}
	}
	return nil
}

// MarkReminded implements usecase.Repository.
//
//archtest:all-owners the reminder job records a sent reminder on contracts the admin sees but may not own
func (r *Repository) MarkReminded(ctx context.Context, id, deadline string) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE contracts_items SET reminded_for = ? WHERE id = ?`, deadline, id); err != nil {
		return fmt.Errorf("mark contract reminded: %w", err)
	}
	return nil
}
