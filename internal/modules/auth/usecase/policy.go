package usecase

import (
	"context"
	"slices"
	"sync"
	"time"
	"unicode/utf8"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

// Policy holds instance-wide rules that admins set (#304). The zero value is
// the core behaviour: no extra rules.
type Policy struct {
	// RequireSecondFactor makes every account without TOTP or a passkey set
	// one up before it can use anything else.
	RequireSecondFactor bool
	// MinPasswordLength raises the minimum above domain.MinPasswordLength.
	MinPasswordLength int
	// MaxSessionAge ends a session this long after sign-in, even while it is
	// in use; zero keeps sliding sessions without an upper limit.
	MaxSessionAge time.Duration
	// MemberModules lists the modules members may use; nil allows all.
	MemberModules []string
}

// alwaysForMembers are modules a member cannot work without, so a policy can
// never take them away. "modules" is the core's /api/modules, which tells the
// interface what to show.
var alwaysForMembers = []string{"auth", "dashboard", "appearance", "onboarding", "search", "licence", "modules"}

// MemberMayUse reports whether members may use a module.
func (p Policy) MemberMayUse(module string) bool {
	return p.MemberModules == nil || slices.Contains(alwaysForMembers, module) || slices.Contains(p.MemberModules, module)
}

// PolicySource provides the current policy (output port). Only the ee/
// admin policies (#304) set one.
type PolicySource interface {
	Policy(ctx context.Context) Policy
}

// Policies holds the source once the server is built; Deps keep a pointer to
// it so every use case sees a source set later.
type Policies struct {
	mu  sync.Mutex
	src PolicySource
}

// Set installs the source.
func (p *Policies) Set(src PolicySource) {
	p.mu.Lock()
	p.src = src
	p.mu.Unlock()
}

// Current returns the policy in force; the zero Policy without a source.
func (p *Policies) Current(ctx context.Context) Policy {
	if p == nil {
		return Policy{}
	}
	p.mu.Lock()
	src := p.src
	p.mu.Unlock()
	if src == nil {
		return Policy{}
	}
	return src.Policy(ctx)
}

// newPassword applies the domain's password rules and the policy's minimum.
func (d Deps) newPassword(ctx context.Context, raw string) (domain.Password, error) {
	pw, err := domain.NewPassword(raw)
	if err != nil {
		return domain.Password{}, err
	}
	if min := d.Policies.Current(ctx).MinPasswordLength; utf8.RuneCountInString(raw) < min {
		return domain.Password{}, kernel.NewValidationErrorf("password", "The password must be at least {min} characters long.", "min", min)
	}
	return pw, nil
}
