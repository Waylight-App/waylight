package kernel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/url"
	"testing"
)

func TestSetupFailure(t *testing.T) {
	wrap := func(err error) error { return &url.Error{Op: "Get", URL: "http://192.168.178.20:2283/api", Err: err} }
	cases := map[string]struct {
		err  error
		want string
	}{
		"blocked":     {wrap(ErrTargetBlocked), "Connections to this address are not allowed."},
		"401":         {fmt.Errorf("%w (401 Unauthorized)", ErrTargetAuth), "The key is not accepted. Was it copied completely?"},
		"certificate": {wrap(x509.UnknownAuthorityError{}), "The certificate is not trusted. Use the http:// address in your home network, or give the service a certificate this server trusts."},
		"plain http":  {wrap(tls.RecordHeaderError{Msg: "first record does not look like a TLS handshake"}), "This address does not speak https. Try it with http://."},
		"refused":     {wrap(&net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}), "Nothing answers at this address. Is the service running?"},
		"timeout":     {fmt.Errorf("get: %w", context.DeadlineExceeded), "Nothing answers at this address. Is the service running?"},
		"dns":         {wrap(&net.DNSError{Err: "no such host", Name: "immich.lan"}), "Nothing answers at this address. Is the service running?"},
		"html":        {fmt.Errorf("%w: invalid character '<'", ErrTargetWrongService), "This is not Immich. Something else runs at this address."},
		"404":         {fmt.Errorf("%w: /api/server/statistics", ErrTargetNotFound), "This is not Immich. Something else runs at this address."},
	}
	for name, c := range cases {
		e := SetupFailure("Immich", c.err)
		if e.Message != c.want || !errors.Is(e, ErrInvalid) || !errors.Is(e.Cause, c.err) {
			t.Errorf("%s: %q", name, e.Message)
		}
	}
	if e := SetupFailure("Immich", ErrTargetWrongService); e.Params["service"] != "Immich" {
		t.Errorf("params %v", e.Params)
	}
}
