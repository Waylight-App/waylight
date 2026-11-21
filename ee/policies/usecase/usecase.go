// Package usecase reads and changes the admin policies (#304).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"fmt"
	"strings"

	licence "waylight/ee/licence/domain"
	"waylight/ee/policies/domain"
	"waylight/internal/shared/kernel"
)

// Store keeps the settings (output port); Load returns the zero Settings
// before the first save.
type Store interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// Licensing tells whether policies are licensed (output port).
type Licensing interface {
	Level(ctx context.Context, feature string) licence.Level
}

var errAdminOnly = kernel.NewPublicError(kernel.ErrForbidden, "Only an admin can see or change the policies.")

// Deps are the ports. Audit records changes in the security log.
type Deps struct {
	Store     Store
	Licensing Licensing
	Audit     kernel.AuditSink
}

// Service implements the use cases.
type Service struct{ d Deps }

// New returns the service.
func New(d Deps) *Service { return &Service{d: d} }

// Current returns the policies in force: none without a licence key; with a
// valid or expired key the saved ones, so an expired key does not quietly
// weaken the rules.
func (s *Service) Current(ctx context.Context) domain.Settings {
	if s.d.Licensing.Level(ctx, licence.FeaturePolicies) == licence.LevelNone {
		return domain.Settings{}
	}
	set, err := s.d.Store.Load(ctx)
	if err != nil {
		return domain.Settings{}
	}
	return set
}

// Get returns the saved policies for admins.
func (s *Service) Get(ctx context.Context) (domain.Settings, error) {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return domain.Settings{}, errAdminOnly
	}
	return s.d.Store.Load(ctx)
}

// Save validates and stores the policies and records the change in the
// security log.
func (s *Service) Save(ctx context.Context, in domain.Settings) (domain.Settings, error) {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return domain.Settings{}, errAdminOnly
	}
	set, err := in.Validate()
	if err != nil {
		return domain.Settings{}, err
	}
	if err := s.d.Store.Save(ctx, set); err != nil {
		return domain.Settings{}, fmt.Errorf("save policies: %w", err)
	}
	if s.d.Audit != nil {
		modules := "all"
		if set.MemberModules != nil {
			modules = strings.Join(set.MemberModules, ",")
		}
		s.d.Audit.Audit(ctx, kernel.AuditEvent{Type: "policies_changed", Detail: fmt.Sprintf(
			"second factor required: %t, minimum password length: %d, session hours: %d, member modules: %s",
			set.RequireSecondFactor, set.MinLength, set.MaxSessionHours, modules)})
	}
	return set, nil
}
