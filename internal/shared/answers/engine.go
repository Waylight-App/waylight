package answers

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

// Request is what a source needs to answer a match.
type Request struct {
	Match Match
	Lang  Lang
	Now   time.Time
}

// Source answers the intents of one module. Its intents name it by ID.
type Source interface {
	ID() string
	Intents() []Intent
	// Entities are the names the source knows (containers, services); nil
	// when its intents have no entity slots.
	Entities(ctx context.Context) ([]Entity, error)
	Resolve(ctx context.Context, r Request) (Answer, error)
}

// Provider is implemented by modules that contribute to the smart search.
type Provider interface {
	AnswerSources() []Source
}

// Engine ranks the intents of its sources and asks the best ones.
type Engine struct {
	sources map[string]Source
	order   []Source
	intents []Intent
	log     *slog.Logger
}

// NewEngine builds the search over sources; log reports failing sources.
func NewEngine(log *slog.Logger, sources ...Source) *Engine {
	e := &Engine{sources: map[string]Source{}, log: log}
	for _, s := range sources {
		e.sources[s.ID()] = s
		e.order = append(e.order, s)
		e.intents = append(e.intents, s.Intents()...)
	}
	return e
}

// Intents are the intents of all sources.
func (e *Engine) Intents() []Intent { return e.intents }

// Answer interprets query and returns up to MaxAnswers answers in lang. A
// failing source gives an answer that says so, so the user sees why
// instead of nothing; a missing slot asks back with suggestions.
func (e *Engine) Answer(ctx context.Context, query string, lang Lang, now time.Time) []Answer {
	matches := Rank(Interpret(query, e.intents, e.entities(ctx), now))
	out := make([]Answer, len(matches))
	var wg sync.WaitGroup
	for i, m := range matches {
		wg.Go(func() { out[i] = e.resolve(ctx, Request{Match: m, Lang: lang, Now: now}) })
	}
	wg.Wait()
	return out
}

func (e *Engine) resolve(ctx context.Context, r Request) Answer {
	m := r.Match
	if len(m.Missing) > 0 {
		names := make([]string, len(m.Suggestions))
		for i, s := range m.Suggestions {
			names[i] = s.Name
		}
		return Answer{
			Kind: Text, Intent: m.Intent.ID, Title: common.T(r.Lang, "which_title"),
			Text: common.T(r.Lang, "which"), Suggestions: names,
		}
	}
	src, ok := e.sources[m.Intent.Source]
	if !ok {
		return e.unavailable(r, errors.New("no source "+m.Intent.Source))
	}
	a, err := src.Resolve(ctx, r)
	if err != nil {
		return e.unavailable(r, err)
	}
	if a.Intent == "" {
		a.Intent = m.Intent.ID
	}
	if a.WidgetKind == "" {
		a.WidgetKind = m.Intent.WidgetKind
	}
	return a
}

func (e *Engine) unavailable(r Request, err error) Answer {
	e.log.Warn("smart search source failed", "intent", r.Match.Intent.ID, "error", err)
	return Answer{
		Kind: Text, Intent: r.Match.Intent.ID, WidgetKind: r.Match.Intent.WidgetKind,
		Title: common.T(r.Lang, "unavailable_title"), Text: common.T(r.Lang, "unavailable"),
	}
}

// entities collects the names of all sources; a failing source adds none.
func (e *Engine) entities(ctx context.Context) []Entity {
	var all []Entity
	// In the order given, so equal matches always resolve the same way.
	for _, s := range e.order {
		list, err := s.Entities(ctx)
		if err != nil {
			e.log.Debug("smart search entities failed", "source", s.ID(), "error", err)
			continue
		}
		all = append(all, list...)
	}
	return all
}
