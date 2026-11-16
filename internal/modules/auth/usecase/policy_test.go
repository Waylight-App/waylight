package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

type fixedPolicy Policy

func (f fixedPolicy) Policy(context.Context) Policy { return Policy(f) }

func withPolicy(e *env, p Policy) *Policies {
	ps := &Policies{}
	ps.Set(fixedPolicy(p))
	e.deps.Policies = ps
	return ps
}

func TestPolicyRaisesTheMinimumPasswordLength(t *testing.T) {
	e := newEnv().withAccount()
	e.deps.MembersEnabled = true
	withPolicy(e, Policy{MinPasswordLength: 16})
	members := NewMembers(e.deps)
	ctxOwner := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: domain.OwnerID, Role: "owner"})
	if _, err := members.Create(ctxOwner, NewMemberRequest{Name: "anna", Password: "twelve-chars"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("short password accepted: %v", err)
	}
	if _, err := members.Create(ctxOwner, NewMemberRequest{Name: "anna", Password: "sixteen-chars-ok"}); err != nil {
		t.Fatalf("long password refused: %v", err)
	}
}

func TestPolicyEndsSessionsAfterTheMaximumAge(t *testing.T) {
	e := newEnv().withAccount()
	ps := withPolicy(e, Policy{MaxSessionAge: 8 * time.Hour})
	res, _ := NewLogin(e.deps).Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip"})
	auth := NewAuthenticateWithPolicies(e.sessions, e.accounts, nil, nil, e.clock, e.deps.TTL, ps, nil)
	for range 7 {
		// In use every hour: the sliding session alone would never end.
		e.clock.now = e.clock.now.Add(time.Hour)
		if _, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: res.SessionToken}); err != nil {
			t.Fatalf("within the maximum age: %v", err)
		}
	}
	e.clock.now = e.clock.now.Add(time.Hour)
	if _, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: res.SessionToken}); !errors.Is(err, domain.ErrSessionInvalid) {
		t.Fatalf("after the maximum age: %v", err)
	}
}

func TestPolicyRequiresASecondFactor(t *testing.T) {
	e := newEnv().withAccount()
	ps := withPolicy(e, Policy{RequireSecondFactor: true})
	res, _ := NewLogin(e.deps).Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip"})
	got, err := NewAuthenticateWithPolicies(e.sessions, e.accounts, nil, nil, e.clock, e.deps.TTL, ps, nil).
		Execute(ctx, AuthenticateRequest{SessionToken: res.SessionToken})
	if err != nil || !got.Principal.SecondFactorMissing {
		t.Fatalf("principal %+v %v", got.Principal, err)
	}
	st, _ := NewGetStatusWithPolicies(e.accounts, e.sessions, e.clock, ps, nil).Execute(ctx, StatusRequest{SessionToken: res.SessionToken})
	if !st.SecondFactorRequired {
		t.Fatal("status does not ask for the second factor")
	}
	authz := NewAuthorizerWithPolicies(nil, ps)
	if err := authz.Authorize(ctx, got.Principal, "GET", "/api/notes"); !errors.Is(err, domain.ErrSecondFactorRequired) {
		t.Fatalf("notes reachable: %v", err)
	}
	if err := authz.Authorize(ctx, got.Principal, "POST", "/api/auth/totp/setup"); err != nil {
		t.Fatalf("setup blocked: %v", err)
	}
}

func TestPolicyLimitsMemberModules(t *testing.T) {
	ps := &Policies{}
	ps.Set(fixedPolicy{MemberModules: []string{"notes"}})
	authz := NewAuthorizerWithPolicies(nil, ps)
	member := kernel.Principal{AccountID: "m1", Role: "member"}
	if err := authz.Authorize(ctx, member, "GET", "/api/todos"); !errors.Is(err, domain.ErrMemberNotAllowed) {
		t.Fatalf("todos reachable: %v", err)
	}
	if err := authz.Authorize(ctx, member, "GET", "/api/dashboard"); errors.Is(err, domain.ErrMemberNotAllowed) {
		t.Fatalf("dashboard is always allowed: %v", err)
	}
	owner := kernel.Principal{AccountID: domain.OwnerID, Role: "owner"}
	if err := authz.Authorize(ctx, owner, "GET", "/api/todos"); err != nil {
		t.Fatalf("owner limited: %v", err)
	}
}
