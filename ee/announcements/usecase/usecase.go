// Package usecase lists, writes and marks announcements (#297).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/ee/announcements/domain"
	"trisse-home/internal/shared/kernel"
)

// Store keeps announcements, read marks and settings (output port).
type Store interface {
	List(ctx context.Context) ([]domain.Announcement, error)
	Get(ctx context.Context, id string) (domain.Announcement, error)
	Save(ctx context.Context, a domain.Announcement) error
	Delete(ctx context.Context, id string) error
	// ReadBy returns the IDs of the announcements accountID marked read.
	ReadBy(ctx context.Context, accountID string) (map[string]bool, error)
	MarkRead(ctx context.Context, accountID, id string, at time.Time) error
	// ForgetAccount removes an account's read marks.
	ForgetAccount(ctx context.Context, accountID string) error
	Settings(ctx context.Context) (domain.Settings, error)
	SaveSettings(ctx context.Context, s domain.Settings) error
}

// Names shows who wrote a notice (output port, the auth module's accounts).
type Names interface {
	Name(ctx context.Context, accountID string) string
}

// Deps are the use cases' collaborators.
type Deps struct {
	Store  Store
	Names  Names
	Alerts kernel.AlertSink
	IDs    kernel.IDGenerator
	Clock  kernel.Clock
	// Audit records writes in the audit log (ee/audit), may be nil.
	Audit func(ctx context.Context, action, id string)
}

// View is an announcement as a reader sees it.
type View struct {
	domain.Announcement
	AuthorName string
	// Read is whether the reader marked it read; always false for guests
	// and kiosk displays, which have no read marks.
	Read bool
	// Editable tells the client to offer editing.
	Editable bool
}

// Input is what a writer sends.
type Input struct {
	Title     string
	Body      string
	Pinned    bool
	ExpiresAt time.Time
	// Notify sends the notice to the notification channels once, on
	// publishing (or on an edit only when asked again).
	Notify bool
}

// Service implements the use cases.
type Service struct{ d Deps }

// New returns the use cases.
func New(d Deps) *Service {
	if d.Alerts == nil {
		d.Alerts = kernel.DiscardAlerts{}
	}
	return &Service{d: d}
}

func canMarkRead(p kernel.Principal) bool {
	return p.AccountID != "" && !p.TabScoped() && !p.IsToken() && !p.IsSystem()
}

// List returns the current notices (archive false) or the expired ones
// (archive true), pinned first and newest first.
func (s *Service) List(ctx context.Context, archive bool) ([]View, error) {
	all, err := s.d.Store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list announcements: %w", err)
	}
	set, err := s.d.Store.Settings(ctx)
	if err != nil {
		return nil, err
	}
	p := kernel.PrincipalFrom(ctx)
	read := map[string]bool{}
	if canMarkRead(p) {
		if read, err = s.d.Store.ReadBy(ctx, p.AccountID); err != nil {
			return nil, err
		}
	}
	now := s.d.Clock.Now()
	var picked []domain.Announcement
	for _, a := range all {
		if a.Active(now) != archive {
			picked = append(picked, a)
		}
	}
	domain.Sort(picked)
	out := make([]View, 0, len(picked))
	names := map[string]string{}
	for _, a := range picked {
		name, ok := names[a.AuthorID]
		if !ok && s.d.Names != nil {
			name = s.d.Names.Name(ctx, a.AuthorID)
			names[a.AuthorID] = name
		}
		out = append(out, View{Announcement: a, AuthorName: name, Read: read[a.ID], Editable: set.MayChange(p, a)})
	}
	return out, nil
}

