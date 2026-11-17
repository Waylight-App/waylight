// Package presenter maps birthdays to JSON (#287).
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/birthdays/domain"
	"waylight/internal/modules/birthdays/usecase"
)

// BirthdayDTO is one birthday with its next occurrence.
type BirthdayDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Month  int    `json:"month"`
	Day    int    `json:"day"`
	Year   int    `json:"year,omitempty"`
	Source string `json:"source"`
	Next   string `json:"next"` // YYYY-MM-DD
	InDays int    `json:"inDays"`
	Age    int    `json:"age,omitempty"`
}

// SettingsDTO never contains the password.
type SettingsDTO struct {
	RemindEnabled bool   `json:"remindEnabled"`
	RemindAt      string `json:"remindAt"`
	CardDAVURL    string `json:"cardDavUrl"`
	Username      string `json:"username"`
	HasPassword   bool   `json:"hasPassword"`
}

// ImportReportDTO says what an import did.
type ImportReportDTO struct {
	Added      int `json:"added"`
	Duplicates int `json:"duplicates"`
}

// Upcoming maps the next birthdays.
func Upcoming(list []domain.Upcoming) []BirthdayDTO {
	out := make([]BirthdayDTO, 0, len(list))
	for _, u := range list {
		out = append(out, BirthdayDTO{ID: u.ID, Name: u.Name, Month: int(u.Month), Day: u.Day, Year: u.Year, Source: u.Source,
			Next: u.Date.Format(time.DateOnly), InDays: u.InDays, Age: u.Age})
	}
	return out
}

// Settings maps the settings view.
func Settings(v usecase.SettingsView) SettingsDTO {
	return SettingsDTO{RemindEnabled: v.Reminder.Enabled, RemindAt: v.Reminder.At, CardDAVURL: v.CardDAVURL, Username: v.Username, HasPassword: v.HasPassword}
}
