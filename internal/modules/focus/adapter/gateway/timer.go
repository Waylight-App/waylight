package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"waylight/internal/modules/focus/domain"
	"waylight/internal/shared/kernel"
	"waylight/internal/shared/sqlutil"
)

// Timer implements usecase.TimerRepository: the principal's own timer.
func (r *Repository) Timer(ctx context.Context) (domain.Timer, bool, error) {
	cond, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	var (
		t                         domain.Timer
		phase                     string
		endsAt                    sql.NullString
		left, total               int64
		work, short, long, longEv int
	)
	err := r.db.QueryRowContext(ctx, `SELECT phase, done, ends_at, left_ms, total_ms, work_min, short_min, long_min, long_every
		FROM focus_timers WHERE `+cond+` LIMIT 1`, args...).
		Scan(&phase, &t.Done, &endsAt, &left, &total, &work, &short, &long, &longEv)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Timer{}, false, nil
	}
	if err != nil {
		return domain.Timer{}, false, fmt.Errorf("query timer: %w", err)
	}
	cfg, cfgErr := domain.NewTimerConfig(work, short, long, longEv)
	if cfgErr != nil {
		// A row that no longer validates starts over instead of failing the widget.
		return domain.Timer{}, false, nil
	}
	t.Phase, t.Config = domain.Phase(phase), cfg
	t.Left, t.Total = time.Duration(left)*time.Millisecond, time.Duration(total)*time.Millisecond
	if endsAt.Valid {
		if t.EndsAt, err = sqlutil.ParseTime(endsAt.String); err != nil {
			return domain.Timer{}, false, err
		}
	}
	return t, true, nil
}

// SaveTimer implements usecase.TimerRepository; the timer belongs to the
// principal.
func (r *Repository) SaveTimer(ctx context.Context, t domain.Timer) error {
	p := kernel.PrincipalFrom(ctx)
	var endsAt any
	if t.Running() {
		endsAt = sqlutil.FormatTime(t.EndsAt)
	}
	c := t.Config
	_, err := r.db.ExecContext(ctx, `INSERT INTO focus_timers (owner_id, phase, done, ends_at, left_ms, total_ms, work_min, short_min, long_min, long_every)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (owner_id) DO UPDATE SET phase = excluded.phase, done = excluded.done, ends_at = excluded.ends_at,
			left_ms = excluded.left_ms, total_ms = excluded.total_ms, work_min = excluded.work_min,
			short_min = excluded.short_min, long_min = excluded.long_min, long_every = excluded.long_every`,
		p.AccountID, string(t.Phase), t.Done, endsAt, t.Left.Milliseconds(), t.Total.Milliseconds(),
		int(c.Work/time.Minute), int(c.Short/time.Minute), int(c.Long/time.Minute), c.LongEvery)
	if err != nil {
		return fmt.Errorf("save timer: %w", err)
	}
	return nil
}
