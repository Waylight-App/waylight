// Package presenter maps phone list entries to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/phonelist/domain"
	"waylight/internal/modules/phonelist/usecase"
)

// EntryDTO is one person of the phone list.
type EntryDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Role  string `json:"role"`
	Phone string `json:"phone"`
	// Dial is the number for a tel: link (digits and a leading +); empty
	// without a phone number.
	Dial  string `json:"dial"`
	Email string `json:"email"`
	Team  string `json:"team"`
	// ReadOnly entries come from the address book and are changed there.
	ReadOnly bool `json:"readOnly"`
}

// Entry maps one entry.
func Entry(e domain.Entry) EntryDTO {
	return EntryDTO{ID: e.ID, Name: e.Name, Role: e.Role, Phone: e.Phone, Dial: domain.DialNumber(e.Phone), Email: e.Email, Team: e.Team, ReadOnly: e.Source == domain.SourceCardDAV}
}

// Entries maps a list of entries.
func Entries(list []domain.Entry) []EntryDTO {
	out := make([]EntryDTO, 0, len(list))
	for _, e := range list {
		out = append(out, Entry(e))
	}
	return out
}

// SettingsDTO is the address book configuration, without the password.
type SettingsDTO struct {
	CardDAVURL  string     `json:"cardDavUrl"`
	Username    string     `json:"username"`
	HasPassword bool       `json:"hasPassword"`
	LastSync    *time.Time `json:"lastSync"`
	LastError   string     `json:"lastError"`
}

// Settings maps the configuration.
func Settings(s usecase.SettingsView) SettingsDTO {
	d := SettingsDTO{CardDAVURL: s.CardDAVURL, Username: s.Username, HasPassword: s.HasPassword, LastError: s.LastError}
	if !s.LastSync.IsZero() {
		t := s.LastSync
		d.LastSync = &t
	}
	return d
}
