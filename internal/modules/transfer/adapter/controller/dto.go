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

// ImportSummaryDTO says what an import from another dashboard adds (#77).
type ImportSummaryDTO struct {
	Tiles     int      `json:"tiles"`
	Bookmarks int      `json:"bookmarks"`
	Widgets   int      `json:"widgets"`
	Skipped   int      `json:"skipped"`
	Tab       string   `json:"tab"`
	Warnings  []string `json:"warnings"`
}

// ForeignImportDTO is returned by POST /api/transfer/import/{format}: the
// merged waylight.yaml, the summary and, when applied, the result.
type ForeignImportDTO struct {
	YAML    string           `json:"yaml"`
	Summary ImportSummaryDTO `json:"summary"`
	Result  *ConfigResultDTO `json:"result,omitempty"`
}
