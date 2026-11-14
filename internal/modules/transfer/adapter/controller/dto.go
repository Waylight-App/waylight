package controller

import (
	"encoding/json"
	"time"
)

// DocumentDTO is the export file.
type DocumentDTO struct {
	Format     string                     `json:"format"`
	Version    int                        `json:"version"`
	ExportedAt time.Time                  `json:"exportedAt"`
	Sections   map[string]json.RawMessage `json:"sections"`
}

// ImportResultDTO is returned by POST /api/transfer/import.
type ImportResultDTO struct {
	Imported []string          `json:"imported"`
	Skipped  []string          `json:"skipped"`
	Failed   map[string]string `json:"failed"`
}

// ConfigResultDTO is returned by POST /api/transfer/config (#88).
type ConfigResultDTO struct {
	Applied []string          `json:"applied"`
	Failed  map[string]string `json:"failed"`
}

// ConfigStatusDTO is returned by GET /api/transfer/config/status (#88).
type ConfigStatusDTO struct {
	// Managed lists the sections CONFIG_FILE owns; empty without managed mode.
	Managed []string `json:"managed"`
}
