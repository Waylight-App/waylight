package controller

// ContractInputDTO is the body of POST /api/contracts and
// PUT /api/contracts/{id}.
type ContractInputDTO struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	// CostCents is the cost per period in the smallest currency unit.
	CostCents int `json:"costCents"`
	// Period is "monthly", "quarterly" or "yearly".
	Period string `json:"period"`
	// TermEnd is the last day of the current term, YYYY-MM-DD.
	TermEnd     string `json:"termEnd"`
	NoticeDays  int    `json:"noticeDays"`
	RenewMonths int    `json:"renewMonths"`
	RemindDays  int    `json:"remindDays"`
	Note        string `json:"note"`
	// Visibility is "private", "everyone" or "group:<id>". Empty uses the
	// default for new records and keeps the visibility of an existing one.
	Visibility string `json:"visibility,omitempty"`
}
