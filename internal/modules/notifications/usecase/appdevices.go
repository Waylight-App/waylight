package usecase

import (
	"context"
	"errors"
	"fmt"

	"waylight/internal/modules/notifications/domain"
	"waylight/internal/shared/kernel"
)

// ErrDeviceGone is returned by a PushRelay when the push service no longer
// knows the device; the device is removed then.
var ErrDeviceGone = errors.New("push device gone")

// AppDeviceRepository persists the push devices (output port).
type AppDeviceRepository interface {
	List(ctx context.Context) ([]domain.AppDevice, error)
	// Save inserts or replaces the device of its session.
	Save(ctx context.Context, d domain.AppDevice) error
	Delete(ctx context.Context, session string) error
}

// PushRelay encrypts a message for a device and hands it to the push relay
// (output port, ADR 0025).
type PushRelay interface {
	Push(ctx context.Context, d domain.AppDevice, m domain.Message) error
}

// AppDeviceInput is what the app registers.
type AppDeviceInput struct {
	Name      string
	Platform  domain.PushPlatform
	Token     string
	PublicKey []byte
}

// AppDevices are the use cases of the push devices (#564).
type AppDevices interface {
	// Register adds or updates the device of the calling app session.
	Register(ctx context.Context, in AppDeviceInput) (domain.AppDevice, error)
	// Unregister removes the device of the calling app session.
	Unregister(ctx context.Context) error
	// List and Remove serve the settings of an admin.
	List(ctx context.Context) ([]domain.AppDevice, error)
	Remove(ctx context.Context, session string) error
	// Publisher delivers messages of the app channel to the devices.
	Publisher
}

// errPushOff is returned while the instance has no relay configured.
var errPushOff = kernel.NewPublicError(kernel.ErrUnavailable, "Push notifications to the app are turned off on this server.")

type appDevices struct {
	repo     AppDeviceRepository
	relay    PushRelay
	sessions kernel.AppSessions
	clock    kernel.Clock
}

// NewAppDevices returns the use cases; relay is nil when push is turned off
// for the instance (PUSH_RELAY_URL empty).
func NewAppDevices(repo AppDeviceRepository, relay PushRelay, sessions kernel.AppSessions, clock kernel.Clock) AppDevices {
	return &appDevices{repo: repo, relay: relay, sessions: sessions, clock: clock}
}

func (uc *appDevices) Register(ctx context.Context, in AppDeviceInput) (domain.AppDevice, error) {
	if uc.relay == nil {
		return domain.AppDevice{}, errPushOff
	}
	p := kernel.PrincipalFrom(ctx)
	if p.AppSession == "" {
		return domain.AppDevice{}, kernel.NewPublicError(kernel.ErrForbidden, "Only the mobile app can register for push notifications.")
	}
	// The notifications are the instance's (ADR 0018), like its channels.
	if !p.IsAdmin() {
		return domain.AppDevice{}, kernel.ErrAdminOnly
	}
	d := domain.AppDevice{Session: p.AppSession, AccountID: p.AccountID, Name: in.Name, Platform: in.Platform,
		Token: in.Token, PublicKey: in.PublicKey, UpdatedAt: uc.clock.Now()}
	if err := d.Validate(); err != nil {
		return domain.AppDevice{}, err
	}
	if err := uc.repo.Save(ctx, d); err != nil {
		return domain.AppDevice{}, fmt.Errorf("save device: %w", err)
	}
	return d, nil
}

func (uc *appDevices) Unregister(ctx context.Context) error {
	p := kernel.PrincipalFrom(ctx)
	if p.AppSession == "" {
		return kernel.NewPublicError(kernel.ErrForbidden, "Only the mobile app can unregister its device.")
	}
	return uc.repo.Delete(ctx, p.AppSession)
}

func (uc *appDevices) List(ctx context.Context) ([]domain.AppDevice, error) {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return nil, kernel.ErrAdminOnly
	}
	return uc.repo.List(ctx)
}

func (uc *appDevices) Remove(ctx context.Context, session string) error {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return kernel.ErrAdminOnly
	}
	return uc.repo.Delete(ctx, session)
}

// Publish implements Publisher for domain.ChannelApp: the message goes to
// every device whose app session still exists. Devices of ended sessions
// and devices the push service no longer knows are removed on the way.
func (uc *appDevices) Publish(ctx context.Context, _ domain.Channel, m domain.Message) error {
	if uc.relay == nil {
		return errPushOff
	}
	devices, err := uc.repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list devices: %w", err)
	}
	var errs []error
	for _, d := range devices {
		active, err := uc.sessions.AppSessionActive(ctx, d.AccountID, d.Session)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !active {
			errs = append(errs, uc.repo.Delete(ctx, d.Session))
			continue
		}
		switch err := uc.relay.Push(ctx, d, m); {
		case errors.Is(err, ErrDeviceGone):
			errs = append(errs, uc.repo.Delete(ctx, d.Session))
		case err != nil:
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
