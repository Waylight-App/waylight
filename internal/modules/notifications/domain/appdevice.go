package domain

import (
	"strings"
	"time"

	"waylight/internal/shared/kernel"
)

// ChannelApp delivers to the mobile app through the push relay (#564,
// ADR 0025). Its devices register themselves; the channel only routes.
const ChannelApp ChannelKind = "app"

// PushPlatform is the push service a device is reached through.
type PushPlatform string

// The push services of Android and iOS.
const (
	PushFCM  PushPlatform = "fcm"
	PushAPNs PushPlatform = "apns"
)

// maxPushToken bounds a device token; FCM tokens are about 160 characters,
// APNs tokens 64.
const maxPushToken = 4096

// AppDevice is a phone that gets push notifications. It belongs to one app
// session and disappears with it.
type AppDevice struct {
	// Session is the public ID of the app session (ADR 0024).
	Session   string
	AccountID string
	Name      string
	Platform  PushPlatform
	Token     string
	// PublicKey is the device's X25519 key; messages are encrypted for it.
	PublicKey []byte
	UpdatedAt time.Time
}

// Validate checks what the app sent.
func (d *AppDevice) Validate() error {
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" || len(d.Name) > 80 {
		d.Name = "Phone"
	}
	if d.Platform != PushFCM && d.Platform != PushAPNs {
		return kernel.NewValidationError("platform", "The platform must be fcm or apns.")
	}
	if d.Token == "" || len(d.Token) > maxPushToken || strings.ContainsAny(d.Token, " \n") {
		return kernel.NewValidationError("token", "The push token is missing or invalid.")
	}
	if len(d.PublicKey) != 32 {
		return kernel.NewValidationError("publicKey", "The public key must be an X25519 key of 32 bytes.")
	}
	return nil
}
