package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/transfer/domain"
	"waylight/internal/shared/kernel"
)

type planSource struct{ plan domain.ImportPlan }

func (p planSource) Parse(map[string][]byte) (domain.ImportPlan, error) { return p.plan, nil }

type fakeConfig struct {
	Config
	applied []byte
}

func (f *fakeConfig) Export(context.Context) ([]byte, error) { return []byte("current"), nil }
func (f *fakeConfig) Apply(_ context.Context, data []byte) (ApplyResult, error) {
	f.applied = data
	return ApplyResult{Applied: []string{"services"}}, nil
}

// genericCodec hands out the current settings and records what is encoded.
type genericCodec struct{ encoded []NamedSection }

func (g *genericCodec) Encode(_ int, s []NamedSection) ([]byte, error) {
	g.encoded = s
	return []byte("merged"), nil
}
func (g *genericCodec) Decode(_ []byte, _ func(string) (string, error), target func(string) (any, bool)) (int, []NamedSection, error) {
	svc, _ := target("services")
	*(svc.(*any)) = []any{map[string]any{"name": "NAS", "url": "https://nas"}}
	bm, _ := target("bookmarks")
	*(bm.(*any)) = []any{}
	return 1, []NamedSection{{Name: "services", Value: svc}, {Name: "bookmarks", Value: bm}}, nil
}

func TestImporter(t *testing.T) {
	plan := domain.ImportPlan{Source: "Homepage", Tiles: []domain.ImportTile{{Name: "Wiki", URL: "https://wiki"}}}
	cfg := &fakeConfig{}
	codec := &genericCodec{}
	uc := NewImporter(map[string]ImportSource{"homepage": planSource{plan}}, cfg, codec, func(string) bool { return true })

	out, err := uc.Import(context.Background(), "Homepage", nil, false)
	if err != nil || out.Result != nil || cfg.applied != nil || out.Summary.Tiles != 1 {
		t.Fatalf("preview: %+v %v", out, err)
	}
	if len(codec.encoded) != 1 || codec.encoded[0].Name != "services" {
		t.Fatalf("only touched sections belong in the result: %+v", codec.encoded)
	}
	if out, _ := uc.Import(context.Background(), "homepage", nil, true); out.Result == nil || string(cfg.applied) != "merged" {
		t.Fatalf("apply: %+v", out)
	}
	if _, err := uc.Import(context.Background(), "excel", nil, false); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("unknown format: %v", err)
	}
}
