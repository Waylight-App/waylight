package usecase

import (
	"context"
	"fmt"
	"slices"
	"time"

	"waylight/internal/modules/contracts/domain"
	"waylight/internal/shared/kernel"
)

// Contracts is the input port of the module; every method is one use case.
type Contracts interface {
	// List returns the visible contracts, the next deadline first;
	// contracts with nothing left to cancel come last.
	List(ctx context.Context) ([]ContractView, error)
	// Create adds a contract; an empty visibility uses the default for new
	// records, private for members and admins.
	Create(ctx context.Context, in domain.Input, visibility string) (ContractView, error)
	// Update changes a contract of the principal; an empty visibility
	// keeps it.
	Update(ctx context.Context, id string, in domain.Input, visibility string) (ContractView, error)
	Delete(ctx context.Context, id string) error
	// Remind sends the reminders that are due (job) and says how many.
	Remind(ctx context.Context) (int, error)
}

// ContractView is a contract with its state today.
type ContractView struct {
	domain.Contract
	Term domain.Term
	// Soon is true inside the reminder period of the next deadline.
	Soon bool
	// Mine says whether the principal created the contract.
	Mine bool
}

// Repository persists contracts (output port, implemented in
// adapter/gateway). Reads return only what the principal in ctx may see;
// Save and Delete change only the principal's own contracts (ADR 0018).
type Repository interface {
	List(ctx context.Context) ([]domain.Contract, error)
	Get(ctx context.Context, id string) (domain.Contract, error)
	Save(ctx context.Context, c domain.Contract) error
	Delete(ctx context.Context, id string) error
	// MarkReminded records the deadline a reminder was sent for, whoever
	// owns the contract; only the reminder job calls it.
	MarkReminded(ctx context.Context, id, deadline string) error
}

// Deps bundles the output ports. Location decides what "today" is
// (nil = time.Local); Alerts receives the reminders (nil = none).
type Deps struct {
	Repo     Repository
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
	Location *time.Location
	Alerts   kernel.AlertSink
}

type uc struct{ d Deps }

// NewContracts returns the contract use cases.
func NewContracts(d Deps) Contracts {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &uc{d: d}
}

func (u *uc) today() time.Time { return u.d.Clock.Now().In(u.d.Location) }

func (u *uc) view(ctx context.Context, c domain.Contract) ContractView {
	today := u.today()
	return ContractView{Contract: c, Term: c.TermOn(today), Soon: c.Soon(today), Mine: c.Access.Mine(kernel.PrincipalFrom(ctx))}
}

// List implements Contracts.
func (u *uc) List(ctx context.Context) ([]ContractView, error) {
	list, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list contracts: %w", err)
	}
	out := make([]ContractView, 0, len(list))
	for _, c := range list {
		out = append(out, u.view(ctx, c))
	}
	slices.SortStableFunc(out, func(a, b ContractView) int {
		if a.Term.Open != b.Term.Open {
			if a.Term.Open {
				return -1
			}
			return 1
		}
		return a.Term.Deadline.Compare(b.Term.Deadline)
	})
	return out, nil
}

// Create implements Contracts.
func (u *uc) Create(ctx context.Context, in domain.Input, visibility string) (ContractView, error) {
	p := kernel.PrincipalFrom(ctx)
	access := kernel.NewAccess(p)
	if visibility != "" {
		v, err := kernel.CheckShare(p, visibility)
		if err != nil {
			return ContractView{}, err
		}
		access.Visibility = v
	}
	c, err := domain.New(u.d.IDs.NewID(), in, access, u.d.Clock.Now())
	if err != nil {
		return ContractView{}, err
	}
	if err := u.d.Repo.Save(ctx, c); err != nil {
		return ContractView{}, fmt.Errorf("save contract: %w", err)
	}
	return u.view(ctx, c), nil
}

// Update implements Contracts.
func (u *uc) Update(ctx context.Context, id string, in domain.Input, visibility string) (ContractView, error) {
	p := kernel.PrincipalFrom(ctx)
	c, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return ContractView{}, err
	}
	if !c.Access.Mine(p) {
		return ContractView{}, kernel.ErrNotOwner
	}
	if visibility != "" {
		if c.Access.Visibility, err = kernel.CheckShare(p, visibility); err != nil {
			return ContractView{}, err
		}
	}
	if c, err = c.Update(in); err != nil {
		return ContractView{}, err
	}
	if err := u.d.Repo.Save(ctx, c); err != nil {
		return ContractView{}, err
	}
	return u.view(ctx, c), nil
}

// Delete implements Contracts.
func (u *uc) Delete(ctx context.Context, id string) error { return u.d.Repo.Delete(ctx, id) }

// Remind implements Contracts. Reminders go to the instance's notification
// channels, so they cover the contracts the admin sees, never members'
// private ones, although the job runs as the system.
func (u *uc) Remind(ctx context.Context) (int, error) {
	if u.d.Alerts == nil {
		return 0, nil
	}
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	list, err := u.d.Repo.List(ctx)
	if err != nil {
		return 0, err
	}
	today, sent := u.today(), 0
	for _, c := range list {
		if !c.ReminderDue(today) {
			continue
		}
		term := c.TermOn(today)
		deadline := term.Deadline.Format(time.DateOnly)
		// Recorded first, so a failing channel never repeats a reminder.
		if err := u.d.Repo.MarkReminded(ctx, c.ID, deadline); err != nil {
			return sent, err
		}
		u.d.Alerts.Alert(ctx, kernel.Alert{Kind: "reminder", Title: c.Name, Body: reminderBody(c, term)})
		sent++
	}
	return sent, nil
}

// reminderBody says by when to cancel and what happens otherwise.
func reminderBody(c domain.Contract, t domain.Term) string {
	body := fmt.Sprintf("Cancel by %s to end the contract on %s.", t.Deadline.Format(time.DateOnly), t.End.Format(time.DateOnly))
	if c.RenewMonths > 0 {
		body += fmt.Sprintf(" Otherwise it renews for %d months.", c.RenewMonths)
	}
	return body
}
