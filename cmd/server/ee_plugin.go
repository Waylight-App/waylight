//go:build ee

package main

import eeweb "trisse-home/ee/web"

// The ee/ frontend plugin (ADR 0021) is served under /ee/ and listed by
// /api/plugins, but only when the bundle was built: a Go-only build of the
// ee features (as in CI's test step) then behaves like a core build in the
// browser instead of pointing the app at a missing module.
func init() {
	if !eeweb.Built() {
		return
	}
	frontendPlugins = append(frontendPlugins, frontendPlugin{prefix: "/ee/", entry: "/ee/plugin.js", files: eeweb.Dist()})
}
