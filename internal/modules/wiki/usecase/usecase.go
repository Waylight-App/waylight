// Package usecase holds the wiki search use cases (#321).
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"trisse-home/internal/modules/wiki/domain"
	"trisse-home/internal/shared/kernel"
)

// Repository stores the connected wikis with encrypted tokens (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Wiki, error)
	Get(ctx context.Context, id string) (domain.Wiki, error)
	Save(ctx context.Context, w domain.Wiki) error
	Delete(ctx context.Context, id string) error
}

// PageSearcher asks one wiki for pages (output port, one adapter per kind).
// Federated, not indexed: nothing a wiki returns is stored.
type PageSearcher interface {
	Search(ctx context.Context, w domain.Wiki, query string, limit int) ([]domain.Page, error)
}

// WikiView is a wiki without its token.
type WikiView struct {
	ID       string
	Name     string
	Kind     domain.Kind
	URL      string
	HasToken bool
}

// WikiInput creates or changes a wiki; an empty Token keeps the stored one.
type WikiInput struct {
	Name  string
	Kind  domain.Kind
	URL   string
	Token string
}

// SearchResult is the merged answer of all wikis; Failed names the wikis
// that did not answer in time, so the panel can say so.
type SearchResult struct {
	Pages  []domain.Page
	Failed []string
}

// Wikis is the input port.
type Wikis interface {
	List(ctx context.Context) ([]WikiView, error)
	Create(ctx context.Context, in WikiInput) (WikiView, error)
	Update(ctx context.Context, id string, in WikiInput) (WikiView, error)
	Delete(ctx context.Context, id string) error
	// Test sends a sample query to one wiki.
	Test(ctx context.Context, id string) error
	Search(ctx context.Context, query string) (SearchResult, error)
}

// Deps bundles the output ports.
type Deps struct {
	Repo      Repository
	Searchers map[domain.Kind]PageSearcher
	// Members are searched for the signed-in person in addition to the
	// connected wikis (#327); nil or empty adds none.
	Members *MemberSources
	IDs       kernel.IDGenerator
	Clock     kernel.Clock
}

// Timing of the federated search.
const (
	// searchTimeout bounds one wiki, so a slow wiki never holds up the
	// panel for long; the others still answer.
	searchTimeout = 4 * time.Second
	// cacheTTL lets repeated keystrokes and reopening the panel reuse an
	// answer (cache-aside, ADR 0006).
	cacheTTL     = time.Minute
	maxCacheSize = 100
	// testQuery is what the test button searches for; any word works.
	testQuery = "trisse"
)

type cached struct {
	at  time.Time
	res SearchResult
}

type wikis struct {
	d     Deps
	mu    sync.Mutex
	cache map[string]cached
}

// New returns the Wikis use case.
func New(d Deps) Wikis { return &wikis{d: d, cache: map[string]cached{}} }

func view(w domain.Wiki) WikiView {
	return WikiView{ID: w.ID, Name: w.Name, Kind: w.Kind, URL: w.URL, HasToken: w.Token != ""}
}

func (uc *wikis) List(ctx context.Context) ([]WikiView, error) {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list wikis: %w", err)
	}
	out := make([]WikiView, 0, len(list))
	for _, w := range list {
		out = append(out, view(w))
	}
	return out, nil
}

func (uc *wikis) Create(ctx context.Context, in WikiInput) (WikiView, error) {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return WikiView{}, fmt.Errorf("list wikis: %w", err)
	}
	if len(list) >= domain.MaxWikis {
		return WikiView{}, domain.ErrTooMany
	}
	w := domain.Wiki{ID: uc.d.IDs.NewID(), Name: in.Name, Kind: in.Kind, URL: in.URL, Token: in.Token}
	return uc.save(ctx, w)
}

func (uc *wikis) Update(ctx context.Context, id string, in WikiInput) (WikiView, error) {
	old, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return WikiView{}, err
	}
	w := domain.Wiki{ID: id, Name: in.Name, Kind: in.Kind, URL: in.URL, Token: in.Token}
	if w.Token == "" {
		// The stored token may only go to the wiki it was entered for.
		if kernel.TargetChanged([]string{old.URL, string(old.Kind)}, []string{strings.TrimRight(strings.TrimSpace(in.URL), "/"), string(in.Kind)}) {
			return WikiView{}, kernel.SecretRequired("token")
		}
		w.Token = old.Token
	}
	return uc.save(ctx, w)
}

