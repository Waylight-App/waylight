package usecase

import (
	"context"
	"fmt"
	"strings"

	"waylight/internal/modules/services/domain"
	"waylight/internal/shared/kernel"
)

// ConfigService is one tile in waylight.yaml (#88). Tiles created from
// container labels are left out: Docker is their source of truth.
type ConfigService struct {
	Name          string
	URL           string
	CheckURL      string
	Icon          string
	Description   string
	SkipTLSVerify bool
	MAC           string
}

// ConfigRepository adds the atomic replacement of the manual tiles.
type ConfigRepository interface {
	Repository
	// ReplaceManual makes list the complete set of manual tiles: tiles
	// whose ID is missing are deleted, the others inserted or updated.
	// Discovered tiles are not touched.
	ReplaceManual(ctx context.Context, list []domain.Service) error
}

// Config exports and applies the services section (input port).
type Config interface {
	Export(ctx context.Context) ([]ConfigService, error)
	Apply(ctx context.Context, list []ConfigService) error
}

type config struct {
	repo ConfigRepository
	ids  kernel.IDGenerator
}

// NewConfig returns the Config use case.
func NewConfig(repo ConfigRepository, ids kernel.IDGenerator) Config {
	return &config{repo: repo, ids: ids}
}

func (uc *config) Export(ctx context.Context) ([]ConfigService, error) {
	list, err := uc.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	out := []ConfigService{}
	for _, s := range list {
		if s.Container != "" {
			continue
		}
		out = append(out, ConfigService{Name: s.Name, URL: s.URL, CheckURL: s.CheckURL, Icon: s.Icon,
			Description: s.Description, SkipTLSVerify: s.SkipTLSVerify, MAC: string(s.MAC)})
	}
	return out, nil
}

var errDuplicateName = kernel.NewValidationError("name", "Each service name can be used only once in the file.")

// Apply keeps the ID of a manual tile with the same name, so its uptime
// history and any widget that points to it survive a re-import.
func (uc *config) Apply(ctx context.Context, in []ConfigService) error {
	existing, err := uc.repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list services: %w", err)
	}
	byName := map[string]string{}
	for _, s := range existing {
		if s.Container == "" {
			byName[strings.ToLower(s.Name)] = s.ID
		}
	}
	seen := map[string]bool{}
	list := make([]domain.Service, 0, len(in))
	for i, c := range in {
		key := strings.ToLower(strings.TrimSpace(c.Name))
		if seen[key] {
			return errDuplicateName
		}
		seen[key] = true
		id, ok := byName[key]
		if !ok {
			id = uc.ids.NewID()
		}
		s, err := domain.NewService(id, c.Name, c.URL, c.CheckURL, c.Icon, c.Description, c.SkipTLSVerify, i)
		if err != nil {
			return err
		}
		if err := s.SetMAC(c.MAC); err != nil {
			return err
		}
		list = append(list, s)
	}
	return uc.repo.ReplaceManual(ctx, list)
}
