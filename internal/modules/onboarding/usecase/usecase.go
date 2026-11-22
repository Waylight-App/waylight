package usecase

import (
	"context"
	"fmt"
	"time"

	"waylight/internal/modules/onboarding/domain"
	"waylight/internal/shared/kernel"
)

// StatusResponse describes the tour state.
type StatusResponse struct {
	Completed   bool
	CompletedAt time.Time
	// Step is the wizard step the account stopped at; empty before it started.
	Step string
}

// Onboarding is the input port of the module; every method is one use case.
type Onboarding interface {
	Status(ctx context.Context) (StatusResponse, error)
	Complete(ctx context.Context) (StatusResponse, error)
	Restart(ctx context.Context) (StatusResponse, error)
	// Resume stores the wizard step and makes the tour pending, e.g. "done"
	// to show the tour alone.
	Resume(ctx context.Context, step string) (StatusResponse, error)
}

// Repository persists the state (output port). Load returns the zero state
// when nothing has been stored yet.
type Repository interface {
	Load(ctx context.Context) (domain.State, error)
	Save(ctx context.Context, s domain.State) error
}

type onboarding struct {
	repo  Repository
	clock kernel.Clock
}

// NewOnboarding returns the onboarding use cases.
func NewOnboarding(repo Repository, clock kernel.Clock) Onboarding {
	return &onboarding{repo: repo, clock: clock}
}

func toResponse(s domain.State) StatusResponse {
	return StatusResponse{Completed: s.Completed(), CompletedAt: s.CompletedAt(), Step: s.Step()}
}

// Status implements Onboarding.
func (uc *onboarding) Status(ctx context.Context) (StatusResponse, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("load onboarding state: %w", err)
	}
	return toResponse(s), nil
}

// Complete implements Onboarding.
func (uc *onboarding) Complete(ctx context.Context) (StatusResponse, error) {
	return uc.update(ctx, func(s *domain.State) error { s.Complete(uc.clock.Now()); return nil })
}

// Restart implements Onboarding.
func (uc *onboarding) Restart(ctx context.Context) (StatusResponse, error) {
	return uc.update(ctx, func(s *domain.State) error { s.Restart(); return nil })
}

// Resume implements Onboarding.
func (uc *onboarding) Resume(ctx context.Context, step string) (StatusResponse, error) {
	return uc.update(ctx, func(s *domain.State) error { return s.Resume(step) })
}

func (uc *onboarding) update(ctx context.Context, change func(*domain.State) error) (StatusResponse, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("load onboarding state: %w", err)
	}
	if err := change(&s); err != nil {
		return StatusResponse{}, err
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return StatusResponse{}, fmt.Errorf("save onboarding state: %w", err)
	}
	return toResponse(s), nil
}
