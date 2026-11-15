package kernel

import "context"

// Mail is a plain-text message to one address, such as a password reset
// link (#296). It is not an alert: it goes to the given address only.
type Mail struct {
	To      string
	Subject string
	Body    string
}

// Mailer sends mail to a single address. The notifications module
// implements it with its e-mail channel; without one, CanSend is false.
type Mailer interface {
	// CanSend reports whether mail can be sent right now.
	CanSend(ctx context.Context) bool
	Send(ctx context.Context, m Mail) error
}

// NoMailer is a Mailer that never sends.
type NoMailer struct{}

// CanSend implements Mailer.
func (NoMailer) CanSend(context.Context) bool { return false }

// Send implements Mailer.
func (NoMailer) Send(context.Context, Mail) error { return ErrUnavailable }
