package usecase

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"waylight/internal/modules/connections/domain"
	"waylight/internal/shared/kernel"
)

// ConfigConnection is one connection in waylight.yaml (#88, #610): service,
// name and address, never the credential.
type ConfigConnection struct {
	Service, Name, URL string
}

// ConfigRepository adds the atomic replacement of all connections.
type ConfigRepository interface {
	Repository
	// ReplaceAll makes list the complete set; connections whose IDs are
	// missing are deleted.
	ReplaceAll(ctx context.Context, list []domain.Connection) error
}

// Config exports and applies the connections section (input port).
type Config interface {
	Export(ctx context.Context) ([]ConfigConnection, error)
	Apply(ctx context.Context, list []ConfigConnection) error
}

type config struct {
	repo  ConfigRepository
	clock kernel.Clock
	ids   kernel.IDGenerator
}

// NewConfig returns the Config use case.
func NewConfig(repo ConfigRepository, clock kernel.Clock, ids kernel.IDGenerator) Config {
	return &config{repo: repo, clock: clock, ids: ids}
}

// Export implements Config, ordered by service and name so the file reads
// the same after every export.
func (c *config) Export(ctx context.Context) ([]ConfigConnection, error) {
	list, err := c.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}
	out := make([]ConfigConnection, 0, len(list))
	for _, x := range list {
		out = append(out, ConfigConnection{Service: x.Service, Name: x.Name, URL: x.URL})
	}
	slices.SortFunc(out, func(a, b ConfigConnection) int {
		return cmp.Or(cmp.Compare(a.Service, b.Service), cmp.Compare(a.Name, b.Name))
	})
	return out, nil
}

// Apply implements Config. Connections are matched by service and name, so
// their IDs (which widgets name) and credentials survive; a credential is
// dropped when the address of its connection changes, so it is never sent
// to another host (#191). Nothing changes when an entry is invalid.
func (c *config) Apply(ctx context.Context, list []ConfigConnection) error {
	if len(list) > domain.MaxConnections {
		return domain.ErrTooMany
	}
	existing, err := c.repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list connections: %w", err)
	}
	out := make([]domain.Connection, 0, len(list))
	for _, in := range list {
		id, cred, created := c.ids.NewID(), domain.Credential{}, c.clock.Now()
		var oldURL string
		for _, o := range existing {
			if o.Service == in.Service && domain.SameName(o.Name, in.Name) {
				id, cred, created, oldURL = o.ID, o.Credential, o.CreatedAt, o.URL
				break
			}
		}
		x, err := domain.NewConnection(id, in.Service, in.Name, in.URL, cred, created)
		if err != nil {
			return err
		}
		if oldURL != "" && kernel.TargetChanged([]string{oldURL}, []string{x.URL}) {
			x.Credential = domain.Credential{}
		}
		if taken(out, x) {
			return domain.ErrNameTaken
		}
		out = append(out, x)
	}
	if err := c.repo.ReplaceAll(ctx, out); err != nil {
		return fmt.Errorf("replace connections: %w", err)
	}
	return nil
}
