package board

import (
	"context"
	"database/sql"
	"io/fs"
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/modules"
	"waylight/internal/modules/board/adapter/controller"
	"waylight/internal/modules/board/adapter/gateway"
	"waylight/internal/modules/board/usecase"
	"waylight/internal/shared/kernel"
)

// compactTextLength shortens the note in the compact detail line, which
// is a single line on a small tile.
const compactTextLength = 80

// Module wires the family board rings and implements modules.Module (#685).
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	uc       usecase.Board
	clock    kernel.Clock
}

// New builds the family board module; loc decides when a note expires.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator, loc *time.Location) *Module {
	deps := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids, Location: loc}
	uc := usecase.NewBoard(deps)
	return &Module{
		seedDemo:     func(ctx context.Context) error { return usecase.SeedDemo(ctx, deps) },
		Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		handlers:     controller.NewHandlers(uc),
		uc:           uc,
		clock:        clock,
	}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "board" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"board"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/board", m.handlers.List)
	r.Protected("POST /api/board", m.handlers.Create)
	r.Protected("PUT /api/board/{id}", m.handlers.Update)
	r.Protected("DELETE /api/board/{id}", m.handlers.Delete)
}

// Jobs implements modules.Module: expired notes are removed hourly, so the
// database does not keep notes nobody sees any more.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "board.cleanup", Interval: time.Hour, Run: func(ctx context.Context) error {
		_, err := m.uc.Cleanup(ctx)
		return err
	}}}
}

// CompactState implements modules.CompactProvider (#597): how many notes
// are on the board, with the first one (pinned or newest) as detail.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	list, err := m.uc.List(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	state := kernel.CompactState{Level: kernel.CompactNone, Count: kernel.IntPtr(len(list)), UpdatedAt: m.clock.Now()}
	switch len(list) {
	case 0:
		state.Title = kernel.NewText("The board is empty")
		state.Count = nil
		state.Empty = true
		return state, nil
	case 1:
		state.Title = kernel.NewText("1 note")
	default:
		state.Title = kernel.NewTextf("{count} notes", "count", len(list))
	}
	state.Detail = kernel.NewTextf("{text}", "text", shorten(list[0].Text))
	return state, nil
}

// shorten keeps the first line of a note within compactTextLength runes.
func shorten(s string) string {
	s, _, _ = strings.Cut(s, "\n")
	if utf8.RuneCountInString(s) <= compactTextLength {
		return s
	}
	return string([]rune(s)[:compactTextLength-1]) + "…"
}
