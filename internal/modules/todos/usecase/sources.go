package usecase

import (
	"context"
	"fmt"

	"waylight/internal/modules/todos/domain"
	"waylight/internal/shared/kernel"
)

// SourceRepository stores the connected external lists with encrypted
// passwords (output port, #169).
type SourceRepository interface {
	List(ctx context.Context) ([]domain.Source, error)
	// Get fails with domain.ErrSourceNotFound for unknown IDs.
	Get(ctx context.Context, id string) (domain.Source, error)
	Save(ctx context.Context, s domain.Source) error
	Delete(ctx context.Context, id string) error
}

// DiscoveredList is a task list found on a server: a CalDAV collection
// (URL) or a Todoist project (ListID).
type DiscoveredList struct {
	URL    string
	ListID string
	Name   string
}

// Providers opens external lists (output port).
type Providers interface {
	// Open returns the to-dos of a list; nothing is requested yet.
	Open(s domain.Source) Repository
	// Discover finds the task lists of an account, starting at url.
	Discover(ctx context.Context, kind domain.SourceKind, url, username, password string) ([]DiscoveredList, error)
}

// SourceInput is a list to connect or change. An empty Password keeps the
// stored one when ID names an existing list.
type SourceInput struct {
	ID       string
	Name     string
	Kind     string
	URL      string
	Username string
	Password string
	ListID   string
	// OAuthState names a finished sign-in (OAuth kinds, #171).
	OAuthState string
}

// SourceView is a list at the use case boundary; it never has the password.
type SourceView struct {
	ID           string
	Name         string
	Kind         string
	URL          string
	Username     string
	ListID       string
	Capabilities domain.Capabilities
}

// Sources manages the connected external lists (input port).
type Sources interface {
	List(ctx context.Context) ([]SourceView, error)
	// Save tests the connection (the list is read once) and stores it.
	Save(ctx context.Context, in SourceInput) (SourceView, error)
	Delete(ctx context.Context, id string) error
	// Discover lists the task lists of an account; in.ID may name a stored
	// list whose password is used when in.Password is empty.
	Discover(ctx context.Context, in SourceInput) ([]DiscoveredList, error)
	// OAuth kinds (#171): start a sign-in, finish it from the provider's
	// redirect, and ask for its state while the browser signs in.
	OAuthInfo() OAuthInfo
	StartOAuth(ctx context.Context, in SourceInput) (OAuthStart, error)
	FinishOAuth(ctx context.Context, state, code, providerError string) error
	OAuthStatus(state string) (string, error)
}

type sources struct {
	d       Deps
	pending *signIns
}

// NewSources returns the Sources use cases.
func NewSources(d Deps) Sources {
	return &sources{d: d, pending: &signIns{m: map[string]*pendingSignIn{}}}
}

func sourceView(s domain.Source) SourceView {
	return SourceView{ID: s.ID, Name: s.Name, Kind: string(s.Kind), URL: s.URL, Username: s.Username, ListID: s.ListID, Capabilities: domain.CapabilitiesOf(s.Kind)}
}

func (uc *sources) List(ctx context.Context) ([]SourceView, error) {
	all, err := uc.d.Sources.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	out := make([]SourceView, 0, len(all))
	for _, s := range all {
		out = append(out, sourceView(s))
	}
	return out, nil
}

// password returns the input password or, when editing, the stored one.
func (uc *sources) password(ctx context.Context, in SourceInput) (string, error) {
	if in.Password != "" || in.ID == "" {
		return in.Password, nil
	}
	old, err := uc.d.Sources.Get(ctx, in.ID)
	if err != nil {
		return "", err
	}
	return old.Password, nil
}

func (uc *sources) Save(ctx context.Context, in SourceInput) (SourceView, error) {
	id := in.ID
	if id == "" {
		all, err := uc.d.Sources.List(ctx)
		if err != nil {
			return SourceView{}, fmt.Errorf("list sources: %w", err)
		}
		if len(all) >= domain.MaxSources {
			return SourceView{}, domain.ErrTooManySources
		}
		id = uc.d.IDs.NewID()
	} else if !domain.ValidSourceID(id) {
		return SourceView{}, domain.ErrSourceNotFound
	} else if _, err := uc.d.Sources.Get(ctx, id); err != nil {
		return SourceView{}, err // PUT never creates a list
	}
	s, err := uc.build(ctx, id, in)
	if err != nil {
		return SourceView{}, err
	}
	if _, err := uc.d.Providers.Open(s).List(ctx); err != nil {
		return SourceView{}, kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	if err := uc.d.Sources.Save(ctx, s); err != nil {
		return SourceView{}, fmt.Errorf("save source: %w", err)
	}
	return sourceView(s), nil
}

// build validates the input into a list; OAuth kinds take their grant from
// the finished sign-in or, when editing, from the stored list.
func (uc *sources) build(ctx context.Context, id string, in SourceInput) (domain.Source, error) {
	kind := domain.SourceKind(in.Kind)
	if !kind.OAuth() {
		pw, err := uc.password(ctx, in)
		if err != nil {
			return domain.Source{}, err
		}
		return domain.NewSource(id, in.Name, kind, in.URL, in.Username, pw, in.ListID)
	}
	if in.OAuthState == "" {
		old, err := uc.d.Sources.Get(ctx, in.ID)
		if err != nil {
			return domain.Source{}, errSignInUnknown
		}
		return domain.NewOAuthSource(id, in.Name, kind, old.Username, in.ListID, old.Grant)
	}
	p, err := uc.signedIn(in.OAuthState)
	if err != nil {
		return domain.Source{}, err
	}
	if p.kind != kind {
		return domain.Source{}, errSignInUnknown
	}
	g := p.grant
	g.ClientSecret = p.clientSecret
	return domain.NewOAuthSource(id, in.Name, kind, p.clientID, in.ListID, g)
}

func (uc *sources) Delete(ctx context.Context, id string) error {
	if !domain.ValidSourceID(id) {
		return domain.ErrSourceNotFound
	}
	s, err := uc.d.Sources.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.d.Sources.Delete(ctx, id); err != nil {
		return err
	}
	if s.Kind.OAuth() && uc.d.OAuth != nil {
		_ = uc.d.OAuth.Revoke(ctx, s.Kind, s.Username, s.Grant) // best effort; the list is gone either way
	}
	return nil
}

func (uc *sources) Discover(ctx context.Context, in SourceInput) ([]DiscoveredList, error) {
	if kind := domain.SourceKind(in.Kind); kind.OAuth() {
		p, err := uc.signedIn(in.OAuthState)
		if err != nil {
			return nil, err
		}
		lists, err := uc.d.Providers.Discover(ctx, kind, "", p.clientID, p.grant.Access)
		if err != nil {
			return nil, kernel.TargetFailure(kernel.ErrUnavailable, err)
		}
		return lists, nil
	}
	pw, err := uc.password(ctx, in)
	if err != nil {
		return nil, err
	}
	// Validate address and credentials like a list, with a placeholder name.
	s, err := domain.NewSource("discover", "discover", domain.SourceKind(in.Kind), in.URL, in.Username, pw, "discover")
	if err != nil {
		return nil, err
	}
	lists, err := uc.d.Providers.Discover(ctx, s.Kind, s.URL, s.Username, s.Password)
	if err != nil {
		return nil, kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	return lists, nil
}
