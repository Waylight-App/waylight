package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/photos"
)

// The Photo frame module registers itself; no shared file lists it (#193).
func init() {
	registerModule("photos", func(d moduleDeps) modules.Module { return photos.New(d.db, d.clock, d.ids) })
}
