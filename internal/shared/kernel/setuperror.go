package kernel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
)

// ErrTargetWrongService means the address answered, but not the way the expected
// service does (an HTML page, a different API). Gateways wrap it when a
// response cannot be decoded (#588).
var ErrTargetWrongService = errors.New("another service answers at this address")

// SetupFailure explains why a connection to service could not be set up, in
// the words of the setup assistant (#588): the cause and the next step, never
// a raw status code or socket error. It is an input error (422), not a 401:
// the browser session is fine, the entered credential is not. The original
// error is kept as Cause for the log.
func SetupFailure(service string, cause error) *PublicError {
	var (
		netErr  net.Error
		dnsErr  *net.DNSError
		opErr   *net.OpError
		unknown x509.UnknownAuthorityError
		invalid x509.CertificateInvalidError
		host    x509.HostnameError
		verify  *tls.CertificateVerificationError
		recHdr  tls.RecordHeaderError
	)
	// Literal templates per cause, so they can be translated (#102).
	var e *PublicError
	switch {
	case errors.Is(cause, ErrTargetBlocked):
		e = NewPublicError(ErrInvalid, "Connections to this address are not allowed.")
	case errors.Is(cause, ErrTargetAuth):
		e = NewPublicError(ErrInvalid, "The key is not accepted. Was it copied completely?")
	case errors.As(cause, &unknown), errors.As(cause, &invalid), errors.As(cause, &host), errors.As(cause, &verify):
		e = NewPublicError(ErrInvalid, "The certificate is not trusted. Use the http:// address in your home network, or give the service a certificate this server trusts.")
	case errors.As(cause, &recHdr):
		// https:// to a port that speaks plain HTTP.
		e = NewPublicError(ErrInvalid, "This address does not speak https. Try it with http://.")
	case errors.Is(cause, context.DeadlineExceeded), errors.As(cause, &netErr) && netErr.Timeout(),
		errors.As(cause, &dnsErr), errors.As(cause, &opErr):
		e = NewPublicError(ErrInvalid, "Nothing answers at this address. Is the service running?")
	default:
		// ErrTargetWrongService, ErrTargetNotFound and any other answer the
		// gateway could not read.
		e = NewPublicErrorf(ErrInvalid, "This is not {service}. Something else runs at this address.", "service", service)
	}
	e.Cause = cause
	return e
}
