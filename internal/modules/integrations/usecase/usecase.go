package usecase

import (
	"context"
	"fmt"

	"waylight/internal/modules/integrations/domain"
	"waylight/internal/shared/kernel"
)

// View is a connection as the API shows it: the names of the credential
// fields that are set, never their values, and the widgets that use it.
type View struct {
	ID, Service, Name, URL string
	Fields                 []string
	// Values holds the fields that are not secret, e.g. the user name.
	Values map[string]string
	UsedBy []domain.Use
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

// Choice is a connection as a widget picker shows it: no address and no
// credential.
type Choice struct {
	ID, Service, Name string
}

// ErrInUse refuses to remove a connection that widgets still name until
// the admin confirms; the list shows which widgets these are.
var ErrInUse = kernel.NewPublicError(kernel.ErrConflict, "Widgets still use this connection. Remove it anyway?")

// Connections is the input port of the module (#610); every method is one
// use case and needs an admin, since connections belong to the instance,
// except Choices.
type Connections interface {
	// Choices lists the connections of a service (every service when it is
	// empty) for a widget's picker; every signed-in account that edits
	// widgets may read them.
	Choices(ctx context.Context, service string) ([]Choice, error)
	// List returns the connections, of one service when service is set.
	List(ctx context.Context, service string) ([]View, error)
	Add(ctx context.Context, in Input) (View, error)
	Update(ctx context.Context, id string, ch Change) (View, error)
	// ReplaceCredential sets new credential fields and keeps name and
	// address; it stays available while config as code manages the rest.
	ReplaceCredential(ctx context.Context, id string, fields map[string]string) (View, error)
	// Remove deletes a connection; one in use needs confirmed.
	Remove(ctx context.Context, id string, confirmed bool) error
	// Check tries the connection with the request the service's settings
	// page makes before saving; nil means the service accepted it.
	Check(ctx context.Context, id string) error
	// Services lists the services a connection can be made to.
	Services(ctx context.Context) ([]domain.Spec, error)
}

// CheckFunc tries an address and credential at the service (see
// modules.ConnectionSpec.Check).
type CheckFunc func(ctx context.Context, url string, credential map[string]string) error

// Service is a spec with the check of the module that owns the service.
type Service struct {
	Spec  domain.Spec
	Check CheckFunc
}

// ErrNoCheck answers a check of a service whose module is switched off.
var ErrNoCheck = kernel.NewPublicError(kernel.ErrUnavailable, "This connection cannot be checked while its module is switched off.")

// Repository persists connections with their credential encrypted (output
// port, implemented in adapter/gateway).
type Repository interface {
	List(ctx context.Context) ([]domain.Connection, error)
	Get(ctx context.Context, id string) (domain.Connection, error)
	Save(ctx context.Context, c domain.Connection) error
	Delete(ctx context.Context, id string) error
}

// Widget is a widget with the connections its settings name
// ("connectionId" or "connectionIds"), none for a widget without.
type Widget struct {
	ID, TabID, Kind string
	ConnectionIDs   []string
	// Mine is true when the widget is on a tab of the signed-in account.
	Mine bool
}

// Deps bundles the output ports.
type Deps struct {
	Repo Repository
	// Widgets lists the widgets of every tab of the instance with their
	// connections; nil when the dashboard is not available, then no
	// connection is in use.
	Widgets func(ctx context.Context) ([]Widget, error)
	// Services are the services of the enabled modules; a connection can
	// only be added for one of them.
	Services []Service
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
}

type uc struct {
	d      Deps
	specs  domain.Specs
	checks map[string]CheckFunc
}

// New returns the use cases.
func New(d Deps) Connections {
	list := make([]domain.Spec, 0, len(d.Services))
	checks := map[string]CheckFunc{}
	for _, s := range d.Services {
		list = append(list, s.Spec)
		if s.Check != nil {
			checks[s.Spec.Service] = s.Check
		}
	}
	return &uc{d: d, specs: domain.NewSpecs(list), checks: checks}
}

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
			out = append(out, u.view(c, uses[c.ID]))
		}
	}
	return out, nil
}

