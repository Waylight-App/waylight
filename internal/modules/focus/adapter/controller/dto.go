package controller

import "time"

// SessionInputDTO is the body of POST /api/focus/sessions: a finished
// focus phase.
type SessionInputDTO struct {
	Minutes int       `json:"minutes"`
	EndedAt time.Time `json:"endedAt"`
}

// TimerConfigInputDTO is the body of PUT /api/focus/timer/config: the
// widget's phase lengths in minutes.
type TimerConfigInputDTO struct {
	Work      int `json:"work"`
	Short     int `json:"short"`
	Long      int `json:"long"`
	LongEvery int `json:"longEvery"`
}
