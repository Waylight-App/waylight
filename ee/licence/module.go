// Package licence is the licence key module of the ee/ features (#307,
// ADR 0017, ADR 0020).
//
// Ring: module wiring. module.go connects the domain, use case and adapter
// packages below and implements modules.Module; it contains no business
// logic. Other ee/ modules receive Licensing and implement modules.Gated
// with it, so the core hides and locks them without importing ee/.
package licence

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"io/fs"

	"trisse-home/ee/licence/adapter/controller"
	"trisse-home/ee/licence/adapter/gateway"
	"trisse-home/ee/licence/domain"
	"trisse-home/ee/licence/usecase"
	"trisse-home/internal/modules"
	"trisse-home/internal/shared/kernel"
)

// Licensing is the port the ee/ features ask.
type Licensing = usecase.Licensing

// Module wires the licence rings and implements modules.Module.
type Module struct {
	modules.Base
	svc      *usecase.Service
	handlers *controller.Handlers
}

// Config of the module. EnvKey is LICENCE_KEY; PublicKey is only set by
// tests, the module otherwise uses the compiled-in key.
type Config struct {
	EnvKey    string
	PublicKey ed25519.PublicKey
}

// New builds the module.
func New(db *sql.DB, clock kernel.Clock, accounts usecase.AccountCounter, cfg Config) *Module {
	pub := cfg.PublicKey
	if pub == nil {
		pub = publicKey
	}
	svc := usecase.New(usecase.Deps{Store: gateway.NewStore(db, clock), Accounts: accounts, Clock: clock,
		PublicKey: pub, EnvKey: cfg.EnvKey})
	return &Module{svc: svc, handlers: controller.NewHandlers(svc)}
}

// Licensing returns the port for the other ee/ modules.
func (m *Module) Licensing() Licensing { return m.svc }

// Name implements modules.Module.
func (m *Module) Name() string { return "licence" }

// DataKind implements modules.Module: one key for the instance.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module. Only admins see or change the key
// (checked in the use cases); every account may ask which features it can use.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/licence", m.handlers.Get)
	r.Protected("PUT /api/licence", m.handlers.Install)
	r.Protected("DELETE /api/licence", m.handlers.Remove)
	r.Protected("GET /api/licence/features", m.handlers.Features)
}

// Gate turns the level of a feature into modules.Access, for ee/ modules that
// implement modules.Gated:
//
//	func (m *Module) Access(ctx context.Context) modules.Access {
//		return licence.Gate(ctx, m.licensing, licence.FeatureBranding)
//	}
func Gate(ctx context.Context, l Licensing, feature string) modules.Access {
	if l == nil {
		return modules.AccessNone
	}
	switch l.Level(ctx, feature) {
	case domain.LevelFull:
		return modules.AccessFull
	case domain.LevelRead:
		return modules.AccessRead
	default:
		return modules.AccessNone
	}
}

// Feature names, re-exported for the other ee/ modules.
const (
	FeatureAnnouncements     = domain.FeatureAnnouncements
	FeatureAbsences          = domain.FeatureAbsences
	FeatureTeamTabs          = domain.FeatureTeamTabs
	FeatureSSO               = domain.FeatureSSO
	FeatureAuditLog          = domain.FeatureAuditLog
	FeaturePolicies          = domain.FeaturePolicies
	FeatureBranding          = domain.FeatureBranding
	FeatureConnectedAccounts = domain.FeatureConnectedAccounts
	FeatureWorkplaceSearch   = domain.FeatureWorkplaceSearch
	FeatureChatApps          = domain.FeatureChatApps
	FeaturePresence          = domain.FeaturePresence
	FeatureDirectorySync     = domain.FeatureDirectorySync
)