// Create publishes a notice.
func (s *Service) Create(ctx context.Context, in Input) (View, error) {
	p := kernel.PrincipalFrom(ctx)
	set, err := s.d.Store.Settings(ctx)
	if err != nil {
		return View{}, err
	}
	if !set.MayWrite(p) {
		return View{}, domain.ErrNotAllowed
	}
	now := s.d.Clock.Now()
	all, err := s.d.Store.List(ctx)
	if err != nil {
		return View{}, err
	}
	active := 0
	for _, a := range all {
		if a.Active(now) {
			active++
		}
	}
	if active >= domain.MaxActive {
		return View{}, domain.ErrTooMany
	}
	a := domain.Announcement{ID: s.d.IDs.NewID(), Title: in.Title, Body: in.Body, Pinned: in.Pinned, AuthorID: p.AccountID,
		ExpiresAt: in.ExpiresAt, CreatedAt: now, UpdatedAt: now}
	if err := a.Validate(); err != nil {
		return View{}, err
	}
	if err := s.d.Store.Save(ctx, a); err != nil {
		return View{}, fmt.Errorf("save announcement: %w", err)
	}
	s.audit(ctx, "create", a.ID)
	if in.Notify {
		s.notify(ctx, a)
	}
	return View{Announcement: a, Editable: true}, nil
}

// Update changes a notice. It notifies again only when in.Notify asks for it.
func (s *Service) Update(ctx context.Context, id string, in Input) (View, error) {
	a, set, err := s.changeable(ctx, id)
	if err != nil {
		return View{}, err
	}
	a.Title, a.Body, a.Pinned, a.ExpiresAt, a.UpdatedAt = in.Title, in.Body, in.Pinned, in.ExpiresAt, s.d.Clock.Now()
	if err := a.Validate(); err != nil {
		return View{}, err
	}
	if err := s.d.Store.Save(ctx, a); err != nil {
		return View{}, fmt.Errorf("save announcement: %w", err)
	}
	s.audit(ctx, "update", a.ID)
	if in.Notify {
		s.notify(ctx, a)
	}
	return View{Announcement: a, Editable: set.MayChange(kernel.PrincipalFrom(ctx), a)}, nil
}

// Delete removes a notice and its read marks.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, _, err := s.changeable(ctx, id); err != nil {
		return err
	}
	if err := s.d.Store.Delete(ctx, id); err != nil {
		return err
	}
	s.audit(ctx, "delete", id)
	return nil
}

// MarkRead marks a notice read for the signed-in person.
func (s *Service) MarkRead(ctx context.Context, id string) error {
	p := kernel.PrincipalFrom(ctx)
	if !canMarkRead(p) {
		return domain.ErrNoReader
	}
	if _, err := s.d.Store.Get(ctx, id); err != nil {
		return err
	}
	return s.d.Store.MarkRead(ctx, p.AccountID, id, s.d.Clock.Now())
}

// Settings returns who may write; admins only.
func (s *Service) Settings(ctx context.Context) (domain.Settings, error) {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return domain.Settings{}, domain.ErrNotAllowed
	}
	return s.d.Store.Settings(ctx)
}

// SaveSettings changes who may write; admins only.
func (s *Service) SaveSettings(ctx context.Context, set domain.Settings) error {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return domain.ErrNotAllowed
	}
	if err := s.d.Store.SaveSettings(ctx, set); err != nil {
		return err
	}
	s.audit(ctx, "settings", "writer-group")
	return nil
}

// ForgetAccount drops a deleted account's read marks; its notices stay.
func (s *Service) ForgetAccount(ctx context.Context, accountID string) error {
	return s.d.Store.ForgetAccount(ctx, accountID)
}

func (s *Service) changeable(ctx context.Context, id string) (domain.Announcement, domain.Settings, error) {
	a, err := s.d.Store.Get(ctx, id)
	if err != nil {
		return domain.Announcement{}, domain.Settings{}, err
	}
	set, err := s.d.Store.Settings(ctx)
	if err != nil {
		return domain.Announcement{}, domain.Settings{}, err
	}
	if !set.MayChange(kernel.PrincipalFrom(ctx), a) {
		return domain.Announcement{}, domain.Settings{}, domain.ErrNotAllowed
	}
	return a, set, nil
}

func (s *Service) notify(ctx context.Context, a domain.Announcement) {
	body := a.Body
	if r := []rune(body); len(r) > 280 {
		body = string(r[:279]) + "…"
	}
	s.d.Alerts.Alert(ctx, kernel.Alert{Kind: "announcement", Title: "📣 " + a.Title, Body: body})
}

func (s *Service) audit(ctx context.Context, action, id string) {
	if s.d.Audit != nil {
		s.d.Audit(ctx, action, id)
	}
}
