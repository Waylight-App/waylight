package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/services/domain"
	"waylight/internal/shared/kernel"
)

type configRepo struct{ memRepo }

func (c *configRepo) ReplaceManual(_ context.Context, list []domain.Service) error {
	var kept []domain.Service
	for _, s := range c.items {
		if s.Container != "" {
			kept = append(kept, s)
		}
	}
	kept = append(kept, list...)
	c.items = kept
	return nil
}

func TestConfigKeepsIDsAndDiscoveredTiles(t *testing.T) {
	ctx := context.Background()
	r := &configRepo{memRepo{items: []domain.Service{
		{ID: "nas", Name: "NAS", URL: "https://nas.lan", CheckURL: "https://nas.lan"},
		{ID: "dock", Name: "Grafana", URL: "https://grafana.lan", CheckURL: "https://grafana.lan", Container: "grafana"},
	}}}
	uc := NewConfig(r, &ids{})
	list, _ := uc.Export(ctx)
	if len(list) != 1 || list[0].Name != "NAS" {
		t.Fatalf("discovered tiles must not be exported: %+v", list)
	}
	if err := uc.Apply(ctx, []ConfigService{{Name: "nas", URL: "https://nas.lan:5001"}, {Name: "Wiki", URL: "https://wiki.lan", MAC: "aa:bb:cc:dd:ee:ff"}}); err != nil {
		t.Fatal(err)
	}
	if len(r.items) != 3 || r.items[1].ID != "nas" || r.items[2].MAC == "" || r.items[0].Container != "grafana" {
		t.Fatalf("%+v", r.items)
	}
	if err := uc.Apply(ctx, []ConfigService{{Name: "A", URL: "https://a"}, {Name: "a", URL: "https://b"}}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("duplicate names: %v", err)
	}
	if err := uc.Apply(ctx, []ConfigService{{Name: "A", URL: "ftp://a"}}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("invalid URL: %v", err)
	}
}
