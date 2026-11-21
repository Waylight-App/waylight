// Package presenter maps focus statistics to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/focus/domain"
)

// DayDTO is the statistic of one day.
type DayDTO struct {
	// Date is YYYY-MM-DD in the server's time zone.
	Date     string `json:"date"`
	Sessions int    `json:"sessions"`
	Minutes  int    `json:"minutes"`
}

// WeekDTO is the answer of GET /api/focus/stats.
type WeekDTO struct {
	// Days are the last seven days including today, oldest first.
	Days []DayDTO `json:"days"`
}

// Week maps the statistics.
func Week(days []domain.Day) WeekDTO {
	out := WeekDTO{Days: make([]DayDTO, 0, len(days))}
	for _, d := range days {
		out.Days = append(out.Days, DayDTO{Date: d.Date.Format(time.DateOnly), Sessions: d.Sessions, Minutes: d.Minutes})
	}
	return out
}

// TimerConfigDTO holds the phase lengths in minutes.
type TimerConfigDTO struct {
	Work      int `json:"work"`
	Short     int `json:"short"`
	Long      int `json:"long"`
	LongEvery int `json:"longEvery"`
}

// TimerDTO is the focus timer of the account (#631). The time left is
// relative, so a client whose clock is off still counts down correctly.
type TimerDTO struct {
	// Phase is work, short or long.
	Phase string `json:"phase"`
	// Done counts the focus sessions finished in the current cycle.
	Done    int  `json:"done"`
	Running bool `json:"running"`
	// RemainingMs is the time left when the answer was written.
	RemainingMs int64 `json:"remainingMs"`
	// TotalMs is the full length of the phase.
	TotalMs int64          `json:"totalMs"`
	Config  TimerConfigDTO `json:"config"`
}

// Timer maps the timer at now.
func Timer(t domain.Timer, now time.Time) TimerDTO {
	c := t.Config
	return TimerDTO{Phase: string(t.Phase), Done: t.Done, Running: t.Running(),
		RemainingMs: t.Remaining(now).Milliseconds(), TotalMs: t.Total.Milliseconds(),
		Config: TimerConfigDTO{Work: int(c.Work / time.Minute), Short: int(c.Short / time.Minute),
			Long: int(c.Long / time.Minute), LongEvery: c.LongEvery}}
}
