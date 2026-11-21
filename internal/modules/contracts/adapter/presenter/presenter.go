// Package presenter maps contracts to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/contracts/usecase"
)

// ContractDTO is one contract in API responses.
type ContractDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	// CostCents is the cost per period in the smallest currency unit.
	CostCents int `json:"costCents"`
	// Period is "monthly", "quarterly" or "yearly".
	Period string `json:"period"`
	// MonthlyCents is the cost spread over one month.
	MonthlyCents int `json:"monthlyCents"`
	// TermEnd is the end of the term as entered, YYYY-MM-DD.
	TermEnd     string `json:"termEnd"`
	NoticeDays  int    `json:"noticeDays"`
	RenewMonths int    `json:"renewMonths"`
	RemindDays  int    `json:"remindDays"`
	Note        string `json:"note"`
	// CurrentEnd and Deadline belong to the next deadline, renewals
	// applied (YYYY-MM-DD).
	CurrentEnd string `json:"currentEnd"`
	Deadline   string `json:"deadline"`
	// Open is false when the contract ends and its last deadline passed.
	Open bool `json:"open"`
	// Soon is true inside the reminder period of the deadline.
	Soon bool `json:"soon"`
	// Visibility is "private", "everyone" or "group:<id>" (ADR 0018).
	Visibility string `json:"visibility"`
	// Mine is true when the viewer created the contract and may change it.
	Mine bool `json:"mine"`
}

// Contract maps one contract.
func Contract(v usecase.ContractView) ContractDTO {
	return ContractDTO{ID: v.ID, Name: v.Name, Provider: v.Provider, CostCents: v.CostCents, Period: string(v.Period), MonthlyCents: v.MonthlyCents(),
		TermEnd: v.TermEndDate(), NoticeDays: v.NoticeDays, RenewMonths: v.RenewMonths, RemindDays: v.RemindDays, Note: v.Note,
		CurrentEnd: v.Term.End.Format(time.DateOnly), Deadline: v.Term.Deadline.Format(time.DateOnly), Open: v.Term.Open, Soon: v.Soon,
		Visibility: string(v.Access.Visibility), Mine: v.Mine}
}

// Contracts maps a list of contracts.
func Contracts(list []usecase.ContractView) []ContractDTO {
	out := make([]ContractDTO, 0, len(list))
	for _, v := range list {
		out = append(out, Contract(v))
	}
	return out
}
