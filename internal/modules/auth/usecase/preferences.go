package usecase

import (
	"context"
	"errors"
	"fmt"
)

// Preferences reads and changes the admin's choice whether their own new
// records start private (input port, ADR 0018 migration step 3). Members
// always start private, so only an admin has the choice.
type Preferences interface {
	PrivateByDefault(ctx context.Context) (bool, error)
	SetPrivateByDefault(ctx context.Context, on bool) error
}

type preferences struct{ deps Deps }

// NewPreferences returns the Preferences use cases.
func NewPreferences(deps Deps) Preferences { return &preferences{deps: deps} }

var errNoPreferences = errors.New("preferences are not available")

// PrivateByDefault implements Preferences.
func (uc *preferences) PrivateByDefault(ctx context.Context) (bool, error) {
	if err := ownerOnly(ctx); err != nil {
		return false, err
	}
	if uc.deps.Preferences == nil {
		return false, errNoPreferences
	}
	return uc.deps.Preferences.PrivateByDefault(ctx, me(ctx))
}

// SetPrivateByDefault implements Preferences. It takes effect with the next
// request, when the principal is built again.
func (uc *preferences) SetPrivateByDefault(ctx context.Context, on bool) error {
	if err := ownerOnly(ctx); err != nil {
		return err
	}
	if uc.deps.Preferences == nil {
		return errNoPreferences
	}
	if err := uc.deps.Preferences.SetPrivateByDefault(ctx, me(ctx), on); err != nil {
		return fmt.Errorf("save preference: %w", err)
	}
	return nil
}
