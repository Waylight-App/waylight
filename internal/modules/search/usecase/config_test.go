package usecase

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/search/domain"
	"trisse-home/internal/shared/kernel"
)

type configRepo struct{ memRepo }

func (c *configRepo) ReplaceShortcuts(_ context.Context, list []domain.Shortcut) error {
	c.shortcuts = list
	return nil
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
