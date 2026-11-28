package controller

// RuleInputDTO maps summaries containing match to a label, color and icon.
type RuleInputDTO struct {
	Match string `json:"match"`
	Label string `json:"label"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

// SettingsInputDTO is the body of PUT /api/waste/settings.
type SettingsInputDTO struct {
	URL   string         `json:"url"`
	Rules []RuleInputDTO `json:"rules"`
}

// ReminderInputDTO is the body of PUT /api/waste/reminder (#866).
type ReminderInputDTO struct {
	Enabled bool `json:"enabled"`
	// Time is "HH:MM" on the evening before a collection.
	Time string `json:"time"`
}