// Choices implements Connections. Guests and kiosk displays do not edit
// widgets, so they get none.
func (u *uc) Choices(ctx context.Context, service string) ([]Choice, error) {
	if p := kernel.PrincipalFrom(ctx); p.TabScoped() || p.IsToken() {
		return nil, kernel.ErrForbidden
	}
	list, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}
	out := []Choice{}
	for _, c := range list {
		if service == "" || c.Service == service {
			out = append(out, Choice{ID: c.ID, Service: c.Service, Name: c.Name})
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
	spec, ok := u.specs[in.Service]
	if !ok {
		return View{}, domain.ErrUnknownService
	}
	cred, err := domain.NewCredential(in.Credential)
	if err != nil {
		return View{}, err
	}
	if err := spec.CheckCredential(cred, true); err != nil {
		return View{}, err
	}
	c, err := domain.NewConnection(u.d.IDs.NewID(), in.Service, in.Name, fixedURL(spec, in.URL), cred, u.d.Clock.Now())
	if err != nil {
		return View{}, err
	}
	if taken(list, c) {
		return View{}, domain.ErrNameTaken
	}
	if err := u.d.Repo.Save(ctx, c); err != nil {
		return View{}, fmt.Errorf("save connection: %w", err)
	}
	return u.view(c, nil), nil
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
	spec := u.specs[c.Service] // zero for a switched-off module: no fixed address
	if err := c.Change(ch.Name, fixedURL(spec, ch.URL)); err != nil {
		return View{}, err
	}
	switch {
	case ch.Credential != nil:
		if c.Credential, err = u.credential(c.Service, ch.Credential); err != nil {
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
	if c.Credential, err = u.credential(c.Service, fields); err != nil {
		return View{}, err
	}
	return u.save(ctx, c)
}

// credential validates entered fields against the service's spec. A
// connection of a service whose module is switched off keeps the general
// rules only, so it stays editable.
func (u *uc) credential(service string, fields map[string]string) (domain.Credential, error) {
	cred, err := domain.NewCredential(fields)
	if err != nil {
		return nil, err
	}
	if spec, ok := u.specs[service]; ok {
		if err := spec.CheckCredential(cred, true); err != nil {
			return nil, err
		}
	}
	return cred, nil
}

// Check implements Connections.
func (u *uc) Check(ctx context.Context, id string) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	c, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return err
	}
	check, ok := u.checks[c.Service]
	if !ok {
		return ErrNoCheck
	}
	if err := u.specs[c.Service].CheckCredential(c.Credential, true); err != nil {
		return err
	}
	return check(ctx, c.URL, c.Credential)
}

// Services implements Connections.
func (u *uc) Services(ctx context.Context) ([]domain.Spec, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	return u.specs.Sorted(), nil
}

// fixedURL replaces the entered address of a service that has only one.
func fixedURL(spec domain.Spec, entered string) string {
	if spec.FixedURL != "" {
		return spec.FixedURL
	}
	return entered
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
	return u.view(c, uses[c.ID]), nil
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
		for _, id := range w.ConnectionIDs {
			if id != "" {
				out[id] = append(out[id], domain.Use{WidgetID: w.ID, TabID: w.TabID, Kind: w.Kind, Mine: w.Mine})
			}
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

func (u *uc) view(c domain.Connection, uses []domain.Use) View {
	if uses == nil {
		uses = []domain.Use{}
	}
	// Without a spec every field counts as secret.
	values := map[string]string{}
	if spec, ok := u.specs[c.Service]; ok {
		values = spec.Visible(c.Credential)
	}
	return View{ID: c.ID, Service: c.Service, Name: c.Name, URL: c.URL, Fields: c.Credential.Fields(), Values: values, UsedBy: uses}
}
