package main

import (
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/board"
)

// The Family board module registers itself; no shared file lists it (#193).
func init() {
	registerModule("board", func(d moduleDeps) modules.Module { return board.New(d.db, d.clock, d.ids, time.Local) })
}
