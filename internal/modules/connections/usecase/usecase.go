package usecase

import (
	"context"
	"fmt"

	"waylight/internal/modules/connections/domain"
	"waylight/internal/shared/kernel"
)

// View is a connection as the API shows it: the names of the credential
// fields that are set, never their values, and the widgets that use it.
type View struct {
	ID, Service, Name, URL string
	Fields                 []string
	UsedBy                 []domain.Use
}

// Input is what an admin enters for a new connection.
type Input struct {
	Service, Name, URL string
	Credential         map[string]string
}

// Change renames a connection or moves it. A nil Credential keeps the
// stored one, which is only allowed while the address stays the same.
type Change struct {
	Name, URL  string
	Credential map[string]string
}

// ErrInUse refuses to remove a connection that widgets still name until
// the admin confirms; the list shows which widgets these are.
var ErrInUse = kernel.NewPublicError(kernel.ErrConflict, "Widgets still use this connection. Remove it anyway?")

// Connections is the input port of the module (#610); every method is one
// use case and needs an admin, since connections belong to the instance.
type Connections interface {
	// List returns the connections, of one service when service is set.
	List(ctx context.Context, service string) ([]View, error)
	Add(ctx context.Context, in Input) (View, error)
	Update(ctx context.Context, id string, ch Change) (View, error)
	// ReplaceCredential sets new credential fields and keeps name and
	// address; it stays available while config as code manages the rest.
	ReplaceCredential(ctx context.Context, id string, fields map[string]string) (View, error)
	// Remove deletes a connection; one in use needs confirmed.
	Remove(ctx context.Context, id string, confirmed bool) error
}

// Repository persists connections with their credential encrypted (output
// port, implemented in adapter/gateway).
type Repository interface {
	List(ctx context.Context) ([]domain.Connection, error)
	Get(ctx context.Context, id string) (domain.Connection, error)
	Save(ctx context.Context, c domain.Connection) error
	Delete(ctx context.Context, id string) error
}

// Widget is a widget with the connection its settings name
// ("connectionId"), empty for none.
type Widget struct {
	ID, TabID, Kind, ConnectionID string
}

// Deps bundles the output ports.
type Deps struct {
	Repo Repository
	// Widgets lists widgets with their connection; nil when the dashboard
	// is not available, then no connection is in use.
	Widgets func(ctx context.Context) ([]Widget, error)
	Clock   kernel.Clock
	IDs     kernel.IDGenerator
}

type uc struct{ d Deps }

// New returns the use cases.
func New(d Deps) Connections { return &uc{d: d} }

func requireAdmin(ctx context.Context) error {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return kernel.ErrForbidden
	}
	return nil
}

// List implements Connections.
func (u *uc) List(ctx context.Context, service string) ([]View, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	list, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}
	uses, err := u.uses(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]View, 0, len(list))
	for _, c := range list {
		if service == "" || c.Service == service {
			out = append(out, view(c, uses[c.ID]))
		}
	}
	return out, nil
}

// Add implements Connections.
func (u *uc) Add(ctx context.Context, in Input) (View, error) {
	if err := requireAdmin(ctx); err != nil {
		return View{}, err
	}
	list, err := u.d.Repo.List(ctx)
	if err != nil {
		return View{}, fmt.Errorf("list connections: %w", err)
	}
	if len(list) >= domain.MaxConnections {
		return View{}, domain.ErrTooMany
	}
	cred, err := domain.NewCredential(in.Credential)
	if err != nil {
		return View{}, err
	}
	c, err := domain.NewConnection(u.d.IDs.NewID(), in.Service, in.Name, in.URL, cred, u.d.Clock.Now())
	if err != nil {
		return View{}, err
	}
	if taken(list, c) {
		return View{}, domain.ErrNameTaken
	}
	if err := u.d.Repo.Save(ctx, c); err != nil {
		return View{}, fmt.Errorf("save connection: %w", err)
	}
	return view(c, nil), nil
}

// Update implements Connections. A stored credential is never sent to a
// new address without being entered again (#191).
func (u *uc) Update(ctx context.Context, id string, ch Change) (View, error) {
	if err := requireAdmin(ctx); err != nil {
		return View{}, err
	}
	c, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return View{}, err
	}
	oldURL := c.URL
	if err := c.Change(ch.Name, ch.URL); err != nil {
		return View{}, err
	}
	switch {
	case ch.Credential != nil:
		if c.Credential, err = domain.NewCredential(ch.Credential); err != nil {
			return View{}, err
		}
	case len(c.Credential) > 0 && kernel.TargetChanged([]string{oldURL}, []string{c.URL}):
		return View{}, kernel.SecretRequired("credential")
	}
	list, err := u.d.Repo.List(ctx)
	if err != nil {
		return View{}, fmt.Errorf("list connections: %w", err)
	}
	if taken(list, c) {
		return View{}, domain.ErrNameTaken
	}
	return u.save(ctx, c)
}

// ReplaceCredential implements Connections.
func (u *uc) ReplaceCredential(ctx context.Context, id string, fields map[string]string) (View, error) {
	if err := requireAdmin(ctx); err != nil {
		return View{}, err
	}
	c, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return View{}, err
	}
	if c.Credential, err = domain.NewCredential(fields); err != nil {
		return View{}, err
	}
	return u.save(ctx, c)
}

// Remove implements Connections.
func (u *uc) Remove(ctx context.Context, id string, confirmed bool) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	if _, err := u.d.Repo.Get(ctx, id); err != nil {
		return err
	}
	if !confirmed {
		uses, err := u.uses(ctx)
		if err != nil {
			return err
		}
		if len(uses[id]) > 0 {
			return ErrInUse
		}
	}
	if err := u.d.Repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}
	return nil
}

func (u *uc) save(ctx context.Context, c domain.Connection) (View, error) {
	if err := u.d.Repo.Save(ctx, c); err != nil {
		return View{}, fmt.Errorf("save connection: %w", err)
	}
	uses, err := u.uses(ctx)
	if err != nil {
		return View{}, err
	}
	return view(c, uses[c.ID]), nil
}

// uses maps connection IDs to the widgets whose settings name them.
func (u *uc) uses(ctx context.Context) (map[string][]domain.Use, error) {
	out := map[string][]domain.Use{}
	if u.d.Widgets == nil {
		return out, nil
	}
	widgets, err := u.d.Widgets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list widgets: %w", err)
	}
	for _, w := range widgets {
		if w.ConnectionID != "" {
			out[w.ConnectionID] = append(out[w.ConnectionID], domain.Use{WidgetID: w.ID, TabID: w.TabID, Kind: w.Kind})
		}
	}
	return out, nil
}

// taken reports whether another connection of the same service has the name.
func taken(list []domain.Connection, c domain.Connection) bool {
	for _, o := range list {
		if o.ID != c.ID && o.Service == c.Service && domain.SameName(o.Name, c.Name) {
			return true
		}
	}
	return false
}

func view(c domain.Connection, uses []domain.Use) View {
	if uses == nil {
		uses = []domain.Use{}
	}
	return View{ID: c.ID, Service: c.Service, Name: c.Name, URL: c.URL, Fields: c.Credential.Fields(), UsedBy: uses}
}
