package usecase

import (
	"context"
	"fmt"

	"waylight/internal/modules/search/domain"
	"waylight/internal/shared/kernel"
)

// ConfigSection is the search part of waylight.yaml (#88).
type ConfigSection struct {
	Provider     string
	OpenInNewTab bool
	Prefixes     domain.CapturePrefixes
	Shortcuts    []ShortcutView
	Keywords     []KeywordView
}

// ConfigRepository adds the atomic replacement of all shortcuts.
type ConfigRepository interface {
	Repository
	ReplaceShortcuts(ctx context.Context, list []domain.Shortcut) error
	ReplaceKeywords(ctx context.Context, list []domain.Keyword) error
}

// Config exports and applies the search section (input port).
type Config interface {
	Export(ctx context.Context) (ConfigSection, error)
	Apply(ctx context.Context, s ConfigSection) error
}

type config struct {
	repo     ConfigRepository
	searcher Searcher
}

// NewConfig returns the Config use case.
func NewConfig(repo ConfigRepository) Config { return &config{repo: repo, searcher: NewSearcher(repo)} }

func (uc *config) Export(ctx context.Context) (ConfigSection, error) {
	st, err := uc.searcher.Settings(ctx)
	if err != nil {
		return ConfigSection{}, err
	}
	return ConfigSection{
		Provider:     st.DefaultProvider,
		OpenInNewTab: st.Bar.OpenInNewTab,
		Prefixes:     domain.CapturePrefixes{Todo: st.Bar.TodoPrefix, Note: st.Bar.NotePrefix, Bookmark: st.Bar.BookmarkPrefix},
		Shortcuts:    st.Shortcuts,
		Keywords:     st.Keywords,
	}, nil
}

var (
	errDuplicateKeyword    = kernel.NewValidationError("keyword", "Each shortcut keyword can be used only once.")
	errDuplicateOwnKeyword = kernel.NewValidationError("keyword", "Each own keyword can be used only once.")
)

// Apply validates everything first, so a mistake in the file changes nothing.
func (uc *config) Apply(ctx context.Context, s ConfigSection) error {
	provider := s.Provider
	if provider == "" {
		provider = domain.DefaultProviderID
	}
	if _, err := domain.ProviderByID(provider); err != nil {
		return err
	}
	pre := s.Prefixes
	if pre == (domain.CapturePrefixes{}) {
		pre = domain.DefaultCapturePrefixes
	}
	bar, err := domain.NewBarOptions(s.OpenInNewTab, pre)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	list := make([]domain.Shortcut, 0, len(s.Shortcuts))
	for _, v := range s.Shortcuts {
		sc, err := domain.NewShortcut(v.Keyword, v.Name, v.URLTemplate)
		if err != nil {
			return err
		}
		if seen[sc.Keyword()] {
			return errDuplicateKeyword
		}
		seen[sc.Keyword()] = true
		list = append(list, sc)
	}
	words := map[string]bool{}
	keywords := make([]domain.Keyword, 0, len(s.Keywords))
	for _, v := range s.Keywords {
		k, err := domain.NewKeyword(v.Keyword, domain.TargetKind(v.Kind), v.Target, v.Label)
		if err != nil {
			return err
		}
		if err := domain.CheckKeywordConflict(k.Word(), bar.Prefixes()); err != nil {
			return err
		}
		if words[k.Word()] {
			return errDuplicateOwnKeyword
		}
		words[k.Word()] = true
		keywords = append(keywords, k)
	}
	if err := uc.repo.ReplaceShortcuts(ctx, list); err != nil {
		return fmt.Errorf("replace shortcuts: %w", err)
	}
	if err := uc.repo.ReplaceKeywords(ctx, keywords); err != nil {
		return fmt.Errorf("replace keywords: %w", err)
	}
	if err := uc.repo.SetDefaultProvider(ctx, provider); err != nil {
		return fmt.Errorf("save default provider: %w", err)
	}
	if err := uc.repo.SetBarOptions(ctx, bar); err != nil {
		return fmt.Errorf("save search field options: %w", err)
	}
	return nil
}
