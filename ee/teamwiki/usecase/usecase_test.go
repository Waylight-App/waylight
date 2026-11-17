package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	licence "waylight/ee/licence/domain"
	"waylight/internal/modules/wiki"
	"waylight/internal/modules/wiki/domain"
	"waylight/internal/shared/kernel"
)

type tokens map[string]error

func (t tokens) AccessToken(_ context.Context, account, _ string) (string, error) {
	if err, ok := t[account]; ok {
		return "", err
	}
	return "tok-" + account, nil
}

type level licence.Level

func (l level) Level(context.Context, string) licence.Level { return licence.Level(l) }

type searcher struct{ token string }

func (s *searcher) Search(_ context.Context, token, q string, _ int) ([]domain.Page, error) {
	s.token = token
	return []domain.Page{{Title: q}}, nil
}

func as(id string) context.Context {
	return kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: id, Role: "member"})
}

func TestSourceUsesTheSearchingPersonsToken(t *testing.T) {
	sr := &searcher{}
	src := NewSource("Confluence", "atlassian", tokens{
		"m2": kernel.NewPublicError(kernel.ErrNotFound, "This account is not connected."),
		"m3": errors.New("database gone"),
	}, level(licence.LevelFull), sr)
	if pages, err := src.Search(as("m1"), "rules", 5); err != nil || len(pages) != 1 || sr.token != "tok-m1" {
		t.Fatalf("%v %v token %q", pages, err, sr.token)
	}
	if _, err := src.Search(as("m2"), "rules", 5); !errors.Is(err, wiki.ErrNotConnected) {
		t.Fatalf("not connected: %v", err)
	}
	if _, err := src.Search(as("m3"), "rules", 5); err == nil || errors.Is(err, wiki.ErrNotConnected) {
		t.Fatalf("real failure hidden: %v", err)
	}
	if _, err := src.Search(context.Background(), "rules", 5); !errors.Is(err, wiki.ErrNotConnected) {
		t.Fatalf("no principal: %v", err)
	}
}

func TestSourceIsSkippedWithoutLicence(t *testing.T) {
	src := NewSource("Notion", "notion", tokens{}, level(licence.LevelNone), &searcher{})
	if _, err := src.Search(as("m1"), "x", 5); !errors.Is(err, wiki.ErrNotConnected) {
		t.Fatalf("got %v", err)
	}
}

func TestSourceLimitsSearchesPerPersonAndMinute(t *testing.T) {
	now := time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC)
	src := NewSource("Slack", "slack", tokens{}, level(licence.LevelFull), &searcher{}).WithLimit(2, func() time.Time { return now })
	for i := range 2 {
		if _, err := src.Search(as("m1"), "x", 5); err != nil {
			t.Fatalf("search %d: %v", i, err)
		}
	}
	if _, err := src.Search(as("m1"), "x", 5); !errors.Is(err, ErrBusy) {
		t.Fatalf("third search: %v", err)
	}
	if _, err := src.Search(as("m2"), "x", 5); err != nil {
		t.Fatalf("another person is limited too: %v", err)
	}
	now = now.Add(61 * time.Second)
	if _, err := src.Search(as("m1"), "x", 5); err != nil {
		t.Fatalf("after a minute: %v", err)
	}
}
