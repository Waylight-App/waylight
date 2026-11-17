// Package presenter maps shopping list items to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"waylight/internal/modules/shopping/domain"
	"waylight/internal/modules/shopping/usecase"
)

// SettingsDTO is returned by the settings endpoints (never the password).
type SettingsDTO struct {
	Configured  bool   `json:"configured"`
	ListURL     string `json:"listUrl"`
	Username    string `json:"username"`
	HasPassword bool   `json:"hasPassword"`
}

// ItemDTO is one list entry.
type ItemDTO struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Done    bool   `json:"done"`
}

// ClearedDTO is returned by POST /api/shopping/clear-completed.
type ClearedDTO struct {
	Deleted int `json:"deleted"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

// Item maps one item.
func Item(it domain.Item) ItemDTO { return ItemDTO{ID: it.ID, Summary: it.Summary, Done: it.Done} }

// Items maps a list of items.
func Items(items []domain.Item) []ItemDTO {
	out := make([]ItemDTO, 0, len(items))
	for _, it := range items {
		out = append(out, Item(it))
	}
	return out
}

// ListDTO is a shopping list kept on waylight (#471).
type ListDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
	// Mine is false for a list someone else shared.
	Mine bool `json:"mine"`
	// Open and Done count the items.
	Open int `json:"open"`
	Done int `json:"done"`
}

// ListsDTO is GET /api/shopping/lists.
type ListsDTO struct {
	Lists []ListDTO `json:"lists"`
	// CalDAV tells whether a CalDAV list is connected.
	CalDAV bool `json:"caldav"`
	// Current is what the widget shows: a list ID, "caldav", or "".
	Current string `json:"current"`
}

// ImportedDTO answers an import.
type ImportedDTO struct {
	Imported int `json:"imported"`
}

// List maps a list.
func List(v usecase.ListView) ListDTO {
	return ListDTO{ID: v.ID, Name: v.Name, Visibility: string(v.Access.Visibility), Mine: v.Mine, Open: v.Open, Done: v.Done}
}

// Lists maps the lists.
func Lists(v usecase.ListsView) ListsDTO {
	out := ListsDTO{Lists: make([]ListDTO, 0, len(v.Lists)), CalDAV: v.CalDAV, Current: v.Current}
	for _, l := range v.Lists {
		out.Lists = append(out.Lists, List(l))
	}
	return out
}
