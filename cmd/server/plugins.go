package main

import "io/fs"

// frontendPlugin is a frontend bundle served next to the app (ADR 0021):
// files under prefix, entry module loaded by the app.
type frontendPlugin struct {
	prefix string // "/ee/"
	entry  string // "/ee/plugin.js"
	files  fs.FS
}

// frontendPlugins is filled by build-specific wiring files (the ee build
// adds its bundle); a core build has none.
var frontendPlugins []frontendPlugin

func pluginURLs() []string {
	out := make([]string, 0, len(frontendPlugins))
	for _, p := range frontendPlugins {
		out = append(out, p.entry)
	}
	return out
}
