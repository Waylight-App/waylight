package gateway

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"trisse-home/ee/ldap/adapter/ldapclient"
	"trisse-home/ee/ldap/domain"
)

// Config tells the gateway how to reach and search the directory.
type Config struct {
	// URL is ldaps://host or ldap://host (StartTLS); plain LDAP is refused.
	URL string
	// BindDN and BindPassword are a service account that may search users.
	BindDN, BindPassword string
	BaseDN               string
	// UserAttribute holds the login name: uid (OpenLDAP) or sAMAccountName (AD).
	UserAttribute string
	// UserClass limits the search to one objectClass (e.g. person); empty: any.
	UserClass string
	TLS       *tls.Config
	Timeout   time.Duration
}

// Directory implements usecase.Directory with the LDAP client.
type Directory struct{ cfg Config }

// NewDirectory returns the gateway.
func NewDirectory(cfg Config) *Directory {
	if cfg.UserAttribute == "" {
		cfg.UserAttribute = "uid"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	return &Directory{cfg: cfg}
}

var errAmbiguous = errors.New("the login name matches more than one directory entry")

// Lookup implements usecase.Directory.
func (d *Directory) Lookup(ctx context.Context, name string) (domain.User, bool, error) {
	c, err := ldapclient.Dial(ctx, d.cfg.URL, d.cfg.TLS, d.cfg.Timeout)
	if err != nil {
		return domain.User{}, false, err
	}
	defer c.Close()
	if err := c.Bind(d.cfg.BindDN, d.cfg.BindPassword); err != nil {
		return domain.User{}, false, fmt.Errorf("service account bind: %w", err)
	}
	entries, err := c.Search(d.cfg.BaseDN, d.cfg.UserAttribute, name, d.cfg.UserClass,
		[]string{"memberOf", "userAccountControl"}, 2)
	if err != nil {
		return domain.User{}, false, err
	}
	switch len(entries) {
	case 0:
		return domain.User{}, false, nil
	case 1:
		e := entries[0]
		return domain.User{DN: e.DN, Groups: e.Attributes["memberof"], Disabled: domain.ADDisabled(e.Get("userAccountControl"))}, true, nil
	default:
		return domain.User{}, false, errAmbiguous
	}
}

// Verify implements usecase.Directory: a bind as the user proves the password.
func (d *Directory) Verify(ctx context.Context, dn, password string) error {
	c, err := ldapclient.Dial(ctx, d.cfg.URL, d.cfg.TLS, d.cfg.Timeout)
	if err != nil {
		return err
	}
	defer c.Close()
	return c.Bind(dn, password)
}