func (uc *wikis) save(ctx context.Context, w domain.Wiki) (WikiView, error) {
	if err := w.Validate(); err != nil {
		return WikiView{}, err
	}
	if err := uc.d.Repo.Save(ctx, w); err != nil {
		return WikiView{}, fmt.Errorf("save wiki: %w", err)
	}
	uc.clearCache()
	return view(w), nil
}

func (uc *wikis) Delete(ctx context.Context, id string) error {
	if err := uc.d.Repo.Delete(ctx, id); err != nil {
		return err
	}
	uc.clearCache()
	return nil
}

func (uc *wikis) Test(ctx context.Context, id string) error {
	w, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return err
	}
	s, ok := uc.d.Searchers[w.Kind]
	if !ok {
		return domain.ErrKind
	}
	ctx, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()
	if _, err := s.Search(ctx, w, testQuery, 1); err != nil {
		return kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	return nil
}

func (uc *wikis) clearCache() {
	uc.mu.Lock()
	uc.cache = map[string]cached{}
	uc.mu.Unlock()
}

// Search asks every wiki at the same time and merges the answers in the
// order of the wikis; a wiki that fails or is slow is only named in Failed.
// Search asks the connected wikis (cached for everyone) and the member
// sources (never cached: each person sees only their own results).
func (uc *wikis) Search(ctx context.Context, query string) (SearchResult, error) {
	res, err := uc.searchWikis(ctx, query)
	if err != nil {
		return SearchResult{}, err
	}
	q, ok := domain.NormaliseQuery(query)
	sources := uc.d.Members.all()
	if !ok || len(sources) == 0 {
		return res, nil
	}
	// The cached result is shared; the member results go into a copy.
	out := SearchResult{Pages: append([]domain.Page{}, res.Pages...), Failed: append([]string{}, res.Failed...)}
	type answer struct {
		pages []domain.Page
		err   error
	}
	answers := make([]answer, len(sources))
	var wg sync.WaitGroup
	for i, src := range sources {
		wg.Go(func() {
			cctx, cancel := context.WithTimeout(ctx, searchTimeout)
			defer cancel()
			pages, err := src.Search(cctx, q, domain.MaxResultsPerWiki)
			answers[i] = answer{pages: pages, err: err}
		})
	}
	wg.Wait()
	for i, a := range answers {
		switch {
		case errors.Is(a.err, ErrNotConnected):
		case a.err != nil:
			out.Failed = append(out.Failed, sources[i].Name())
		default:
			for j, p := range a.pages {
				if j == domain.MaxResultsPerWiki {
					break
				}
				p.Wiki = sources[i].Name()
				out.Pages = append(out.Pages, p)
			}
		}
	}
	return out, nil
}

func (uc *wikis) searchWikis(ctx context.Context, query string) (SearchResult, error) {
	q, ok := domain.NormaliseQuery(query)
	if !ok {
		return SearchResult{Pages: []domain.Page{}}, nil
	}
	key := strings.ToLower(q)
	now := uc.d.Clock.Now()
	uc.mu.Lock()
	if c, hit := uc.cache[key]; hit && now.Sub(c.at) < cacheTTL {
		uc.mu.Unlock()
		return c.res, nil
	}
	uc.mu.Unlock()

	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return SearchResult{}, fmt.Errorf("list wikis: %w", err)
	}
	type answer struct {
		pages []domain.Page
		err   error
	}
	answers := make([]answer, len(list))
	var wg sync.WaitGroup
	for i, w := range list {
		s, ok := uc.d.Searchers[w.Kind]
		if !ok {
			answers[i].err = domain.ErrKind
			continue
		}
		wg.Go(func() {
			cctx, cancel := context.WithTimeout(ctx, searchTimeout)
			defer cancel()
			pages, err := s.Search(cctx, w, q, domain.MaxResultsPerWiki)
			answers[i] = answer{pages: pages, err: err}
		})
	}
	wg.Wait()
	res := SearchResult{Pages: []domain.Page{}}
	for i, a := range answers {
		if a.err != nil {
			res.Failed = append(res.Failed, list[i].Name)
			continue
		}
		for j, p := range a.pages {
			if j == domain.MaxResultsPerWiki {
				break
			}
			p.Wiki = list[i].Name
			res.Pages = append(res.Pages, p)
		}
	}
	// Only complete answers are cached, so a wiki that was slow once is
	// asked again on the next keystroke.
	if len(res.Failed) == 0 {
		uc.mu.Lock()
		if len(uc.cache) >= maxCacheSize {
			uc.cache = map[string]cached{}
		}
		uc.cache[key] = cached{at: now, res: res}
		uc.mu.Unlock()
	}
	return res, nil
}
