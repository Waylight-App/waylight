package gateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/kernel"
	"trisse-home/internal/shared/sqlutil"
)

// AccountRepository implements usecase.AccountRepository (#282). Secrets
// never leave this type unencrypted on their way to the database.
type AccountRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewAccountRepository returns the repository.
func NewAccountRepository(db *sql.DB, cipher Cipher) *AccountRepository {
	return &AccountRepository{db: db, cipher: cipher}
}

const accountColumns = `id, provider, client_id, name, secret_encrypted, access_encrypted, refresh_encrypted, expiry, calendars`

type calendarJSON struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color,omitempty"`
	Writable bool   `json:"writable,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
}

func (r *AccountRepository) scan(row interface{ Scan(...any) error }) (domain.Account, error) {
	var a domain.Account
	var provider, secret, access, refresh, expiry, cals string
	if err := row.Scan(&a.ID, &provider, &a.ClientID, &a.Name, &secret, &access, &refresh, &expiry, &cals); err != nil {
		return a, err
	}
	a.Provider = domain.Provider(provider)
	var err error
	for _, f := range []struct {
		dst *string
		src string
	}{{&a.Grant.ClientSecret, secret}, {&a.Grant.Access, access}, {&a.Grant.Refresh, refresh}} {
		if f.src == "" {
			continue
		}
		if *f.dst, err = r.cipher.Decrypt(f.src); err != nil {
			return a, fmt.Errorf("decrypt account token: %w", err)
		}
	}
	a.Grant.Expiry, _ = time.Parse(time.RFC3339, expiry)
	var list []calendarJSON
	if err := json.Unmarshal([]byte(cals), &list); err != nil {
		return a, fmt.Errorf("decode calendars: %w", err)
	}
	for _, c := range list {
		a.Calendars = append(a.Calendars, domain.AccountCalendar(c))
	}
	return a, nil
}

// List implements usecase.AccountRepository, oldest first.
func (r *AccountRepository) List(ctx context.Context) ([]domain.Account, error) {
	cond, args := mine(ctx)
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+accountColumns+` FROM calendar_accounts WHERE `+cond+` ORDER BY created_at`, args...)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()
	var out []domain.Account
	for rows.Next() {
		a, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Get implements usecase.AccountRepository.
func (r *AccountRepository) Get(ctx context.Context, id string) (domain.Account, error) {
	cond, args := mine(ctx)
	a, err := r.scan(r.db.QueryRowContext(ctx, `SELECT `+accountColumns+` FROM calendar_accounts WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	return a, err
}

func (r *AccountRepository) encrypt(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	return r.cipher.Encrypt(s)
}

// Save implements usecase.AccountRepository (insert or update).
func (r *AccountRepository) Save(ctx context.Context, a domain.Account) error {
	secret, err := r.encrypt(a.Grant.ClientSecret)
	if err != nil {
		return err
	}
	access, err := r.encrypt(a.Grant.Access)
	if err != nil {
		return err
	}
	refresh, err := r.encrypt(a.Grant.Refresh)
	if err != nil {
		return err
	}
	list := make([]calendarJSON, 0, len(a.Calendars))
	for _, c := range a.Calendars {
		list = append(list, calendarJSON(c))
	}
	cals, err := json.Marshal(list)
	if err != nil {
		return err
	}
	expiry := ""
	if !a.Grant.Expiry.IsZero() {
		expiry = a.Grant.Expiry.UTC().Format(time.RFC3339)
	}
	account := kernel.PrincipalFrom(ctx).SettingsAccount()
	owned, ownedArgs := sqlutil.OwnedBy(kernel.PrincipalFor(account), "calendar_accounts")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	_, err = r.db.ExecContext(ctx, `INSERT INTO calendar_accounts (`+accountColumns+`, owner_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, secret_encrypted = excluded.secret_encrypted,
		access_encrypted = excluded.access_encrypted, refresh_encrypted = excluded.refresh_encrypted,
		expiry = excluded.expiry, calendars = excluded.calendars WHERE `+owned,
		append([]any{a.ID, string(a.Provider), a.ClientID, a.Name, secret, access, refresh, expiry, string(cals), account}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("save account: %w", err)
	}
	return nil
}

// Delete implements usecase.AccountRepository.
func (r *AccountRepository) Delete(ctx context.Context, id string) error {
	cond, args := mine(ctx)
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM calendar_accounts WHERE id = ? AND `+cond, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}
