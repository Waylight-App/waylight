package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/kernel"
	"trisse-home/internal/shared/sqlutil"
)

// SubscriptionRepository stores ICS subscriptions with encrypted URLs.
type SubscriptionRepository struct {
	db     *sql.DB
	cipher Cipher
}

// mine is the condition for the calendars of the account whose settings
// apply (ADR 0018): an account's own, and the admin's for guests, kiosk
// displays and background work. There is no fallback from a member to the
// admin's calendars; they are private.
func mine(ctx context.Context) (string, []any) {
	return sqlutil.OwnedBy(kernel.PrincipalFor(kernel.PrincipalFrom(ctx).SettingsAccount()), "")
}

// NewSubscriptionRepository returns a repository backed by db.
func NewSubscriptionRepository(db *sql.DB, cipher Cipher) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, cipher: cipher}
}

const subscriptionColumns = `id, name, url_encrypted, color, refresh_minutes`

func (r *SubscriptionRepository) scan(row interface{ Scan(...any) error }) (domain.Subscription, error) {
	var s domain.Subscription
	var enc string
	var refresh int
	if err := row.Scan(&s.ID, &s.Name, &enc, &s.Color, &refresh); err != nil {
		return domain.Subscription{}, err
	}
	url, err := r.cipher.Decrypt(enc)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("decrypt subscription URL (was SECRET_KEY changed?): %w", err)
	}
	s.URL, s.RefreshInterval = url, time.Duration(refresh)*time.Minute
	return s, nil
}

// List implements usecase.SubscriptionRepository in creation order.
func (r *SubscriptionRepository) List(ctx context.Context) ([]domain.Subscription, error) {
	cond, args := mine(ctx)
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+subscriptionColumns+` FROM calendar_subscriptions WHERE `+cond+` ORDER BY position, name`, args...)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}
	defer rows.Close()
	var out []domain.Subscription
	for rows.Next() {
		s, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// Get implements usecase.SubscriptionRepository.
func (r *SubscriptionRepository) Get(ctx context.Context, id string) (domain.Subscription, error) {
	cond, args := mine(ctx)
	s, err := r.scan(r.db.QueryRowContext(ctx, `SELECT `+subscriptionColumns+` FROM calendar_subscriptions WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Subscription{}, domain.ErrSubscriptionNotFound
	}
	return s, err
}

// Save implements usecase.SubscriptionRepository (insert or update).
func (r *SubscriptionRepository) Save(ctx context.Context, s domain.Subscription) error {
	enc, err := r.cipher.Encrypt(s.URL)
	if err != nil {
		return fmt.Errorf("encrypt subscription URL: %w", err)
	}
	cond, args := mine(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(kernel.PrincipalFor(kernel.PrincipalFrom(ctx).SettingsAccount()), "calendar_subscriptions")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	_, err = r.db.ExecContext(ctx, `INSERT INTO calendar_subscriptions (id, name, url_encrypted, color, refresh_minutes, position, owner_id)
		VALUES (?, ?, ?, ?, ?, (SELECT COALESCE(MAX(position), -1) + 1 FROM calendar_subscriptions WHERE `+cond+`), ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, url_encrypted = excluded.url_encrypted, color = excluded.color,
			refresh_minutes = excluded.refresh_minutes WHERE `+owned,
		append(append(append([]any{s.ID, s.Name, enc, s.Color, int(s.RefreshInterval / time.Minute)}, args...), kernel.PrincipalFrom(ctx).SettingsAccount()), ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert subscription: %w", err)
	}
	return nil
}

// Delete implements usecase.SubscriptionRepository.
func (r *SubscriptionRepository) Delete(ctx context.Context, id string) error {
	cond, args := mine(ctx)
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	if _, err := r.db.ExecContext(ctx, `DELETE FROM calendar_subscriptions WHERE id = ? AND `+cond, append([]any{id}, args...)...); err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	return nil
}
