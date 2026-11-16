//go:build ee

package main

import (
	"os"
	"strings"

	"trisse-home/ee/sso"
	"trisse-home/internal/modules"
)

// Single sign-on for members (#302), part of the ee/ features. Without
// OIDC_ISSUER it has nothing to do; without a licence key it lets nobody but
// the owner in, as a core build does.
func init() {
	registerModule("sso", func(d moduleDeps) modules.Module {
		return sso.New(d.db, d.clock, d.auth, licenceModule(d).Licensing(), sso.Policy{
			Emails:      envList("OIDC_MEMBER_EMAILS"),
			Groups:      envList("OIDC_MEMBER_GROUPS"),
			SyncGroups:  envList("OIDC_SYNC_GROUPS"),
			AdminGroups: envList("OIDC_ADMIN_GROUPS"),
		})
	})
}

// envList reads a comma-separated list; empty entries are dropped.
func envList(name string) []string {
	var out []string
	for _, v := range strings.Split(os.Getenv(name), ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
