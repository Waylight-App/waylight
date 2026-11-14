package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/shared/kernel"
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

// DiscoveredList is a task list found on a server.
type DiscoveredList struct {
	URL  string
	Name string
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
}

// SourceView is a list at the use case boundary; it never has the password.
type SourceView struct {
	ID           string
	Name         string
	Kind         string
	URL          string
	Username     string
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
}

type sources struct{ d Deps }

// NewSources returns the Sources use cases.
func NewSources(d Deps) Sources { return &sources{d: d} }

func sourceView(s domain.Source) SourceView {
	return SourceView{ID: s.ID, Name: s.Name, Kind: string(s.Kind), URL: s.URL, Username: s.Username, Capabilities: domain.CapabilitiesOf(s.Kind)}
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
	pw, err := uc.password(ctx, in)
	if err != nil {
		return SourceView{}, err
	}
	s, err := domain.NewSource(id, in.Name, domain.SourceKind(in.Kind), in.URL, in.Username, pw)
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

func (uc *sources) Delete(ctx context.Context, id string) error {
	if !domain.ValidSourceID(id) {
		return domain.ErrSourceNotFound
	}
	if _, err := uc.d.Sources.Get(ctx, id); err != nil {
		return err
	}
	return uc.d.Sources.Delete(ctx, id)
}

func (uc *sources) Discover(ctx context.Context, in SourceInput) ([]DiscoveredList, error) {
	pw, err := uc.password(ctx, in)
	if err != nil {
		return nil, err
	}
	// Validate address and credentials like a list, with a placeholder name.
	s, err := domain.NewSource("discover", "discover", domain.SourceKind(in.Kind), in.URL, in.Username, pw)
	if err != nil {
		return nil, err
	}
	lists, err := uc.d.Providers.Discover(ctx, s.Kind, s.URL, s.Username, s.Password)
	if err != nil {
		return nil, kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	return lists, nil
}
