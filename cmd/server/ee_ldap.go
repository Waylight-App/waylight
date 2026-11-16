//go:build ee

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"trisse-home/ee/ldap"
	"trisse-home/internal/modules"
)

// Sign-in with an LDAP or Active Directory password (#396), part of the ee/
// features. Without LDAP_URL it does nothing; without a licence key that
// includes directory-sync every name is checked locally.
func init() {
	registerModule("ldap", func(d moduleDeps) modules.Module {
		return ldap.New(d.db, d.auth, licenceModule(d).Licensing(), ldap.Config{
			URL:           os.Getenv("LDAP_URL"),
			BindDN:        os.Getenv("LDAP_BIND_DN"),
			BindPassword:  os.Getenv("LDAP_BIND_PASSWORD"),
			BaseDN:        os.Getenv("LDAP_BASE_DN"),
			UserAttribute: os.Getenv("LDAP_USER_ATTRIBUTE"),
			UserClass:     os.Getenv("LDAP_USER_CLASS"),
			TLS:           ldapTLS(os.Getenv("LDAP_CA_FILE")),
		}, ldap.Policy{MemberGroups: envList("LDAP_MEMBER_GROUPS"), SyncGroups: envList("LDAP_SYNC_GROUPS"),
			AdminGroups: envList("LDAP_ADMIN_GROUPS")})
	})
}

// ldapTLS trusts the system roots plus an optional CA file, for directories
// with certificates from an internal CA (common for Active Directory).
func ldapTLS(caFile string) *tls.Config {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if caFile == "" {
		return cfg
	}
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	pem, err := os.ReadFile(caFile) //nolint:gosec // path set by the operator
	if err != nil || !pool.AppendCertsFromPEM(pem) {
		// A wrong CA file makes every connection fail, which the log shows; the
		// server still starts so the owner can sign in and fix it.
		fmt.Fprintf(os.Stderr, "ldap: cannot use LDAP_CA_FILE %q\n", caFile)
	}
	cfg.RootCAs = pool
	return cfg
}
