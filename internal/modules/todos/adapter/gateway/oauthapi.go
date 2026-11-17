package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"sync"
	"time"

	"waylight/internal/modules/todos/domain"
	"waylight/internal/modules/todos/usecase"
	"waylight/internal/shared/kernel"
)

// oauthAPI is the HTTP side shared by the providers that sign in with OAuth
// (#171): bearer tokens from the TokenSource, JSON bodies, status mapping
// and a short per-list cache so several widgets share one request.
type oauthAPI struct {
	client *http.Client
	tokens usecase.TokenSource
	now    func() time.Time

	mu    sync.Mutex
	cache map[string]cachedList
}

type cachedList struct {
	at    time.Time
	todos []domain.Todo
}

// oauthMinRefresh is the shortest time between two loads of one list.
const oauthMinRefresh = 30 * time.Second

func newOAuthAPI(client *http.Client, tokens usecase.TokenSource) *oauthAPI {
	return &oauthAPI{client: client, tokens: tokens, now: time.Now, cache: map[string]cachedList{}}
}

func (a *oauthAPI) cached(key string) ([]domain.Todo, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	c, ok := a.cache[key]
	if !ok || a.now().Sub(c.at) >= oauthMinRefresh {
		return nil, false
	}
	return c.todos, true
}

func (a *oauthAPI) store(key string, todos []domain.Todo) {
	a.mu.Lock()
	a.cache[key] = cachedList{at: a.now(), todos: todos}
	a.mu.Unlock()
}

func (a *oauthAPI) forget(key string) {
	a.mu.Lock()
	delete(a.cache, key)
	a.mu.Unlock()
}

// do sends one request with the list's access token (or token, for
// discovery before the list exists). It returns the response headers.
func (a *oauthAPI) do(ctx context.Context, s domain.Source, token, method, target string, body any, dst any, headers map[string]string) (http.Header, error) {
	if token == "" {
		t, err := a.tokens.AccessToken(ctx, s)
		if err != nil {
			return nil, err
		}
		token = t
	}
	var r io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
		r = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, r)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read answer: %w", err)
	}
	switch {
	case res.StatusCode == http.StatusUnauthorized:
		return nil, domain.ErrSignInAgain
	case res.StatusCode == http.StatusForbidden:
		return nil, fmt.Errorf("access refused: %w", kernel.ErrTargetAuth)
	case res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusGone:
		return nil, domain.ErrTodoNotFound
	case res.StatusCode == http.StatusPreconditionFailed || res.StatusCode == http.StatusConflict:
		return nil, domain.ErrVersionConflict
	case res.StatusCode == http.StatusTooManyRequests:
		return nil, domain.ErrRateLimited
	case res.StatusCode >= 300:
		return nil, fmt.Errorf("server answered %s", res.Status)
	}
	if dst != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, dst); err != nil {
			return nil, fmt.Errorf("parse answer: %w", err)
		}
	}
	return res.Header, nil
}

// hashVersion derives a to-do version from fields that change with it.
func hashVersion(parts ...string) int {
	h := fnv.New32a()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{0})
	}
	return int(h.Sum32()&0x7fffffff) | 1
}

// sortTodos orders by due date (none last), then creation, and numbers them.
func sortTodos(todos []domain.Todo) []domain.Todo {
	slicesSortTodos(todos)
	for i := range todos {
		todos[i].MoveTo(i)
	}
	return todos
}
