package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/search/domain"
	"waylight/internal/shared/kernel"
)

type configRepo struct{ memRepo }

func (c *configRepo) ReplaceShortcuts(_ context.Context, list []domain.Shortcut) error {
	c.shortcuts = list
	return nil
}

func (c *configRepo) ReplaceKeywords(_ context.Context, list []domain.Keyword) error {
	c.keywords = list
	return nil
}

func TestSearchConfigKeywords(t *testing.T) {
	ctx := context.Background()
	r := &configRepo{}
	uc := NewConfig(r)
	if err := uc.Apply(ctx, ConfigSection{Keywords: []KeywordView{{Keyword: "NAS", Kind: "url", Target: "https://nas.lan"}}}); err != nil {
		t.Fatal(err)
	}
	got, _ := uc.Export(ctx)
	if len(got.Keywords) != 1 || got.Keywords[0].Keyword != "nas" || got.Keywords[0].Target != "https://nas.lan" {
		t.Fatalf("%+v", got.Keywords)
	}
	for name, s := range map[string]ConfigSection{
		"prefix":    {Keywords: []KeywordView{{Keyword: "t", Kind: "query", Target: "x"}}},
		"duplicate": {Keywords: []KeywordView{{Keyword: "a", Kind: "query", Target: "x"}, {Keyword: "A", Kind: "query", Target: "y"}}},
		"invalid":   {Keywords: []KeywordView{{Keyword: "a", Kind: "url", Target: "nas.lan"}}},
	} {
		if err := uc.Apply(ctx, s); !errors.Is(err, kernel.ErrInvalid) || len(r.keywords) != 1 {
			t.Errorf("%s: want a validation error and no change, got %v (%d keywords)", name, err, len(r.keywords))
		}
	}
}

func TestSearchConfig(t *testing.T) {
	ctx := context.Background()
	r := &configRepo{}
	uc := NewConfig(r)
	err := uc.Apply(ctx, ConfigSection{Provider: "startpage", OpenInNewTab: true,
		Shortcuts: []ShortcutView{{Keyword: "!YT", Name: "YouTube", URLTemplate: "https://www.youtube.com/results?search_query={query}"}}})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := uc.Export(ctx)
	if got.Provider != "startpage" || !got.OpenInNewTab || got.Prefixes != domain.DefaultCapturePrefixes || got.Shortcuts[0].Keyword != "yt" {
		t.Fatalf("%+v", got)
	}
	dup := ConfigSection{Shortcuts: []ShortcutView{
		{Keyword: "w", Name: "A", URLTemplate: "https://a/{query}"}, {Keyword: "W", Name: "B", URLTemplate: "https://b/{query}"}}}
	if err := uc.Apply(ctx, dup); !errors.Is(err, kernel.ErrInvalid) || r.provider != "startpage" {
		t.Fatalf("duplicate keyword must change nothing: %v", err)
	}
	if err := uc.Apply(ctx, ConfigSection{Provider: "altavista"}); err == nil {
		t.Fatal("unknown provider accepted")
	}
}
