// Package web embeds the built frontend plugin of the ee/ features
// (ADR 0021). dist/ is produced by the npm build in this directory; the
// committed dist/.gitkeep keeps `go build -tags ee` working without it, and
// the server then lists the plugin but has no entry module to serve.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Dist returns the built plugin files rooted at dist/.
func Dist() fs.FS {
	sub, _ := fs.Sub(dist, "dist") // the directory is embedded, so Sub cannot fail
	return sub
}

// Built reports whether the npm build produced the entry module.
func Built() bool {
	_, err := fs.Stat(Dist(), "plugin.js")
	return err == nil
}
